# Visual Final Bundle

final integration gate: lead-2 four-branch merge @ d4831b9 + portrait manifest extension 37d6f84 + JPEG q80 capture alignment f86b74e (worker-21)

Legacy replay requires `legacy-capture.mjs --out <final-dir> --root <repo-root> --baseline <repo-root>/src/utils/media/testdata/visual/baseline --clock-erratum <repo-root>/src/utils/media/testdata/visual/capture-clock-erratum.json --base-url <legacy-url> --playwright <playwright-module>`. It replays manifest entries (or `--id gacha` for diagnostics) at each manifest DPR, takes Playwright locator JPEGs at default quality, waits 3000ms for Gacha, and uses an advancing fake wall clock anchored to the erratum for Gacha ECharts. Calendar and State use their fixed erratum clocks.

Current replay requires `visual-final -spec-dir <canonical-spec-dir> -out <final-dir> -baseline <baseline-manifest> -resource-manifest <resource-manifest> -clock-erratum <capture-clock-erratum> -legacy-script <legacy-capture.mjs>`.
