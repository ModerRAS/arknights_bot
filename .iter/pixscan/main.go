// Scan v2: find LIGHT cell extents (value cells / panels) to settle attr row height.
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func load(p string) image.Image {
	f, err := os.Open(p)
	if err != nil {
		fmt.Println("ERR", p, err)
		os.Exit(1)
	}
	defer f.Close()
	im, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("ERR decode", p, err)
		os.Exit(1)
	}
	return im
}

func lightRuns(im image.Image, x, y0, y1, thr int) [][2]int {
	var runs [][2]int
	start := -1
	for y := y0; y <= y1; y++ {
		r, g, b, _ := im.At(x, y).RGBA()
		lum := (r + g + b) / 3 >> 8
		light := int(lum) > thr
		if light && start < 0 {
			start = y
		}
		if !light && start >= 0 {
			runs = append(runs, [2]int{start, y - 1})
			start = -1
		}
	}
	if start >= 0 {
		runs = append(runs, [2]int{start, y1})
	}
	return runs
}

func lightRunsH(im image.Image, y, x0, x1, thr int) [][2]int {
	var runs [][2]int
	start := -1
	for x := x0; x <= x1; x++ {
		r, g, b, _ := im.At(x, y).RGBA()
		lum := (r + g + b) / 3 >> 8
		light := int(lum) > thr
		if light && start < 0 {
			start = x
		}
		if !light && start >= 0 {
			runs = append(runs, [2]int{start, x - 1})
			start = -1
		}
	}
	if start >= 0 {
		runs = append(runs, [2]int{start, x1})
	}
	return runs
}

func fmtRuns(runs [][2]int, scale float64) string {
	s := ""
	for _, r := range runs {
		s += fmt.Sprintf("[%d-%d => %.1f-%.1f] ", r[0], r[1], float64(r[0])/scale, float64(r[1]+1)/scale)
	}
	return s
}

func main() {
	base := load(`C:\WorkSpace\Golang\arknights_bot-satori-w3-help-boxes\src\utils\media\testdata\visual\baseline\images\operator.jpg`)
	newp := load(`C:\WorkSpace\Golang\arknights_bot-satori-w3-help-boxes\src\utils\media\testdata\visual\final\new\operator.png`)
	const sc = 1.5
	fmt.Println("== ATTR value cell col1 (x=210img) light runs y 0..400img thr=150 ==")
	fmt.Println("base:", fmtRuns(lightRuns(base, 210, 0, 400, 150), sc))
	fmt.Println("new :", fmtRuns(lightRuns(newp, 210, 0, 400, 150), sc))
	fmt.Println("== ATTR value cell col2 (x=480img) thr=150 ==")
	fmt.Println("base:", fmtRuns(lightRuns(base, 480, 0, 400, 150), sc))
	fmt.Println("new :", fmtRuns(lightRuns(newp, 480, 0, 400, 150), sc))
	fmt.Println("== ATTR value cells horizontal (y=60img) x 100..800img thr=150 ==")
	fmt.Println("base:", fmtRuns(lightRunsH(base, 60, 100, 800, 150), sc))
	fmt.Println("new :", fmtRuns(lightRunsH(newp, 60, 100, 800, 150), sc))
	fmt.Println("== POTENTIAL panel (x=30img) light runs y 250..480img thr=100 ==")
	fmt.Println("base:", fmtRuns(lightRuns(base, 30, 250, 480, 100), sc))
	fmt.Println("new :", fmtRuns(lightRuns(newp, 30, 250, 480, 100), sc))
	fmt.Println("== PB panel (x=30img) y 690..980img thr=100 ==")
	fmt.Println("base:", fmtRuns(lightRuns(base, 30, 690, 980, 100), sc))
	fmt.Println("new :", fmtRuns(lightRuns(newp, 30, 690, 980, 100), sc))
	fmt.Println("== NAMES panel (x=30img) y 960..1160img thr=100 ==")
	fmt.Println("base:", fmtRuns(lightRuns(base, 30, 960, 1160, 100), sc))
	fmt.Println("new :", fmtRuns(lightRuns(newp, 30, 960, 1160, 100), sc))
	fmt.Println("== PB panel horizontal (y=740img) x 0..500img thr=100 ==")
	fmt.Println("base:", fmtRuns(lightRunsH(base, 740, 0, 500, 100), sc))
	fmt.Println("new :", fmtRuns(lightRunsH(newp, 740, 0, 500, 100), sc))
	fmt.Println("== NAMES panel horizontal (y=1000img) x 0..500img thr=100 ==")
	fmt.Println("base:", fmtRuns(lightRunsH(base, 1000, 0, 500, 100), sc))
	fmt.Println("new :", fmtRuns(lightRunsH(newp, 1000, 0, 500, 100), sc))
	fmt.Println("== SKILL badges row (y=190img) x 1050..1650img: mid-gray runs thr 100..180 ==")
	midRuns := func(im image.Image, y, x0, x1 int) [][2]int {
		var runs [][2]int
		start := -1
		for x := x0; x <= x1; x++ {
			r, g, b, _ := im.At(x, y).RGBA()
			lum := (r + g + b) / 3 >> 8
			mid := int(lum) > 100 && int(lum) < 200
			if mid && start < 0 {
				start = x
			}
			if !mid && start >= 0 {
				runs = append(runs, [2]int{start, x - 1})
				start = -1
			}
		}
		if start >= 0 {
			runs = append(runs, [2]int{start, x1})
		}
		return runs
	}
	fmt.Println("base:", fmtRuns(midRuns(base, 190, 1050, 1650), sc))
	fmt.Println("new :", fmtRuns(midRuns(newp, 190, 1050, 1650), sc))
	fmt.Println("== SKILL badge vertical (x=1330img) y 170..210img mid ==")
	midRunsV := func(im image.Image, x, y0, y1 int) [][2]int {
		var runs [][2]int
		start := -1
		for y := y0; y <= y1; y++ {
			r, g, b, _ := im.At(x, y).RGBA()
			lum := (r + g + b) / 3 >> 8
			mid := int(lum) > 100 && int(lum) < 200
			if mid && start < 0 {
				start = y
			}
			if !mid && start >= 0 {
				runs = append(runs, [2]int{start, y - 1})
				start = -1
			}
		}
		if start >= 0 {
			runs = append(runs, [2]int{start, y1})
		}
		return runs
	}
	fmt.Println("base:", fmtRuns(midRunsV(base, 1330, 170, 210), sc))
	fmt.Println("new :", fmtRuns(midRunsV(newp, 1330, 170, 210), sc))
}
