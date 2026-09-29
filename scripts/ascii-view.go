// ASCII luminance view of old/new renders using Go's decoder (harness ground truth).
// Usage: go run ascii-view.go <scene> [cols rows]
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

func lum(img image.Image, x, y int) uint32 {
	r, g, b, _ := img.At(x, y).RGBA()
	// ITU-R BT.601 luma: sum scaled to 0..255 (16-bit channels)
	return (299*r + 587*g + 114*b) >> 18
}

func ascii(img image.Image, cols, rows int, name string) {
	b := img.Bounds()
	W, H := b.Dx(), b.Dy()
	fmt.Printf("--- %s (%dx%d) ---\n", name, W, H)
	const ramp = " .:-=+*#%@"
	for r := 0; r < rows; r++ {
		line := make([]byte, cols)
		for c := 0; c < cols; c++ {
			// average luma over block
			x0, y0 := c*W/cols, r*H/rows
			x1, y1 := (c+1)*W/cols, (r+1)*H/rows
			var s uint64
			n := 0
			for y := y0; y < y1; y += 2 {
				for x := x0; x < x1; x += 2 {
					s += uint64(lum(img, x, y))
					n++
				}
			}
			v := int(s / uint64(n))
			idx := v * (len(ramp) - 1) / 255
			line[c] = ramp[idx]
		}
		fmt.Println(string(line))
	}
}

func main() {
	scene := os.Args[1]
	cols, rows := 60, 20
	if len(os.Args) > 3 {
		cols, _ = strconv.Atoi(os.Args[2])
		rows, _ = strconv.Atoi(os.Args[3])
	}
	base := "src/utils/media/testdata/visual/final"
	oldData, err := os.ReadFile(base + "/old/" + scene + ".jpg")
	if err != nil {
		fmt.Println("read old:", err)
		os.Exit(1)
	}
	newData, err := os.ReadFile(base + "/new/" + scene + ".png")
	if err != nil {
		fmt.Println("read new:", err)
		os.Exit(1)
	}
	oldImg, _, err := image.Decode(bytes.NewReader(oldData))
	if err != nil {
		fmt.Println("decode old:", err)
		os.Exit(1)
	}
	newImg, _, err := image.Decode(bytes.NewReader(newData))
	if err != nil {
		fmt.Println("decode new:", err)
		os.Exit(1)
	}
	ascii(oldImg, cols, rows, "OLD "+scene)
	ascii(newImg, cols, rows, "NEW "+scene)
}