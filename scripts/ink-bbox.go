// ink-bbox.go: ink bounding box of an image region (alpha>20 & non-near-white)
// usage: go run ./scripts/ink-bbox.go <png> <x0> <y0> <x1> <y1> [showGrid]
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("usage: ink-bbox.go <png> x0 y0 x1 y1 [showGrid]")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("ERR", err)
		os.Exit(1)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		fmt.Println("ERR", err)
		os.Exit(1)
	}
	at := func(x, y int) (int, int) {
		r, g, b, _ := img.At(x, y).RGBA()
		br, bg2, bb, _ := img.At(bounds.Min.X, bounds.Min.Y).RGBA()
		absi := func(v int) int {
			if v < 0 {
				return -v
			}
			return v
		}
		if absi(int(r>>8)-int(br>>8)) > 20 || absi(int(g>>8)-int(bg2>>8)) > 20 || absi(int(b>>8)-int(bb>>8)) > 20 {
			return 1, int((r + g + b) / (3 * 257))
		}
		return 0, 0
	}
	x0, _ := strconv.Atoi(os.Args[2])
	y0, _ := strconv.Atoi(os.Args[3])
	x1, _ := strconv.Atoi(os.Args[4])
	y1, _ := strconv.Atoi(os.Args[5])
	show := len(os.Args) > 6 && os.Args[6] == "1"
	minX, minY, maxX, maxY := x1, y1, -1, -1
	bounds := img.Bounds()
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			v, _ := at(x, y)
			if v == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < 0 {
		fmt.Println("no ink in region")
		return
	}
	fmt.Printf("ink bbox: x=%d..%d (w=%d) y=%d..%d (h=%d)\n", minX, maxX, maxX-minX+1, minY, maxY, maxY-minY+1)
	if show {
		const RAMP = " .:-=+*#%@"
		for y := y0; y < y1; y++ {
			row := make([]byte, 0, x1-x0)
			for x := x0; x < x1; x++ {
				v, a := at(x, y)
				if v == 0 {
					row = append(row, '.')
				} else {
					row = append(row, RAMP[(a*len(RAMP))/256])
				}
			}
			fmt.Printf("%4d |%s|\n", y, string(row))
		}
	}
}