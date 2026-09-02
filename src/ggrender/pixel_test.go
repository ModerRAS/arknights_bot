package ggrender

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

type ReportEntry struct {
	Scene      string  `json:"scene"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Format     string  `json:"format"`
	Scale      float64 `json:"scale"`
	HashOld    string  `json:"hashOld"`
	HashNew    string  `json:"hashNew"`
	Hash       string  `json:"hash"`
	Score      float64 `json:"score"`
	Similarity float64 `json:"similarity"`
	BBox       [4]int  `json:"bbox"`
	Passed     bool    `json:"passed"`
	OldPath    string  `json:"oldPath"`
	NewPath    string  `json:"newPath"`
	DiffPath   string  `json:"diffPath"`
}

type manifestFile struct {
	Entries []struct {
		ID         string  `json:"id"`
		Baseline   string  `json:"baseline"`
		Sha256     string  `json:"sha256"`
		Scale      float64 `json:"scale"`
		Format     string  `json:"format"`
		PixelWidth int     `json:"pixelWidth"`
		PixelHeight int    `json:"pixelHeight"`
		BBox       struct {
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"bbox"`
	} `json:"entries"`
}

// unified similarity: 1 - sum(|dR|+|dG|+|dB|+|dA|)/(w*h*4*255)
func similarityNormalized(old, new *image.RGBA) (float64, [4]int) {
	w, h := old.Bounds().Dx(), old.Bounds().Dy()
	var sum int64
	minX, minY := w, h
	maxX, maxY := -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			co := old.RGBAAt(x, y)
			cn := new.RGBAAt(x, y)
			dr := int(co.R) - int(cn.R)
			if dr < 0 {
				dr = -dr
			}
			dg := int(co.G) - int(cn.G)
			if dg < 0 {
				dg = -dg
			}
			db := int(co.B) - int(cn.B)
			if db < 0 {
				db = -db
			}
			da := int(co.A) - int(cn.A)
			if da < 0 {
				da = -da
			}
			if dr+dg+db+da != 0 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
			sum += int64(dr + dg + db + da)
		}
	}
	total := int64(w * h * 4 * 255)
	var sim float64
	if total > 0 {
		sim = 1.0 - float64(sum)/float64(total)
	} else {
		sim = 1
	}
	var bbox [4]int
	if maxX >= 0 {
		bbox = [4]int{minX, minY, maxX, maxY}
	} else {
		bbox = [4]int{0, 0, 0, 0}
	}
	return sim, bbox
}

func TestGGPixelParity(t *testing.T) {
	if len(Scenes) != 16 {
		t.Fatalf("scene集合必须精确等于16, 实际 %d: %v", len(Scenes), Scenes)
	}
	seen := make(map[string]bool)
	for _, s := range Scenes {
		if seen[s] {
			t.Fatalf("scene重复: %s", s)
		}
		seen[s] = true
	}
	oldSet := make(map[string]struct{})
	newSet := make(map[string]struct{})
	for _, s := range Scenes {
		oldSet[normalizeScene(s)] = struct{}{}
		newSet[normalizeScene(s)] = struct{}{}
	}
	if len(oldSet) != 16 || len(newSet) != 16 {
		t.Fatalf("old/new scene集合必须均为16, old=%d new=%d", len(oldSet), len(newSet))
	}
	for k := range oldSet {
		if _, ok := newSet[k]; !ok {
			t.Fatalf("scene集合不一致，缺失 %s", k)
		}
	}

	// 冻结 Playwright baseline 来自 feat/satori-renderer@4bda363 的 manifest
	repoRoot := filepath.Join("..", "..")
	manifestPath := filepath.Join(repoRoot, "src", "ggrender", "testdata", "visual", "baseline", "manifest.json")
	// fallback to relative when running from src/ggrender
	if _, err := os.Stat(manifestPath); err != nil {
		manifestPath = filepath.Join("testdata", "visual", "baseline", "manifest.json")
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("读取 manifest 失败 %s: %v", manifestPath, err)
	}
	var mf manifestFile
	if err := json.Unmarshal(manifestBytes, &mf); err != nil {
		t.Fatalf("解析 manifest 失败: %v", err)
	}
	manifestMap := make(map[string]struct {
		Sha256     string
		Scale      float64
		Format     string
		PixelWidth int
		PixelHeight int
		Baseline   string
	})
	for _, e := range mf.Entries {
		manifestMap[e.ID] = struct {
			Sha256     string
			Scale      float64
			Format     string
			PixelWidth int
			PixelHeight int
			Baseline   string
		}{Sha256: e.Sha256, Scale: e.Scale, Format: e.Format, PixelWidth: e.PixelWidth, PixelHeight: e.PixelHeight, Baseline: e.Baseline}
	}
	if len(manifestMap) != 16 {
		t.Fatalf("manifest entries 必须为16, 实际 %d", len(manifestMap))
	}

	baselineDir := filepath.Join(filepath.Dir(manifestPath))
	outRoot := filepath.Join(repoRoot, "tmp", "pixel-compare")
	_ = os.MkdirAll(outRoot, 0755)
	// also ensure testdata/visual/baseline exists for harness
	_ = os.MkdirAll(filepath.Join("testdata", "visual", "baseline", "images"), 0755)

	var entries []ReportEntry
	var failed []string

	for _, scene := range Scenes {
		sceneDir := filepath.Join(outRoot, scene)
		_ = os.MkdirAll(sceneDir, 0755)
		oldPath := filepath.Join(sceneDir, "old.png")
		newPath := filepath.Join(sceneDir, "new.png")
		diffPath := filepath.Join(sceneDir, "diff.png")
		heatmapPath := filepath.Join(sceneDir, "heatmap.png")

		// manifest entry
		ent, ok := manifestMap[scene]
		if !ok {
			t.Fatalf("manifest 缺少场景 %s", scene)
		}

		// 加载冻结 Playwright old baseline (JPEG)
		baselineRel := ent.Baseline // e.g. images/base.jpg
		baselineAbs := filepath.Join(baselineDir, baselineRel)
		// ensure file exists
		oldBytes, err := os.ReadFile(baselineAbs)
		if err != nil {
			t.Fatalf("读取 Playwright baseline 失败 %s (%s): %v", scene, baselineAbs, err)
		}
		// manifest hash gate
		shOld := sha256.Sum256(oldBytes)
		hashOldHex := hex.EncodeToString(shOld[:])
		if hashOldHex != ent.Sha256 {
			t.Fatalf("manifest hash gate 失败 %s: 文件 %s sha %s != manifest %s", scene, baselineAbs, hashOldHex, ent.Sha256)
		}
		// 解码 old (JPEG)
		imgOld, err := jpeg.Decode(bytes.NewReader(oldBytes))
		if err != nil {
			// fallback generic decode
			imgOld, _, err = image.Decode(bytes.NewReader(oldBytes))
			if err != nil {
				t.Fatalf("解码 old baseline %s: %v", scene, err)
			}
		}
		// 写入 out old.png 供检查（转为 PNG）
		{
			var buf bytes.Buffer
			_ = png.Encode(&buf, imgOld)
			_ = os.WriteFile(oldPath, buf.Bytes(), 0644)
		}

		// 生成 new via gg 独立
		dcNew, err := RenderGGContext(scene, nil)
		if err != nil {
			t.Fatalf("场景 %s gg渲染失败: %v", scene, err)
		}
		imgNewRaw := dcNew.Image()
		var bufNew bytes.Buffer
		_ = png.Encode(&bufNew, imgNewRaw)
		newBytes := bufNew.Bytes()
		shNew := sha256.Sum256(newBytes)
		hashNewHex := hex.EncodeToString(shNew[:])
		_ = os.WriteFile(newPath, newBytes, 0644)

		imgNew, _, err := image.Decode(bytes.NewReader(newBytes))
		if err != nil {
			t.Fatalf("解码 new %s: %v", scene, err)
		}

		bOld := imgOld.Bounds()
		bNew := imgNew.Bounds()
		if bOld.Dx() != bNew.Dx() || bOld.Dy() != bNew.Dy() {
			// 尺寸不等直接失败（诚实红灯），不再计算相似度
			t.Errorf("场景 %s 尺寸不同: old %dx%d new %dx%d (manifest pixel %dx%d)", scene, bOld.Dx(), bOld.Dy(), bNew.Dx(), bNew.Dy(), ent.PixelWidth, ent.PixelHeight)
			failed = append(failed, fmt.Sprintf("%s size mismatch old %dx%d new %dx%d", scene, bOld.Dx(), bOld.Dy(), bNew.Dx(), bNew.Dy()))
			// 仍记录 entry 以便报告
			entries = append(entries, ReportEntry{
				Scene: scene, Width: bNew.Dx(), Height: bNew.Dy(), Format: ent.Format, Scale: ent.Scale,
				HashOld: hashOldHex, HashNew: hashNewHex, Hash: hashNewHex,
				Score: 0, Similarity: 0, BBox: [4]int{0, 0, 0, 0}, Passed: false,
				OldPath: filepath.ToSlash(oldPath), NewPath: filepath.ToSlash(newPath), DiffPath: filepath.ToSlash(diffPath),
			})
			continue
		}
		w, h := bOld.Dx(), bOld.Dy()
		rgbaOld := imageToRGBA(imgOld)
		rgbaNew := imageToRGBA(imgNew)
		sim, bbox := similarityNormalized(rgbaOld, rgbaNew)

		// diff/heatmap: 相同像素半透明，差异红
		diffImg := image.NewRGBA(bOld)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				co := rgbaOld.RGBAAt(x, y)
				cn := rgbaNew.RGBAAt(x, y)
				if co == cn {
					diffImg.SetRGBA(x, y, color.RGBA{R: co.R, G: co.G, B: co.B, A: 60})
				} else {
					diffImg.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				}
			}
		}
		var diffBuf bytes.Buffer
		_ = png.Encode(&diffBuf, diffImg)
		_ = os.WriteFile(diffPath, diffBuf.Bytes(), 0644)
		_ = os.WriteFile(heatmapPath, diffBuf.Bytes(), 0644)

		passed := sim >= 0.99
		entry := ReportEntry{
			Scene: scene, Width: w, Height: h, Format: ent.Format, Scale: ent.Scale,
			HashOld: hashOldHex, HashNew: hashNewHex, Hash: hashNewHex,
			Score: sim, Similarity: sim, BBox: bbox, Passed: passed,
			OldPath: filepath.ToSlash(oldPath), NewPath: filepath.ToSlash(newPath), DiffPath: filepath.ToSlash(diffPath),
		}
		entries = append(entries, entry)
		if !passed {
			failed = append(failed, fmt.Sprintf("%s %.5f <0.99 bbox=%v", scene, sim, bbox))
		}
		t.Logf("scene %-12s %dx%d scale=%.1f similarity=%.5f bbox=%v hashOld=%s hashNew=%s passed=%v", scene, w, h, ent.Scale, sim, bbox, hashOldHex[:12], hashNewHex[:12], passed)
	}

	reportJSONPath := filepath.Join(outRoot, "report.json")
	jb, _ := json.MarshalIndent(entries, "", "  ")
	_ = os.WriteFile(reportJSONPath, jb, 0644)

	reportMdPath := filepath.Join(outRoot, "report.md")
	var md bytes.Buffer
	md.WriteString("# Pixel Parity Report (gg vs Playwright frozen baseline)\n\n")
	md.WriteString(fmt.Sprintf("Scenes: %d (manifest: %s)\n\n", len(entries), manifestPath))
	md.WriteString("| scene | WxH | scale | format | similarity | hashOld | hashNew | bbox | passed |\n")
	md.WriteString("|-------|-----|-------|--------|------------|---------|---------|------|--------|\n")
	for _, e := range entries {
		md.WriteString(fmt.Sprintf("| %s | %dx%d | %.1f | %s | %.5f | %s | %s | %v | %v |\n",
			e.Scene, e.Width, e.Height, e.Scale, e.Format, e.Score, e.HashOld[:12], e.HashNew[:12], e.BBox, e.Passed))
	}
	if len(failed) > 0 {
		md.WriteString("\n## Failed (honest red)\n\n")
		for _, f := range failed {
			md.WriteString("- " + f + "\n")
		}
	} else {
		md.WriteString("\nAll 16 scenes PASSED (>=0.99).\n")
	}
	_ = os.WriteFile(reportMdPath, md.Bytes(), 0644)

	if len(failed) > 0 {
		for _, f := range failed {
			t.Logf("FAIL %s", f)
		}
		// 诚实红灯：虽未达标但 harness 已证实非伪 1.0，仍让测试失败以提示后续优化
		t.Fatalf("像素相似度未达标 (honest red): %d/%d 失败: %v", len(failed), len(entries), failed)
	}
}

// TestGGPixelParity_Negative 故意扰动 new 使相似度 <0.99，证明 harness 非伪 1.0
func TestGGPixelParity_Negative(t *testing.T) {
	scene := "box"
	dc, err := RenderGGContext(scene, nil)
	if err != nil {
		t.Fatalf("render %s: %v", scene, err)
	}
	img := dc.Image()
	// 先深拷贝原图：imageToRGBA 对 *image.RGBA 是别名快路径，直接用会与被扰动图像共享像素
	b := img.Bounds()
	orig := image.NewRGBA(b)
	draw.Draw(orig, b, img, b.Min, draw.Src)
	rgba := imageToRGBA(img)
	w, h := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	// 扰动：左上角 w/3 x h/3 区域填充纯红（区域随画布尺寸成比例，保证任意画布尺寸下扰动幅度足够）
	pw, ph := w/3, h/3
	if pw < 50 {
		pw = 50
	}
	if ph < 50 {
		ph = 50
	}
	for y := 0; y < ph && y < h; y++ {
		for x := 0; x < pw && x < w; x++ {
			rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	// 与原图对比应 <0.99
	sim, _ := similarityNormalized(orig, rgba)
	if sim >= 0.99 {
		t.Fatalf("负向测试失败: 扰动后相似度仍 %.5f >=0.99, harness 可能伪 1.0", sim)
	}
	t.Logf("negative test passed: perturbed %s similarity=%.5f <0.99 (honest harness)", scene, sim)
}
