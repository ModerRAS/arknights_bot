package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareImagesAlignsGlobalTranslation(t *testing.T) {
	old := testImage(160, 120)
	addAnchors(old)
	fillTestRect(old, image.Rect(60, 45, 100, 75), color.RGBA{20, 40, 220, 255})
	new := translateImage(old, globalOffset{DX: 4, DY: 3}, canvasBackground(old))

	comparison := compareTestImages(t, pngTestData(t, old), pngTestData(t, new))
	if comparison.Offset != (globalOffset{DX: -4, DY: -3}) || comparison.AlignedScore != 1 || !comparison.LocalPass {
		t.Fatalf("offset=%+v aligned=%f localPass=%t", comparison.Offset, comparison.AlignedScore, comparison.LocalPass)
	}
}

func TestCompareImagesRejectsLocalObjectShift(t *testing.T) {
	old, new := testImage(600, 500), testImage(600, 500)
	addAnchors(old)
	addAnchors(new)
	fillTestRect(old, image.Rect(280, 220, 322, 265), color.RGBA{20, 40, 220, 255})
	fillTestRect(new, image.Rect(286, 220, 328, 265), color.RGBA{20, 40, 220, 255})

	comparison := compareTestImages(t, pngTestData(t, old), pngTestData(t, new))
	if comparison.LocalPass || len(comparison.Meaningful) == 0 {
		t.Fatalf("local shift accepted: %+v", comparison)
	}
}

func TestZeroTranslationPreservesRGBAAndScore(t *testing.T) {
	img := testImage(80, 60)
	fillTestRect(img, image.Rect(12, 12, 36, 38), color.RGBA{10, 20, 230, 255})
	data := pngTestData(t, img)
	decoded, err := decodeRGBA(data, "test")
	if err != nil {
		t.Fatal(err)
	}
	translated := translateImage(decoded, globalOffset{}, canvasBackground(decoded))
	if !bytes.Equal(translated.Pix, decoded.Pix) {
		t.Fatal("zero offset changed decoded RGBA bytes")
	}
	comparison := compareTestImages(t, data, data)
	if !comparison.ZeroByteIdentity || comparison.RawScore != comparison.AlignedScore || comparison.Offset != (globalOffset{}) {
		t.Fatalf("identity=%t raw=%f aligned=%f offset=%+v", comparison.ZeroByteIdentity, comparison.RawScore, comparison.AlignedScore, comparison.Offset)
	}
	if got := pixelDelta(color.RGBA{B: 255, A: 255}, color.RGBA{A: 255}); got != 255 {
		t.Fatalf("blue delta=%d, want 255", got)
	}
}

func TestSelectFinalOffsetKeepsZeroWhenSuggestionWorsens(t *testing.T) {
	img := testImage(80, 60)
	addAnchors(img)
	if got := selectFinalOffset(img, img, globalOffset{DX: 1}, canvasBackground(img)); got != (globalOffset{}) {
		t.Fatalf("offset=%+v, want zero", got)
	}
}

func TestCompareImagesRejectsEdgeObjectLoss(t *testing.T) {
	old, new := testImage(200, 120), testImage(200, 120)
	fillTestRect(old, image.Rect(0, 30, 42, 75), color.RGBA{20, 40, 220, 255})
	fillTestRect(new, image.Rect(0, 30, 34, 75), color.RGBA{20, 40, 220, 255})
	comparison := compareTestImages(t, pngTestData(t, old), pngTestData(t, new))
	if comparison.LocalPass {
		t.Fatal("edge object loss accepted")
	}
}

func TestCompareImagesRejectsDimensionMismatch(t *testing.T) {
	_, err := compareImages(pngTestData(t, testImage(20, 20)), pngTestData(t, testImage(21, 20)), filepath.Join(t.TempDir(), "raw.png"), filepath.Join(t.TempDir(), "raw-heat.png"), filepath.Join(t.TempDir(), "aligned.png"), filepath.Join(t.TempDir(), "aligned-heat.png"))
	if err == nil {
		t.Fatal("dimension mismatch accepted")
	}
}

func TestCompareImagesAcceptsJPEGMicroNoiseWithStableAnchors(t *testing.T) {
	img := testImage(160, 120)
	addAnchors(img)
	fillTestRect(img, image.Rect(50, 40, 110, 80), color.RGBA{96, 96, 96, 255})
	comparison := compareTestImages(t, pngTestData(t, img), jpegTestData(t, img, 100))
	if !comparison.LocalPass || comparison.Offset != (globalOffset{}) {
		t.Fatalf("jpeg micro-noise rejected: offset=%+v meaningful=%+v", comparison.Offset, comparison.Meaningful)
	}
}

func TestKnownAnchorSignaturesRemainMeaningful(t *testing.T) {
	blank := testImage(64, 64)
	cases := []struct {
		page string
		box  rect
	}{
		{"calendar", rect{Width: 2546, Height: 38}},
		{"calendar", rect{Width: 2546, Height: 25}},
		{"lottery", rect{Width: 139, Height: 138}},
		{"lottery", rect{Width: 42, Height: 45}},
	}
	for _, tc := range cases {
		t.Run(tc.page, func(t *testing.T) {
			pass, meaningful, _ := localGate(tc.page, blank, blank, []diffComponent{{BBox: tc.box, Pixels: 1, MeanDelta: 1}}, canvasBackground(blank))
			if pass || len(meaningful) != 1 {
				t.Fatalf("%s component %+v ignored", tc.page, tc.box)
			}
		})
	}
}

func TestCompareBundleErrorsForFailedPage(t *testing.T) {
	out := t.TempDir()
	for _, dir := range []string{"old", "new", "compare"} {
		if err := os.Mkdir(filepath.Join(out, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	entries := make(map[string]baselineEntry, len(expected))
	pages := make(map[string]spec, len(expected))
	for _, id := range expected {
		entries[id] = baselineEntry{ID: id, Scale: 1, Format: "jpeg", Baseline: "old/" + id + ".jpg", PixelWidth: 64, PixelHeight: 64}
		pages[id] = spec{ID: id, Component: id, Width: 64, Height: 64, Scale: 1, Props: jsonTestProps}
	}
	old := testImage(64, 64)
	fillTestRect(old, image.Rect(10, 10, 52, 55), color.RGBA{20, 40, 220, 255})
	oldData := pngTestData(t, old)
	entries["base"] = baselineEntry{ID: "base", Scale: 1, Format: "jpeg", Baseline: "old/base.jpg", SHA256: sha256Hex(oldData), PixelWidth: 64, PixelHeight: 64}
	if err := os.WriteFile(filepath.Join(out, "old", "base.jpg"), oldData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "new", "base.png"), pngTestData(t, testImage(64, 64)), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := compareBundle(out, []byte("manifest"), []byte("resources"), filepath.Join(out, "clock.json"), []byte("clock"), []byte("script"), entries, pages, nil, "test")
	if err == nil {
		t.Fatal("failed page did not fail bundle")
	}
}

var jsonTestProps = []byte(`{}`)

func testImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fillTestRect(img, img.Bounds(), color.RGBA{220, 220, 220, 255})
	return img
}

func addAnchors(img *image.RGBA) {
	bounds := img.Bounds()
	for _, box := range []image.Rectangle{
		image.Rect(8, 8, 20, 20),
		image.Rect(bounds.Max.X-20, 8, bounds.Max.X-8, 20),
		image.Rect(8, bounds.Max.Y-20, 20, bounds.Max.Y-8),
		image.Rect(bounds.Max.X-20, bounds.Max.Y-20, bounds.Max.X-8, bounds.Max.Y-8),
	} {
		fillTestRect(img, box, color.RGBA{20, 40, 220, 255})
	}
}

func fillTestRect(img *image.RGBA, box image.Rectangle, c color.RGBA) {
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func pngTestData(t *testing.T, img image.Image) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func jpegTestData(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := jpeg.Encode(&data, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func compareTestImages(t *testing.T, oldData, newData []byte) imageComparison {
	t.Helper()
	dir := t.TempDir()
	comparison, err := compareImages(oldData, newData, filepath.Join(dir, "raw.png"), filepath.Join(dir, "raw-heat.png"), filepath.Join(dir, "aligned.png"), filepath.Join(dir, "aligned-heat.png"))
	if err != nil {
		t.Fatal(err)
	}
	return comparison
}
