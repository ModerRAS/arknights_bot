// Per-region local registration (analysis helper, NOT scoring code).
// Usage: go run local-region.go <scene> <maxOff> <x,y,w,h> [x,y,w,h ...]
// Regions in device px (from renderer geometry). Prints best (dx,dy) per region
// with delta improvement %. Positive dx/dy = new content must move right/down.
package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"strings"
)

type region struct{ x0, y0, x1, y1 int; name string }

func main() {
	scene := os.Args[1]
	maxOff, _ := strconv.Atoi(os.Args[2])
	base := "src/utils/media/testdata/visual/final"
	oldData, _ := os.ReadFile(base + "/old/" + scene + ".jpg")
	newData, _ := os.ReadFile(base + "/new/" + scene + ".png")
	oldImg, _, _ := image.Decode(bytes.NewReader(oldData))
	newImg, _, _ := image.Decode(bytes.NewReader(newData))
	var regions []region
	for _, arg := range os.Args[3:] {
		parts := strings.Split(arg, ",")
		name := parts[0]
		x, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])
		w, _ := strconv.Atoi(parts[3])
		h, _ := strconv.Atoi(parts[4])
		regions = append(regions, region{x, y, x + w, y + h, name})
	}
	regionDelta := func(r region, dx, dy int) uint64 {
		var s uint64
		for y := r.y0; y < r.y1; y++ {
			for x := r.x0; x < r.x1; x++ {
				or, og, ob, _ := oldImg.At(x, y).RGBA()
				nr, ng, nb, _ := newImg.At(x+dx, y+dy).RGBA()
				s += uint64(max8(abs8(or>>8, nr>>8), abs8(og>>8, ng>>8), abs8(ob>>8, nb>>8)))
			}
		}
		return s
	}
	for _, r := range regions {
		baseD := regionDelta(r, 0, 0)
		bestD, bestDX, bestDY := baseD, 0, 0
		for dy := -maxOff; dy <= maxOff; dy++ {
			for dx := -maxOff; dx <= maxOff; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				d := regionDelta(r, dx, dy)
				if d < bestD {
					bestD, bestDX, bestDY = d, dx, dy
				}
			}
		}
		impr := float64(baseD-bestD) / float64(baseD) * 100
		fmt.Printf("%-12s (%d,%d %dx%d) base=%d best=%d (dx=%d,dy=%d) impr=%.1f%%\n",
			r.name, r.x0, r.y0, r.x1-r.x0, r.y1-r.y0, baseD, bestD, bestDX, bestDY, impr)
	}
}

func abs8(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
func max8(a, b, c uint32) uint32 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}