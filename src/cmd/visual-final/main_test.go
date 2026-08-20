package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"testing"
)

func TestFinalReportRejectsUnacceptedPage(t *testing.T) {
	report := finalReport{Pages: []finalPage{{ID: "enemy", OldGate: true, Pass: false}}}
	if err := validateFinalReport(report); err == nil {
		t.Fatal("expected rejected page to fail the final bundle gate")
	}
}

func TestCompareImagesAlignsOneGlobalTranslation(t *testing.T) {
	old := testImage(64, 64, image.Rect(8, 8, 28, 28), image.Rect(36, 24, 54, 46))
	candidate := translateImage(old, globalOffset{DX: 3, DY: -2}, canvasBackground(old))
	result, err := compareImages(testPNG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if result.RawScore >= 1 {
		t.Fatalf("raw score=%f, want translation penalty", result.RawScore)
	}
	if result.AlignedScore != 1 || result.Offset != (globalOffset{DX: -3, DY: 2}) || !result.LocalPass {
		t.Fatalf("aligned=%f offset=%+v local=%v", result.AlignedScore, result.Offset, result.LocalPass)
	}
}

func TestCompareImagesAlignsDarkCanvasTranslation(t *testing.T) {
	old := darkTestImage(64, 64, image.Rect(10, 10, 30, 30), image.Rect(40, 30, 55, 50))
	candidate := translateImage(old, globalOffset{DX: -2, DY: 3}, canvasBackground(old))
	result, err := compareImages(testPNG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if result.AlignedScore != 1 || result.Offset != (globalOffset{DX: 2, DY: -3}) || !result.LocalPass {
		t.Fatalf("aligned=%f offset=%+v local=%v", result.AlignedScore, result.Offset, result.LocalPass)
	}
}

func TestSelectFinalOffsetNeverRegressesRawSimilarity(t *testing.T) {
	old := testImage(64, 64, image.Rect(8, 8, 28, 28), image.Rect(40, 40, 60, 60))
	candidate := translateImage(old, globalOffset{DX: 2, DY: 0}, canvasBackground(old))
	offset := selectFinalOffset(old, candidate, globalOffset{DX: -2, DY: 0}, canvasBackground(old))
	if offset != (globalOffset{DX: -2, DY: 0}) {
		t.Fatalf("offset=%+v, want exact-score winner", offset)
	}
	if got, raw := imageSimilarity(old, translateImage(candidate, offset, canvasBackground(old))), imageSimilarity(old, candidate); got < raw {
		t.Fatalf("aligned=%f below raw=%f", got, raw)
	}
}

func TestCompareImagesRejectsOneObjectShiftAfterRegistration(t *testing.T) {
	old := testImage(64, 64, image.Rect(2, 2, 34, 62), image.Rect(40, 40, 60, 60))
	candidate := image.NewRGBA(old.Bounds())
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			candidate.Set(x, y, old.At(x, y))
		}
	}
	for y := 40; y < 60; y++ {
		for x := 40; x < 60; x++ {
			candidate.Set(x, y, color.White)
		}
	}
	for y := 40; y < 60; y++ {
		for x := 44; x < 64; x++ {
			candidate.Set(x, y, color.Black)
		}
	}
	result, err := compareImages(testPNG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Offset != (globalOffset{}) || result.LocalPass || len(result.Components) == 0 {
		t.Fatalf("aligned=%f offset=%+v local=%v components=%d", result.AlignedScore, result.Offset, result.LocalPass, len(result.Components))
	}
}

func TestCompareImagesAcceptsJPEGNoiseSpeckles(t *testing.T) {
	old := testImage(256, 256, image.Rect(32, 32, 224, 224))
	candidate := image.NewRGBA(old.Bounds())
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			candidate.Set(x, y, old.At(x, y))
		}
	}
	for _, point := range [][2]int{{1, 1}, {3, 250}, {250, 3}, {252, 252}, {7, 127}, {127, 7}} {
		candidate.Set(point[0], point[1], color.RGBA{R: 180, G: 180, B: 180, A: 255})
	}
	result, err := compareImages(testJPEG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.LocalPass {
		t.Fatalf("JPEG/PNG noise should not be an object failure: components=%d", len(result.Components))
	}
}

func TestCompareImagesIgnoresNoiseSpeckles(t *testing.T) {
	old := testImage(64, 64, image.Rect(8, 8, 56, 56))
	candidate := image.NewRGBA(old.Bounds())
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			candidate.Set(x, y, old.At(x, y))
		}
	}
	for _, point := range [][2]int{{1, 1}, {3, 60}, {60, 3}, {61, 61}, {5, 35}, {35, 5}} {
		candidate.Set(point[0], point[1], color.RGBA{R: 180, G: 180, B: 180, A: 255})
	}
	result, err := compareImages(testPNG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.LocalPass || len(result.Components) == 0 {
		t.Fatalf("local=%v components=%d", result.LocalPass, len(result.Components))
	}
}

func TestCompareImagesPadsEdgesInsteadOfCropping(t *testing.T) {
	old := testImage(32, 32, image.Rect(5, 5, 20, 20), image.Rect(28, 8, 32, 24))
	candidate := translateImage(old, globalOffset{DX: 1, DY: 0}, canvasBackground(old))
	result, err := compareImages(testPNG(old), testPNG(candidate), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Offset != (globalOffset{DX: -1, DY: 0}) || result.AlignedScore >= 1 || result.AlignedBBox == nil {
		t.Fatalf("offset=%+v aligned=%f bbox=%+v", result.Offset, result.AlignedScore, result.AlignedBBox)
	}
}

func TestCompareImagesRejectsDimensionChanges(t *testing.T) {
	_, err := compareImages(testPNG(testImage(16, 16, image.Rect(1, 1, 4, 4))), testPNG(testImage(15, 16, image.Rect(1, 1, 4, 4))), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err == nil {
		t.Fatal("expected dimension mismatch")
	}
}

func darkTestImage(width, height int, blocks ...image.Rectangle) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 18, G: 22, B: 28, A: 255})
		}
	}
	for _, block := range blocks {
		for y := block.Min.Y; y < block.Max.Y; y++ {
			for x := block.Min.X; x < block.Max.X; x++ {
				img.Set(x, y, color.RGBA{R: 230, G: 230, B: 230, A: 255})
			}
		}
	}
	return img
}

func testImage(width, height int, blocks ...image.Rectangle) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}
	for _, block := range blocks {
		for y := block.Min.Y; y < block.Max.Y; y++ {
			for x := block.Min.X; x < block.Max.X; x++ {
				img.Set(x, y, color.Black)
			}
		}
	}
	return img
}

func testJPEG(img image.Image) []byte {
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 80}); err != nil {
		panic(err)
	}
	return out.Bytes()
}

func testPNG(img image.Image) []byte {
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		panic(err)
	}
	return out.Bytes()
}
