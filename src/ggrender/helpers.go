// Package ggrender provides pure-Go gg rendering for 16 bot scenes.
// Replaces Playwright/Chromium HTML screenshots with Skia-style gg drawing.
// ponytail: minimal helpers reused across scenes, no extra abstraction.
package ggrender

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

var (
	// AssetRoot absolute path for assets.
	AssetRoot = "C:/WorkSpace/Golang/arknights_bot/assets"
	// FontCandidates tried in order.
	FontCandidates = []string{
		"C:/WorkSpace/Golang/arknights_bot/assets/font/NotoSansHans-Regular.ttf",
		"C:/Windows/Fonts/msyh.ttc",
		"C:/Windows/Fonts/simhei.ttf",
		"C:/Windows/Fonts/msyh.ttf",
	}
)

var amiyaPath = filepath.Join(AssetRoot, "common", "amiya.png")
var tagRe = regexp.MustCompile(`<[^>]*>`)

// StripHTML removes tags.
func StripHTML(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	repl := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&nbsp;", " ", "&#39;", "'", "&quot;", "\"")
	return strings.TrimSpace(repl.Replace(s))
}

func itoa(i int) string { return strconv.Itoa(i) }

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func rarityColor(r int) (int, int, int) {
	switch r {
	case 6:
		return 240, 180, 40
	case 5:
		return 170, 110, 220
	case 4:
		return 90, 160, 230
	case 3:
		return 150, 150, 150
	default:
		return 150, 150, 150
	}
}

// LoadImage loads local image.
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

// Decode decodes any registered format.
func Decode(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

// tryLocal loads local asset or fallback.
func tryLocal(rel string) image.Image {
	if img, err := LoadImage(filepath.Join(AssetRoot, rel)); err == nil {
		return img
	}
	if img, err := LoadImage(amiyaPath); err == nil {
		return img
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

// fetch downloads remote image with 6s timeout (mirrors poc/ggrender/helpers.go).
func fetch(url string) (image.Image, error) {
	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return Decode(resp.Body)
}

// FetchImage tries remote URL with 6s timeout, falls back to local asset, then 1x1 pixel.
// Matches poc logic and Playwright's resource error fallback.
func FetchImage(url, fallbackPath string) image.Image {
	if url != "" {
		if img, err := fetch(url); err == nil {
			return img
		}
	}
	if img, err := LoadImage(fallbackPath); err == nil {
		return img
	}
	if img, err := LoadImage(amiyaPath); err == nil {
		return img
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

// ScaleContain fits within w*h.
func ScaleContain(img image.Image, w, h int) *image.RGBA {
	srcW, srcH := img.Bounds().Dx(), img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}
	scale := math.Min(float64(w)/float64(srcW), float64(h)/float64(srcH))
	nw, nh := int(math.Round(float64(srcW)*scale)), int(math.Round(float64(srcH)*scale))
	if nw <= 0 {
		nw = 1
	}
	if nh <= 0 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// ScaleCover covers w*h.
func ScaleCover(img image.Image, w, h int) *image.RGBA {
	srcW, srcH := img.Bounds().Dx(), img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}
	scale := math.Max(float64(w)/float64(srcW), float64(h)/float64(srcH))
	nw, nh := int(math.Round(float64(srcW)*scale)), int(math.Round(float64(srcH)*scale))
	tmp := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.NearestNeighbor.Scale(tmp, tmp.Bounds(), img, img.Bounds(), draw.Over, nil)
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	sx, sy := (nw-w)/2, (nh-h)/2
	draw.Draw(out, out.Bounds(), tmp, image.Pt(sx, sy), draw.Over)
	return out
}

// ScaleExact stretches to w*h.
func ScaleExact(img image.Image, w, h int) *image.RGBA {
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// FillBackground fills canvas.
func FillBackground(dc *gg.Context, r, g, b int) {
	dc.SetRGB255(r, g, b)
	dc.Clear()
}

// RoundRect fills rounded rect.
func RoundRect(dc *gg.Context, x, y, w, h, r float64) {
	dc.DrawRoundedRectangle(x, y, w, h, r)
	dc.Fill()
}

// StrokeRoundRect strokes rounded rect.
func StrokeRoundRect(dc *gg.Context, x, y, w, h, r float64) {
	dc.DrawRoundedRectangle(x, y, w, h, r)
	dc.Stroke()
}

// LoadDefaultFont tries candidates.
func LoadDefaultFont(dc *gg.Context, size float64) error {
	var last error
	for _, p := range FontCandidates {
		if err := dc.LoadFontFace(p, size); err == nil {
			return nil
		} else {
			last = err
		}
	}
	return last
}

// LoadFont loads specific path.
func LoadFont(dc *gg.Context, path string, size float64) error {
	return dc.LoadFontFace(path, size)
}

// EncodePNG returns PNG bytes.
func EncodePNG(dc *gg.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AssetPath joins.
func AssetPath(rel string) string { return filepath.Join(AssetRoot, rel) }

func profFirst(p string) string {
	if p == "" {
		return "?"
	}
	return string([]rune(p)[0:1])
}

// DrawPortraitTile draws portrait tile with remote fetch (6s timeout) + local fallback.
// Uses same cache strategy as poc: portraitURL remotes, profession icon local.
func DrawPortraitTile(dc *gg.Context, x, y, w, h int, portraitURL, profession string, rarity, level int, name string) {
	// portrait: remote fetch with fallback to rarity-colored placeholder
	var port image.Image
	if portraitURL != "" {
		port = FetchImage(portraitURL, amiyaPath)
	} else {
		port = tryLocal("common/amiya.png")
	}
	dc.DrawImage(ScaleCover(port, w, h), x, y)
	// subtle rarity tint overlay
	r, g, b := rarityColor(rarity)
	dc.SetRGBA255(r, g, b, 22)
	dc.DrawRectangle(float64(x), float64(y), float64(w), float64(h))
	dc.Fill()
	// profession icon if local exists else letter
	if ic, err := LoadImage(filepath.Join(AssetRoot, "box", profession+".png")); err == nil {
		dc.DrawImage(ScaleExact(ic, 15, 15), x+3, y+3)
	} else {
		dc.SetRGB255(255, 255, 255)
		_ = LoadDefaultFont(dc, 11)
		dc.DrawStringAnchored(profFirst(profession), float64(x+10), float64(y+10), 0.5, 0.5)
	}
	// rarity dots top-right
	dc.SetRGB255(r, g, b)
	for i := 0; i < rarity && i < 6; i++ {
		dc.DrawCircle(float64(x+w-8-i*7), float64(y+8), 2.5)
		dc.Fill()
	}
	// level badge
	cx, cy := float64(x+12), float64(y+h-12)
	dc.SetRGBA255(0, 0, 0, 170)
	dc.DrawCircle(cx, cy, 10)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	_ = LoadDefaultFont(dc, 10)
	dc.DrawStringAnchored(itoa(level), cx, cy, 0.5, 0.5)
	// name bar
	dc.SetRGBA255(0, 0, 0, 180)
	dc.DrawRectangle(float64(x), float64(y+h-13), float64(w), 13)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	_ = LoadDefaultFont(dc, 9)
	if name != "" {
		dc.DrawStringAnchored(name, float64(x+w/2), float64(y+h)-6.5, 0.5, 0.5)
	}
}

// ProgressBar draws rounded progress.
func ProgressBar(dc *gg.Context, x, y, w, h float64, frac float64, r, g, b int) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	dc.SetRGBA255(255, 255, 255, 40)
	dc.DrawRoundedRectangle(x, y, w, h, h/2)
	dc.Fill()
	dc.SetRGB255(r, g, b)
	dc.DrawRoundedRectangle(x, y, w*frac, h, h/2)
	dc.Fill()
}

// SectionTitle draws label.
func SectionTitle(dc *gg.Context, s string, x, y float64) {
	dc.SetRGB255(120, 200, 220)
	_ = LoadDefaultFont(dc, 16)
	dc.DrawString(s, x, y)
}

// drawHeaderLabel draws top label bar used by Box/Missing etc.
func drawHeaderLabel(dc *gg.Context, title string, mainW, labelH int) {
	label := tryLocal("help/label.png")
	lw, lh := label.Bounds().Dx(), label.Bounds().Dy()
	if lw > 0 && lh > 0 {
		dc.DrawImage(ScaleContain(label, mainW, labelH), 0, 0)
	} else {
		dc.SetRGB255(60, 60, 70)
		dc.DrawRectangle(0, 0, float64(mainW), float64(labelH))
		dc.Fill()
	}
	dc.SetRGB255(255, 255, 255)
	_ = LoadDefaultFont(dc, 26)
	dc.DrawString(title, 25, float64(labelH)-18)
}

// fillRoundedCard helper
func fillRoundedCard(dc *gg.Context, x, y, w, h, r float64, alpha int) {
	dc.SetRGBA255(255, 255, 255, alpha)
	RoundRect(dc, x, y, w, h, r)
}

// drawText helpers with font fallback ignore error
func setFont(dc *gg.Context, size float64) { _ = LoadDefaultFont(dc, size) }

func drawString(dc *gg.Context, s string, x, y float64) {
	dc.DrawString(s, x, y)
}

func drawStringAnchored(dc *gg.Context, s string, x, y, ax, ay float64) {
	dc.DrawStringAnchored(s, x, y, ax, ay)
}

// measure helper
func measure(dc *gg.Context, s string) (float64, float64) { return dc.MeasureString(s) }

// color helper for help rarity? reuse

// imageToRGBA converts any image to RGBA for comparison.
func imageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, img, b.Min, draw.Src)
	return dst
}

// solidColor returns 1x1 color image.
func solidColor(r, g, b, a int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

var _ = color.RGBA{}
var _ = bytes.Buffer{}
