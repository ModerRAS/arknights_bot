// png-insets.go: print ink bbox top/bottom insets for PNG assets (1 line each)
// usage: go run ./scripts/png-insets.go <file>...
package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
)

func main() {
	for _, path := range os.Args[1:] {
		f, err := os.Open(path)
		if err != nil {
			fmt.Println(path, "ERR", err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			fmt.Println(path, "ERR", err)
			continue
		}
		b := img.Bounds()
		minX, minY, maxX, maxY := b.Max.X, b.Max.Y, -1, -1
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a > 20*257 {
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
		}
		fmt.Printf("%s %dx%d art bbox=(%d,%d)-(%d,%d) topInset=%d bottomInset=%d leftInset=%d rightInset=%d\n",
			path, b.Dx(), b.Dy(), minX, minY, maxX, maxY, minY-b.Min.Y, b.Max.Y-1-maxY, minX-b.Min.X, b.Max.X-1-maxX)
	}
}