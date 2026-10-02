# task1-instrument-compare — gg vs satori vs yoga 像素比较机制对照

```yaml
provenance: measured        # 本报告全部机制性结论均来自实际读取源码 + 实际执行 grep/git；
                            # 唯一 inferred 项已在正文逐条标注。
measured_at:
  generated_at: 2026-10-02T10:38:24Z (UTC)
  author: worker-1 (Pi Team, boss-1/lead-2 委派)
  # 三棵被读工作树 —— 工作树路径 / 分支 / commit 三件套
  gg:      {worktree: C:/WorkSpace/Golang/arknights_bot-gg-card-atomic,
            branch:   feat/gg-render-card-atomic,
            commit:   d939c38563cbfeccaeb941618a3ea8c558e175e9}
  satori:  {worktree: C:/WorkSpace/Golang/arknights_bot-satori,
            branch:   feat/satori-renderer,
            commit:   b9466612aac06080976e9b17d305db692fd139bc,
            note: "origin/feat/satori-renderer = eabf373；本地领先 2 提交，未推送
                   (git merge-base --is-ancestor b946661 origin/feat/satori-renderer exit=1)。
                   git merge-base --is-ancestor eb55d5a b946661 exit=0 → eb55d5a 为父。"}
  yoga:    {worktree: C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go,   # 禁碰树，只读，零改动
            branch:   feat/yoga-skia-go-renderer,
            commit:   370426246b500e9349a09e898a449dfe4d176ede}
original_path: C:/WorkSpace/Golang/.lead2-reports/task1-instrument-compare.md
scope_note: 全部否定性结论的枚举范围 = `git worktree list` 输出的全部 22 棵工作树
            （实测 `git worktree list | wc -l` = 22；不是 21），glob 见 §6 取证附录。
mutation: 零 commit / 零 push / 零门禁代码改动 / 零工作树写入。
```

---

## §1 表 A — 机制可移植

> 判据（读代码得出，非推断）：加进来**只需新增输出或新增字段**，不碰 `similarityNormalized` 的分母、不碰 `sim < 0.99`、不碰零面积 bbox 守卫。
>
> **两栏必须分开读**（Boss 边界 1）：**出处栏**说这东西从哪来、是不是事后归档；**可用性栏**说它能不能用。出处是 `retroactive-migration` **不推出**「没用」——按本项目「按有没有诊断分档挑实现目标」的方法，诊断指标的合法角色是**挑目标**，不是**判过线**。

| # | 机制 | gg | satori | yoga | **可用性**（诊断层 / 门禁层） | **出处**（归档性质 + 函数 + 工作树/分支/commit） |
|---|---|---|---|---|---|---|
| A1 | **真·逐通道量值 heatmap**（`peak = max(dr,dg,db,da)`，红色强度编码幅度） | **无** | 有 | 有 | 诊断 **可** / 门禁 **不适用**（产物，不是判据） | `diffImagePair()` @ yoga `3704262` + satori `b946661` |
| A2 | **diff 连通域分解**（flood fill 4 邻接，阈值 `componentDeltaThreshold=32`；产出 bbox / 像素数 / meanDelta） | 无 | 有 | 有 | 诊断 **可** / 门禁 **不适用** | `diffComponents()` @ yoga `3704262` + satori `b946661` |
| A3 | **连通域合并**（`componentMergePadding=48`；并集 bbox + 像素累加） | 无 | 有 | 有 | 诊断 **可** / 门禁 **不适用** | `mergeComponents()` @ yoga `3704262` + satori `b946661` |
| A4 | **微噪声分类器**（pixels≤64 且 ≤8×8 且 meanDelta<64） | 无 | 有 | 有 | **诊断可**（标注哪些连通域被判为噪声）；**进门禁即不可** → 见 B4 | `microNoise()` @ yoga `3704262` + satori `b946661` |
| A5 | **alignment 作为诊断**：生成 aligned diff / heatmap、量残差分布与全局偏移量 | 无 | 有 | 有 | **诊断层合法可移植**（只出图与偏移量，不动 `gatePassed` 路径）→ 评分层见 **B1** | `registerOffset()` / `selectFinalOffset()` / `translateImage()` @ yoga `3704262` + satori `b946661` |
| A6 | **局部配准诊断**（每个连通域 ±8px 搜位移，报 `improvement`，阈值 0.15） | 无 | 有 | 有 | 诊断 **可**（回答「哪些物体整体错位」） / 进门禁见 **B4** | `localRegistration()` @ yoga `3704262` + satori `b946661` |
| A7 | **ink 掩膜结构指标**（`inkOv` Jaccard / `rowCorr` / `colCorr` Pearson / `flatBgPct`） | **有**（在 `.audit/`，离线脚本，**不在门禁路径**） | 无 | 无 | **挑目标 合法**；**判过线 不合法**（`pixel_test.go` 判定路径里不许出现） | **事后归档**（`provenance: retroactive-migration`，原始产物在仓库外 `_lead4/`）→ `mask_overlap.py::main()` @ gg `d939c38` |
| A8 | **structural verdict 四值判定** | **有**（同上，离线） | 无 | 无 | 同 A7 | 同 A7，`mask_overlap.py::main()` @ gg `d939c38` |
| A9 | **误差集中度分析**（`top5_err_share_pct` / `top10_err_share_pct` / `error_budget_split`） | **有**（`.audit/`，离线） | 无 | 无 | 同 A7 | **事后归档** → `missing-decomp.py::main()` @ gg `d939c38`（该脚本自述**直接 import 复用** A7 的 ink 定义，非独立出处） |
| A10 | **背景 oracle 天花板**（背景像素置基线值后重算 sim 上界） | **有**（`.audit/`，离线） | 无 | 无 | 同 A7 | **事后归档** → `missing-decomp.py::sim_ceiling` @ gg `d939c38` |
| A11 | **比较器自身的负向对照单测**（全局平移对齐 / 局部物体平移 / 边缘物体丢失 / 尺寸不等 / JPEG 微噪声 / 零位移保序） | 无（1 正 + 2 负；2 个负向只测 perturb 与零面积，不测比较器内部） | 有（9 个） | 有（9 个） | 诊断/回归 **可** / 门禁 **不适用** | `main_test.go` 的 9 个 `Test*` @ yoga `3704262` + satori `b946661` |
| A12 | **raw 与 aligned 双分数并列记录** | 无（只有 raw） | 有 | 有 | 诊断 **可**（并列记录，不替换判定用的那个） | `imageComparison` 结构体 @ yoga `3704262` + satori `b946661` |
| A13 | 基线 sha256 门禁 | 有（`hashOldHex != ent.Sha256` → `t.Fatalf`，整测中止） | 有（`item.OldGate`，逐条记录，不中止） | 有 | 两侧等价，**gg 不缺** | `TestGGPixelParity()` @ gg `d939c38` / `visual-final/main.go:289` @ satori `b946661` |
| A14 | 尺寸不等直接判红 | 有 | 有（返回 error） | 有 | 两侧等价，**gg 不缺** | 同上 |

**A1 的具体证据（gg 缺的那一格）**：gg `src/ggrender/pixel_test.go` 第 299–300 行连续两次写**同一个** `diffBuf.Bytes()`：

```go
_ = os.WriteFile(diffPath,     diffBuf.Bytes(), 0644)
_ = os.WriteFile(heatmapPath, diffBuf.Bytes(), 0644)
```

即 gg 的 diff 图只有「同/不同」二值（不同→`{255,0,0,255}`，相同→原色 `A=60`），**不含幅度**；heatmap 是 diff 的**同字节别名**（不是第二张图）。satori/yoga 的 heatmap 是 `{peak,0,0,255}`，`peak = max(dr,dg,db,da)`。〔measured〕

---

## §2 表 B — 仪器不可移植

> 判据：加进来必然改动「**被比较的像**」或「**分数的含义**」，或必须引入从基线反推的常量 → 所有已记录分数作废。
> **本表任何一条都不得由本报告代为落地；B1 已按 Boss 边界 3 停下上交。**

| # | 机制 | gg | satori | yoga | 为什么不可移植（依据读到的代码） | 与表 A 的关系 | 出处 |
|---|---|---|---|---|---|---|---|
| **B1** | **alignment 作为评分**：把配准塞进相似度计算**再出分** | 无 | 有 | 有 | 它替换了**被比较的像**：`translateImage(newImg, offset, background)` 之后才 `diffImagePair(oldImg, aligned, …)`，`AlignedScore` 是对**被平移过的图**打的分。搬进 gg 就等于 gg 的 `sim` 不再是「原图对原图」，仓库里每一条已记录分数的语义改变、全部作废。**已停下，交 Boss 决定，本报告不动手。** | 同一套 `registerOffset` 函数的**两种用法**：A5 = 只出诊断图（合法）／B1 = 进分数（不可移植）。**二者必须分开批** | `registerOffset()` / `selectFinalOffset()` / `translateImage()` @ yoga `3704262` + satori `b946661` |
| **B2** | **`LocalPass` 作为门禁合取项** | 无 | 有 | 有 | `item.Pass = item.OldGate && comparison.AlignedScore >= threshold && comparison.LocalPass` — 在 0.99 之外**再加一个与 sim 无关的独立通过条件**。gg 的 `gatePassed(sim, bbox)` 签名无法在不扩签名、不改判定路径的前提下容纳它。**且其结果依赖配准后的像**，故与 B1 耦合：B1 不批则 B2 也不可单独批 | 无对应的 A 行（`LocalPass` 本身不是诊断） | `visual-final/main.go:310` @ satori `b946661` + yoga `3704262` |
| **B3** | **`pageAnchorSignatures` 硬编码每页像素签名**（calendar `2546×38` / `2546×25`、lottery `139×138` / `42×45`） | 无 | 有 | 有 | 这些常量是**从冻结基线量出来的**，直接违反「不许从基线反推取值」红线；且换场景即失效，不可泛化 | 无对应 A 行 | `pageAnchorSignatures` / `anchorComponent()` @ yoga `3704262` + satori `b946661` |
| **B4** | **`microNoise` 过滤 + 局部配准进门禁** | 无 | 有 | 有 | 它把 ≤64px / ≤8×8 / meanDelta<64 的差异**当作不存在**。放进门禁 = 悄悄放宽有效阈值，改的是「什么算误差」，与改 0.99 等效 | **A4 / A6 的门禁层**；A4/A6 的诊断层仍合法 | `localGate()` @ yoga `3704262` + satori `b946661` |
| **B5** | **`abs8` 的 16bit→8bit 还原路径** | 无（直接 `RGBAAt` 取 8bit 精确作差） | 有 | 有 | 原则上换口径：satori/yoga 门禁路径上的 `abs8` 是 `>>8` **移位截断**，gg 是精确 8bit 差值。**但对 8bit 解码输入已穷举证明两者等价**（见 §3「等价性证明」），故不构成隐性口径差；只有在引入 8bit 以外的位深时才会分叉 | 无 | **门禁路径**：`abs8()` @ `src/cmd/visual-final/main.go:681-686`（`>>8`），satori `b946661` + yoga `3704262`，两树同内容<br>**测试路径**：`abs8()` @ `src/utils/media/compare_satori.go`（`/0x101`），**仅 yoga `3704262` 有** |
| **B6** | **以 aligned 图作为报告证据** | 无 | 有 | 有 | 若 gg 报告 aligned diff 作为门禁证据，证据与被门禁判的像不再对应（B1 的传染） | **A5 的门禁层** | `comparePageImages()` @ yoga `3704262` + satori `b946661` |

---

## §3 公式并列

**gg** — `similarityNormalized(old, new *image.RGBA) (float64, [4]int)`，`src/ggrender/pixel_test.go`，gg `feat/gg-render-card-atomic` @ `d939c38`：

```
每像素  delta = |ΔR| + |ΔG| + |ΔB| + |ΔA|      // RGBAAt() 的 8bit 值直接作差取绝对值
全图    sum   = Σ delta
分母    total = int64(w * h * 4 * 255)
分数    sim   = 1 - float64(sum)/float64(total)   // total>0；否则 sim = 1
bbox    有任一 delta≠0 的像素 → 其外接矩形；否则 [0,0,0,0]
```

**satori** — `similarity()` + `pixelDelta()` + `abs8()`，工作树 `C:/WorkSpace/Golang/arknights_bot-satori`、分支 `feat/satori-renderer`、commit `b946661`，文件 `src/cmd/visual-final/main.go`：
- `similarity()` @ **`:436-438`**（分母行 `:437`）
- `pixelDelta()` @ `:648-652`（分子，四通道）
- `abs8()` @ `:681-686`（`>>8` 移位截断）

**yoga** — 同名同结构的函数，工作树 `C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go`、分支 `feat/yoga-skia-go-renderer`、commit `3704262`，文件 `src/cmd/visual-final/main.go`，**行号与 satori 侧完全相同**（`:437` / `:648-652` / `:681-686`）。

> **定位更正（读代码时踩过的坑）**：`src/utils/media/compare_satori.go` **只存在于 yoga 树**（satori 树 0 命中，yoga 树 1 命中），文件名里的 "satori" 是误导 —— 它不是 satori 线的文件。且它**不是 `main.go` 的逐字副本**（两者 `abs8` 实现不同，见下）。

两树公式原文：
```
每像素  delta = abs8(R) + abs8(G) + abs8(B) + abs8(A)   // 四通道，与 4*255 分母对称
          abs8(a,b) = (|a.RGBA() - b.RGBA()|) >> 8      // 门禁路径实况
全图    total = Σ delta
分母    total = bounds.Dx() * bounds.Dy() * 4 * 255
分数    score = 1 - float64(total)/float64(Dx*Dy*4*255)
```

**`abs8` 存在两个版本，务必区分（这是本报告被自我更正的一处）**：
- `>>8` 移位截断 —— 在 **`main.go`**，**satori 与 yoga 两树都有**，**这一份才在门禁路径上**（`visual-final` 是 `package main`，自带一份比较器）。
- `/0x101` 精确除法 —— 在 **`compare_satori.go`**，**仅 yoga 有**，注释写着 "identical to ggrender's unified diff formula" 与 "(The old >>8 truncation path is removed; numerics are unchanged)"。**但它不在门禁路径上**：其 8 个调用点全在 `src/utils/media/yoga_skia_test.go`，属测试路径。
- ⇒ **引用口径证据时不要引 `compare_satori.go` 的注释**，门禁口径的出处是 `main.go:681-686`。

### 是否同口径：**同口径。**〔measured〕

- 三者都是 **4 通道 L1 之和 / (w × h × 4 × 255)**：4 通道、满量程 255、除法式 `1 - 误差/预算` 完全一致。历史担心的「3 通道分母 vs 4 通道总分母」**没有发生**。
- 交叉印证：`missing-decomp-1b433bc.json` 落盘的仪器定义原文写作 `1 - sum(|dR|+|dG|+|dB|+|dA|) / (w*h*4*255)`。
- satori 与 yoga 的两个 `main.go` **内容相同、行尾不同**（不是「逐字相同」）：`diff --strip-trailing-cr` **exit 0**（内容一致）；但 satori 侧 30832 bytes / **0 个 CR**（LF），yoga 侧 31629 bytes / **797 个 CR**（CRLF），裸 `diff` 因此报 `1,797c1,797`（全文皆变）。**这正是 AGENTS.md 警告的 CRLF 假差异**。〔measured〕
- `compare_satori.go`（仅 yoga）**与 `main.go` 的比较器并不相同** —— 差在 `abs8`（`/0x101` vs `>>8`）。它是一个被改过 `abs8` 的独立副本，不是「逐字复用副本」。〔measured〕

**等价性证明（`>>8` 不是隐性口径差）**：〔measured〕
- 8bit 解码输入下 `color.RGBA()` 展开为 `v*0x101`，故 8bit 差值 `d` 的 16bit 差为 `257*d`。
- 穷举 `d ∈ [0,255]`：`exceptions = len([d for d in range(256) if (257*d)>>8 != (257*d)//0x101]) = 0`；校验 `0 + 256 == 256` ✓。
- ⇒ 门禁路径的 `>>8` 与测试路径的 `/0x101` 对 8bit 输入**恒等**，两者不构成口径差；`compare_satori.go` 注释里 "numerics are unchanged" 的说法与本穷举结果一致。
- 探测尺子自检：`printf 'x\r\ny\n' | od -c` 先验证能看见 `\r`；字节差用 `wc -c` 减 `tr -d '\r' | wc -c`（不用 `grep -c $'\r$'`，该形态在本项目已被记为不可信）。

### 但必须同时写清的分歧：**公式同口径 ≠ 分数可比。**〔measured〕

satori/yoga 的 `similarity()` 作用在**配准后**的 `aligned` 图上；gg 作用在**原始未配准**的图上。相同的是那个除法式子，**不是同一个被测量**。全部语义差异来自 **B1**（被比较的像不同）与 **B2 / B4**（额外通过条件不同）——即：**gg 缺的不是一个更精确的公式，是一条「比较对象」与「通过判据」上的额外机制。**

---

## §4 未证实的方向（4 条假设逐条结案）

**1. 「satori / yoga 各自有独立比较 harness，路径待确认」→ 已确认。**〔measured〕
- satori：`arknights_bot-satori` @ `b946661` 的 `src/cmd/visual-final/main.go`（`threshold = 0.99` 在第 26 行；`comparePageImages()` 在第 340 行）+ `src/cmd/visual-final/main_test.go`（9 个比较器单测）。
- yoga：`arknights_bot-satori-yoga-skia-go` @ `3704262` 的**同一对文件**（`src/cmd/visual-final/main.go` + `main_test.go`），内容与 satori 侧相同（行尾不同，见 §3）。yoga 另有 `src/utils/media/compare_satori.go`（**仅 yoga 有**，在测试路径上，且 `abs8` 与 `main.go` 不同），以及产物 `src/utils/media/testdata/visual/final-yoga-skia/compare/*.aligned.diff.png`。
- 附带核实：satori 树**没有** `src/ggrender`（`ls -d` exit=2）。yoga 树**有** `src/ggrender/pixel_test.go`，但那是 gg 线历史的旧副本（仍是 `passed := sim >= 0.99`，无 `gatePassed`、无零面积守卫），**不是 yoga 自己的门禁**。〔measured〕
- 两树另有 `src/cmd/visual-regression`，产出 `report.json` + `heatmap`。

**2. 「satori/yoga 有 gg 没有的 rowCorr / colCorr / inkOv」→ 已确认，但方向与假设相反。**〔measured〕
- 这些指标住在 **gg 自己的** `.audit/content-mask-overlap-20260930/mask_overlap.py`（gg 工作树 `arknights_bot-gg-card-atomic`、分支 `feat/gg-render-card-atomic`、commit `d939c38` 的 `main()`），且是**离线脚本、不在门禁路径**。
- 公式原文（脚本头部 `Definition` 段）：`bg` = 精确众数 24-bit RGB（**无量化**）；ink 掩膜 = `|pixel − bg|₁ > 30`；`inkOv% = 100·|A∧B| / |A∨B|`（并集分母 Jaccard）；`rowCorr`/`colCorr` = per-row / per-column ink 计数的 Pearson r；`flatBg% = 100·max(众数像素数)/(H·W)`。
- ** Boss 自查项的答复 —— 这些指标到底出自谁：**
  - **出处是单一起点**：脚本头部自述 `origin: C:/WorkSpace/Golang/_lead4/false-friend-scan-16.json`，且 `original_script: NEVER PERSISTED to disk`，定义本身是从 pi 会话 transcript **逐字恢复**的。原始产物仍可考：`C:/WorkSpace/Golang/_lead4/false-friend-scan-16.json`（3381 bytes，实测存在）。〔measured〕
  - **`.audit/` 下这些文件是事后分析归档，不是任何一条线的仪器。** 它们既不能被称作 satori 线的仪器，也不能被当作已有仪器的一部分。**但按 Boss 边界 1，这不推出它们没用** —— 见 §1 表 A 的「可用性」栏：作为**挑目标**的诊断合法，作为**判过线**的门禁不合法。〔inferred，依据 §1 表 A 使用规则〕
  - **有没有第二个独立出处？没有。** 仓库外另有 `_regen/mask_overlap.py` 与 `_regen/mask_overlap_new.py`，二者与入库的 `.audit/.../mask_overlap.py` **逐字节相同**（工作副本，非独立出处）；`_regen/mask_overlap_committed.py` 与之差 24 行（是加 `UNTESTABLE` 支前后的变体）。`_lead1/tools/headhunt_tiles.py` 载有**同一定义文字与同一套阈值**但自带 `def corr()`，属**同一定义的第二个实现**（换场景：headhunt），不是独立出处。〔measured〕
  - `missing-decomp.py`（A9/A10）自述「复用已入库的 ink 掩膜定义（直接 import，不另造一套）」→ 它**派生自** A7/A6，同一条出处链，不是独立第二源。〔measured〕

**3. 「存在 structural verdict，取值集合待确认」→ 已确认，但取值集合需更正。**〔measured〕
- 判定函数：`mask_overlap.py::main()` 内联表达式，gg `feat/gg-render-card-atomic` @ `d939c38`：

```
v = "UNTESTABLE"     if flatBgPct > 99.9
    "STRUCT-RELATED" if (inkOv >= 50 and max(|rowCorr|,|colCorr|) >= 0.5)
    ("UNRELATED/FAKE" if (inkOv < 50 and max(|rowCorr|,|colCorr|) < 0.5) else "BORDERLINE")
```

- **更正**：`UNRELATED` 与 `FAKE` 不是两个取值，代码里是**单个合写字符串 `UNRELATED/FAKE`**。实际取值集合为 4 个：`UNTESTABLE` / `STRUCT-RELATED` / `UNRELATED/FAKE` / `BORDERLINE`。`UNTESTABLE` 是本仓库加的扩展支（脚本头部自述非 2026-09-30 原始规则；原始三值规则会把 `depot` 判成 `UNRELATED/FAKE`）。
- 两版实测计数（`report-1b433bc.json` @ `1b433bc17b45` / `report.json` @ `ec440bf`）：
  - @1b433bc：`STRUCT-RELATED` 6 / `BORDERLINE` 2 / `UNRELATED/FAKE` 7 / `UNTESTABLE` 1 → 校验 6+2+7+1 = 16 ✓
  - @ec440bf：`STRUCT-RELATED` 4 / `BORDERLINE` 2 / `UNRELATED/FAKE` 9 / `UNTESTABLE` 1 → 校验 4+2+9+1 = 16 ✓
  - `UNRELATED/FAKE` 的 7 个成员实测：`box, box-detail, box-summary, enemy, gacha, lottery, recruit`。

**4. 「误差集中度分析」与「背景 oracle 天花板」→ 两者均已确认，且都只存在于 gg 线。**〔measured〕
- 误差集中度：`missing-decomp-1b433bc/missing-decomp.py::main()` 的 `top5_share` / `top10_share`（对 COLS×ROWS 网格切块、按误差降序取前 5/10 的占比）。落盘实测：`tiles.top5_err_share_pct` = 10.916、`tiles.top10_err_share_pct` = 20.71；另有 `error_budget_split`（`content_px_err_share_pct` = 99.555 / `flat_bg_px_err_share_pct` = 0.445）。该脚本只处理 **`missing`** 一个场景。
- 背景 oracle 天花板：同脚本的 `ceiling_if_background_oracle`（`sim_ceiling`），落盘字段名一致。
- 否定结论（举证方式见 §6）：在 satori 与 yoga 两树上 `grep -rn -E 'oracle|ceiling|集中度'` 的全部命中**均与该机制无关** —— 仅 yoga `src/skia/backend.go:34` 的英文注释 "pixel oracle"（skia 测试夹具）、yoga `src/skia/doc.go:13` 的 "Ponytail ceilings"（代码债标记）、satori 的**编译产物** `src/visual-final.exe`（二进制字符串命中，**不是源码**）。无任何 Go/Python 源码级实现。

---

## §5 与 AGENTS.md / HANDOVER-gg.md 不符的事实

| # | 文档路径 | 原文断言 | 实际事实 | 实际事实的出处 |
|---|---|---|---|---|
| D1 | `C:\WorkSpace\Golang\AGENTS.md:55` | 「实例：card 的 **216,347** 个色对、最大仅 **0.49%** —— 弥散」—— 读起来像一条已落盘、可复算的实测 | `216347` / `216,347` / `0.49%` 在**任何工作树内都搜不到**（跨全部 22 棵的 grep，唯一命中是该文档自身 + 仓库外 `_lead1/card-gap-layout.md:194,330`）。仓库内唯一的集中度实测是 **`missing`** 场景的 `top5=10.916%` / `top10=20.71%`；而 `0.445%` 那个数是 `missing` 的 `flat_bg_px_err_share_pct`，**数值与场景都不对**（0.445 ≠ 0.49，且是 missing 不是 card） | 跨 22 棵 grep（§6 N5）；`.audit/missing-decomp-1b433bc/missing-decomp-1b433bc.json` 的 `scene`="missing"、`tiles.top5_err_share_pct`、`error_budget_split.flat_bg_px_err_share_pct` |
| D2 | 任务假设（与 AGENTS.md 措辞易混） | 易读成「rowCorr / colCorr / inkOv / verdict 是 satori 或 yoga 线的方法，gg 缺」 | **方向相反**：四样全在 gg 线自己的 `.audit/` 里，且属**事后归档**（`retroactive-migration`），不属任何一条线的仪器 | §4 第 2、3 条的穷举命令与 exit code |
| D3 | `AGENTS.md`「bbox 空守卫」段 | 表述为 card-atomic 的当前状态，读者会以为工作树里都是带 `gatePassed` 的版本 | 同名文件 `src/ggrender/pixel_test.go` 在两棵工作树里是**两个不同仪器版本**：gg 工作树 `arknights_bot-gg-card-atomic` @ `d939c38` 有 `gatePassed(sim,bbox)`；yoga 工作树 `arknights_bot-satori-yoga-skia-go` @ `3704262` 仍是旧的 `passed := sim >= 0.99`、无 `gatePassed`、无零面积守卫。正是 AGENTS.md 自己那条「文件名只标识到工作树」的实例 | 两树该文件 `diff`：18 行删除、117–129 行 `gatePassed` 整块仅存在于 gg 侧 |
| D4 | `AGENTS.md`「bbox 空守卫」段 | 「card-atomic HEAD 已推进到 `ec440bf`」 | HEAD 已继续前进到 `d939c38`（`chore(audit): archive the baseline-as-render-input check as a standing check`）。**属快照过期，非事实错误** | `git log -1` @ gg `arknights_bot-gg-card-atomic` |
| D5 | — | — | **未发现腐坏**（逐条实测全部对上）：AGENTS.md 关于两版分数表、`manifest.json` 字段清单（`entries` + `templateTreeSHA256` / `assetTreeSHA256` / `font` / `browserDeterminism` / `gachaTsUnit` / `legacyRuntime`，**manifest 内确无零阈值字段**）、陷阱 7 个及其成员名单、`base` = BORDERLINE、`depot` 不可测、`card` 两版差 4.75e-05（`0.9601980124080882` @ec440bf vs `0.9601504863664215` @1b433bc，差 **4.7526e-05**）| 两份 report 的 `measurement_point` + 逐场景字段比对：两版 `scenes` 各 16 条，逐场景比对仅 **3** 条不同（`card` / `headhunt` / `operator`），13 条逐字段相同；校验 3+13 = 16 ✓ |

### §5.1 缺 `provenance` 标记的 `.audit/` 归档文件（单列，不混进上表）

审计范围：`arknights_bot-gg-card-atomic`（分支 `feat/gg-render-card-atomic`，commit `d939c38`）下 `.audit/` 的**全部 12 个文件**，逐个 `grep -c 'retroactive-migration'` 与 `grep -c 'original_path'`。结果 **9 OK / 3 MISSING**。

| 文件 | `provenance: retroactive-migration` | `original_path` | 依项目规矩的定性 |
|---|---|---|---|
| `.audit/00447f3/report.md` | **缺** | **缺** | **缺标记**。同目录的 `report.json` 两样俱全（`provenance=retroactive-migration`, `original_path` 存在），故这是 markdown 渲染件漏标，不是「只有结果没有尺子」 |
| `.audit/missing-decomp-1b433bc/missing-decomp.py` | **缺** | 有（散文中写明：原始产物写到 `_lead2/`） | **缺标记**。它自身是仓库外定义的归档（头注释即声明输出写到 `_lead2/`），按规矩应带显式 `provenance: retroactive-migration`；实际只有散文式 original_path，机器读不到 |
| `.audit/standing-checks/baseline-as-render-input/README.md`<br>`.audit/standing-checks/baseline-as-render-input/check_baseline_as_render_input.py` | **缺** | **缺** | **规矩是否适用存疑** —— `standing-checks/` 是常驻自检脚本而非「仓库外测量结果的归档」，可能本就不适用 retroactive-migration 规矩。**本报告不断言它违规**，只记录它没有标记 |

配套事实（正面）：`.audit/content-mask-overlap-20260930/` 的 4 个文件（`mask_overlap.py` / `report.json` / `report-1b433bc.json` / `report.md`）与 `.audit/missing-decomp-1b433bc/` 的 3 个 JSON **全部两样俱全**，分别指向 `_lead4/false-friend-scan-16.json` 与 `_lead2/<same filename>`。且 `.audit/` **未被 gitignore、已入库**（`git check-ignore` exit=1 = 未忽略；`git ls-files --error-unmatch` exit=0 = 已跟踪）。

---

## §6 取证附录

**工作树枚举范围（否定性结论的依据）**：以下每条否定结论都把 `git worktree list` 的**全部 22 棵**（实测 `wc -l` = 22）逐棵喂给同一条 grep，而不是只看 2–3 棵。

```bash
git -C arknights_bot worktree list | wc -l        # → 22
git -C arknights_bot worktree list | awk '{print $1}' | while read -r wt; do
  grep -rl --include='*.go' --include='*.py' --include='*.mjs' --include='*.js' -E "$NAME" "$wt" 2>/dev/null
done
```

**尺子自检**（结论碰巧对不等于尺子能用 —— 先证明尺子会响）：

- **阳性对照**：同一 grep 形状打已知阳性文件 → `grep -rn -E 'rowCorr' arknights_bot-gg-card-atomic/.audit/content-mask-overlap-20260930/mask_overlap.py` → 2 行命中，**exit 0**。
- **阴性对照**：同一 grep 形状打 satori/yoga 两树 → **exit 1 且输出为空**。
- **尺子失效的一次实录**：首轮批量审计用了 `p=$(grep -c ... || echo 0)`，而 `grep -c` 无命中时**同时**输出 `0` 并 exit 1，`|| echo 0` 又追加一个 `0`，变量变成 `"0\n0"`，导致 printf 输出错行。**已作废该轮结果并重跑**，改用 `[ "$pr" -ne 0 ] && p=0` 归一；重跑结果与逐条复核（§5.1 表内每格均单独 `grep -c` + 打印 exit code）一致。**这条兜底正是本项目明令禁用的 `||` 形态** —— 它把「查询失败」和「结果为空」压成同一信号，此处它压的是「读到 0」和「grep 失败」。
- **已用但本报告未依赖**：`|| echo none`（全程未用）；`MSYS_NO_PATHCONV=1`（本报告未用 `git show <rev>:<path>`，全部读工作树文件，故无类路径转换风险）；`ctx_execute` / `ctx_batch_execute`（全程未用于取证，按纪律走 bash；唯一一次误用风险自查已排除）。

**逐条否定结论的举证方式**（Boss 边界：举证栏空 = 未举证 = 退回重做）：

| 否定结论 | 举证方式 | 枚举范围 | 结果 |
|---|---|---|---|
| satori 线**无** `rowCorr` / `colCorr` / `inkOv` | 穷举扫了全量：22 棵工作树 × 4 个源码 glob（`*.go` `*.py` `*.mjs` `*.js`）逐棵 grep，**并跑阳性对照** | `git worktree list` 全 22 棵 | 命中 4 个文件，全在 `.audit/`（gg 与 consolidate 两树的归档副本）；**无任何 `src/` 或 `renderer/` 下的源码命中** |
| yoga 线**无** `rowCorr` / `colCorr` / `inkOv` | 同上（同一枚举） | 同上 | 同上（yoga 树 0 命中） |
| satori 线**无** `STRUCT-RELATED` | 同上 | 同上 | 命中 2 个文件，均为 `.audit/` 归档副本；源码 0 命中 |
| satori 线**无** oracle / ceiling / 集中度 | 穷举 `grep -rn -E 'oracle|ceiling|集中度'` 覆盖两树**全部文件类型** | satori + yoga 两树全量 | 命中 3 处，全为无关（注释 2 + 二进制 1） |
| satori 线**无** `src/ggrender` | `ls -d`（未用 grep，无「另一种叫法」风险） | 单点 | **exit 2** = 不存在 |
| `216347` / `0.49%` **不存在于**任何工作树 | 穷举：跨 22 棵工作树 × `*.py` `*.json` `*.md` `*.txt` | 全 22 棵 | exit 1（除文档自身与仓库外 `_lead1/`） |
| satori / yoga 线**无** `similarityNormalized` 的独立实现 | 同一 grep 形状的**阳性对照在 yoga 树命中** `src/ggrender/pixel_test.go`（exit 0）→ 证明尺子能出信号，satori 树的 exit 1 是真阴性 | satori + yoga 两树 | satori 树 exit 1；yoga 树的命中是 gg 历史的旧副本（见 D3） |

**「另一种叫法」已排除**：A7/A8 是 Python 分析脚本里的局部变量名（`ov` / `rc` / `cc` / `corr()`），不是通用选择器或类名——换叫法仍会命中 `corr(`、`Jaccard`、`.sum(axis=1)` 这些结构性写法；两树均无此类写法。〔inferred，但有 §6 N5 的穷举结果支撑〕

---

## 一句话总结

**gg 在「仪器严格性」上不比 satori/yoga 差 —— 三条相似度公式同口径，都是 4 通道 L1/(w·h·4·255)**（gg `pixel_test.go::similarityNormalized` / satori `main.go::similarity` @ `b946661` / yoga `main.go::similarity` @ `3704262`；两树的 `main.go` 内容相同、行尾不同）；差距全部在「比较的对象」与「通过的判据」上。 最大的三项是：①**被比较的像是否配准**（B1，唯一真正改变分数语义的一项，已按 Boss 边界停下上交、不动手）②**是否有一个与相似度无关的结构通过条件**（B2/B4）③**诊断分辨率** —— gg 的 heatmap 是 diff 的逐字副本（A1），satori/yoga 有真量值 heatmap + 连通域分解（A1–A4）。反方向需更正的是：`rowCorr`/`colCorr`/`inkOv`/`verdict`/误差集中度/背景 oracle 这些**不是 satori 的仪器、也不属任何一条线的仪器**（是 gg 树里 `retroactive-migration` 的事后归档），但**它们仍然有用** —— 合法角色是**按「有没有诊断」分档挑实现目标**，非法角色才是进 `gatePassed` 判定路径。