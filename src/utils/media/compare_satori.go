// ponytail: reuse from satori/src/cmd/visual-final/main.go
package media

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
)

type rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type globalOffset struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

type diffComponent struct {
	BBox      rect    `json:"bbox"`
	Pixels    int     `json:"pixels"`
	MeanDelta float64 `json:"meanDelta"`
}

type localShift struct {
	BBox        rect         `json:"bbox"`
	Offset      globalOffset `json:"offset"`
	Improvement float64      `json:"improvement"`
}

type imageComparison struct {
	RawScore         float64
	RawBBox          *rect
	AlignedScore     float64
	AlignedBBox      *rect
	Offset           globalOffset
	LocalPass        bool
	Meaningful       []diffComponent
	ObjectShifts     []localShift
	ZeroByteIdentity bool
}

const registrationMaxOffset = 16
const localRegistrationMaxOffset = 8
const componentDeltaThreshold = 32
const componentMergePadding = 48

type anchorSignature struct{ Width, Height int }

var pageAnchorSignatures = map[string][]anchorSignature{
	"calendar": {{Width: 2546, Height: 38}, {Width: 2546, Height: 25}},
	"lottery":  {{Width: 139, Height: 138}, {Width: 42, Height: 45}},
}

func compareImages(oldData, newData []byte, rawDiffPath, rawHeatmapPath, alignedDiffPath, alignedHeatmapPath string) (imageComparison, error) {
	return comparePageImages("", oldData, newData, rawDiffPath, rawHeatmapPath, alignedDiffPath, alignedHeatmapPath)
}

func comparePageImages(pageID string, oldData, newData []byte, rawDiffPath, rawHeatmapPath, alignedDiffPath, alignedHeatmapPath string) (imageComparison, error) {
	oldImg, err := decodeRGBA(oldData, "old")
	if err != nil {
		return imageComparison{}, err
	}
	newImg, err := decodeRGBA(newData, "new")
	if err != nil {
		return imageComparison{}, err
	}
	if oldImg.Bounds().Size() != newImg.Bounds().Size() {
		return imageComparison{}, fmt.Errorf("dimension mismatch old=%dx%d new=%dx%d", oldImg.Bounds().Dx(), oldImg.Bounds().Dy(), newImg.Bounds().Dx(), newImg.Bounds().Dy())
	}
	rawScore, rawBBox, _, err := diffImagePair(oldImg, newImg, rawDiffPath, rawHeatmapPath)
	if err != nil {
		return imageComparison{}, err
	}
	background := canvasBackground(oldImg)
	offset := selectFinalOffset(oldImg, newImg, registerOffset(oldImg, newImg, background), background)
	aligned := translateImage(newImg, offset, background)
	alignedScore, alignedBBox, components, err := diffImagePair(oldImg, aligned, alignedDiffPath, alignedHeatmapPath)
	if err != nil {
		return imageComparison{}, err
	}
	localPass, meaningful, shifts := localGate(pageID, oldImg, aligned, components, background)
	return imageComparison{RawScore: rawScore, RawBBox: rawBBox, AlignedScore: alignedScore, AlignedBBox: alignedBBox, Offset: offset, LocalPass: localPass, Meaningful: meaningful, ObjectShifts: shifts, ZeroByteIdentity: offset == (globalOffset{}) && bytes.Equal(aligned.Pix, newImg.Pix)}, nil
}

func decodeRGBA(data []byte, label string) (*image.RGBA, error) {
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", label, err)
	}
	rgba := image.NewRGBA(decoded.Bounds())
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			rgba.Set(x, y, decoded.At(x, y))
		}
	}
	return rgba, nil
}

func selectFinalOffset(oldImg, newImg image.Image, suggested globalOffset, background color.Color) globalOffset {
	if suggested == (globalOffset{}) {
		return globalOffset{}
	}
	if imageSimilarity(oldImg, translatedImage(newImg, suggested, background)) > imageSimilarity(oldImg, newImg) && anchorImproves(oldImg, newImg, suggested, background) {
		return suggested
	}
	return globalOffset{}
}

func anchorImproves(oldImg, newImg image.Image, offset globalOffset, background color.Color) bool {
	for _, roi := range anchorROIs(oldImg.Bounds()) {
		if regionDelta(oldImg, newImg, roi, offset, background) < regionDelta(oldImg, newImg, roi, globalOffset{}, background) {
			return true
		}
	}
	return false
}

func anchorROIs(bounds image.Rectangle) []image.Rectangle {
	size := minInt(32, minInt(bounds.Dx()/2, bounds.Dy()/2))
	if size == 0 {
		return nil
	}
	return []image.Rectangle{
		image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+size, bounds.Min.Y+size),
		image.Rect(bounds.Max.X-size, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+size),
		image.Rect(bounds.Min.X, bounds.Max.Y-size, bounds.Min.X+size, bounds.Max.Y),
		image.Rect(bounds.Max.X-size, bounds.Max.Y-size, bounds.Max.X, bounds.Max.Y),
	}
}

func preferredOffset(candidate, current globalOffset) bool {
	candidateDistance := absInt(candidate.DX) + absInt(candidate.DY)
	currentDistance := absInt(current.DX) + absInt(current.DY)
	if candidateDistance != currentDistance {
		return candidateDistance < currentDistance
	}
	if candidate.DX != current.DX {
		return candidate.DX < current.DX
	}
	return candidate.DY < current.DY
}

func imageSimilarity(oldImg, newImg image.Image) float64 {
	bounds := oldImg.Bounds()
	var total uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			total += pixelDelta(oldImg.At(x, y), newImg.At(x, y))
		}
	}
	return similarity(total, bounds)
}

func similarity(total uint64, bounds image.Rectangle) float64 {
	return 1 - float64(total)/float64(bounds.Dx()*bounds.Dy()*4*255)
}

func registerOffset(oldImg, newImg image.Image, background color.Color) globalOffset {
	bounds := oldImg.Bounds()
	step := maxInt(1, maxInt(bounds.Dx(), bounds.Dy())/256)
	best, bestScore := globalOffset{}, ^uint64(0)
	for dy := -registrationMaxOffset; dy <= registrationMaxOffset; dy++ {
		for dx := -registrationMaxOffset; dx <= registrationMaxOffset; dx++ {
			var score uint64
			for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
				for x := bounds.Min.X; x < bounds.Max.X; x += step {
					score += pixelDelta(oldImg.At(x, y), translatedAt(newImg, x-dx, y-dy, background))
				}
			}
			candidate := globalOffset{DX: dx, DY: dy}
			if score < bestScore || (score == bestScore && preferredOffset(candidate, best)) {
				best, bestScore = candidate, score
			}
		}
	}
	return best
}

func translatedImage(src image.Image, offset globalOffset, background color.Color) image.Image {
	if offset == (globalOffset{}) {
		return src
	}
	return translateImage(src, offset, background)
}

func translateImage(src image.Image, offset globalOffset, background color.Color) *image.RGBA {
	if offset == (globalOffset{}) {
		if rgba, ok := src.(*image.RGBA); ok {
			return rgba
		}
	}
	bounds := src.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			out.Set(x, y, translatedAt(src, x-offset.DX, y-offset.DY, background))
		}
	}
	return out
}

func canvasBackground(img image.Image) color.Color {
	return img.At(img.Bounds().Min.X, img.Bounds().Min.Y)
}

func translatedAt(src image.Image, x, y int, background color.Color) color.Color {
	if !image.Pt(x, y).In(src.Bounds()) {
		return background
	}
	return src.At(x, y)
}

func diffImagePair(oldImg, newImg image.Image, diffPath, heatmapPath string) (float64, *rect, []diffComponent, error) {
	bounds := oldImg.Bounds()
	diff, heatmap := image.NewRGBA(bounds), image.NewRGBA(bounds)
	peaks := make([]uint8, bounds.Dx()*bounds.Dy())
	minX, minY, maxX, maxY := bounds.Dx(), bounds.Dy(), -1, -1
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			or, og, ob, oa := oldImg.At(x, y).RGBA()
			nr, ng, nb, na := newImg.At(x, y).RGBA()
			dr, dg, db, da := abs8(or, nr), abs8(og, ng), abs8(ob, nb), abs8(oa, na)
			diff.SetRGBA(x, y, color.RGBA{dr, dg, db, 255})
			peak := max8(max8(dr, dg), max8(db, da))
			heatmap.SetRGBA(x, y, color.RGBA{peak, 0, 0, 255})
			peaks[y*bounds.Dx()+x] = peak
			if peak > 0 {
				minX, minY, maxX, maxY = minInt(minX, x), minInt(minY, y), maxInt(maxX, x), maxInt(maxY, y)
			}
		}
	}
	if err := satoriWritePNG(diffPath, diff); err != nil {
		return 0, nil, nil, err
	}
	if err := satoriWritePNG(heatmapPath, heatmap); err != nil {
		return 0, nil, nil, err
	}
	score := imageSimilarity(oldImg, newImg)
	if maxX < 0 {
		return score, nil, nil, nil
	}
	return score, &rect{X: minX, Y: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}, diffComponents(peaks, bounds.Dx(), bounds.Dy()), nil
}

func localGate(pageID string, oldImg, aligned image.Image, components []diffComponent, background color.Color) (bool, []diffComponent, []localShift) {
	var meaningful []diffComponent
	var shifts []localShift
	for _, component := range mergeComponents(components) {
		if !anchorComponent(pageID, component) && microNoise(component) {
			continue
		}
		meaningful = append(meaningful, component)
		offset, improvement := localRegistration(oldImg, aligned, component.BBox, background)
		if offset != (globalOffset{}) && improvement >= 0.15 {
			shifts = append(shifts, localShift{BBox: component.BBox, Offset: offset, Improvement: improvement})
		}
	}
	return len(meaningful) == 0, meaningful, shifts
}

func anchorComponent(pageID string, component diffComponent) bool {
	for _, signature := range pageAnchorSignatures[pageID] {
		if component.BBox.Width >= signature.Width && component.BBox.Height >= signature.Height {
			return true
		}
	}
	return false
}

func microNoise(component diffComponent) bool {
	return component.Pixels <= 64 && component.BBox.Width <= 8 && component.BBox.Height <= 8 && component.MeanDelta < 64
}

func mergeComponents(components []diffComponent) []diffComponent {
	var merged []diffComponent
	for _, component := range components {
		for index := 0; index < len(merged); index++ {
			if !rectanglesNear(component.BBox, merged[index].BBox, componentMergePadding) {
				continue
			}
			merged[index].BBox = unionRect(component.BBox, merged[index].BBox)
			merged[index].Pixels += component.Pixels
			component = merged[index]
			merged = append(merged[:index], merged[index+1:]...)
			index = -1
		}
		merged = append(merged, component)
	}
	return merged
}

func rectanglesNear(a, b rect, padding int) bool {
	return a.X-padding <= b.X+b.Width && b.X-padding <= a.X+a.Width && a.Y-padding <= b.Y+b.Height && b.Y-padding <= a.Y+a.Height
}

func unionRect(a, b rect) rect {
	minX, minY := minInt(a.X, b.X), minInt(a.Y, b.Y)
	maxX, maxY := maxInt(a.X+a.Width, b.X+b.Width), maxInt(a.Y+a.Height, b.Y+b.Height)
	return rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}

func localRegistration(oldImg, aligned image.Image, box rect, background color.Color) (globalOffset, float64) {
	bounds := oldImg.Bounds()
	roi := image.Rect(maxInt(bounds.Min.X, box.X-4), maxInt(bounds.Min.Y, box.Y-4), minInt(bounds.Max.X, box.X+box.Width+4), minInt(bounds.Max.Y, box.Y+box.Height+4))
	base := regionDelta(oldImg, aligned, roi, globalOffset{}, background)
	if base == 0 {
		return globalOffset{}, 0
	}
	best, bestDelta := globalOffset{}, base
	for dy := -localRegistrationMaxOffset; dy <= localRegistrationMaxOffset; dy++ {
		for dx := -localRegistrationMaxOffset; dx <= localRegistrationMaxOffset; dx++ {
			candidate := globalOffset{DX: dx, DY: dy}
			delta := regionDelta(oldImg, aligned, roi, candidate, background)
			if delta < bestDelta || (delta == bestDelta && preferredOffset(candidate, best)) {
				best, bestDelta = candidate, delta
			}
		}
	}
	return best, float64(base-bestDelta) / float64(base)
}

func regionDelta(oldImg, aligned image.Image, roi image.Rectangle, offset globalOffset, background color.Color) uint64 {
	var total uint64
	for y := roi.Min.Y; y < roi.Max.Y; y++ {
		for x := roi.Min.X; x < roi.Max.X; x++ {
			total += pixelDelta(oldImg.At(x, y), translatedAt(aligned, x-offset.DX, y-offset.DY, background))
		}
	}
	return total
}

func diffComponents(peaks []uint8, width, height int) []diffComponent {
	seen := make([]bool, len(peaks))
	var components []diffComponent
	for start, peak := range peaks {
		if seen[start] || peak < componentDeltaThreshold {
			continue
		}
		queue := []int{start}
		seen[start] = true
		minX, minY, maxX, maxY, count, total := start%width, start/width, start%width, start/width, 0, uint64(0)
		for len(queue) > 0 {
			index := queue[0]
			queue = queue[1:]
			x, y := index%width, index/width
			count++
			total += uint64(peaks[index])
			minX, minY, maxX, maxY = minInt(minX, x), minInt(minY, y), maxInt(maxX, x), maxInt(maxY, y)
			for _, step := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+step[0], y+step[1]
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				next := ny*width + nx
				if !seen[next] && peaks[next] >= componentDeltaThreshold {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		components = append(components, diffComponent{BBox: rect{X: minX, Y: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}, Pixels: count, MeanDelta: float64(total) / float64(count)})
	}
	return components
}

func pixelDelta(a, b color.Color) uint64 {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return uint64(abs8(ar, br)) + uint64(abs8(ag, bg)) + uint64(abs8(ab, bb)) + uint64(abs8(aa, ba))
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
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

func satoriWritePNG(path string, img image.Image) error {
	if err := os.MkdirAll(osPathDir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func osPathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}

func abs8(a, b uint32) uint8 {
	// Standard decoders store 8-bit pixels; color.RGBA() re-expands to 16-bit
	// as v*0x101, so dividing the 16-bit delta by 0x101 recovers the exact
	// 8-bit channel delta — identical to ggrender's unified diff formula.
	// (The old >>8 truncation path is removed; numerics are unchanged.)
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return uint8(d / 0x101)
}
func max8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}
