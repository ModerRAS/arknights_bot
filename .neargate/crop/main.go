// Local helper: crop region from images and paste side-by-side into out.png
// usage: go run . out.png x y w h img1 [img2 ...]  (coords in pixels)
package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
)

func decode(data []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return img
}

func main() {
	out := os.Args[1]
	var x, y, w, h int
	fmtSscan(os.Args[2:6], &x, &y, &w, &h)
	images := make([]image.Image, 0, len(os.Args)-6)
	for _, p := range os.Args[6:] {
		data, err := os.ReadFile(p)
		if err != nil {
			panic(err)
		}
		images = append(images, decode(data))
	}
	canvas := image.NewRGBA(image.Rect(0, 0, w*len(images)+4*(len(images)-1), h))
	for i, img := range images {
		off := i * (w + 4)
		for yy := 0; yy < h; yy++ {
			for xx := 0; xx < w; xx++ {
				canvas.Set(off+xx, yy, img.At(x+xx, y+yy))
			}
		}
	}
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, canvas); err != nil {
		panic(err)
	}
}

func fmtSscan(s []string, v ...*int) {
	for i := range s {
		n := 0
		for _, c := range s[i] {
			n = n*10 + int(c-'0')
		}
		*v[i] = n
	}
}
