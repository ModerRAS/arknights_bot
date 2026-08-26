package skia

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ponytail: freetype 粗测 (opentype is superset at x/image v0.11.0), P2 换 SkShaper
var (
	textFontOnce sync.Once
	textFont     *opentype.Font
	textFontErr  error
	textFaces    = map[float64]font.Face{}
	textFacesMu  sync.Mutex
)

func loadTextFont() (*opentype.Font, error) {
	textFontOnce.Do(func() {
		cands := []string{
			FontPath,
			filepath.Join("C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go", FontPath),
			filepath.Join("../../", FontPath),
			filepath.Join("../", FontPath),
		}
		if r := os.Getenv("SKIA_REPO_ROOT"); r != "" {
			cands = append([]string{filepath.Join(r, FontPath)}, cands...)
		}
		if wd, err := os.Getwd(); err == nil {
			for _, rel := range []string{FontPath, filepath.Join("..", FontPath), filepath.Join("../..", FontPath)} {
				cands = append(cands, filepath.Join(wd, rel))
			}
		}
		var data []byte
		var err error
		for _, p := range cands {
			data, err = os.ReadFile(p)
			if err == nil {
				tf, perr := opentype.Parse(data)
				if perr == nil {
					textFont = tf
					textFontErr = nil
					return
				}
				err = perr
			}
		}
		textFontErr = err
	})
	return textFont, textFontErr
}

func getFace(size float64) (font.Face, error) {
	f, err := loadTextFont()
	if err != nil {
		return nil, err
	}
	textFacesMu.Lock()
	defer textFacesMu.Unlock()
	if face, ok := textFaces[size]; ok {
		return face, nil
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil, err
	}
	textFaces[size] = face
	return face, nil
}

func MeasureText(text string, fontSize float64) float64 {
	if text == "" {
		return 0
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	face, err := getFace(fontSize)
	if err != nil || face == nil {
		return float64(len([]rune(text))) * fontSize * 0.6
	}
	d := &font.Drawer{Face: face}
	adv := d.MeasureString(text)
	return float64(adv) / 64
}

func (c *Canvas) DrawText(text string, x, y float32, fontSize float64, col color.RGBA, scale float64) {
	if text == "" || c == nil || c.img == nil {
		return
	}
	pxSize := fontSize * scale
	if pxSize < 1 {
		pxSize = 12 * scale
	}
	face, err := getFace(pxSize)
	if err != nil {
		return
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	dotX := fixed.I(int(x*float32(scale) + 0.5))
	dotY := fixed.I(int(y*float32(scale) + float32(ascent) + 0.5))
	d := &font.Drawer{
		Dst:  c.img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: dotX * 64, Y: dotY * 64},
	}
	d.DrawString(text)
	_ = metrics
}

var _ = fixed.Int26_6(0)
