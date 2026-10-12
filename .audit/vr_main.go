package main

import (
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
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type manifest struct {
	TemplateTreeSHA256 string  `json:"templateTreeSHA256"`
	AssetTreeSHA256    string  `json:"assetTreeSHA256"`
	Entries            []entry `json:"entries"`
}

type entry struct {
	ID       string  `json:"id"`
	Baseline string  `json:"baseline"`
	SHA256   string  `json:"sha256"`
	Scale    float64 `json:"scale"`
	Format   string  `json:"format"`
}

type result struct {
	ID              string  `json:"id"`
	Baseline        string  `json:"baseline"`
	BaselineSHA256  string  `json:"baselineSha256"`
	Candidate       string  `json:"candidate,omitempty"`
	CandidateSHA256 string  `json:"candidateSha256,omitempty"`
	Heatmap         string  `json:"heatmap,omitempty"`
	BaselineFormat  string  `json:"baselineFormat,omitempty"`
	CandidateFormat string  `json:"candidateFormat,omitempty"`
	Similarity      float64 `json:"similarity"`
	Pass            bool    `json:"pass"`
	Error           string  `json:"error,omitempty"`
}

type report struct {
	Threshold         float64  `json:"threshold"`
	MinimumSimilarity float64  `json:"minimumSimilarity"`
	Results           []result `json:"results"`
}

func main() {
	manifestPath := flag.String("manifest", "utils/media/testdata/visual/baseline/manifest.json", "legacy baseline manifest")
	candidateDir := flag.String("candidate-dir", ".visual-regression/new", "new-renderer PNG/JPEG output directory")
	outDir := flag.String("out", ".visual-regression", "report and heatmap output directory")
	threshold := flag.Float64("threshold", 99, "minimum whole-image similarity percentage")
	jobs := flag.Int("jobs", 1, "parallel image comparisons")
	verifyOnly := flag.Bool("verify", false, "verify baseline and template/asset hashes without candidate images")
	flag.Parse()

	if *jobs < 1 {
		fatal(errors.New("jobs must be positive"))
	}
	m, err := readManifest(*manifestPath)
	if err != nil {
		fatal(err)
	}
	if err := verifyManifest(m, *manifestPath); err != nil {
		fatal(err)
	}
	if *verifyOnly {
		return
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal(err)
	}

	results := compareAll(m.Entries, filepath.Dir(*manifestPath), *candidateDir, *outDir, *threshold, *jobs)
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	minimum := 100.0
	for _, r := range results {
		if r.Similarity < minimum {
			minimum = r.Similarity
		}
	}
	data, err := json.MarshalIndent(report{Threshold: *threshold, MinimumSimilarity: minimum, Results: results}, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "report.json"), append(data, '\n'), 0o644); err != nil {
		fatal(err)
	}
	for _, r := range results {
		fmt.Printf("%s: %.4f%%\n", r.ID, r.Similarity)
		if !r.Pass {
			fatal(errors.New("visual regression threshold failed; see " + filepath.Join(*outDir, "report.json")))
		}
	}
}

func readManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, err
	}
	if len(m.Entries) == 0 {
		return manifest{}, errors.New("manifest has no entries")
	}
	return m, nil
}

func verifyManifest(m manifest, manifestPath string) error {
	seen := make(map[string]bool, len(m.Entries))
	baseDir := filepath.Dir(manifestPath)
	for _, e := range m.Entries {
		if e.ID == "" || e.Baseline == "" || e.SHA256 == "" || e.Scale <= 0 {
			return fmt.Errorf("invalid manifest entry %#v", e)
		}
		if seen[e.ID] {
			return fmt.Errorf("duplicate manifest id %q", e.ID)
		}
		seen[e.ID] = true
		if e.Format != "jpeg" && e.Format != "png" {
			return fmt.Errorf("%s: unsupported baseline format %q", e.ID, e.Format)
		}
		actual, err := fileSHA256(filepath.Join(baseDir, e.Baseline))
		if err != nil {
			return fmt.Errorf("%s: %w", e.ID, err)
		}
		if !strings.EqualFold(actual, e.SHA256) {
			return fmt.Errorf("%s: baseline SHA-256 mismatch", e.ID)
		}
	}
	if m.TemplateTreeSHA256 == "" && m.AssetTreeSHA256 == "" {
		return nil
	}
	root, err := moduleRoot(filepath.Dir(manifestPath))
	if err != nil {
		return err
	}
	for _, pair := range []struct{ want, dir string }{{m.TemplateTreeSHA256, "../template"}, {m.AssetTreeSHA256, "../assets"}} {
		if pair.want == "" {
			continue
		}
		actual, err := treeSHA256(filepath.Join(root, pair.dir))
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, pair.want) {
			return fmt.Errorf("%s SHA-256 changed after legacy baseline generation", filepath.Base(pair.dir))
		}
	}
	return verifyFixtureIDs(filepath.Join(root, "utils", "media", "testdata", "visual", "fixtures.json"), m.Entries)
}

type fixtureIDs struct {
	Scenes []struct {
		ID string `json:"id"`
	} `json:"scenes"`
}

func verifyFixtureIDs(path string, entries []entry) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var fixtures fixtureIDs
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return err
	}
	if len(fixtures.Scenes) != 16 || len(entries) != 16 {
		return fmt.Errorf("fixture scenes=%d and manifest entries=%d; both must be 16", len(fixtures.Scenes), len(entries))
	}
	ids := make(map[string]bool, 16)
	for _, scene := range fixtures.Scenes {
		if scene.ID == "" || ids[scene.ID] {
			return fmt.Errorf("invalid or duplicate fixture id %q", scene.ID)
		}
		ids[scene.ID] = true
	}
	for _, e := range entries {
		if !ids[e.ID] {
			return fmt.Errorf("manifest id %q missing from fixtures", e.ID)
		}
		delete(ids, e.ID)
	}
	for id := range ids {
		return fmt.Errorf("fixture id %q missing from manifest", id)
	}
	return nil
}

func compareAll(entries []entry, baselineDir, candidateDir, outDir string, threshold float64, jobs int) []result {
	work := make(chan entry)
	results := make(chan result, len(entries))
	var wg sync.WaitGroup
	for range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range work {
				results <- compare(e, baselineDir, candidateDir, outDir, threshold)
			}
		}()
	}
	go func() {
		for _, e := range entries {
			work <- e
		}
		close(work)
		wg.Wait()
		close(results)
	}()
	out := make([]result, 0, len(entries))
	for r := range results {
		out = append(out, r)
	}
	return out
}

func compare(e entry, baselineDir, candidateDir, outDir string, threshold float64) result {
	r := result{ID: e.ID, Baseline: e.Baseline, BaselineSHA256: e.SHA256}
	base, baseFormat, err := decode(filepath.Join(baselineDir, e.Baseline))
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.BaselineFormat = baseFormat
	candidatePath, err := candidateFile(candidateDir, e.ID)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Candidate = candidatePath
	if r.CandidateSHA256, err = fileSHA256(candidatePath); err != nil {
		r.Error = err.Error()
		return r
	}
	candidate, candidateFormat, err := decode(candidatePath)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.CandidateFormat = candidateFormat
	if base.Bounds().Size() != candidate.Bounds().Size() {
		r.Error = fmt.Sprintf("dimension mismatch: baseline=%v candidate=%v", base.Bounds().Size(), candidate.Bounds().Size())
		return r
	}

	heat := image.NewRGBA(base.Bounds())
	var total uint64
	bounds := base.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, ba := base.At(x, y).RGBA()
			cr, cg, cb, ca := candidate.At(x, y).RGBA()
			delta := abs(br, cr) + abs(bg, cg) + abs(bb, cb) + abs(ba, ca)
			total += uint64(delta)
			level := uint8(delta / 4 >> 8)
			heat.SetRGBA(x, y, color.RGBA{R: level, A: 255})
		}
	}
	heatPath := filepath.Join(outDir, e.ID+".heatmap.png")
	if err := writePNG(heatPath, heat); err != nil {
		r.Error = err.Error()
		return r
	}
	r.Heatmap = heatPath
	max := float64(bounds.Dx() * bounds.Dy() * 4 * 65535)
	r.Similarity = 100 * (1 - float64(total)/max)
	r.Pass = r.Similarity >= threshold
	return r
}

func candidateFile(dir, id string) (string, error) {
	for _, ext := range []string{".png", ".jpeg", ".jpg"} {
		path := filepath.Join(dir, id+ext)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s: candidate PNG/JPEG not found", id)
}

func decode(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", err
	}
	if format != "png" && format != "jpeg" {
		return nil, "", fmt.Errorf("%s: unsupported image format %q", path, format)
	}
	return img, format, nil
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func treeSHA256(root string) (string, error) {
	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, path := range files {
		digest, err := fileSHA256(path)
		if err != nil {
			return "", err
		}
		rel, _ := filepath.Rel(root, path)
		fmt.Fprintf(h, "%s\x00%s\n", filepath.ToSlash(rel), digest)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func moduleRoot(start string) (string, error) {
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
	}
}

func abs(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "visual-regression:", err)
	os.Exit(1)
}
