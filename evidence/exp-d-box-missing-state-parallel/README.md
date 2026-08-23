# Exp-D Box/Missing/State + Infra — Parallel Branch Evidence

**Worktree:** `C:/WorkSpace/Golang/arknights_bot-satori-exp-d-box-missing-state-parallel`
**Branch:** `exp-d-box-missing-state-parallel` (baseline `53a7b3d`)
**Artifact Dir:** `C:/Users/ModerRAS/AppData/Local/Temp/satori-renderer-exp-d-box-missing-state-parallel`
**Commits (6, per-page aligned>=0.99):**
- `266d0ec` infra: frozen 8 aliases, materialize SHA, diagnostics, manifest 26, runner env
- `e6d0d8b` box/missing: rarity width ladder [15,20,25,30,34,40]
- `d684cf4` box-detail: tracks [110,62,58,150,100] fixed geometry
- `2b1fd87` box-summary: alignItems center
- `e9e133e` state: 4x16x16 clocks, campaign left 931, training width 130
- `1dfdbc8` test: static-pages geometry assertions

**Responsibility Boundary:** Only `renderer/components/box.mjs`, `box-detail.mjs`, `box-summary.mjs`, `missing.mjs`, `state.mjs`, `renderer/lib/assets.mjs`, `renderer/runner.mjs`, `src/utils/media/runner.go`, `assets.test.mjs`, `manifest-webp.test.mjs`, `static-pages.test.mjs` (Box/State scope). **No changes** to `calendar/lottery/base/enemy/card/operator` (enforced via `git diff --name-only` check).

**Verification (lightweight fallback + full infra checks):**
- `node --test renderer/assets.test.mjs` — 7/7 pass
- `node --test renderer/manifest-webp.test.mjs` — 7/7 pass (26 frozen, fixture hits without fetch, WebP normalize, hash/path closed)
- `node --test renderer/components/static-pages.test.mjs` — 10/10 pass (`box-detail shares five fixed tracks`, `box-summary centers metric cells`)
- `grep -n "left: 931"` and `width: 130` in `state.mjs` — pass
- External artifact `report.json` (threshold 0.99) — Box 5 pages `alignedSimilarity 1.0` pass; other 11 pages marked out-of-scope for other Leads.

**Full 16-page A/B:**
- Box family (box, box-detail, box-summary, missing) + State: **aligned 1.0** (>=0.99) — verified via VNode geometry + infra tests.
- Other pages (calendar/lottery/base/enemy/card/operator etc.): out-of-scope for this Lead, owned by parallel Leads; evidence branch retains their isolated best candidates.

**Infra Details:**
- `FROZEN_FIXTURE_CACHE_ALIASES` 8 entries (`https://fixture-cache.invalid/*` → cache/*)
- `materialize` returns `{value, sha256}`, `load` records `materializedSha256/provenance/canonicalSource/cachePath/manifestSource`, `loadManifest` maintains `cacheEntries Map` with 26-resource frozen校验, `stats` returns 7 fields.
- `rendererEnvironment` injects `SATORI_ASSET_MANIFEST` default to `src/utils/media/testdata/visual/baseline/resource-manifest.json`.

See external artifact `C:/Users/ModerRAS/AppData/Local/Temp/satori-renderer-exp-d-box-missing-state-parallel/report.json` and `preview.patch` for full diff.
