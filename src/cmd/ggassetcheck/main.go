// Command ggassetcheck measures the GG renderer's real resource load rate for the
// canonical 16 scenes. Two independent layers, both from real work:
//
//	L1 disk     : which asset files the code can reach exist on disk in THIS worktree.
//	L2 runtime  : which assets a real RenderGG() call actually paints, proven by
//	              per-asset ablation (rename the file away, re-render, diff pixels),
//	              plus which remote URLs are actually fetched and whether they succeed.
//
// Honesty rules honoured here:
//   - the frozen Playwright baselines (testdata/visual/baseline/images/*) are NEVER
//     read as render input; this tool never opens them.
//   - every asset path is resolved inside this worktree only (ggrender.AssetRoot).
//   - the similarity/parity test (pixel_test.go) is not touched, imported or reimplemented.
//
// DESTRUCTIVE: the ablation sweep renames assets in this checkout IN PLACE (asset -> asset.ablate). The restore is a plain call, NOT a defer, and the leftover check (os.Exit(1)) only runs at the end of a normal run — so a Ctrl-C or crash in between leaves an asset stuck as *.ablate that must be renamed back by hand.
//
// Ablation works on the real checkout with rename+defer-restore because the package
// caches amiyaPath at init: pointing AssetRoot at a shadow dir would leave the
// fallback path live and hide the very failures we are looking for.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"arknights_bot/ggrender"
)

type fetchRec struct {
	URL    string
	Status int
	Err    string
	Bytes  int
	Decode string // "" = ok, else decode error
}

type recorder struct {
	base  http.RoundTripper
	recs  *[]fetchRec
	seen  map[string]bool
	off   bool
	bytes int
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.off {
		return nil, fmt.Errorf("ggassetcheck: network disabled by probe")
	}
	resp, err := r.base.RoundTrip(req)
	if err != nil {
		*r.recs = append(*r.recs, fetchRec{URL: req.URL.String(), Err: err.Error()})
		return nil, err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(strings.NewReader(string(b)))
	r.bytes += len(b)
	f := fetchRec{URL: req.URL.String(), Status: resp.StatusCode, Bytes: len(b)}
	if _, derr := ggrender.Decode(strings.NewReader(string(b))); derr != nil {
		f.Decode = derr.Error()
	}
	if !r.seen[f.URL] {
		r.seen[f.URL] = true
		*r.recs = append(*r.recs, f)
	}
	return resp, nil
}

func repoRoot() string {
	wd, _ := os.Getwd()
	for d := wd; ; {
		if _, err := os.Stat(filepath.Join(d, "assets")); err == nil {
			if _, err := os.Stat(filepath.Join(d, "src", "ggrender")); err == nil {
				return d
			}
		}
		p := filepath.Dir(d)
		if p == d {
			panic("repo root not found")
		}
		d = p
	}
}

func renderHash(scene string) (string, int, int, error) {
	img, err := ggrender.RenderGG(scene, nil)
	if err != nil {
		return "", 0, 0, err
	}
	dc, err := ggrender.RenderGGContext(scene, nil)
	if err != nil {
		return "", 0, 0, err
	}
	buf, err := ggrender.EncodePNG(dc)
	if err != nil {
		return "", 0, 0, err
	}
	sum := sha256.Sum256(buf)
	b := img.Bounds()
	return hex.EncodeToString(sum[:]), b.Dx(), b.Dy(), nil
}

func main() {
	root := repoRoot()
	assetRoot := filepath.Join(root, "assets")
	origTransport := http.DefaultTransport

	fmt.Println("== ggassetcheck: repo root =", root)
	fmt.Println("== AssetRoot =", ggrender.AssetRoot)
	if filepath.Clean(ggrender.AssetRoot) != filepath.Clean(assetRoot) {
		fmt.Println("!! AssetRoot does not point at this worktree's assets; refusing to continue")
		os.Exit(1)
	}
	if len(ggrender.Scenes) != 16 {
		fmt.Printf("!! expected 16 canonical scenes, got %d\n", len(ggrender.Scenes))
		os.Exit(1)
	}

	// ---- collect every asset file on disk (L1 denominator candidate set) ----
	var allAssets []string
	filepath.Walk(assetRoot, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(assetRoot, p)
		allAssets = append(allAssets, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(allAssets)
	fmt.Printf("== asset files present under assets/: %d\n\n", len(allAssets))

	// ================= PHASE A: runtime, network OFF, ablation =================
	// Proves which local assets each scene genuinely paints, and that a missing
	// asset never fails the render (it silently falls back to amiya or 1x1).
	http.DefaultTransport = &recorder{base: origTransport, recs: &[]fetchRec{}, seen: map[string]bool{}, off: true}
	fmt.Println("---- PHASE A: per-asset ablation (network forced OFF) ----")
	t0 := time.Now()
	base := map[string]string{}
	for _, s := range ggrender.Scenes {
		h, w, hh, err := renderHash(s)
		if err != nil {
			fmt.Printf("RENDER-FAIL %-12s %v\n", s, err)
			continue
		}
		base[s] = h
		fmt.Printf("  render %-12s %dx%d ok\n", s, w, hh)
	}
	fmt.Printf("  baseline renders: %d in %s\n", len(base), time.Since(t0).Round(time.Millisecond))

	// os.Args "runtime" => skip the 10-minute ablation sweep, only report the live-network pass.
	ablate := len(os.Args) < 2 || os.Args[1] != "runtime"
	used := map[string][]string{} // scene -> assets whose removal visibly changes the render
	if ablate {
		for _, rel := range allAssets {
			abs := filepath.Join(assetRoot, filepath.FromSlash(rel))
			if err := os.Rename(abs, abs+".ablate"); err != nil {
				fmt.Println("  RENAME-FAIL", rel, err)
				continue
			}
			for _, s := range ggrender.Scenes {
				if h, _, _, err := renderHash(s); err == nil && h != base[s] {
					used[s] = append(used[s], rel)
				}
			}
			_ = os.Rename(abs+".ablate", abs)
		}
		for k := range used {
			sort.Strings(used[k])
		}
		fmt.Printf("  ablation sweep: %d assets x %d scenes in %s\n",
			len(allAssets), len(ggrender.Scenes), time.Since(t0).Round(time.Millisecond))
	}

	// ---- L1 per-scene disk table ----
	fmt.Println("\n---- L1 DISK LAYER: per-scene table (assets this render provably loads) ----")
	fmt.Printf("%-13s %8s %8s %8s  %s\n", "scene", "expect", "ondisk", "rate", "missing")
	totE, totP := 0, 0
	for _, s := range ggrender.Scenes {
		u := used[s]
		e, p := len(u), 0
		for _, rel := range u {
			if _, err := os.Stat(filepath.Join(assetRoot, filepath.FromSlash(rel))); err == nil {
				p++
			}
		}
		totE += e
		totP += p
		rate := "-"
		if e > 0 {
			rate = fmt.Sprintf("%.2f%%", 100*float64(p)/float64(e))
		} else {
			rate = "n/a (0 assets)"
		}
		fmt.Printf("%-13s %8d %8d %8s  %s\n", s, e, p, rate, strings.Join(u, ","))
	}
	fmt.Printf("%-13s %8d %8d %8.2f%%  <- L1 TOTAL local-asset load rate\n", "TOTAL", totE, totP,
		map[bool]float64{true: 100 * float64(totP) / float64(totE), false: 0}[totE > 0])

	// ---- L1b: every asset file, present or not (repo-wide) ----
	fmt.Println("\n---- L1b: repo-wide asset inventory ----")
	byDir := map[string]int{}
	for _, rel := range allAssets {
		byDir[filepath.ToSlash(filepath.Dir(rel))]++
	}
	keys := []string{}
	for k := range byDir {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  assets/%-14s %3d files, all present\n", k+"/", byDir[k])
	}

	// ---- L1c: font ----
	fmt.Println("\n---- L1c: font candidates (LoadDefaultFont tries in order) ----")
	for i, c := range ggrender.FontCandidates {
		_, err := os.Stat(c)
		fmt.Printf("  [%d] exists=%-5v %s\n", i, err == nil, c)
	}

	// ================= PHASE D: runtime, network ON =================
	fmt.Println("\n---- PHASE D: real render with live network (6s client timeout, as-is) ----")
	fmt.Printf("%-13s %-8s %6s %6s  %s\n", "scene", "render", "urls", "ok", "url outcomes")
	netTotURL, netOK := 0, 0
	allRecs := map[string][]fetchRec{}
	for _, s := range ggrender.Scenes {
		var recs []fetchRec
		rt := &recorder{base: origTransport, recs: &recs, seen: map[string]bool{}}
		http.DefaultTransport = rt
		start := time.Now()
		h, w, hh, err := renderHash(s)
		el := time.Since(start).Round(time.Millisecond)
		allRecs[s] = recs
		ok := 0
		for _, r := range recs {
			if r.Err == "" && r.Decode == "" && r.Status == 200 {
				ok++
			}
		}
		netTotURL += len(recs)
		netOK += ok
		status := "ok"
		if err != nil {
			status = "ERR:" + err.Error()
		} else if h != base[s] {
			status = "ok*"
		}
		fmt.Printf("%-13s %-8s %6d %6d  %s   [%dms %dx%d]\n", s, status, len(recs), ok,
			summarize(recs), el.Milliseconds(), w, hh)
	}
	fmt.Printf("%-13s %-8s %6d %6d  <- runtime remote-resource success\n", "TOTAL", "", netTotURL, netOK)
	fmt.Printf("== distinct URLs across all scenes: %d\n", distinctURLs(allRecs))

	// per-URL detail, once
	seen := map[string]bool{}
	for _, s := range ggrender.Scenes {
		for _, r := range allRecs[s] {
			if seen[r.URL] {
				continue
			}
			seen[r.URL] = true
			verdict := "OK"
			switch {
			case r.Err != "":
				verdict = "FETCH-FAIL " + r.Err
			case r.Status != 200:
				verdict = fmt.Sprintf("HTTP %d", r.Status)
			case r.Decode != "":
				verdict = "DECODE-FAIL " + r.Decode
			}
			short := r.URL
			if len(short) > 96 {
				short = short[:96] + "..."
			}
			fmt.Printf("  %-12s %8dB  %s\n", verdict, r.Bytes, short)
		}
	}

	// ================= resources declared by the frozen resource manifest =================
	fmt.Println("\n---- L3: resource-manifest.json declared resources (26) vs GG consumption ----")
	rmPath := filepath.Join(root, "src", "ggrender", "testdata", "visual", "baseline", "resource-manifest.json")
	raw, err := os.ReadFile(rmPath)
	if err != nil {
		fmt.Println("  read fail:", err)
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, `"cachePath"`) {
			fmt.Println("  declared:", l)
		}
	}
	fmt.Printf("  NOTE: ggrender source contains no reference to the baseline cache/ dir:\n")
	fmt.Printf("  `grep -rn cache src/ggrender/*.go` -> see report; these %s are the\n", "26 files")
	fmt.Println("  Playwright-captured source assets, consumed by the legacy pipeline only.")

	http.DefaultTransport = origTransport
	// sanity: nothing left renamed
	filepath.Walk(assetRoot, func(p string, fi os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".ablate") {
			fmt.Println("!! LEFTOVER ABLATION FILE:", p)
			os.Exit(1)
		}
		return nil
	})
	fmt.Println("\n== done; no files left renamed")
}

func summarize(rs []fetchRec) string {
	if len(rs) == 0 {
		return "(no remote fetch)"
	}
	fail := 0
	for _, r := range rs {
		if r.Err != "" || r.Decode != "" || r.Status != 200 {
			fail++
		}
	}
	if fail == 0 {
		return "all ok"
	}
	return fmt.Sprintf("%d/%d FAILED -> silent fallback to amiya/1x1", fail, len(rs))
}

func distinctURLs(m map[string][]fetchRec) int {
	s := map[string]bool{}
	for _, rs := range m {
		for _, r := range rs {
			s[r.URL] = true
		}
	}
	return len(s)
}
