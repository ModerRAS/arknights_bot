package media

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type visualFixtures struct {
	Scenes []struct {
		ID               string   `json:"id"`
		Template         string   `json:"template"`
		Caller           string   `json:"caller"`
		FixtureSource    string   `json:"fixtureSource"`
		GachaTSUnit      string   `json:"gachaTsUnit"`
		Scale            float64  `json:"scale"`
		LegacyFormat     string   `json:"legacyFormat"`
		CandidateFormats []string `json:"candidateFormats"`
	} `json:"scenes"`
	NewRendererAcceptance []struct {
		ID string `json:"id"`
	} `json:"newRendererAcceptance"`
}

func TestVisualFixturesCoverEveryScreenshotEntry(t *testing.T) {
	data, err := os.ReadFile("testdata/visual/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures visualFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"base": "Base.tmpl", "box": "Box.tmpl", "box-detail": "BoxDetail.tmpl", "box-summary": "BoxSummary.tmpl",
		"calendar": "Calendar.tmpl", "card": "Card.tmpl", "depot": "Depot.tmpl", "enemy": "Enemy.tmpl",
		"gacha": "Gacha.tmpl", "headhunt": "Headhunt.tmpl", "help": "Help.tmpl", "lottery": "Lottery.tmpl",
		"missing": "Missing.tmpl", "operator": "Operator.tmpl", "recruit": "Recruit.tmpl", "state": "State.tmpl",
	}
	got := make(map[string]bool, len(fixtures.Scenes))
	for _, scene := range fixtures.Scenes {
		if scene.Template != want[scene.ID] || scene.Caller == "" || scene.FixtureSource == "" || scene.Scale <= 0 || scene.LegacyFormat != "jpeg" || len(scene.CandidateFormats) != 2 {
			t.Fatalf("invalid visual fixture %#v", scene)
		}
		if got[scene.ID] {
			t.Fatalf("duplicate visual fixture %q", scene.ID)
		}
		got[scene.ID] = true
		if scene.ID == "gacha" && scene.GachaTSUnit != "milliseconds" {
			t.Fatalf("gacha fixture must preserve legacy millisecond Date semantics, got %q", scene.GachaTSUnit)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("fixture count=%d, want=%d", len(got), len(want))
	}
	for id := range want {
		if !got[id] {
			t.Errorf("missing fixture for %s", id)
		}
	}
	if len(fixtures.NewRendererAcceptance) != 9 {
		t.Fatalf("acceptance scenarios=%d, want=9", len(fixtures.NewRendererAcceptance))
	}
}

func TestVisualRegressionComparesPNGAndJPEG(t *testing.T) {
	root := t.TempDir()
	baselineDir := filepath.Join(root, "baseline")
	candidateDir := filepath.Join(root, "candidate")
	outDir := filepath.Join(root, "out")
	for _, dir := range []string{baselineDir, candidateDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 20), G: uint8(y * 20), B: 90, A: 255})
		}
	}
	writeJPEG(t, filepath.Join(baselineDir, "jpeg.jpg"), img)
	writePNG(t, filepath.Join(baselineDir, "png.png"), img)
	copyFile(t, filepath.Join(candidateDir, "jpeg.jpg"), filepath.Join(baselineDir, "jpeg.jpg"))
	copyFile(t, filepath.Join(candidateDir, "png.png"), filepath.Join(baselineDir, "png.png"))

	manifestPath := filepath.Join(baselineDir, "manifest.json")
	m := map[string]any{"entries": []map[string]any{
		{"id": "jpeg", "baseline": "jpeg.jpg", "sha256": fileHash(t, filepath.Join(baselineDir, "jpeg.jpg")), "scale": 1.5, "format": "jpeg"},
		{"id": "png", "baseline": "png.png", "sha256": fileHash(t, filepath.Join(baselineDir, "png.png")), "scale": 1, "format": "png"},
	}}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	goPath := filepath.Join(runtime.GOROOT(), "bin", "go")
	cmd := exec.Command(goPath, "run", "../../cmd/visual-regression", "-manifest", manifestPath, "-candidate-dir", candidateDir, "-out", outDir, "-jobs", "2")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("visual-regression failed: %v\n%s", err, output)
	}
	for _, path := range []string{"jpeg.heatmap.png", "png.heatmap.png", "report.json"} {
		if _, err := os.Stat(filepath.Join(outDir, path)); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func writeJPEG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func copyFile(t *testing.T, dst, src string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileHash(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
