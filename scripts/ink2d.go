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

func main() {
	scene := os.Args[1]
	imgIdx := 0
	if len(os.Args) > 6 {
		imgIdx, _ = strconv.Atoi(os.Args[6])
	}
	x0, _ := strconv.Atoi(os.Args[2])
	y0, _ := strconv.Atoi(os.Args[3])
	x1, _ := strconv.Atoi(os.Args[4])
	y1, _ := strconv.Atoi(os.Args[5])
	base := "src/utils/media/testdata/visual/final"
	paths := []string{base + "/old/" + scene + ".jpg", base + "/new/" + scene + ".png"}
	data, _ := os.ReadFile(paths[imgIdx])
	img, _, _ := image.Decode(bytes.NewReader(data))
	br, bg, bb, _ := img.At(0, 0).RGBA()
	ink := func(x, y int) bool {
		r, g, b, _ := img.At(x, y).RGBA()
		return abs8(r>>8, br>>8) > 20 || abs8(g>>8, bg>>8) > 20 || abs8(b>>8, bb>>8) > 20
	}
	fmt.Printf("=== %s %s %d..%d x %d..%d ===\n", scene, map[int]string{0: "OLD", 1: "NEW"}[imgIdx], x0, x1, y0, y1)
	for y := y0; y < y1; y++ {
		line := make([]byte, x1-x0)
		for x := x0; x < x1; x++ {
			if ink(x, y) {
				r, g, b, _ := img.At(x, y).RGBA()
				lum := (299*r + 587*g + 114*b) >> 18
				if lum > 200 {
					line[x-x0] = '#'
				} else if lum > 120 {
					line[x-x0] = '+'
				} else {
					line[x-x0] = '.'
				}
			} else {
				line[x-x0] = ' '
			}
		}
		fmt.Printf("%3d |%s|\n", y, string(line))
	}
}
func abs8(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
