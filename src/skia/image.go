package skia

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/webp"
)

// Limits mirror renderer/lib/assets.mjs DEFAULT_ASSET_LIMITS.
const (
	MaxImageBytes    = 16 * 1024 * 1024
	MaxDecodedBytes  = 16 * 1024 * 1024
	MaxDecodedWidth  = 4096
	MaxDecodedHeight = 4096
	MaxDecodedPixels = 16 * 1024 * 1024
	CacheMaxEntries  = 128
	CacheTTL         = 5 * time.Minute
)

// Allowed MIME set — must match assets.mjs ALLOWED_MIME.
var allowedMIME = map[string]bool{
	"image/png":     true,
	"image/jpeg":    true,
	"image/webp":    true,
	"image/gif":     true,
	"image/avif":    true,
	"image/svg+xml": true,
}

// Frozen 8 fixture aliases (renderer/lib/assets.mjs).
var frozenFixtureAliases = map[string]string{
	"https://fixture-cache.invalid/card/secretary-painting.png":   "card-secretary-1024.png",
	"https://fixture-cache.invalid/card/player-avatar.png":        "card-player-portrait-180x360.png",
	"https://fixture-cache.invalid/operator/amiya-painting.png":   "operator-painting-1024.png",
	"https://fixture-cache.invalid/operator/building-skill-icon.png": "operator-building-36.png",
	"https://fixture-cache.invalid/operator/skill-icon.png":       "operator-skill-128.png",
	"https://fixture-cache.invalid/enemy/originium-slug.png":      "enemy-originium-slug-158.png",
	"https://fixture-cache.invalid/recruit-amiya.png":             "amiya-avatar.webp",
	"https://fixture-cache.invalid/depot-lmd.png":                 "depot-lmd.png",
}

// Image is the SkImage equivalent: holds decoded bytes + metadata.
// In skia-tag mode, SkData::MakeWithCopy + SkImage::MakeFromEncoded produces same fields plus native handle.
type Image struct {
	Bytes  []byte // normalized (WebP→PNG when needed)
	MIME   string
	Width  int
	Height int
	SHA256 string
}

// lruEntry mirrors createAssetLoader cache entry (value + expiry).
type lruEntry struct {
	img       *Image
	expiresAt time.Time
}

// Loader validates SHA/bytes/MIME triple and caches 128/5min.
// Ponytail: global mutex, per-key pending via sync.Map — add sharded locks if contended.
type Loader struct {
	mu       sync.Mutex
	cache    map[string]*lruEntry
	order    []string // LRU order, 0 = oldest
	pending  sync.Map // string -> chan struct{}
	manifest map[string]manifestEntry
	root     string
}

type manifestEntry struct {
	CachePath string
	SHA256    string
	Bytes     int
	MIME      string
}

func NewLoader(repoRoot string) (*Loader, error) {
	ld := &Loader{
		cache: make(map[string]*lruEntry),
		root:  repoRoot,
	}
	// Try to load frozen manifest (26 resources).
	manifestPath := filepath.Join(repoRoot, "src/utils/media/testdata/visual/baseline/resource-manifest.json")
	if alt := os.Getenv("YOGA_SKIA_ASSET_MANIFEST"); alt != "" {
		manifestPath = alt
		if !filepath.IsAbs(manifestPath) {
			manifestPath = filepath.Join(repoRoot, manifestPath)
		}
	}
	if err := ld.loadManifest(manifestPath); err != nil {
		// Non-fatal in stub mode — tests that need manifest will skip gracefully.
		_ = err
	}
	return ld, nil
}

func (ld *Loader) loadManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Use generic JSON parse without importing large struct — minimal.
	// Reuse simple scan: look for cachePath/sha256 pairs via stdlib json.
	// Ponytail: delegate to encoding/json minimal unmarshal.
	type res struct {
		CachePath string `json:"cachePath"`
		SHA256    string `json:"sha256"`
		Bytes     int    `json:"bytes"`
		MIME      string `json:"mime"`
		Alias     string `json:"requestAlias"`
		SourceURL string `json:"sourceURL"`
	}
	var doc struct {
		Status    string `json:"status"`
		Resources []res  `json:"resources"`
	}
	if err := jsonUnmarshal(data, &doc); err != nil {
		return err
	}
	if doc.Status != "frozen" || len(doc.Resources) != 26 {
		return fmt.Errorf("manifest must be frozen with 26 resources")
	}
	m := make(map[string]manifestEntry, 26*2+8)
	for _, r := range doc.Resources {
		e := manifestEntry{CachePath: r.CachePath, SHA256: strings.ToLower(r.SHA256), Bytes: r.Bytes, MIME: r.MIME}
		m[r.Alias] = e
		m[r.SourceURL] = e
		// Also index by basename for fixture alias resolution
		m[filepath.Base(r.CachePath)] = e
	}
	// 8 fixture aliases point into manifest by basename
	for alias, base := range frozenFixtureAliases {
		if e, ok := m[base]; ok {
			m[alias] = e
		}
	}
	ld.manifest = m
	return nil
}

// Load validates and decodes an asset source (data URI, local path, or manifest alias).
// It mirrors renderer/lib/assets.mjs materialize() + validateImageBytes() + normalizeWebp().
func (ld *Loader) Load(source string) (*Image, error) {
	if source == "" {
		return nil, errors.New("ASSET_SOURCE_MISSING")
	}
	// Deduplicate concurrent loads for same source (like pending Map in assets.mjs)
	ch := make(chan struct{})
	actual, loaded := ld.pending.LoadOrStore(source, ch)
	if loaded {
		<-actual.(chan struct{})
		if img := ld.getCached(source); img != nil {
			return img, nil
		}
	} else {
		defer func() { close(ch); ld.pending.Delete(source) }()
	}
	if img := ld.getCached(source); img != nil {
		return img, nil
	}
	img, err := ld.loadUncached(source)
	if err != nil {
		return nil, err
	}
	ld.putCached(source, img)
	return img, nil
}

func (ld *Loader) getCached(key string) *Image {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	e, ok := ld.cache[key]
	if !ok || time.Now().After(e.expiresAt) {
		if ok {
			delete(ld.cache, key)
			// remove from order
			for i, k := range ld.order {
				if k == key {
					ld.order = append(ld.order[:i], ld.order[i+1:]...)
					break
				}
			}
		}
		return nil
	}
	// Move to MRU
	for i, k := range ld.order {
		if k == key {
			ld.order = append(ld.order[:i], ld.order[i+1:]...)
			ld.order = append(ld.order, key)
			break
		}
	}
	return e.img
}

func (ld *Loader) putCached(key string, img *Image) {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	// evict expired
	now := time.Now()
	for k, e := range ld.cache {
		if now.After(e.expiresAt) {
			delete(ld.cache, k)
		}
	}
	// evict LRU if over capacity
	for len(ld.cache) >= CacheMaxEntries && len(ld.order) > 0 {
		oldest := ld.order[0]
		ld.order = ld.order[1:]
		delete(ld.cache, oldest)
	}
	ld.cache[key] = &lruEntry{img: img, expiresAt: now.Add(CacheTTL)}
	// order dedup
	for i, k := range ld.order {
		if k == key {
			ld.order = append(ld.order[:i], ld.order[i+1:]...)
			break
		}
	}
	ld.order = append(ld.order, key)
}

func (ld *Loader) loadUncached(source string) (*Image, error) {
	// 1) Manifest alias fast path — validates SHA/bytes/MIME before SkData
	if ld.manifest != nil {
		if entry, ok := ld.manifest[source]; ok {
			cacheRoot := filepath.Join(ld.root, "src/utils/media/testdata/visual/baseline")
			abs := filepath.Join(cacheRoot, entry.CachePath)
			// Also handle basename-only entries
			if _, err := os.Stat(abs); err != nil {
				abs = filepath.Join(cacheRoot, filepath.Base(entry.CachePath))
			}
			b, err := os.ReadFile(abs)
			if err != nil {
				return nil, fmt.Errorf("ASSET_MANIFEST_CACHE_MISSING: %w", err)
			}
			if len(b) != entry.Bytes {
				return nil, fmt.Errorf("ASSET_MANIFEST_HASH: bytes %d != %d", len(b), entry.Bytes)
			}
			sum := sha256.Sum256(b)
			if hex.EncodeToString(sum[:]) != entry.SHA256 {
				return nil, fmt.Errorf("ASSET_MANIFEST_HASH: sha mismatch for %s", entry.CachePath)
			}
			if err := validateMagic(b, entry.MIME); err != nil {
				return nil, err
			}
			// WebP normalization
			norm, mime2, ww, hh, err := maybeNormalizeWebP(b, entry.MIME)
			if err != nil {
				return nil, err
			}
			sha2 := sha256.Sum256(norm)
			return &Image{Bytes: norm, MIME: mime2, Width: ww, Height: hh, SHA256: hex.EncodeToString(sha2[:])}, nil
		}
		// Remote not in manifest while manifest is frozen → fatal (mirrors assets.mjs)
		if isRemote(source) {
			return nil, errors.New("ASSET_MANIFEST_MISSING: remote asset absent from frozen manifest")
		}
	}
	// 2) data: URI
	if strings.HasPrefix(source, "data:") {
		b, mime, err := parseDataURI(source)
		if err != nil {
			return nil, err
		}
		if err := validateMagic(b, mime); err != nil {
			return nil, err
		}
		norm, mime2, w, h, err := maybeNormalizeWebP(b, mime)
		if err != nil {
			return nil, err
		}
		sha := sha256.Sum256(norm)
		return &Image{Bytes: norm, MIME: mime2, Width: w, Height: h, SHA256: hex.EncodeToString(sha[:])}, nil
	}
	// 3) local file (relative to repo root, like renderer assets)
	if !isRemote(source) {
		rel := strings.TrimPrefix(source, "/")
		rel = strings.TrimPrefix(rel, "file://")
		abs := filepath.Join(ld.root, rel)
		// Prevent directory escape (isInside)
		if !isInside(ld.root, abs) {
			return nil, errors.New("ASSET_LOCAL_ESCAPE")
		}
		b, err := os.ReadFile(abs)
		if err != nil {
			return nil, fmt.Errorf("ASSET_NOT_FOUND: %w", err)
		}
		if len(b) > MaxImageBytes {
			return nil, errors.New("ASSET_TOO_LARGE")
		}
		mime := mimeForPath(abs)
		if mime == "" {
			// fallback detect
			if d := detectMIME(b); d != "" {
				mime = d
			} else {
				return nil, errors.New("ASSET_MIME_UNSUPPORTED")
			}
		}
		if !allowedMIME[mime] {
			return nil, fmt.Errorf("ASSET_MIME_UNSUPPORTED: %s", mime)
		}
		if err := validateMagic(b, mime); err != nil {
			return nil, err
		}
		norm, mime2, w, h, err := maybeNormalizeWebP(b, mime)
		if err != nil {
			return nil, err
		}
		sha := sha256.Sum256(norm)
		return &Image{Bytes: norm, MIME: mime2, Width: w, Height: h, SHA256: hex.EncodeToString(sha[:])}, nil
	}
	return nil, errors.New("ASSET_INVALID_SOURCE")
}

func maybeNormalizeWebP(b []byte, mime string) ([]byte, string, int, int, error) {
	if mime != "image/webp" {
		w, h, _ := imageDimensions(b, mime)
		return b, mime, w, h, nil
	}
	// Try Go native WebP decode (no cgo), then re-encode as PNG — mirrors skia.png()
	// If skia_use_libwebp is true, caller can skip this and use b directly with SkCodec.
	img, err := webp.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("ASSET_WEBP_DECODE: %w", err)
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 || w > MaxDecodedWidth || h > MaxDecodedHeight || w*h > MaxDecodedPixels {
		return nil, "", 0, 0, errors.New("ASSET_DECODE_LIMIT")
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, "", 0, 0, err
	}
	if buf.Len() > MaxDecodedBytes {
		return nil, "", 0, 0, errors.New("ASSET_DECODE_LIMIT")
	}
	return buf.Bytes(), "image/png", w, h, nil
}

func parseDataURI(s string) ([]byte, string, error) {
	// data:[mime][;base64],payload — minimal, matches assets.mjs parseDataUri
	comma := strings.Index(s, ",")
	if comma < 0 {
		return nil, "", errors.New("ASSET_INVALID_SOURCE")
	}
	meta := s[5:comma]
	payload := s[comma+1:]
	semi := strings.Index(meta, ";")
	mime := meta
	isBase64 := false
	if semi >= 0 {
		mime = meta[:semi]
		isBase64 = meta[semi+1:] == "base64"
	}
	mime = strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0]))
	if !allowedMIME[mime] {
		return nil, "", fmt.Errorf("ASSET_MIME_UNSUPPORTED: %s", mime)
	}
	var b []byte
	if isBase64 {
		// strip whitespace like assets.mjs
		payload = strings.ReplaceAll(payload, " ", "")
		payload = strings.ReplaceAll(payload, "\n", "")
		payload = strings.ReplaceAll(payload, "\r", "")
		decoded, err := decodeBase64(payload)
		if err != nil {
			return nil, "", err
		}
		b = decoded
	} else {
		b = []byte(payload)
	}
	if len(b) > MaxImageBytes {
		return nil, "", errors.New("ASSET_TOO_LARGE")
	}
	return b, mime, nil
}

func decodeBase64(s string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	// try raw without padding
	b, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}
	return nil, fmt.Errorf("ASSET_INVALID_SOURCE: base64: %w", err)
}

func detectMIME(b []byte) string {
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return "image/png"
	}
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return "image/jpeg"
	}
	if len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WEBP" {
		return "image/webp"
	}
	if len(b) >= 6 && (string(b[0:6]) == "GIF87a" || string(b[0:6]) == "GIF89a") {
		return "image/gif"
	}
	text := strings.TrimSpace(string(b[:min(512, len(b))]))
	if strings.HasPrefix(strings.ToLower(text), "<svg") || strings.Contains(strings.ToLower(text[:min(512, len(text))]), "<svg") {
		return "image/svg+xml"
	}
	return ""
}

func validateMagic(b []byte, mime string) error {
	if !strings.HasPrefix(mime, "image/") {
		return nil
	}
	detected := detectMIME(b)
	if detected == "" || detected != mime {
		// Allow jpeg alias
		if mime == "image/jpeg" && detected == "image/jpeg" {
			return nil
		}
		return fmt.Errorf("ASSET_MAGIC_MISMATCH: declared %s != detected %s", mime, detected)
	}
	return nil
}

func imageDimensions(b []byte, mime string) (int, int, error) {
	var cfg image.Config
	var err error
	switch mime {
	case "image/png":
		cfg, err = png.DecodeConfig(bytes.NewReader(b))
	case "image/jpeg":
		cfg, err = jpeg.DecodeConfig(bytes.NewReader(b))
	case "image/webp":
		img, e := webp.Decode(bytes.NewReader(b))
		if e != nil {
			return 0, 0, e
		}
		bounds := img.Bounds()
		return bounds.Dx(), bounds.Dy(), nil
	default:
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func mimeForPath(p string) string {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".avif":
		return "image/avif"
	case ".svg":
		return "image/svg+xml"
	default:
		return ""
	}
}

func isRemote(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

func isInside(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func jsonUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
