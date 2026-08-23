package skia

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"sync"
)

// Expected metrics for assets/font/NotoSansHans-Regular.ttf (see appendix font-stack).
const (
	FontPath           = "assets/font/NotoSansHans-Regular.ttf"
	FontSHA256         = "96843f2d70e886fbe41eaccc33cb35ead02553cd57bc45fd3ddc03d0368313a5"
	FontUnitsPerEm     = 1000
	FontNumGlyphs      = 30888
	FontHeadAsc        = 880
	FontHeadDesc       = -120
	FontLineGap        = 500
	FontMagicTrueType  = "\x00\x01\x00\x00" // sfnt 1.0
)

// Typeface is the Go handle for SkTypeface (stub or real).
type Typeface struct {
	Data        []byte
	SHA256      string
	UnitsPerEm  int
	NumGlyphs   int
	Ascender    int
	Descender   int
	LineGap     int
	// Fallback chain — nil means no emoji fallback loaded.
	Fallback *Typeface
}

var (
	fontOnce sync.Once
	fontMemo *Typeface
	fontErr  error
)

// LoadTypeface loads NotoSansHans, validates magic+SHA+metrics, and caches.
// Equivalent to SkData::MakeWithCopy + SkTypeface::MakeFromData + SkFontMgr register.
func LoadTypeface(path string) (*Typeface, error) {
	if path == "" {
		path = FontPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 4 || string(data[:4]) != FontMagicTrueType {
		// Also allow OTTO/CFF fallback gracefully, but our file is TrueType.
		return nil, errors.New("font magic mismatch: want TrueType 00 01 00 00")
	}
	sha := sha256.Sum256(data)
	hexSHA := hex.EncodeToString(sha[:])
	if hexSHA != FontSHA256 {
		return nil, errors.New("font SHA256 mismatch: " + hexSHA)
	}
	units, glyphs, asc, desc, gap, err := parseFontMetrics(data)
	if err != nil {
		return nil, err
	}
	if units != FontUnitsPerEm || glyphs != FontNumGlyphs {
		return nil, errors.New("font metrics mismatch")
	}
	tf := &Typeface{
		Data:       data,
		SHA256:     hexSHA,
		UnitsPerEm: units,
		NumGlyphs:  glyphs,
		Ascender:   asc,
		Descender:  desc,
		LineGap:    gap,
	}
	// Emoji fallback — ponytail: try optional NotoColorEmoji.ttf, ignore if absent.
	if fb, err := tryLoadFallback(); err == nil {
		tf.Fallback = fb
	}
	return tf, nil
}

// LoadTypefaceCached — sync.Once singleton mirroring yoga-render fontPromise.
func LoadTypefaceCached() (*Typeface, error) {
	fontOnce.Do(func() {
		fontMemo, fontErr = LoadTypeface(FontPath)
	})
	return fontMemo, fontErr
}

func tryLoadFallback() (*Typeface, error) {
	// Optional asset, not in repo by default; presence enables emoji shaping.
	candidates := []string{"assets/font/NotoColorEmoji.ttf", "assets/font/NotoSansSymbols.ttf"}
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if len(b) < 4 {
			continue
		}
		sha := sha256.Sum256(b)
		return &Typeface{Data: b, SHA256: hex.EncodeToString(sha[:])}, nil
	}
	return nil, errors.New("no fallback font")
}

// parseFontMetrics reads head/hhea/maxp without external deps.
func parseFontMetrics(data []byte) (unitsPerEm, numGlyphs, asc, desc, gap int, err error) {
	if len(data) < 12 {
		return 0, 0, 0, 0, 0, errors.New("font too small")
	}
	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	if len(data) < 12+numTables*16 {
		return 0, 0, 0, 0, 0, errors.New("font table directory truncated")
	}
	var headOff, hheaOff, maxpOff int = -1, -1, -1
	for i := 0; i < numTables; i++ {
		base := 12 + i*16
		tag := string(data[base : base+4])
		off := int(binary.BigEndian.Uint32(data[base+8 : base+12]))
		switch tag {
		case "head":
			headOff = off
		case "hhea":
			hheaOff = off
		case "maxp":
			maxpOff = off
		}
	}
	if headOff < 0 || hheaOff < 0 || maxpOff < 0 {
		return 0, 0, 0, 0, 0, errors.New("missing head/hhea/maxp")
	}
	// head: unitsPerEm at +18 (uint16), bbox at +36..44
	if headOff+54 > len(data) {
		return 0, 0, 0, 0, 0, errors.New("head truncated")
	}
	unitsPerEm = int(binary.BigEndian.Uint16(data[headOff+18 : headOff+20]))
	// hhea: asc at +4 (int16), desc at +6 (int16), lineGap at +8 (int16)
	if hheaOff+10 > len(data) {
		return 0, 0, 0, 0, 0, errors.New("hhea truncated")
	}
	asc = int(int16(binary.BigEndian.Uint16(data[hheaOff+4 : hheaOff+6])))
	desc = int(int16(binary.BigEndian.Uint16(data[hheaOff+6 : hheaOff+8])))
	gap = int(int16(binary.BigEndian.Uint16(data[hheaOff+8 : hheaOff+10])))
	// maxp: numGlyphs at +4 (uint16)
	if maxpOff+6 > len(data) {
		return 0, 0, 0, 0, 0, errors.New("maxp truncated")
	}
	numGlyphs = int(binary.BigEndian.Uint16(data[maxpOff+4 : maxpOff+6]))
	return unitsPerEm, numGlyphs, asc, desc, gap, nil
}

// Shape is the HarfBuzz-equivalent stub: returns glyph count for a string.
// Real Skia path (//go:build skia) would call SkShaper::Shape → SkTextBlob.
// Ponytail: naive rune count is enough for layout metrics in stub mode;
// upgrade when variable shaping (ligatures/Arabic) matters.
func (tf *Typeface) Shape(text string) int {
	if tf == nil {
		return 0
	}
	// ponytail: global count, per-run shaping if complex scripts arrive
	return len([]rune(text))
}

// Contains reports whether the typeface (or fallback) likely covers rune r.
// Stub: NotoSansHans covers 30888 glyphs including CJK/Symbols/●, but not emoji.
func (tf *Typeface) Contains(r rune) bool {
	if tf == nil {
		return false
	}
	if r >= 0x1F300 && r <= 0x1FAFF { // Emoji range
		return tf.Fallback != nil && len(tf.Fallback.Data) > 0
	}
	return true // 30k CJK covers business chars
}
