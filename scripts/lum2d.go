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
	imgIdx, _ := strconv.Atoi(os.Args[6])
	x0, _ := strconv.Atoi(os.Args[2])
	y0, _ := strconv.Atoi(os.Args[3])
	x1, _ := strconv.Atoi(os.Args[4])
	y1, _ := strconv.Atoi(os.Args[5])
	base := "src/utils/media/testdata/visual/final"
	paths := []string{base + "/old/" + scene + ".jpg", base + "/new/" + scene + ".png"}
	data, _ := os.ReadFile(paths[imgIdx])
	img, _, _ := image.Decode(bytes.NewReader(data))
	fmt.Printf("=== %s %s %d..%d x %d..%d (lum ramp) ===\n", scene, map[int]string{0: "OLD", 1: "NEW"}[imgIdx], x0, x1, y0, y1)
	const ramp = " .:-=+*#%@"
	for y := y0; y < y1; y++ {
		line := make([]byte, x1-x0)
		for x := x0; x < x1; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			lum := (299*r + 587*g + 114*b) >> 18
			line[x-x0] = ramp[int(lum)*(len(ramp)-1)/255]
		}
		fmt.Printf("%3d |%s|\n", y, string(line))
	}
}
