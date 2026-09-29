// Local iteration helper (not committed). Mirrors the visual-final gate formula:
// abs8 per-channel deltas over RGBA, similarity = 1 - total/(w*h*4*255).
// Aligned score uses a brute-force offset search in [-16,16] (the official gate
// validates a suggested offset instead; brute force is an upper bound — final
// confirmation always re-runs src/cmd/visual-final).
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
)

const (
	componentDeltaThreshold = 32
	componentMergePadding   = 48
)

type rect struct{ X, Y, W, H int }

type component struct {
	BBox      rect
	Pixels    int
	MeanDelta float64
}

func main() {
	baselineDir, candidateDir, outDir := os.Args[1], os.Args[2], os.Args[3]
	ids := os.Args[4:]
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	for _, id := range ids {
		oldData, err := os.ReadFile(filepath.Join(baselineDir, id+".jpg"))
		if err != nil {
			fmt.Printf("%s: %v\n", id, err)
			continue
		}
		newData, err := os.ReadFile(filepath.Join(candidateDir, id+".png"))
		if err != nil {
			fmt.Printf("%s: %v\n", id, err)
			continue
		}
		oldImg := decodeRGBA(oldData)
		newImg := decodeRGBA(newData)
		if oldImg.Bounds().Size() != newImg.Bounds().Size() {
			fmt.Printf("%s: dimension mismatch\n", id)
			continue
		}
		raw := imageSimilarity(oldImg, newImg)
		bestDX, bestDY, bestScore := 0, 0, raw
		for dy := -16; dy <= 16; dy++ {
			for dx := -16; dx <= 16; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				if s := imageSimilarity(oldImg, translateImage(newImg, dx, dy)); s > bestScore {
					bestScore, bestDX, bestDY = s, dx, dy
				}
			}
		}
		aligned := translateImage(newImg, bestDX, bestDY)
		writeHeatmap(filepath.Join(outDir, id+".heatmap.png"), oldImg, aligned)
		comps := meaningfulComponents(oldImg, aligned)
		fmt.Printf("%s: raw=%.4f%% aligned=%.4f%% offset=(%d,%d) components=%d\n", id, raw*100, bestScore*100, bestDX, bestDY, len(comps))
		for _, c := range comps {
			fmt.Printf("    bbox=(%d,%d,%dx%d) px=%d meanDelta=%.1f\n", c.BBox.X, c.BBox.Y, c.BBox.W, c.BBox.H, c.Pixels, c.MeanDelta)
		}
	}
}

func decodeRGBA(data []byte) *image.RGBA {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	rgba := image.NewRGBA(decoded.Bounds())
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			rgba.Set(x, y, decoded.At(x, y))
		}
	}
	return rgba
}

func abs8(a, b uint32) uint8 {
	if a > b {
		return uint8((a - b) >> 8)
	}
	return uint8((b - a) >> 8)
}

func pixelDelta(a, b color.Color) uint64 {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return uint64(abs8(ar, br)) + uint64(abs8(ag, bg)) + uint64(abs8(ab, bb)) + uint64(abs8(aa, ba))
}

func imageSimilarity(oldImg, newImg image.Image) float64 {
	bounds := oldImg.Bounds()
	var total uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			total += pixelDelta(oldImg.At(x, y), newImg.At(x, y))
		}
	}
	return 1 - float64(total)/float64(bounds.Dx()*bounds.Dy()*4*255)
}

// translateImage shifts new by (dx,dy); vacated area fills with the old image's
// corner background approximation (solid color from pixel 0,0 of old).
func translateImage(img image.Image, dx, dy int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	bg := img.At(bounds.Min.X, bounds.Min.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			sx, sy := x-dx, y-dy
			if sx < bounds.Min.X || sx >= bounds.Max.X || sy < bounds.Min.Y || sy >= bounds.Max.Y {
				out.Set(x, y, bg)
				continue
			}
			out.Set(x, y, img.At(sx, sy))
		}
	}
	return out
}

func writeHeatmap(path string, oldImg, aligned image.Image) error {
	bounds := oldImg.Bounds()
	heat := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			or, og, ob, oa := oldImg.At(x, y).RGBA()
			nr, ng, nb, na := aligned.At(x, y).RGBA()
			dr, dg, db, da := abs8(or, nr), abs8(og, ng), abs8(ob, nb), abs8(oa, na)
			peak := max8(max8(dr, dg), max8(db, da))
			heat.SetRGBA(x, y, color.RGBA{peak, 0, 0, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, heat)
}

func max8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func meaningfulComponents(oldImg, aligned image.Image) []component {
	bounds := oldImg.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	peaks := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			or, og, ob, oa := oldImg.At(x, y).RGBA()
			nr, ng, nb, na := aligned.At(x, y).RGBA()
			dr, dg, db, da := abs8(or, nr), abs8(og, ng), abs8(ob, nb), abs8(oa, na)
			peaks[y*w+x] = max8(max8(dr, dg), max8(db, da))
		}
	}
	var comps []component
	seen := make([]bool, len(peaks))
	for start, peak := range peaks {
		if seen[start] || peak < componentDeltaThreshold {
			continue
		}
		queue := []int{start}
		seen[start] = true
		minX, minY, maxX, maxY, count, total := start%w, start/w, start%w, start/w, 0, uint64(0)
		for len(queue) > 0 {
			index := queue[0]
			queue = queue[1:]
			x, y := index%w, index/w
			count++
			total += uint64(peaks[index])
			minX, minY = minInt(minX, x), minInt(minY, y)
			maxX, maxY = maxInt(maxX, x), maxInt(maxY, y)
			for _, step := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+step[0], y+step[1]
				if nx < 0 || nx >= w || ny < 0 || ny >= h {
					continue
				}
				next := ny*w + nx
				if !seen[next] && peaks[next] >= componentDeltaThreshold {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		comps = append(comps, component{
			BBox:      rect{X: minX, Y: minY, W: maxX - minX + 1, H: maxY - minY + 1},
			Pixels:    count,
			MeanDelta: float64(total) / float64(count),
		})
	}
	return mergeComponents(comps)
}

func mergeComponents(comps []component) []component {
	sort.Slice(comps, func(i, j int) bool { return comps[i].Pixels > comps[j].Pixels })
	var merged []component
	for _, c := range comps {
		hit := false
		for i := range merged {
			if near(c.BBox, merged[i].BBox) {
				merged[i].BBox = union(c.BBox, merged[i].BBox)
				merged[i].Pixels += c.Pixels
				hit = true
				break
			}
		}
		if !hit {
			merged = append(merged, c)
		}
	}
	return merged
}

func near(a, b rect) bool {
	pad := componentMergePadding
	return a.X-pad <= b.X+b.W && b.X-pad <= a.X+a.W && a.Y-pad <= b.Y+b.H && b.Y-pad <= a.Y+a.H
}

func union(a, b rect) rect {
	minX, minY := minInt(a.X, b.X), minInt(a.Y, b.Y)
	maxX, maxY := maxInt(a.X+a.W, b.X+b.W), maxInt(a.Y+a.H, b.Y+b.H)
	return rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
