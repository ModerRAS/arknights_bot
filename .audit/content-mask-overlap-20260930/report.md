# 内容掩膜（ink）重叠率 —— 定义归档与 card 判定

## 1. 这份文档解决什么

`inkOv`（内容掩膜重叠率）此前只有**结果**、没有**仪器**：`_lead4/false-friend-scan-16.json`
里有 16 行数值，但生成它的脚本从未落盘。`C:/WorkSpace/Golang/_lead4/` 的目录清单里
只有 8 个 PNG + 3 个 JSON + 1 个 MD，**没有任何脚本文件**。

这正是 AGENTS.md 取证纪律那一节当天新写入的条款所描述的失效形态：
**「阈值型判据必须连仪器定义一起落盘。只留整数等于只留结果没留尺子。」**

本目录是那份「尺子」的追认归档。

## 2. 出处与 provenance

| 项 | 值 |
|---|---|
| `provenance` | `retroactive-migration` |
| `original_path` | `C:/WorkSpace/Golang/_lead4/false-friend-scan-16.json` |
| 原始脚本路径 | **无**（从未落盘） |
| 定义恢复来源 | `C:/Users/ModerRAS/.pi/agent/sessions/--C--WorkSpace-Golang--/2026-09-30T01-03-02-285Z_01a0efd6-1e8c-761a-a158-2582d00bd136.jsonl`（lead-4 会话，`toolCallId: call_function_6xsn6ccngb3p_1`） |
| 为什么是追认 | 生成脚本从未写进磁盘；定义是从**产出该 JSON 的那次会话记录**里逐字恢复的。所以这份归档本身就是一次事后追认，必须带 `provenance` 标记，否则就变成「归档了证据、留下了同样的病」。 |
| 本目录不含任何图像 | 冻结基线图一律不入库 |

## 3. 定义（逐字，与原实现一致）

- **背景**：每张图各自的**精确众数 24-bit RGB**（`np.unique` 计数取 `argmax`，**不做量化**）。
- **ink 掩膜**：`np.abs(pixel - bg).sum(axis=2) > 30` —— R+G+B 的 **L1** 距离，阈值 **30**。
- **重叠率**：`inkOv = 100 * |A ∩ B| / |A ∪ B|` —— **Jaccard**，分母是并集。
- **rowCorr / colCorr**：逐行 / 逐列 ink 计数的 Pearson 相关系数。
- **flatBgPct**：`100 * max(两图众数像素计数) / (H * W)`。
- **判定规则（先判 `UNTESTABLE`，再判三档）**：
  - `UNTESTABLE`：**`flatBgPct > 99.9`**（整张画布单一纯色 → ink 掩膜按定义是空集）——
    **优先于**下面三档，因为掩膜为空时「重叠低」是空洞为真，corr 与 ink 指标不携带信息
  - `STRUCT-RELATED`：`inkOv >= 50` **且** `max(|rowCorr|, |colCorr|) >= 0.5`
  - `UNRELATED/FAKE`：`inkOv < 50` **且** `max(|rowCorr|, |colCorr|) < 0.5`
  - `BORDERLINE`：其余

  **`UNTESTABLE` 这一支是本仓的扩展，不属于 2026-09-30 的原版**（原版只有三档，
  因此把 depot 报成 `UNRELATED/FAKE`）。五个度量字段不受该扩展影响。

  **硬边界：`flatBgPct > 99.9`，不得放宽。** 在这 16 行里它只命中 depot（100.0），
  与次高者 calendar（87.0）相距 **13.0pp**。`box-detail 76.1`、`base 76.2`、`state 70.7`、
  `enemy 69.8`、`box-summary 53.0` 这些「高但不到 100」的行是**可测的**，
  判为「可测但弱相关」，**不是** `UNTESTABLE`。

## 4. 互证结果

同一份定义、两套独立实现（不同代码、不同工作树、不同 commit 的输入图）：

| 校验点 | lead-4 @ `1b433bc` | 本次复现 @ `ec440bf` | 一致 |
|---|---|---|---|
| calendar inkOv | 86.7 | **86.7** | 是 |
| calendar rowCorr / colCorr / flatBg | +0.373 / +1.000 / 87.0 | +0.373 / +1.000 / 87.0 | 是 |
| state inkOv | 71.8 | **71.8** | 是 |
| state rowCorr / colCorr / flatBg | +0.950 / +0.978 / 70.7 | +0.950 / +0.978 / 70.7 | 是 |

全 16 行里 **14 行的全部 5 个度量字段逐位相同**。剩下 2 行（`headhunt`、`operator`）不同，
原因是这两个场景的 **similarity 本身** 在两个 commit 上就不同（即输入图不同），
不是定义分歧：

| 场景 | sim @1b433bc | inkOv @1b433bc | sim @ec440bf | inkOv @ec440bf |
|---|---|---|---|---|
| headhunt | 0.96476 | 98.3 | 0.79983 | 29.4 |
| operator | 0.84554 | 76.8 | 0.69063 | 25.0 |

**结论：这是互证，不是单方测量。** 定义可用。

## 5. 16 场景全表（@ `ec440bf` card-atomic）

| 场景 | sim | 过 0.99 | inkOv% | rowCorr | colCorr | flatBg% | 判定 |
|---|---|---|---|---|---|---|---|
| calendar | 0.99432 | 是 | 86.7 | +0.373 | +1.000 | 87.0 | STRUCT-RELATED |
| card | 0.96020 | 否 | **77.6** | +0.909 | +0.929 | 23.0 | **STRUCT-RELATED** |
| state | 0.99093 | 是 | 71.8 | +0.950 | +0.978 | 70.7 | STRUCT-RELATED |
| missing | 0.87614 | 否 | 50.8 | +0.610 | +0.442 | 37.1 | STRUCT-RELATED |
| base | 0.95900 | 否 | 28.1 | +0.568 | +0.296 | 76.2 | BORDERLINE |
| help | 0.83943 | 否 | 35.9 | +0.615 | +0.004 | 43.7 | BORDERLINE |
| box | 0.82504 | 否 | 42.1 | +0.375 | -0.323 | 35.2 | UNRELATED/FAKE |
| recruit | 0.61351 | 否 | 39.9 | +0.431 | +0.051 | 48.6 | UNRELATED/FAKE |
| box-summary | 0.88098 | 否 | 15.5 | +0.122 | +0.152 | 53.0 | UNRELATED/FAKE |
| enemy | 0.91951 | 否 | 12.2 | -0.201 | +0.059 | 69.8 | UNRELATED/FAKE |
| gacha | 0.90251 | 否 | 9.1 | +0.097 | +0.230 | 53.2 | UNRELATED/FAKE |
| box-detail | 0.87844 | 否 | 9.9 | +0.025 | +0.075 | 76.1 | UNRELATED/FAKE |
| lottery | 0.98358 | 否 | 8.2 | +0.048 | -0.021 | 45.3 | UNRELATED/FAKE |
| depot | 0.91728 | 否 | 0.0 | +0.081 | -0.010 | 100.0 | **UNTESTABLE**（见 5.1） |
| headhunt | 0.79983 | 否 | 29.4 | +0.206 | -0.034 | 68.8 | UNRELATED/FAKE |
| operator | 0.69063 | 否 | 25.0 | +0.076 | -0.296 | 53.5 | UNRELATED/FAKE |

计数：`STRUCT-RELATED` 4 / `BORDERLINE` 2 / `UNRELATED-FAKE` 9 / **`UNTESTABLE` 1**。

### 5.1 `depot` 的 `inkOv = 0.0` 表示**不可测**，verdict 记为 `UNTESTABLE`

`depot` 的 `flatBgPct = 100.0`：整张画布是**单一纯色**，众数像素覆盖 100% 画布。
在这种输入下 ink 掩膜按定义必为空集，而空集与任何掩膜的交集恒为空、并集恒等于对方，
所以 `inkOv` **必然**是 0.0 —— 这是定义的退化结果，不是关于 depot 内容的测量结论。

**不可测 ≠ 不合格。** 这一行不得被读成「depot 造假」或「depot 结构无关」：
它只能读成「本指标在纯色画布上没有分辨力」。要判 depot 必须换一个能产生 ink 的测量口径。

**归档口径已统一：** 判定规则新增 `UNTESTABLE` 支（见第 3 节，硬条件 `flatBgPct > 99.9`），
`depot` 的 `verdict` 字段在 `report.json` 与 `report-1b433bc.json` 中均为 **`UNTESTABLE`**，
并带 `verdict_reason` 字段。**同一份归档里不再存在「机器字段说 FAKE、人读正文说不可测」的两个说法。**
（历史记录：`2026-09-30` 原版只有三档，因此当时报的是 `UNRELATED/FAKE`。）

### 5.2 `lottery` 是全表最危险的一行

`sim 0.98358` 是 16 行里**离 0.99 门禁最近**的非过线项，而它的
`inkOv = 8.2%`（全表第二低）、`max(|rowCorr|,|colCorr|) = 0.048`（全表最低）→ `UNRELATED/FAKE`。

**分数最接近过线、结构相关性最低。** 下一次有人拿 lottery 去试 0.99 门禁时，
这一行就是拦住他的东西：它离门禁只差 0.00642，但那个分数是深色扁平背景撑出来的，
继续往上堆相似度只会把一个结构无关的页面推过线，不会让它更接近基线。

### 5.3 两套独立实现互证的边界

全 16 行里 **14 行的 5 个度量字段逐位相同**。剩下 2 行（`headhunt`、`operator`）不同，
原因是**输入渲染图在两个 commit 上本身就不同**（`similarity` 也不同：
`0.96476 → 0.79983`、`0.84554 → 0.69063`），**不是定义分歧**。

这个区分必须留着：不存在两套定义在打架，只有两个 commit 各自渲出了不同的图。
定义侧的一致性由 calendar（86.7 / +0.373 / +1.000 / 87.0）与 state（71.8 / +0.950 / +0.978 / 70.7）
两行的逐位相同单独证明。

## 6. card 的判定

- **判别线**：`inkOv >= 50` 且 `max(|rowCorr|,|colCorr|) >= 0.5` → 结构相关。
- **实测**：`inkOv = 77.6%`（判别线的 1.55 倍），`rowCorr = +0.909`、`colCorr = +0.929`（判别线 0.5 的 1.8 倍），`flatBgPct = 23.0%`。
- **三档结论：真进展。**
- 与 Boss 转述的假高分反例并列：`base` 是 `0.95900` / `inkOv 28.1%` / `flatBg 76.2%`；
  `card` 是 `0.96020` / `inkOv 77.6%` / `flatBg 23.0%`。**两者分数几乎相同（0.959 vs 0.960），
  结构相关性完全相反**——card 的画布只有 23% 是平铺背景，base 有 76%。这正面证实了
  「深色扁平背景主导分数」的机制，同时也证实 **card 的 0.96020 不是被背景撑出来的**。
- `lottery` 0.98358 是本次扫描里分数最高的非过门禁项，但 `inkOv` 只有 8.2% → **假高分**。
  这是全表里最值得警惕的一行：它是「离门禁最近」的那一行。

## 6.1 测量点与提交点的分离

**测量点是 `ec440bf`，不是本文件所在的提交。** 本目录在 `feat/gg-render-card-atomic` 上入库时，
该分支 HEAD 已由他人推进到 `5aa2ac3`（两个 `.audit/` 文档提交）。
实测 `git diff --stat ec440bf..5aa2ac3 -- src/ggrender/scene_card.go src/ggrender/render.go`
输出为空，故 card 的渲染输出未变，本文件的数对 `5aa2ac3` 仍然成立 —— 但这是**另一次核验的结论**，
不是「本文件测的就是当前 HEAD」。任何后来者引用本文件时必须读这一行。

## 7. 边界声明

- 本目录**不含任何冻结基线图像**，只有定义代码 + 数值结果。
- 脚本只读 harness 已经写出的 `tmp/pixel-compare/<scene>/{old,new}.png`，
  **不 import 仓库任何代码**，**不参与渲染**，**不修改任何评分代码**。
- `content_mask_overlap` **没有**被实现进 harness；`pixel_test.go` 一个字节未动。
  门禁仍然只算 `similarityNormalized`，本指标是**门禁之外的一次性判别**。

## 8. 第二次测量：`1b433bc`（PR #5 源分支 tip）

> 本节为**追加**，第 1–7 节一字未改。第 4/5 节的表与全部数字是 `ec440bf` 上的测量；
> 本节是 `1b433bc` 上的另一次测量，**两者的差异归因见 8.3**。

### 8.1 测量点与三次确认

| 项 | 值 |
|---|---|
| `measurement_point.commit` | `1b433bc17b453523c7bb11ce23eb29a89f644064`（`consolidate/gg-mainline` tip，PR #5 源分支） |
| 跑法 | 临时 detached worktree，跑完 `git worktree remove`；未切 card-atomic |
| 命令（**必须从仓库外执行**，见下方说明**） | 1) `cd <临时 detached worktree>/src && go test ./ggrender/ -run 'TestGGPixelParity$' -v -count=1`（7.67s）<br>2) `git show 1646f40:.audit/content-mask-overlap-20260930/mask_overlap.py > /c/WorkSpace/Golang/_regen/mask_overlap.py`<br>3) `python -B /c/WorkSpace/Golang/_regen/mask_overlap.py <worktree-root> base box box-detail box-summary calendar card depot enemy gacha headhunt help lottery missing operator recruit state` |
| 脚本位置 | **不得在本目录（归档目录）内运行或 `import`**。归档目录只写入、不执行。详见 `report.json` 的 `regeneration` 节。 |
| 产物 | `report-1b433bc.json`（与本文件同目录，已随 `1646f40` 入库） |
| 图像 | 零 |

**三次确认（这是本次重跑的全部意义）：**

1. **定义正确** —— 定义来自 lead-4 会话 transcript 的逐字恢复，与 lead-4 当年**手写脚本**是两条独立路径。
2. **实现正确** —— 恢复出来的脚本复算出的数，与 lead-4 当年那份裸 JSON **16/16 行、六个字段零差异**
   （`sim` / `inkOv` / `rowCorr` / `colCorr` / `flatBgPct` / `verdict`）。
3. **测量点正确** —— 输入渲染图取自 `1b433bc`，而 lead-4 当年也用 `1b433bc`。

三者的交集是唯一变量：只有「测量点」被换掉过，两条代码路径给出同一个数。

### 8.2 `1b433bc` 上的 16 行

| 场景 | sim | 过 0.99 | inkOv% | rowCorr | colCorr | flatBg% | 判定 |
|---|---|---|---|---|---|---|---|
| headhunt | 0.96476 | 否 | 98.3 | +0.965 | +0.887 | 4.6 | STRUCT-RELATED |
| calendar | 0.99432 | 是 | 86.7 | +0.373 | +1.000 | 87.0 | STRUCT-RELATED |
| card | 0.96015 | 否 | 77.6 | +0.909 | +0.929 | 23.0 | STRUCT-RELATED |
| operator | 0.84554 | 否 | 76.8 | +0.882 | +0.602 | 28.0 | STRUCT-RELATED |
| state | 0.99093 | 是 | 71.8 | +0.950 | +0.978 | 70.7 | STRUCT-RELATED |
| missing | 0.87614 | 否 | 50.8 | +0.610 | +0.442 | 37.1 | STRUCT-RELATED |
| base | 0.95900 | 否 | 28.1 | +0.568 | +0.296 | 76.2 | BORDERLINE |
| help | 0.83943 | 否 | 35.9 | +0.615 | +0.004 | 43.7 | BORDERLINE |
| box | 0.82504 | 否 | 42.1 | +0.375 | -0.323 | 35.2 | UNRELATED/FAKE |
| recruit | 0.61351 | 否 | 39.9 | +0.431 | +0.051 | 48.6 | UNRELATED/FAKE |
| box-summary | 0.88098 | 否 | 15.5 | +0.122 | +0.152 | 53.0 | UNRELATED/FAKE |
| enemy | 0.91951 | 否 | 12.2 | -0.201 | +0.059 | 69.8 | UNRELATED/FAKE |
| gacha | 0.90251 | 否 | 9.1 | +0.097 | +0.230 | 53.2 | UNRELATED/FAKE |
| box-detail | 0.87844 | 否 | 9.9 | +0.025 | +0.075 | 76.1 | UNRELATED/FAKE |
| lottery | 0.98358 | 否 | 8.2 | +0.048 | -0.021 | 45.3 | UNRELATED/FAKE |
| depot | 0.91728 | 否 | 0.0 | +0.081 | -0.010 | 100.0 | **UNTESTABLE**（见 5.1） |

计数：**STRUCT-RELATED 6 / BORDERLINE 2 / UNRELATED-FAKE 7 / `UNTESTABLE` 1**（门禁仍 2/16 通过）。

### 8.3 与第 4/5 节（`ec440bf`）的差异归因

`1b433bc` 是 `ec440bf` 的**后代，领先 26 个提交**
（`git merge-base --is-ancestor ec440bf 1b433bc` exit 0；反向 exit 1；
`git rev-list --count ec440bf..1b433bc` = 26，反向 = 0；期间 `render.go +111/-20`、`pixel_test.go +203`）。

| 行 | 度量字段 | sim | 归因 |
|---|---|---|---|
| `headhunt` | 五个全变（98.3→29.4、+0.965→+0.206、+0.887→-0.034、4.6→68.8、STRUCT-RELATED→UNRELATED/FAKE） | 0.96476 vs 0.79983 | 输入渲染图不同 |
| `operator` | 五个全变（76.8→25.0、+0.882→+0.076、+0.602→-0.296、28.0→53.5、STRUCT-RELATED→UNRELATED/FAKE） | 0.84554 vs 0.69063 | 输入渲染图不同 |
| `card` | **全同**（77.6 / +0.909 / +0.929 / 23.0 / STRUCT-RELATED） | 0.9601504863664215 vs 0.9601980124080882（首个差异在**小数点后第 5 位**：第 5 位为 `5` vs `9`） | 渲染几乎相同，仅有亚阈值色差 |
| 其余 13 行 | 全同 | 全同 | — |

**两处 verdict 翻转的原因是输入代码不同，不是定义分歧。** 本指标在两个测量点上都是同一把尺子。

### 8.4 两份测量点的计数对照（引用时务必带测量点）

| 测量点 | STRUCT-RELATED | BORDERLINE | UNRELATED/FAKE | 门禁通过 |
|---|---|---|---|---|
| `ec440bf`（本文件第 4/5 节、`report.json`） | 4 | 2 | 10 | 2/16 |
| `1b433bc`（本节、`report-1b433bc.json`） | **6** | 2 | **8** | 2/16 |

PR #5 的源分支是 `consolidate/gg-mainline`，因此**对外应引用 `1b433bc` 这一组，
即 STRUCT-RELATED 6 / BORDERLINE 2 / UNRELATED-FAKE 7 / UNTESTABLE 1**。
（Boss 在 PR #5 表格里数出的 6/2/7/1 与此一致；此前写的「8/16」未把 `depot`
从 FAKE 里拆出来。）

### 8.5 在 `1b433bc` 上追加的三条限定

- `lottery 0.98358` 仍是 `UNRELATED/FAKE`（inkOv 8.2、maxCorr 0.048）—— 判据不变，仍是全表最危险的一行。
- `missing 50.8` 只高出 50 线 **0.8pp**，且 `colCorr 0.442` 本身低于 0.5，判定完全靠 `max(0.610, 0.442) = 0.610 >= 0.5` 撑住 —— **全表最脆弱的一个 STRUCT-RELATED 判定**。
- `headhunt 98.3` 的 `flatBg` 只有 **4.6%**，高分**不是**背景给的，与 `lottery` 机制相反 —— 可作「分数高低 ≠ 结构相关性」的对照样本。
