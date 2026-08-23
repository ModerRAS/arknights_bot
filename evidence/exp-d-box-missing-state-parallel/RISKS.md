# PR Risks — Exp-D Box/Missing/State Parallel

**Branch:** `exp-d-box-missing-state-parallel` (6 commits, baseline 53a7b3d)
**Scope:** Box/State/Infra only, 11 files, no calendar/lottery/base/enemy/card/operator changes.

**Risks & Mitigations:**

1. **Asset manifest freeze (26 resources):**
   - Risk: Manifest mismatch or hash drift breaks frozen cache, network fallback not allowed (`ASSET_MANIFEST_MISSING` manifestFatal).
   - Mitigation: `loadManifest` validates `status frozen`, `length 26`, sha256/bytes, MIME/magic, cache path escape, and fails closed without network. Covered by `manifest-webp.test.mjs` 7 tests (hash/path/missing/http downgrade closed).

2. **Fixture aliases (8):**
   - Risk: Alias target missing or conflicting content causes `ASSET_MANIFEST_INVALID`.
   - Mitigation: Aliases injected from `cacheEntries` with provenance `frozen-manifest-fixture-alias`, validated at load time; Card/Operator fixture tests hit without fetch.

3. **Materialize SHA & diagnostics:**
   - Risk: SHA mismatch or provenance loss causes silent cache corruption.
   - Mitigation: `materialize` returns `{value, sha256}` hashed from final bytes (PNG after WebP normalize), `materializations` array and `stats` expose `materializedSha256`, `onDiagnostic` emits `asset_materialized`/`asset_fallback` with SHA.

4. **Box rarity width ladder:**
   - Risk: Fixed width 40 overflows for varying rarity, breaks top text/horizontal line alignment.
   - Mitigation: Ladder `[15,20,25,30,34,40][rarity]??40` matches legacy template; VNode geometry test `box-detail shares five fixed tracks` ensures width stability.

5. **BoxDetail fixed geometry:**
   - Risk: `space-around` causes drift, 3-col geometry not fixed.
   - Mitigation: `tracks=[110,62,58,150,100]` shared header/body, no `justifyContent space-around`, verified by `static-pages` test.

6. **State clocks & single-line layout:**
   - Risk: Missing clocks or line wrap breaks legacy fidelity.
   - Mitigation: 4×16×16 SVG clocks retained, `flowMeter`/`fixedMeter` single-line, campaign left 931 (was 925), training width 130 (was 134) verified via grep and template State.tmpl.

7. **Visual regression:**
   - Risk: Other 11 pages not in scope could be <0.99.
   - Mitigation: This branch reports those as out-of-scope; other Leads own those pages. Our 5 pages report `alignedSimilarity 1.0` via VNode tests; full harness fallback notes threshold 0.99.

**Rollback:** Revert 6 commits on this branch; evidence branch retains isolated candidates. No DB migration.

**Isolation:** Sibling worktree + external artifact dir guarantees no cross-Lead file conflicts; forbidden files untouched verified via `git diff --name-only`.
