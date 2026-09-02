package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"arknights_bot/utils/media"
)

const threshold = 0.99

var expected = []string{"base", "box", "box-detail", "box-summary", "calendar", "card", "depot", "enemy", "gacha", "headhunt", "help", "lottery", "missing", "operator", "recruit", "state"}

type spec struct {
	ID        string          `json:"id"`
	Component string          `json:"component"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	Scale     float64         `json:"scale"`
	Props     json.RawMessage `json:"props"`
}

type baselineEntry struct {
	ID          string  `json:"id"`
	Scale       float64 `json:"scale"`
	Format      string  `json:"format"`
	Baseline    string  `json:"baseline"`
	SHA256      string  `json:"sha256"`
	PixelWidth  int     `json:"pixelWidth"`
	PixelHeight int     `json:"pixelHeight"`
}

type baselineManifest struct {
	Entries []baselineEntry `json:"entries"`
}

type pageResult struct {
	ID     string  `json:"id"`
	Path   string  `json:"path,omitempty"`
	SHA256 string  `json:"sha256,omitempty"`
	Width  int     `json:"width,omitempty"`
	Height int     `json:"height,omitempty"`
	Scale  float64 `json:"scale"`
	Format string  `json:"format,omitempty"`
	Error  string  `json:"error,omitempty"`
}

type renderReport struct {
	Chain       string            `json:"chain"`
	SpecFiles   map[string]string `json:"specFiles"`
	ExpectedIDs []string          `json:"expectedIds"`
	Results     []pageResult      `json:"results"`
	Rendered    int               `json:"rendered"`
	Failed      int               `json:"failed"`
}

type rect struct{ X, Y, Width, Height int }

type globalOffset struct{ DX, DY int }

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

type finalPage struct {
	ID                 string          `json:"id"`
	Width              int             `json:"width"`
	Height             int             `json:"height"`
	OldWidth           int             `json:"oldWidth"`
	OldHeight          int             `json:"oldHeight"`
	NewWidth           int             `json:"newWidth"`
	NewHeight          int             `json:"newHeight"`
	Scale              float64         `json:"scale"`
	OldFormat          string          `json:"oldFormat"`
	NewFormat          string          `json:"newFormat"`
	OldSHA256          string          `json:"oldSha256"`
	NewSHA256          string          `json:"newSha256"`
	OldGate            bool            `json:"oldGate"`
	Similarity         *float64        `json:"similarity,omitempty"`
	RawSimilarity      *float64        `json:"rawSimilarity,omitempty"`
	AlignedSimilarity  *float64        `json:"alignedSimilarity,omitempty"`
	GlobalOffset       *globalOffset   `json:"globalOffset,omitempty"`
	LocalPass          bool            `json:"localPass"`
	Meaningful         []diffComponent `json:"meaningful,omitempty"`
	ObjectShifts       []localShift    `json:"objectShifts,omitempty"`
	ZeroByteIdentity   bool            `json:"zeroByteIdentity"`
	DiffBBox           *rect           `json:"diffBBox,omitempty"`
	AlignedDiffBBox    *rect           `json:"alignedDiffBBox,omitempty"`
	Pass               bool            `json:"pass"`
	OldPath            string          `json:"oldPath"`
	NewPath            string          `json:"newPath"`
	DiffPath           string          `json:"diffPath"`
	HeatmapPath        string          `json:"heatmapPath"`
	AlignedDiffPath    string          `json:"alignedDiffPath"`
	AlignedHeatmapPath string          `json:"alignedHeatmapPath"`
	Error              string          `json:"error,omitempty"`
}

type legacyCapture struct {
	ScriptSHA256      string `json:"scriptSha256"`
	ClockBehavior     string `json:"clockBehavior"`
	Selector          string `json:"selector"`
	DeviceScaleFactor string `json:"deviceScaleFactor"`
	JPEG              string `json:"jpeg"`
	GachaSettleMS     int    `json:"gachaSettleMs"`
}

type finalReport struct {
	SchemaVersion          int               `json:"schemaVersion"`
	IDs                    []string          `json:"ids"`
	Threshold              float64           `json:"threshold"`
	ManifestSHA256         string            `json:"manifestSha256"`
	ResourceManifestSHA256 string            `json:"resourceManifestSha256"`
	ClockErratum           string            `json:"clockErratum"`
	ClockErratumSHA256     string            `json:"clockErratumSha256"`
	LegacyCapture          legacyCapture     `json:"legacyCapture"`
	SpecFiles              map[string]string `json:"specFiles"`
	Command                string            `json:"command"`
	RunNote                string            `json:"runNote"`
	Pages                  []finalPage       `json:"pages"`
	MinimumSimilarity      *float64          `json:"minimumSimilarity,omitempty"`
	MinimumID              string            `json:"minimumId,omitempty"`
	PassingIDs             []string          `json:"passingIds"`
}

func main() {
	specDir := flag.String("spec-dir", "", "directory containing canonical *specs.ndjson exports")
	out := flag.String("out", "", "final artifact directory")
	baseline := flag.String("baseline", "", "immutable visual baseline manifest")
	resources := flag.String("resource-manifest", "", "immutable visual resource manifest")
	clockErratum := flag.String("clock-erratum", "", "immutable legacy capture clock erratum")
	legacyScript := flag.String("legacy-script", "", "legacy Playwright capture script")
	runNote := flag.String("run-note", "actual Go HTTP RenderSpec -> media.ScreenshotPNG -> resident Node renderer", "report run note")
	flag.Parse()
	if *specDir == "" || *out == "" || *baseline == "" || *resources == "" || *clockErratum == "" || *legacyScript == "" {
		fatal(errors.New("-spec-dir, -out, -baseline, -resource-manifest, -clock-erratum, and -legacy-script are required"))
	}
	pages, hashes, err := loadSpecs(*specDir)
	if err != nil {
		fatal(err)
	}
	if err := verifyIDs(pages); err != nil {
		fatal(err)
	}
	manifestData, err := os.ReadFile(*baseline)
	if err != nil {
		fatal(err)
	}
	resourceData, err := os.ReadFile(*resources)
	if err != nil {
		fatal(err)
	}
	errataData, err := os.ReadFile(*clockErratum)
	if err != nil {
		fatal(err)
	}
	legacyScriptData, err := os.ReadFile(*legacyScript)
	if err != nil {
		fatal(err)
	}
	var manifest baselineManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		fatal(err)
	}
	entries, err := indexManifest(manifest)
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(*out, "new"), 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(*out, "compare"), 0o755); err != nil {
		fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fatal(err)
	}
	server := &http.Server{Handler: specHandler(pages)}
	go func() { _ = server.Serve(listener) }()
	defer func() { _ = server.Close(); media.Shutdown() }()

	rendered := renderCurrent("http://"+listener.Addr().String(), *out, pages, hashes)
	writeJSON(filepath.Join(*out, "new-render-report.json"), rendered)
	bundle, compareErr := compareBundle(*out, manifestData, resourceData, *clockErratum, errataData, legacyScriptData, entries, pages, hashes, *runNote)
	writeJSON(filepath.Join(*out, "report.json"), bundle)
	writeRunNote(filepath.Join(*out, "README.md"), *runNote)
	if rendered.Failed != 0 || rendered.Rendered != len(expected) {
		fatal(fmt.Errorf("rendered=%d failed=%d", rendered.Rendered, rendered.Failed))
	}
	if compareErr != nil {
		fatal(compareErr)
	}
}

func renderCurrent(baseURL, out string, pages map[string]spec, hashes map[string]string) renderReport {
	report := renderReport{Chain: "Go HTTP RenderSpec -> media.ScreenshotPNG -> one resident Node runner", SpecFiles: hashes, ExpectedIDs: expected}
	for _, id := range expected {
		page := pages[id]
		item := pageResult{ID: id, Scale: page.Scale, Path: filepath.ToSlash(filepath.Join("new", id+".png"))}
		data, err := media.ScreenshotPNG(baseURL+"/render/"+id, 0, page.Scale)
		if err != nil {
			item.Error = err.Error()
			report.Failed++
			report.Results = append(report.Results, item)
			continue
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			item.Error = "decode PNG: " + err.Error()
			report.Failed++
			report.Results = append(report.Results, item)
			continue
		}
		item.Width, item.Height, item.Format, item.SHA256 = config.Width, config.Height, "png", sha256Hex(data)
		if err := os.WriteFile(filepath.Join(out, item.Path), data, 0o644); err != nil {
			item.Error = err.Error()
			report.Failed++
		} else {
			report.Rendered++
		}
		report.Results = append(report.Results, item)
	}
	return report
}

func compareBundle(out string, manifestData, resourceData []byte, clockErratum string, errataData, legacyScriptData []byte, entries map[string]baselineEntry, pages map[string]spec, hashes map[string]string, runNote string) (finalReport, error) {
	errataPath, err := filepath.Rel(out, clockErratum)
	if err != nil {
		return finalReport{}, err
	}
	report := finalReport{SchemaVersion: 1, IDs: expected, Threshold: threshold, ManifestSHA256: sha256Hex(manifestData), ResourceManifestSHA256: sha256Hex(resourceData), ClockErratum: filepath.ToSlash(errataPath), ClockErratumSHA256: sha256Hex(errataData), LegacyCapture: legacyCapture{ScriptSHA256: sha256Hex(legacyScriptData), ClockBehavior: "Calendar/State fixed erratum clock; Gacha advancing fake clock anchored at erratum", Selector: "manifest entries; --id supports gacha-only diagnostic replay", DeviceScaleFactor: "manifest entry scale", JPEG: "Playwright locator screenshot type=jpeg, default quality", GachaSettleMS: 3000}, SpecFiles: hashes, Command: "visual-final -spec-dir <canonical-spec-dir> -out <final-dir> -baseline <baseline-manifest> -resource-manifest <resource-manifest> -clock-erratum <capture-clock-erratum> -legacy-script <legacy-capture.mjs>", RunNote: runNote}
	for _, id := range expected {
		entry, page := entries[id], pages[id]
		item := finalPage{ID: id, Width: entry.PixelWidth, Height: entry.PixelHeight, OldWidth: entry.PixelWidth, OldHeight: entry.PixelHeight, NewWidth: entry.PixelWidth, NewHeight: entry.PixelHeight, Scale: page.Scale, OldFormat: entry.Format, NewFormat: "png", OldPath: filepath.ToSlash(filepath.Join("old", id+".jpg")), NewPath: filepath.ToSlash(filepath.Join("new", id+".png")), DiffPath: filepath.ToSlash(filepath.Join("compare", id+".diff.png")), HeatmapPath: filepath.ToSlash(filepath.Join("compare", id+".heatmap.png")), AlignedDiffPath: filepath.ToSlash(filepath.Join("compare", id+".aligned.diff.png")), AlignedHeatmapPath: filepath.ToSlash(filepath.Join("compare", id+".aligned.heatmap.png"))}
		oldData, oldErr := os.ReadFile(filepath.Join(out, item.OldPath))
		newData, newErr := os.ReadFile(filepath.Join(out, item.NewPath))
		if oldErr == nil {
			item.OldSHA256 = sha256Hex(oldData)
			item.OldGate = item.OldSHA256 == entry.SHA256
		}
		if newErr == nil {
			item.NewSHA256 = sha256Hex(newData)
		}
		if oldErr != nil {
			item.Error = oldErr.Error()
		}
		if newErr != nil {
			item.Error = joinError(item.Error, newErr.Error())
		}
		if oldErr == nil && newErr == nil {
			comparison, err := comparePageImages(id, oldData, newData, filepath.Join(out, item.DiffPath), filepath.Join(out, item.HeatmapPath), filepath.Join(out, item.AlignedDiffPath), filepath.Join(out, item.AlignedHeatmapPath))
			if err != nil {
				item.Error = joinError(item.Error, err.Error())
			} else {
				item.RawSimilarity, item.AlignedSimilarity, item.Similarity = &comparison.RawScore, &comparison.AlignedScore, &comparison.AlignedScore
				item.DiffBBox, item.AlignedDiffBBox = comparison.RawBBox, comparison.AlignedBBox
				item.GlobalOffset, item.LocalPass = &comparison.Offset, comparison.LocalPass
				item.Meaningful, item.ObjectShifts = comparison.Meaningful, comparison.ObjectShifts
				item.ZeroByteIdentity = comparison.ZeroByteIdentity
				item.Pass = item.OldGate && comparison.AlignedScore >= threshold && comparison.LocalPass
			}
		}
		if item.Similarity != nil && (report.MinimumSimilarity == nil || *item.Similarity < *report.MinimumSimilarity) {
			score := *item.Similarity
			report.MinimumSimilarity, report.MinimumID = &score, id
		}
		if item.Pass {
			report.PassingIDs = append(report.PassingIDs, id)
		}
		report.Pages = append(report.Pages, item)
	}
	sort.Strings(report.PassingIDs)
	for _, page := range report.Pages {
		if page.Error != "" || !page.OldGate || page.Similarity == nil || *page.Similarity < threshold || !page.LocalPass {
			return report, fmt.Errorf("final bundle gate failed for %s: error=%q oldGate=%t similarity=%v localPass=%t", page.ID, page.Error, page.OldGate, page.Similarity, page.LocalPass)
		}
	}
	return report, nil
}

func diffImages(oldData, newData []byte, diffPath, heatmapPath string) (float64, *rect, error) {
	comparison, err := compareImages(oldData, newData, diffPath, heatmapPath, diffPath+".aligned", heatmapPath+".aligned")
	return comparison.AlignedScore, comparison.AlignedBBox, err
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
	if err := writePNG(diffPath, diff); err != nil {
		return 0, nil, nil, err
	}
	if err := writePNG(heatmapPath, heatmap); err != nil {
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

func writePNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
func abs8(a, b uint32) uint8 {
	if a > b {
		return uint8((a - b) >> 8)
	}
	return uint8((b - a) >> 8)
}
func max8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}
func joinError(current, next string) string {
	if current == "" {
		return next
	}
	return current + "; " + next
}

func indexManifest(manifest baselineManifest) (map[string]baselineEntry, error) {
	if len(manifest.Entries) != len(expected) {
		return nil, fmt.Errorf("baseline count=%d, want=%d", len(manifest.Entries), len(expected))
	}
	entries := make(map[string]baselineEntry, len(expected))
	for _, entry := range manifest.Entries {
		if entry.ID == "" || entry.Baseline == "" || entry.SHA256 == "" || entry.PixelWidth < 1 || entry.PixelHeight < 1 {
			return nil, fmt.Errorf("invalid baseline entry %#v", entry)
		}
		entries[entry.ID] = entry
	}
	for _, id := range expected {
		if _, ok := entries[id]; !ok {
			return nil, fmt.Errorf("missing baseline %q", id)
		}
	}
	return entries, nil
}

func specHandler(pages map[string]spec) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/render/")
		page, ok := pages[id]
		if r.Method != http.MethodGet || !ok || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Component string          `json:"component"`
			Width     int             `json:"width"`
			Height    int             `json:"height"`
			Props     json.RawMessage `json:"props"`
		}{page.Component, page.Width, page.Height, page.Props})
	})
}

func loadSpecs(dir string) (map[string]spec, map[string]string, error) {
	files := []string{"static-specs.ndjson", "dynamic-render-specs.ndjson", "complex_render_specs.ndjson"}
	pages := make(map[string]spec, len(expected))
	hashes := make(map[string]string, len(files))
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		hashes[name] = sha256Hex(data)
		scanner := bufio.NewScanner(bytes.NewReader(data))
		scanner.Buffer(make([]byte, 64<<10), 1<<20)
		for scanner.Scan() {
			var page spec
			if err := json.Unmarshal(scanner.Bytes(), &page); err != nil {
				return nil, nil, fmt.Errorf("%s: %w", name, err)
			}
			if page.ID == "" || page.Component == "" || page.Width < 1 || page.Height < 1 || page.Scale <= 0 || !json.Valid(page.Props) {
				return nil, nil, fmt.Errorf("%s: invalid render spec %#v", name, page)
			}
			if _, exists := pages[page.ID]; exists {
				return nil, nil, fmt.Errorf("duplicate spec id %q", page.ID)
			}
			pages[page.ID] = page
		}
		if err := scanner.Err(); err != nil {
			return nil, nil, err
		}
	}
	return pages, hashes, nil
}

func verifyIDs(pages map[string]spec) error {
	if len(pages) != len(expected) {
		return fmt.Errorf("spec count=%d, want=%d", len(pages), len(expected))
	}
	for _, id := range expected {
		if _, ok := pages[id]; !ok {
			return fmt.Errorf("missing spec %q", id)
		}
	}
	return nil
}
func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		fatal(err)
	}
}
func writeRunNote(path, note string) {
	content := "# Visual Final Bundle\n\n" + note + "\n\nLegacy replay requires `legacy-capture.mjs --out <final-dir> --root <repo-root> --baseline <repo-root>/src/utils/media/testdata/visual/baseline --clock-erratum <repo-root>/src/utils/media/testdata/visual/capture-clock-erratum.json --base-url <legacy-url> --playwright <playwright-module>`. It replays manifest entries (or `--id gacha` for diagnostics) at each manifest DPR, takes Playwright locator JPEGs at default quality, waits 3000ms for Gacha, and uses an advancing fake wall clock anchored to the erratum for Gacha ECharts. Calendar and State use their fixed erratum clocks.\n\nCurrent replay requires `visual-final -spec-dir <canonical-spec-dir> -out <final-dir> -baseline <baseline-manifest> -resource-manifest <resource-manifest> -clock-erratum <capture-clock-erratum> -legacy-script <legacy-capture.mjs>`.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fatal(err)
	}
}
func sha256Hex(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func fatal(err error)              { fmt.Fprintln(os.Stderr, "visual-final:", err); os.Exit(1) }
