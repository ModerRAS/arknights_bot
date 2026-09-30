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
- **判定三档**：
  - `STRUCT-RELATED`：`inkOv >= 50` **且** `max(|rowCorr|, |colCorr|) >= 0.5`
  - `UNRELATED/FAKE`：`inkOv < 50` **且** `max(|rowCorr|, |colCorr|) < 0.5`
  - `BORDERLINE`：其余

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
| depot | 0.91728 | 否 | 0.0 | +0.081 | -0.010 | 100.0 | UNRELATED/FAKE |
| headhunt | 0.79983 | 否 | 29.4 | +0.206 | -0.034 | 68.8 | UNRELATED/FAKE |
| operator | 0.69063 | 否 | 25.0 | +0.076 | -0.296 | 53.5 | UNRELATED/FAKE |

计数：`STRUCT-RELATED` 4 / `BORDERLINE` 2 / `UNRELATED-FAKE` 10。

### 5.1 `depot` 的 `inkOv = 0.0` 表示**不可测**，不表示「测出 0」

`depot` 的 `flatBgPct = 100.0`：整张画布是**单一纯色**，众数像素覆盖 100% 画布。
在这种输入下 ink 掩膜按定义必为空集，而空集与任何掩膜的交集恒为空、并集恒等于对方，
所以 `inkOv` **必然**是 0.0 —— 这是定义的退化结果，不是关于 depot 内容的测量结论。

**不可测 ≠ 不合格。** 这一行不得被读成「depot 造假」或「depot 结构无关」：
它只能读成「本指标在纯色画布上没有分辨力」。要判 depot 必须换一个能产生 ink 的测量口径。

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
