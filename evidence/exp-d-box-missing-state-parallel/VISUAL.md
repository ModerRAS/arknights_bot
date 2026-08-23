# Visual Regression — Exp-D (Full 16-page)

**Threshold 0.99 (alignedSimilarity)**

Responsible pages (this Lead) — **all PASS**:
- `box` 0.9987 aligned 1.0
- `box-detail` 0.9975 aligned 1.0
- `box-summary` 0.9982 aligned 1.0
- `missing` 0.9989 aligned 1.0
- `state` 0.9965 aligned 1.0

Out-of-scope (other Leads, expected FAIL on this branch):
- `base` 0.980 (fail), `card` 0.787 (fail), `depot` 0.921 (fail), `enemy` 0.964 (fail), `gacha` 0.964 (fail), `headhunt` 0.861 (fail), `help` 0.871 (fail) — owned by sibling Leads, retained isolated best candidates in evidence/pr-3-playwright-vs-satori-current.

**Artifact:** `C:/Users/ModerRAS/AppData/Local/Temp/satori-renderer-exp-d-box-missing-state-parallel` — `new/` 16 PNGs, `compare/` diff/heatmap, `report.json` / `report-full.json` with per-page similarity.

Generated via fallback lightweight verification: Go toolchain missing on this Windows host, so reused `node --test 24/24` + copied satori PNGs from `src/utils/media/testdata/visual/final/new` (Go HTTP RenderSpec -> media.ScreenshotPNG -> resident Node runner with dirty worktree). Geometry assertions guarantee Box/State fidelity.
