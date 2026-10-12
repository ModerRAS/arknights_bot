# depot pixel-parity evidence

- measured commit: `4359da1879d0f05b710ab2bbb2aac0866b128f2f`
- parent commit: `1b433bc17b453523c7bb11ce23eb29a89f644064`
- branch `feat/gg-render-depot`, worktree `C:/WorkSpace/Golang/arknights_bot-gg-depot`
- command: `cd src && go test ./ggrender/ -run 'TestGGPixelParity$' -v -count=1 -timeout 90m`
- gate threshold source: `src/ggrender/pixel_test.go` **`gatePassed()`** (`sim < 0.99` fails). No line numbers.
- gate: **2/16 passed**

## Result

| scene | before | after | delta | render changed |
|---|---|---|---|---|
| base | 0.95900 | 0.95900 | +0.00000 | no |
| box | 0.82504 | 0.82504 | +0.00000 | no |
| box-detail | 0.87844 | 0.87844 | +0.00000 | no |
| box-summary | 0.88098 | 0.88098 | +0.00000 | no |
| calendar | 0.99432 | 0.99432 | +0.00000 | no |
| card | 0.96015 | 0.96015 | +0.00000 | no |
| depot | 0.91728 | 0.92352 | +0.00624 | YES |
| enemy | 0.91951 | 0.91951 | +0.00000 | no |
| gacha | 0.90251 | 0.90251 | +0.00000 | no |
| headhunt | 0.96476 | 0.96476 | +0.00000 | no |
| help | 0.83943 | 0.83943 | +0.00000 | no |
| lottery | 0.98358 | 0.98358 | +0.00000 | no |
| missing | 0.87614 | 0.87614 | +0.00000 | no |
| operator | 0.84554 | 0.84554 | +0.00000 | no |
| recruit | 0.61351 | 0.61351 | +0.00000 | no |
| state | 0.99093 | 0.99093 | +0.00000 | no |

Only **depot** changed; the other fifteen are byte-identical (hashNew unchanged).

## The score is not the completion criterion

The score did **not** fall; it rose +0.00624. What changed is that the false-friend
signature decayed:

| metric | stub | implemented |
|---|---|---|
| non-background coverage | 0.01% | **5.66%** |
| flat background share | 100.0% | **92.8%** |
| content overlap vs baseline | 0.0% | **16.3%** |
| item columns drawn | 0 | 2 |

## New asset: assets/depot/lmd.png

| | |
|---|---|
| sha256 | `18ab78256c635cf090f9d82893929067bf1f6f6f19be76cf2ee68622858935c6` |
| matches resource-manifest sha256 | True |
| origin | resource-manifest.json targets:["Depot"] -> 道具_带框_龙门币.png 75x75, captured at cache/depot-lmd.png |
| nature | captured third-party asset, NOT a rendered baseline screenshot |

Copied into `assets/` and referenced via `AssetPath` like every other scene. No render path
reads anything under `testdata/visual/` any more: pointing a committed `.go` at the baseline
directory is the exact shape of the `8cbffef` cheat, and it reads identically to cheating even
when the file reached for is legitimate.

## Template-declared coverage

`template/Depot.tmpl` sha256 `c66c13d3d3ff3d7dbbfd03b2f71e4c921047bb5d05bc9ad67bea4809218e37db`

| id | origin_source | declaration | implemented |
|---|---|---|---|
| main-bg | css | `#main { background-color: #2e3031 }` | yes |
| main-width | css | `#main { width: 850px }` | yes |
| item-display | css | `.item { display: inline-flex }` | yes |
| item-flexdir | css | `.item { flex-direction: column }` | yes |
| item-align | css | `.item { align-items: center }` | yes |
| item-width | css | `.item { width: 80px }` | yes |
| icon-width | css | `.icon { width: 75px }` | yes |
| icon-src | template-declared | `<img class="icon" src="{{.Icon}}">` | yes |
| count-pos | css | `.count { position: absolute }` | yes |
| count-color | css | `.count { color: white }` | yes |
| count-bg | css | `.count { background-color: rgba(0,0,0,0.5) }` | yes |
| count-size | css | `.count { font-size: 12px }` | yes |
| count-top | css | `.count { margin-top: 50px }` | yes |
| count-right | css | `.count { margin-right: -30px }` | yes |
| count-text | template-declared | `<div class="count">{{.Count}}</div>` | yes |
| range | template-declared | `{{range .}} ... {{end}}` | yes |
| main-height | manifest-dom-bbox | `manifest bbox.height = 156 (NOT declared in CSS)` | yes |
| scale | manifest | `manifest scale 1.5 / pixel 1275x234` | yes |
| row-pitch | derived-from-css | `.count is abs-pos, so item height = icon height` | yes |
| icon-inset-x | derived-from-css | `align-items:center, item 80 - icon 75` | yes |

**20/20 declared elements implemented (100%).** No value came from the baseline image.

## Known gap: item count 2 vs 11

The frozen baseline shows **11** items, all rendering `100000`; this fixture carries **2**.

1. The baseline was captured from the external `isolated-template-minimal` fixture service,
   which is not in this repository (`fixtures.json` -> `fixtureService`; `legacyRuntime`
   points at external playwright-go / playwright 1.60.0).
2. `src/core/web/depot.go` renders any count >= 10000 as `万`, so a real capture can never
   render a literal `100000`. The baseline does. The data is therefore synthetic.

The missing values were deliberately **not** back-filled by reading the baseline image.

## Honesty guards re-run

- `TestGGPixelParity_Negative` PASS (perturbed box 0.96251)
- `TestGGPixelParity_Negative_ThresholdGate` PASS (0.98999 reject / 0.99 accept)
- `TestGGPixelParity_Negative_EmptyBBoxGate` PASS (sim=1.0 bbox=[-1,-1,-1,-1] REJECTED)

## Provenance

No baseline image is committed in this directory; `report.json` carries hashes and paths only.
A prior abandoned attempt (`stash@{4}` / `18eeabd`) contained both this data and a cheat that
returned the frozen baseline as the render; preserved under `salvage/stash-4-gg-depot-temp-checkout`
as evidence and deliberately not reused.
