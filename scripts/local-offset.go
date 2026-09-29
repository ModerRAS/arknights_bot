// Per-block local registration analysis (analysis helper, NOT scoring code).
// Usage: go run local-offset.go <scene> <blockW> <blockH> [maxOff]
// For each block with hot pixels: find best (dx,dy) in [-maxOff,maxOff] that
// minimizes summed peak channel delta; prints offset + before/after delta sum.
package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
)

func peakDelta(oldImg, newImg image.Image, x, y int) uint32 {
	or, og, ob, _ := oldImg.At(x, y).RGBA()
	nr, ng, nb, _ := newImg.At(x, y).RGBA()
	d := max8(abs8(or>>8, nr>>8), abs8(og>>8, ng>>8), abs8(ob>>8, nb>>8))
	return d
}

func blockDelta(oldImg, newImg image.Image, x0, y0, x1, y1 int, dx, dy int) uint64 {
	var s uint64
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			or, og, ob, _ := oldImg.At(x, y).RGBA()
			nr, ng, nb, _ := newImg.At(x+dx, y+dy).RGBA()
			s += uint64(max8(abs8(or>>8, nr>>8), abs8(og>>8, ng>>8), abs8(ob>>8, nb>>8)))
		}
	}
	return s
}

func main() {
	scene := os.Args[1]
	bw, _ := strconv.Atoi(os.Args[2])
	bh, _ := strconv.Atoi(os.Args[3])
	maxOff := 8
	if len(os.Args) > 4 {
		maxOff, _ = strconv.Atoi(os.Args[4])
	}
	base := "src/utils/media/testdata/visual/final"
	oldData, _ := os.ReadFile(base + "/old/" + scene + ".jpg")
	newData, _ := os.ReadFile(base + "/new/" + scene + ".png")
	oldImg, _, _ := image.Decode(bytes.NewReader(oldData))
	newImg, _, _ := image.Decode(bytes.NewReader(newData))
	b := oldImg.Bounds()
	W, H := b.Dx(), b.Dy()
	cols, rows := (W+bw-1)/bw, (H+bh-1)/bh
	fmt.Printf("scene=%s %dx%d blocks=%dx%d maxOff=%d\n", scene, W, H, cols, rows, maxOff)
	// global baseline: delta without any offset
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x0, y0 := c*bw, r*bh
			x1, y1 := minInt(W, x0+bw), minInt(H, y0+bh)
			baseD := blockDelta(oldImg, newImg, x0, y0, x1, y1, 0, 0)
			// only report blocks with meaningful heat (>=32 threshold fraction)
			hot := 0
			total := 0
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					total++
					if peakDelta(oldImg, newImg, x, y) >= 32 {
						hot++
					}
				}
			}
			if hot*10 < total { // <10% hot -> skip
				continue
			}
			bestD, bestDX, bestDY := baseD, 0, 0
			for dy := -maxOff; dy <= maxOff; dy++ {
				for dx := -maxOff; dx <= maxOff; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					d := blockDelta(oldImg, newImg, x0, y0, x1, y1, dx, dy)
					if d < bestD {
						bestD, bestDX, bestDY = d, dx, dy
					}
				}
			}
			impr := float64(baseD-bestD) / float64(baseD) * 100
			hotPct := hot * 100 / total
			if bestDX != 0 || bestDY != 0 {
				fmt.Printf("block(%d,%d) xy=(%d,%d) %dx%d hot=%d%% baseDelta=%d bestDelta=%d (dx=%d,dy=%d) impr=%.1f%%\n",
					c, r, x0, y0, x1-x0, y1-y0, hotPct, baseD, bestD, bestDX, bestDY, impr)
			}
		}
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
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}