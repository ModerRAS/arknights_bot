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
	"os"
	"path/filepath"
	"strings"
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
		ID          string  `json:"id"`
		Baseline    string  `json:"baseline"`
		Sha256      string  `json:"sha256"`
		Scale       float64 `json:"scale"`
		Format      string  `json:"format"`
		PixelWidth  int     `json:"pixelWidth"`
		PixelHeight int     `json:"pixelHeight"`
		BBox        struct {
			X      float64 `json:"x"`
			Y      float64 `json:"y"`
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"bbox"`
	} `json:"entries"`
}

// ---------------------------------------------------------------------------
// 为什么这里没有 manifest 树哈希校验（Task 1 结论：不可实施）
// ---------------------------------------------------------------------------
//
// 缺陷 #6 是「信任锚自证」：把冻结基线图换成自制图、同时改 manifest.json 里
// 该条目的 sha256，代码层零防线。下面记录的 hash gate 只校验
// 「baseline 文件的 sha256 == 同一条 manifest 记录里的 sha256」——两者可同改。
//
// 曾经考虑用 manifest.json 自带的 templateTreeSHA256 / assetTreeSHA256 做锚。
// **已穷举验证，该测量不可实施。** 理由有两条，第二条更根本：
//
//  1. 值不可复现。算法本身是明确且可复现的 —— 它在 8873add 引入的
//     src/cmd/visual-regression/main.go 的 treeSHA256() 里：filepath.WalkDir
//     递归收集目录下全部文件，sort.Strings 后把每行
//     `相对路径\x00<该文件内容的 hex sha256>\n` 喂进 sha256。照此实现逐 commit
//     重算 feat/satori-renderer 的**全部 473 个 commit**（tOK=473 aOK=473），
//     templateTree 出现 69 种互不相同的值、assetTree 21 种，与 manifest 里那两个
//     字面量**零交集**（0 命中）。4bda363、8873add、8873add^ 与本树 HEAD 也
//     个个对不上。那两个字面量的来源已不可考，无法写成一份能自动判定真伪的校验。
//
//  2. **即便算法可复现，也没有可对齐的参照物。** 那两个字段的根是
//     <repoRoot>/template 与 <repoRoot>/assets。template/ 两树一致（各 20 个
//     文件），但 assets/ **gg 树 102 个文件 vs satori 树 100 个文件**——文件数
//     本身就对不上，无论用什么顺序、哈希什么内容都对不齐。
//
//     另有一条更直接的观察：**这两个字段的覆盖范围根本不包含冻结基线图。**
//     它们的两个根是 template/ 与 assets/，而「基线有没有被换过」问的是
//     baseline/images/*.jpg。把这两个字面量写成常量锁住，锁的是两个与基线图无关的
//     字段，会给人虚假的安全感，而缺陷 #6 依旧畅通。故不写。
//
// **本 harness 不校验基线完整性。** 缺陷 #6 未被修复，属已记录在案的已知缺口。
// 它的防线完全在仓库之外：manifest.json 必须始终处于 git 跟踪之下，且任何改动
// 都必须经过人工审查。**这是流程纪律，不是密码学锚点。** 当前没有任何机制能
// 自动区分「合法地重抓了基线」与「把基线换成了自己画的答案」。
//
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
		// 哨兵：全画布逐像素零差异。x∈[0,w-1]、y∈[0,h-1]，所以 -1 不可能是任何
		// 真实差异像素的坐标，[4]int{-1,-1,-1,-1} 与「差异恰好落在单个像素上的
		// [x0,y0,x0,y0]」在数值上不可能撞值（旧的 [0,0,0,0] 会与位于 (0,0) 的
		// 单像素差异撞成同一个值）。这里只改返回值的附带信息：sum/total/sim 的
		// 计算本体一行未动，任何场景的分数不变。
		bbox = zeroDiffBBox
	}
	return sim, bbox
}

// zeroDiffBBox 是「全画布零差异」的哨兵值，永不可能与任何真实 diff bbox 混淆。
var zeroDiffBBox = [4]int{-1, -1, -1, -1}

// gatePassed 是唯一的门禁判定点。相似度达标还不够：diff bbox 必须非空。
//
// bbox 是**闭区间**像素坐标 [x0 y0 x1 y1]（不是 [x y w h]），所以 bbox[2]-bbox[0]
// 是「跨度减一」而不是宽度。零跨度的情形有两种，必须分开说，且两种都判失败
// （fail-closed）——但诊断信息不能再说假话：
//
//   - 零差异：similarityNormalized 走哨兵分支返回 [-1 -1 -1 -1]。
//   - 单像素差异：返回 [x0 y0 x0 y0]，跨度为 0。
//
// 无论哪一种，similarityNormalized 都**已经把整个画布逐像素比较完了**
// （外层循环恒定扫满 w*h）。旧文案「no pixel actually compared」在这两种情形下
// 都是事实错误。
func gatePassed(sim float64, bbox [4]int) (bool, string) {
	if bbox == zeroDiffBBox {
		return false, fmt.Sprintf("zero-diff bbox=%v (identical images: 0 differing pixels; the full canvas WAS compared; gate rejects vacuous exact-parity)", bbox)
	}
	if bbox[2]-bbox[0] <= 0 || bbox[3]-bbox[1] <= 0 {
		return false, fmt.Sprintf("degenerate diff bbox=%v (zero span: every differing pixel shares one column or one row; the full canvas WAS compared)", bbox)
	}
	if sim < 0.99 {
		return false, fmt.Sprintf("%.5f <0.99 bbox=%v", sim, bbox)
	}
	return true, ""
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
		Sha256      string
		Scale       float64
		Format      string
		PixelWidth  int
		PixelHeight int
		Baseline    string
	})
	for _, e := range mf.Entries {
		manifestMap[e.ID] = struct {
			Sha256      string
			Scale       float64
			Format      string
			PixelWidth  int
			PixelHeight int
			Baseline    string
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

		passed, reason := gatePassed(sim, bbox)
		entry := ReportEntry{
			Scene: scene, Width: w, Height: h, Format: ent.Format, Scale: ent.Scale,
			HashOld: hashOldHex, HashNew: hashNewHex, Hash: hashNewHex,
			Score: sim, Similarity: sim, BBox: bbox, Passed: passed,
			OldPath: filepath.ToSlash(oldPath), NewPath: filepath.ToSlash(newPath), DiffPath: filepath.ToSlash(diffPath),
		}
		entries = append(entries, entry)
		if !passed {
			failed = append(failed, fmt.Sprintf("%s %s", scene, reason))
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

// TestGGPixelParity_Negative_ThresholdGate 锁死 0.99 阈值（Task 2）。
//
// 覆盖度缺口：TestGGPixelParity_Negative 只调 similarityNormalized、不调 gatePassed，
// 所以把 gatePassed 里的 0.99 改成 0.5 那个测试照样全绿。本测试调**真实的
// gatePassed**，而不是它的表达式副本，因此具备变异检测能力。
func TestGGPixelParity_Negative_ThresholdGate(t *testing.T) {
	// 非退化 bbox——否则会被零跨度守卫先拦掉，根本走不到阈值判定，测不到阈值。
	nonDegenerate := [4]int{0, 0, 63, 63}

	// 1) 介于 0.5 与 0.99 之间的分数必须被拒。这些值全部低于任何合理的阈值，
	//    把 0.99 改小到其中任何一个之上就会让本组断言变红。
	for _, sim := range []float64{0.50, 0.60, 0.75, 0.90, 0.95, 0.98, 0.9899} {
		passed, reason := gatePassed(sim, nonDegenerate)
		if passed {
			t.Fatalf("阈值被放宽: sim=%.4f < 0.99 且 bbox 非退化, gatePassed 却返回 true", sim)
		}
		if !strings.Contains(reason, "<0.99") {
			t.Fatalf("sim=%.4f 的失败原因须写明 <0.99, 实际 %q", sim, reason)
		}
	}

	// 2) 边界：恰好 0.99 放行（门禁语义是 sim >= 0.99），再低一丝就拒。
	if passed, reason := gatePassed(0.99, nonDegenerate); !passed {
		t.Fatalf("sim=0.99 且 bbox 非退化应放行, 实际拒绝: %s", reason)
	}
	if passed, _ := gatePassed(0.98999, nonDegenerate); passed {
		t.Fatalf("阈值被放宽: sim=0.98999 低于 0.99, gatePassed 却返回 true")
	}

	// 3) 不对「高分一定被拒」下断言。门禁语义就是 sim >= 0.99 放行，所以 0.9995
	//    会被放行——那不是 bug，是已记录的残余风险 #7（以基线作底 + 少量扰动，
	//    预算约 1.33% 画布可全错仍过）。若断言 0.9995 必须被拒，是把错误信念写进
	//    测试；若反过来断言它必须被放行，日后有人合法收紧阈值反而会红。两者皆不可取，
	//    故此处不设断言，只留注释备查。
	t.Logf("0.99 threshold locked via real gatePassed: 0.98999->reject, 0.99->accept, 0.9995->accept (known risk #7 budget, deliberately not asserted)")
}

// TestGGPixelParity_Negative_EmptyBBoxGate 零跨度 bbox 不得通过门禁，
// 并锁住「零差异」与「单像素差异」的区分（Task 3）。
func TestGGPixelParity_Negative_EmptyBBoxGate(t *testing.T) {
	// 两张完全相同的图：similarityNormalized 走 maxX<0 分支，产出哨兵 bbox
	// 且 sim=1.0。这是 8873add 以来最典型的伪通过场景。
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 7, A: 255})
		}
	}
	sim, bbox := similarityNormalized(img, img)
	if sim < 0.99 {
		t.Fatalf("前置条件不成立: 相同图 sim=%.5f 应为 1.0", sim)
	}
	if bbox != zeroDiffBBox {
		t.Fatalf("零差异必须走哨兵 bbox, 实际 %v", bbox)
	}
	passed, reason := gatePassed(sim, bbox)
	if passed {
		t.Fatalf("门禁伪通过: 零差异 sim=%.5f bbox=%v 不应通过", sim, bbox)
	}
	if !strings.Contains(reason, "zero-diff") {
		t.Fatalf("失败信息须写明 zero-diff, 实际 %q", reason)
	}
	// 旧文案「no pixel actually compared」是事实错误：画布每个像素都比较过了。
	if strings.Contains(reason, "no pixel actually compared") {
		t.Fatalf("失败信息不得再声称「没有像素被比较」, 实际 %q", reason)
	}

	// Task 3 核心：单像素差异必须与零差异**可区分**，且同样被拒（行为未放宽）。
	// 差异点刻意放在 (37, 41)——旧哨兵 [0 0 0 0] 与位于 (0,0) 的单像素差异会撞值。
	one := image.NewRGBA(img.Bounds())
	draw.Draw(one, one.Bounds(), img, one.Bounds().Min, draw.Src)
	one.SetRGBA(37, 41, color.RGBA{R: 1, G: 2, B: 3, A: 4})
	simOne, bboxOne := similarityNormalized(img, one)
	if bboxOne != [4]int{37, 41, 37, 41} {
		t.Fatalf("单像素差异的 bbox 应为 [37 41 37 41], 实际 %v", bboxOne)
	}
	if bboxOne == zeroDiffBBox {
		t.Fatalf("单像素差异与零差异不可区分, 哨兵失效: %v", bboxOne)
	}
	if simOne >= 1.0 {
		t.Fatalf("前置条件不成立: 单像素差异 sim=%.5f 应 < 1.0", simOne)
	}
	passedOne, reasonOne := gatePassed(simOne, bboxOne)
	if passedOne {
		t.Fatalf("门禁伪通过: 单像素差异 bbox=%v 不应通过", bboxOne)
	}
	if !strings.Contains(reasonOne, "degenerate") {
		t.Fatalf("单像素差异须报 degenerate 而非 zero-diff, 实际 %q", reasonOne)
	}
	if strings.Contains(reasonOne, "no pixel actually compared") {
		t.Fatalf("单像素差异的失败信息不得声称「没有像素被比较」, 实际 %q", reasonOne)
	}

	// 宽或高单独为 0 同样要拦
	for _, b := range [][4]int{{10, 10, 10, 40}, {10, 10, 40, 10}} {
		if ok, _ := gatePassed(1.0, b); ok {
			t.Fatalf("门禁伪通过: 零跨度 bbox=%v 不应通过", b)
		}
	}
	// 反向：非空 bbox 且分数达标不应被守卫误伤
	if ok, _ := gatePassed(0.995, [4]int{0, 0, 10, 10}); !ok {
		t.Fatalf("非空 bbox 且分数达标时不应被守卫拦截")
	}
	t.Logf("zero-diff rejected: sim=%.5f bbox=%v reason=%s", sim, bbox, reason)
	t.Logf("single-pixel diff rejected and distinguishable: sim=%.5f bbox=%v reason=%s", simOne, bboxOne, reasonOne)
}
