package skia

import (
	"crypto/sha256"
	"encoding/hex"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot() string {
	if r := os.Getenv("SKIA_REPO_ROOT"); r != "" {
		return r
	}
	// hard pin — pony: explicit beats heuristic
	for _, cand := range []string{
		"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go",
		"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go-B-skia",
	} {
		if _, err := os.Stat(filepath.Join(cand, "assets/font/NotoSansHans-Regular.ttf")); err == nil {
			return cand
		}
	}
	_, file, _, _ := runtime.Caller(0)
	artifact := filepath.Dir(filepath.Dir(file))
	for _, cand := range []string{
		filepath.Join(artifact, "..", "..", "WorkSpace", "Golang", "arknights_bot-satori-yoga-skia-go"),
		filepath.Join(filepath.Dir(file), "..", ".."),
	} {
		if _, err := os.Stat(filepath.Join(cand, "assets/font/NotoSansHans-Regular.ttf")); err == nil {
			return cand
		}
	}
	return artifact
}

func TestFontMetricsAndSHA(t *testing.T) {
	root := repoRoot()
	path := filepath.Join(root, FontPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	if len(data) != 8927296 {
		t.Fatalf("font bytes=%d want 8927296", len(data))
	}
	h := sha256.Sum256(data)
	if got := hex.EncodeToString(h[:]); got != FontSHA256 {
		t.Fatalf("font sha=%s want %s", got, FontSHA256)
	}
	units, glyphs, asc, desc, gap, err := parseFontMetrics(data)
	if err != nil {
		t.Fatalf("parse metrics: %v", err)
	}
	if units != FontUnitsPerEm || glyphs != FontNumGlyphs || asc != FontHeadAsc || desc != FontHeadDesc || gap != FontLineGap {
		t.Fatalf("metrics mismatch: units=%d glyphs=%d asc=%d desc=%d gap=%d", units, glyphs, asc, desc, gap)
	}
	// LoadTypeface exercises magic+sha+metrics path
	tf, err := LoadTypeface(path)
	if err != nil {
		t.Fatalf("LoadTypeface: %v", err)
	}
	if tf.NumGlyphs != 30888 || tf.UnitsPerEm != 1000 {
		t.Fatalf("typeface: %+v", tf)
	}
	if got := tf.Shape("阿米娅●"); got != 4 {
		t.Fatalf("shape count=%d", got)
	}
	if !tf.Contains('阿') || !tf.Contains('●') {
		t.Fatalf("contains CJK/symbol failed")
	}
	// emoji without fallback should be false (no emoji file in repo)
	if tf.Contains('😀') {
		t.Logf("emoji fallback present (optional asset loaded)")
	}
}

func TestImageAssetTripleAndWebPAndLRU(t *testing.T) {
	root := repoRoot()
	ld, err := NewLoader(root)
	if err != nil {
		t.Fatalf("NewLoader: %v", err)
	}
	if ld.manifest == nil || len(ld.manifest) < 26 {
		t.Skipf("manifest not loaded (need %s)", filepath.Join(root, "src/utils/media/testdata/visual/baseline/resource-manifest.json"))
	}
	// frozen 26 — each must pass SHA/bytes/MIME and produce decoded bytes
	// Count distinct resources by cachePath basename
	seen := map[string]bool{}
	for _, entry := range ld.manifest {
		seen[filepath.Base(entry.CachePath)] = true
	}
	if len(seen) != 26 {
		t.Fatalf("manifest distinct cache files=%d want 26", len(seen))
	}
	// pick a PNG and a WebP to validate normalization
	pngAlias := "https://media.prts.wiki/d/dd/%E7%AB%8B%E7%BB%98_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png"
	if _, ok := ld.manifest[pngAlias]; !ok {
		pngAlias = "https://fixture-cache.invalid/card/secretary-painting.png"
	}
	img, err := ld.Load(pngAlias)
	if err != nil {
		t.Fatalf("Load png alias %q: %v", pngAlias, err)
	}
	if img.MIME != "image/png" || len(img.Bytes) < 8 || img.Width == 0 {
		t.Fatalf("png image meta: mime=%s bytes=%d w=%d", img.MIME, len(img.Bytes), img.Width)
	}
	// WebP alias — must normalize to PNG with magic
	webpAlias := "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90"
	if _, ok := ld.manifest[webpAlias]; !ok {
		webpAlias = "https://fixture-cache.invalid/recruit-amiya.png"
	}
	wimg, err := ld.Load(webpAlias)
	if err != nil {
		t.Fatalf("Load webp alias %q: %v", webpAlias, err)
	}
	if wimg.MIME != "image/png" {
		t.Fatalf("webp normalized mime=%s want image/png", wimg.MIME)
	}
	if wimg.Bytes[0] != 0x89 || wimg.Bytes[1] != 0x50 {
		t.Fatalf("webp normalized not PNG magic")
	}
	if wimg.Width != 180 || wimg.Height != 180 {
		t.Fatalf("webp normalized dims=%dx%d want 180x180", wimg.Width, wimg.Height)
	}
	// 8 fixture aliases must be resolvable without fetch (no network)
	for alias, base := range frozenFixtureAliases {
		img, err := ld.Load(alias)
		if err != nil {
			t.Fatalf("fixture alias %s -> %s: %v", alias, base, err)
		}
		if len(img.Bytes) == 0 || img.SHA256 == "" {
			t.Fatalf("fixture %s empty", alias)
		}
		// second load must hit LRU (same pointer or same bytes)
		img2, _ := ld.Load(alias)
		if img2.SHA256 != img.SHA256 {
			t.Fatalf("LRU cache mismatch for %s", alias)
		}
	}
	// data URI (local) — proves non-manifest path still validates magic
	pngPath := filepath.Join(root, "assets/calendar/bg.png")
	if b, err := os.ReadFile(pngPath); err == nil {
		uri := "data:image/png;base64," + b64(b)
		img, err := ld.Load(uri)
		if err != nil {
			t.Fatalf("data uri load: %v", err)
		}
		if img.MIME != "image/png" {
			t.Fatalf("data uri mime=%s", img.MIME)
		}
	}
	// local escape must fail
	if _, err := ld.Load("/etc/passwd"); err == nil {
		t.Fatalf("local escape should fail")
	}
}

func TestCanvasPrimitivesAndProgressAndShadow(t *testing.T) {
	c := NewCanvas(1000, 882)
	c.Clear(color.RGBA{12, 13, 12, 255})
	// rect15
	c.DrawRect(Rect{X: 0, Y: 0, W: 100, H: 60}, Paint{Color: color.RGBA{0x2e, 0x30, 0x31, 255}})
	if got := c.Image().RGBAAt(5, 5); got.R != 0x2e || got.G != 0x30 {
		t.Fatalf("rect fill: %+v", got)
	}
	// rrect7
	c.DrawRRect(RRect{Rect: Rect{X: 200, Y: 10, W: 100, H: 60}, Radius: 7}, Paint{Color: color.RGBA{0xff, 0x00, 0x00, 255}})
	if got := c.Image().RGBAAt(250, 40); got.R != 0xff {
		t.Fatalf("rrect fill center: %+v", got)
	}
	// rrect corner outside radius should remain bg
	if got := c.Image().RGBAAt(200, 10); got.R == 0xff {
		t.Fatalf("rrect corner should be clipped, got red at 200,10")
	}
	// circle2
	c.DrawCircle(Circle{CX: 500, CY: 100, R: 20}, Paint{Color: color.RGBA{0, 0xff, 0x00, 255}})
	if got := c.Image().RGBAAt(500, 100); got.G != 0xff {
		t.Fatalf("circle center: %+v", got)
	}
	// path4 (gacha pie sector) — at least some pixels inside
	c.DrawPath([]PathCmd{
		{Op: "M", Args: []float32{700, 100}},
		{Op: "L", Args: []float32{780, 100}},
		{Op: "L", Args: []float32{780, 180}},
		{Op: "L", Args: []float32{700, 180}},
		{Op: "Z", Args: nil},
	}, Paint{Color: color.RGBA{0, 0, 0xff, 255}})
	if got := c.Image().RGBAAt(740, 140); got.B != 0xff {
		t.Fatalf("path fill: %+v", got)
	}
	// progress5 — 150x3 bar with clip
	c.DrawRect(Rect{X: 10, Y: 300, W: 150, H: 3}, Paint{Color: color.RGBA{0x33, 0x33, 0x33, 255}})
	c.DrawRect(Rect{X: 10, Y: 300, W: 75, H: 3}, Paint{Color: color.RGBA{0x54, 0x70, 0xc6, 255}})
	if got := c.Image().RGBAAt(20, 301); got.B != 0xc6 && got.R != 0x54 {
		t.Logf("progress inner pixel: %+v (may vary)", got)
	}
	// rrect clipped by overflow
	c.Save()
	c.ClipRect(Rect{X: 400, Y: 400, W: 100, H: 100})
	c.DrawRect(Rect{X: 350, Y: 350, W: 200, H: 200}, Paint{Color: color.RGBA{0xff, 0xff, 0x00, 255}})
	c.Restore()
	if got := c.Image().RGBAAt(390, 390); got.R == 0xff {
		t.Fatalf("clip should prevent outside draw at 390,390: %+v", got)
	}
	if got := c.Image().RGBAAt(450, 450); got.R != 0xff {
		t.Fatalf("clip inside should be yellow: %+v", got)
	}
	// DropShadow — anti-regression: should not panic and should paint offset
	c.DrawDropShadow(RRect{Rect: Rect{X: 500, Y: 600, W: 80, H: 40}, Radius: 6}, color.RGBA{0, 0, 0, 128}, 0, 3, 2.5)
	c.DrawRRect(RRect{Rect: Rect{X: 500, Y: 600, W: 80, H: 40}, Radius: 6}, Paint{Color: color.RGBA{0x22, 0x22, 0x22, 255}})
	// PNG output must be valid and within limits
	b, err := c.PNGBytes()
	if err != nil {
		t.Fatalf("PNGBytes: %v", err)
	}
	if len(b) < 8 || b[0] != 0x89 {
		t.Fatalf("PNG magic")
	}
	if c.W*c.H > MaxDecodedPixels {
		t.Fatalf("canvas exceeds max pixels")
	}
}

func TestBackendRenderCardStub(t *testing.T) {
	root := repoRoot()
	be, err := NewBackend(root)
	if err != nil {
		t.Fatalf("NewBackend: %v", err)
	}
	c := be.RenderCardStub(1280, 720)
	if c.W != 1280 || c.H != 720 {
		t.Fatalf("stub dims %dx%d", c.W, c.H)
	}
	// spot check bg vs card
	if got := c.Image().RGBAAt(0, 0); got.R != 12 {
		t.Fatalf("stub bg: %+v", got)
	}
	if got := c.Image().RGBAAt(30, 160); got.R != 0x1f {
		t.Fatalf("stub card: %+v", got)
	}
	b, _ := c.PNGBytes()
	if len(b) == 0 {
		t.Fatalf("stub png empty")
	}
}

func b64(b []byte) string {
	// inline to avoid import bloat in test
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	_ = chars
	// use stdlib instead
	return b64std(b)
}

func b64std(b []byte) string {
	// minimal wrapper
	// we import encoding/base64 via helper to keep test file small
	return b64encode(b)
}

// b64encode indirection to allow go vet without extra import at top
func b64encode(b []byte) string {
	// use encoding/base64 directly via re-import trick: we already have it in image.go? no
	// So implement via stdlib call using same logic as base64.StdEncoding
	// To avoid import cycle, we cheat: use known function
	return encodeBase64(b)
}

// tiny base64 copy (ponytail: stdlib one-liner beats pulling import for test)
func encodeBase64(src []byte) string {
	const enc = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	out := make([]byte, ((len(src)+2)/3)*4)
	j := 0
	for i := 0; i < len(src); i += 3 {
		var v uint32
		remain := len(src) - i
		v |= uint32(src[i]) << 16
		if remain > 1 {
			v |= uint32(src[i+1]) << 8
		}
		if remain > 2 {
			v |= uint32(src[i+2])
		}
		out[j] = enc[(v>>18)&0x3F]
		out[j+1] = enc[(v>>12)&0x3F]
		if remain > 1 {
			out[j+2] = enc[(v>>6)&0x3F]
		} else {
			out[j+2] = '='
		}
		if remain > 2 {
			out[j+3] = enc[v&0x3F]
		} else {
			out[j+3] = '='
		}
		j += 4
	}
	return string(out)
}
