// Vertical ink profile per x-range: old vs new. Usage:
// go run ink-profile.go <scene> <x0,x1,y0,y1:name> [more ranges]
// Prints per-y ink counts (pixels differing from canvas bg) for old and new.
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

func main() {
	scene := os.Args[1]
	base := "src/utils/media/testdata/visual/final"
	oldData, _ := os.ReadFile(base + "/old/" + scene + ".jpg")
	newData, _ := os.ReadFile(base + "/new/" + scene + ".png")
	oldImg, _, _ := image.Decode(bytes.NewReader(oldData))
	newImg, _, _ := image.Decode(bytes.NewReader(newData))
	bg := oldImg.At(0, 0) // canvas bg (46,48,49)
	isInk := func(img image.Image, x, y int) bool {
		r, g, bl, _ := img.At(x, y).RGBA()
		br, bg2, bb, _ := bg.RGBA()
		return abs8(r>>8, br>>8) > 20 || abs8(g>>8, bg2>>8) > 20 || abs8(bl>>8, bb>>8) > 20
	}
	for _, arg := range os.Args[2:] {
		parts := strings.Split(arg, ",")
		name := parts[0]
		x0, _ := strconv.Atoi(parts[1])
		x1, _ := strconv.Atoi(parts[2])
		y0, _ := strconv.Atoi(parts[3])
		y1, _ := strconv.Atoi(parts[4])
		fmt.Printf("=== %s x=%d..%d y=%d..%d (y: oldCount/newCount) ===\n", name, x0, x1, y0, y1)
		for y := y0; y < y1; y++ {
			oc, nc := 0, 0
			for x := x0; x < x1; x++ {
				if isInk(oldImg, x, y) {
					oc++
				}
				if isInk(newImg, x, y) {
					nc++
				}
			}
			if oc > 0 || nc > 0 {
				mark := " "
				if oc > 0 && nc == 0 {
					mark = "<"
				} else if nc > 0 && oc == 0 {
					mark = ">"
				}
				fmt.Printf("  y=%3d old=%3d new=%3d %s\n", y, oc, nc, mark)
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