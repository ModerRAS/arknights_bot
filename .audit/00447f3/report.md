# 像素门禁报告 —— 事后追认归档

**本文件是一份事后追认（retroactive migration）：它在被归档的那次运行之后才被创建，
不是与该次运行同期存在的证据。**

被归档的对象是**本次审计在 `00447f3` 上重新复现出来的** harness 报告，
而不是迁移前既已存在于 `tmp/pixel-compare/report.json` 的那份旧文件（旧文件见
「旧 `tmp/` 报告」一节，本次**不迁移**）。

> **`tmp/` 从未被任何 commit 跟踪**（`.gitignore` 的 `tmp/` 规则覆盖该路径），
> 因此该路径上的内容在历史上**无法被独立复核**。
> **归档这个动作本身不构成对来源的任何证明。**

---

## 一、有效性警告（强制字段缺失，不可删、不可降级为脚注）

**当前 harness 不计算内容掩膜重叠率。** `report.json` 中每个场景的
`content_mask_overlap` 字段**显式为 `null`**，并带 `content_mask_overlap_reason`。
键**没有被省略**——省略与"明确记为算不出"是两回事，省略会让下一个人以为没这回事。

依据：`src/ggrender/pixel_test.go` 只定义三个函数
`similarityNormalized`、`TestGGPixelParity`、`TestGGPixelParity_Negative`，
`git grep -niE 'overlap|mask|iou|structur' <sha> -- src/ggrender/pixel_test.go`
在 `00447f3` 与 `ec440bf` 两处均 **exit 1、输出为空**。

本 UI 家族存在已知反例：**两张结构上完全无关的页面，因画布约 70% 是相近暗色背景，
也能拿到约 `0.959`，而内容掩膜重叠率只有约 28%。判别法：内容掩膜重叠 < 50% 即结构无关。**

因此，**本报告中的 `similarity` 单独不足以证明任何渲染与冻结 Playwright 基线在结构上对应。
`card = 0.9601980124080882` 正属于这一类分数。
它的"确定性"已由本次复现验证；它的"有效性"未经检验，目前无法排除它正是一个假高分。**
本归档中任何数字都不得被引用为"card 与基线相符"的证据。
补齐该指标需要改评分代码，属红线，不在本次范围内。

---

## 二、三态对照：`lottery` 的分界点落在 `0674ab4`

全部为**实测**（各自干净 detached worktree 重跑），非推断。

| commit | parent | card score | card hashNew | lottery score | lottery hashNew |
|---|---|---|---|---|---|
| `00447f3` | `871515b` | `0.9601980124080882` | `26fd5664a1c2` | `0.9835808689865397` | `37a15ba596de` |
| `ec440bf` | `00447f3` | `0.9601980124080882` | `26fd5664a1c2` | `0.9835808689865397` | `37a15ba596de` |
| `0674ab4` | `ec440bf` | `0.9601980124080882` | `26fd5664a1c2` | `0.9839251621741464` | `3375f3edff2f` |

**分界点：`0674ab4`。** `00447f3` 与 `ec440bf` 的报告**逐字节相同**
（sha256 均为 `329e6cf`）——`ec440bf` 只改了 `src/ggrender/pixel_test.go` 的门禁逻辑，
未动任何 renderer 或 asset，渲染输出不变。`lottery` 的差异**只在 `0674ab4` 出现**。

**card 在三态下完全一致**（实测，非推理）：`hashNew=26fd5664a1c2…`、
`score=0.9601980124080882`。
**后果很重要：card 这一行无法区分这三棵树，因此不得用 card 来给一份报告定代。**

---

## 三、归属

| 项 | 值 |
|---|---|
| 产出 commit | `00447f336083815486db3610957e63c41b3c28fa` |
| 标题 | `feat(gg): honest card rebuild similarity=0.96020` |
| 提交时间 | `2026-09-02T16:36:39+08:00` |
| 依据 | checkout 到该 commit 重跑 harness。报告自身**无 commit sha 字段、无时间戳字段**，归属完全建立在本次复现上。 |

**确定性**：`00447f3` 上独立跑了两次 harness，两次产出的 report.json **逐字节相同**。

---

## 四、复现命令（分状态写死）

`00447f3`（本次归档对象）：

```
git worktree add --detach C:/WorkSpace/Golang/_audit-repro-00447f3 00447f3
cd C:/WorkSpace/Golang/_audit-repro-00447f3/src
go test ./ggrender/ -run TestGGPixelParity -v
```

`ec440bf`（对照态）：

```
git worktree add --detach C:/WorkSpace/Golang/_audit-repro-ec440bf ec440bf
cd C:/WorkSpace/Golang/_audit-repro-ec440bf/src
go test ./ggrender/ -run TestGGPixelParity -v
```

`0674ab4`（对照态，旧报告的真正来源态）：

```
git worktree add --detach C:/WorkSpace/Golang/_audit-repro-0674ab4 0674ab4
cd C:/WorkSpace/Golang/_audit-repro-0674ab4/src
go test ./ggrender/ -run TestGGPixelParity -v
```

清理：

```
git worktree remove C:/WorkSpace/Golang/_audit-repro-00447f3
git worktree remove C:/WorkSpace/Golang/_audit-repro-ec440bf
git worktree remove C:/WorkSpace/Golang/_audit-repro-0674ab4
```

**实测（`go1.26.6 windows/amd64`）：** 三次 `go test` 退出码均为 **1**。
这是**预期结果**——16 场景中 14 个低于 `0.99` 门禁（honest red），不是 harness 故障。
三次均为 **16 场景 / 2 个 passed**（`calendar`、`state`）。
三个临时 detached worktree 均已移除。

---

## 五、旧 `tmp/` 报告（**不迁移**，仅记录关系）

迁移前已存在于 `tmp/pixel-compare/report.json` 的那份文件**本次不迁移**，
它与本次复现的关系如下（实测）：

- **16 个场景中 15 个逐位吻合**（含 card 全部字段）。
- **1 个不一致：`lottery`。**

| 项 | 旧 `tmp/` 报告 | 本次 `00447f3` 复现 |
|---|---|---|
| `lottery` score | `0.9839251621741464` | `0.9835808689865397` |
| `lottery` hashNew 前缀 | `3375f3edff2f` | `37a15ba596de` |
| 绝对差 | `0.00034429318760664795` | — |

**成因：代码状态差异。** 旧报告**逐字节等于 `0674ab4` 的复现结果**
（sha256 均为 `dc983be`），而不等于 `00447f3` 的复现结果，
故它产生于 `ec440bf` 之后的状态。`0674ab4` 相对其父的唯一实质改动是 `func RenderLottery`
（改用新增的 `assets/font/msyh-regular.ttf`，`setFontP` 取代 `setFont`，并移动标题）。
该 commit 尚不含这套 msyh 改动，所以 `00447f3` 复现的 `lottery` 行与之不同，
而未被该改动触及的 card 行逐位吻合。

**（未验证假设，勿当结论）** 字体解析回落到系统 `C:/Windows/Fonts/`
（而非未入库的 msyh 资源）可能也是叠加因素之一。上面的差异已被代码改动**完整解释**，
且在各自干净 worktree 中**确定性复现**，因此它**不是**环境抖动。
字体回落对旧报告是否另有贡献**未经测试**，在查清前不得写成结论。

---

## 六、跨机可复现性缺口（已知缺口，非场景缺陷，不进分数表）

**任何依赖字体的场景都不可跨机复现。** 本次复现中字体栈回落到系统字体
`C:/Windows/Fonts/`（见 `src/ggrender/helpers.go` 中列出的 `msyh.ttc` / `msyh.ttf`），
而不是任何随仓库分发的 msyh 资源。系统字体集不同的机器，会为每个依赖字体的场景
渲染出不同像素。**本文件的数字只对本机有效。**
该项是**已知缺口**，不是某个场景的缺陷，也不是门禁输入。

---

## 七、门禁来源

判定收敛在 `src/ggrender/pixel_test.go` 的
**`func gatePassed(sim float64, bbox [4]int) (bool, string)`**。
（按**函数名**引用；行号每次改动都会漂，不是稳定引用。）

它做两件事：

1. **零面积 bbox 守卫** —— 在比较分数之前先拦下宽度或高度 `<= 0` 的 bbox，
   使一次"没有比过任何差异像素"的空比较无法空洞通过（由 `ec440bf` 引入）。
2. **阈值** —— `sim < 0.99` 判失败。

**它不计算内容掩膜重叠率**，因此不覆盖结构性对应问题（见第一节）。

---

## 八、分数表（全精度，未四舍五入）

共 **16** 个场景，`passed` **2** 个 / 失败 **14** 个。
`content_mask_overlap` 全部为 `null`（见第一节）。

| scene | score | bbox | hashNew | content_mask_overlap | passed |
|---|---|---|---|---|---|
| `base` | `0.9589964211295751` | `[0, 0, 1664, 917]` | `1aafad90d714` | `null` | `false` |
| `box` | `0.8250372524631744` | `[0, 0, 1049, 535]` | `55d299b08d3a` | `null` | `false` |
| `box-detail` | `0.8784374903269025` | `[0, 0, 721, 278]` | `92d22b58aade` | `null` | `false` |
| `box-summary` | `0.8809750314141331` | `[0, 0, 1349, 722]` | `45e63a36e647` | `null` | `false` |
| `calendar` | `0.99432366536721` | `[0, 0, 2863, 1619]` | `e7939ca718d0` | `null` | `true` |
| `card` | `0.9601980124080882` | `[0, 0, 1279, 711]` | `26fd5664a1c2` | `null` | `false` |
| `depot` | `0.9172816240959263` | `[0, 0, 1247, 233]` | `5de23d50dc14` | `null` | `false` |
| `enemy` | `0.9195071737682478` | `[0, 0, 983, 476]` | `bc7b5af135de` | `null` | `false` |
| `gacha` | `0.9025141429411271` | `[0, 0, 1499, 1322]` | `e43865641c03` | `null` | `false` |
| `headhunt` | `0.7998304534119018` | `[0, 0, 1048, 575]` | `b109c04ed677` | `null` | `false` |
| `help` | `0.839432182957432` | `[0, 0, 989, 2048]` | `e984bc7e9ffa` | `null` | `false` |
| `lottery` | `0.9835808689865397` | `[0, 0, 1472, 1666]` | `37a15ba596de` | `null` | `false` |
| `missing` | `0.8761380684114441` | `[0, 0, 1049, 535]` | `01b88d9597b9` | `null` | `false` |
| `operator` | `0.6906304752178649` | `[0, 0, 1799, 1199]` | `b50d020fb104` | `null` | `false` |
| `recruit` | `0.6135057512531993` | `[0, 0, 1349, 533]` | `9ef36c3b0894` | `null` | `false` |
| `state` | `0.9909337226595012` | `[32, 32, 1079, 507]` | `79f31688bbe0` | `null` | `true` |

> 上表为**搬运**，数值逐位取自本次复现产出的报告，未做任何四舍五入或"修正"。

---

## 声明

- `tmp/` 下的原始 harness 输出**从未被 git 跟踪**，迁移/归档前**无法被独立复核**；
  **归档动作本身不构成对来源的证明。**
- 本归档**不含任何图片或像素数据**；冻结基线图未入库。
- 未修改 `pixel_test.go` 或任何评分代码；未实现内容掩膜重叠率指标；
  未加载基线图作为渲染背景；未跨工作树引用资源参与渲染或评分。
- 位置引用一律使用函数名。


---

## 十一、归属（补记：目录名不承载归属，故在此写足）

以下两条是**纯归属陈述**。它们只回答「这份东西从哪来」，
**不蕴含任何关于分数可信度的推论**。归属与有效性是两件独立的事，
下文第十三节单独陈述有效性，此处不借归属暗示有效性，也不反过来。

**其一：旧 `tmp/` 报告无法归属到任何 commit —— 这是事实，不是待查项。**
card 线那份迁移前已存在于 `tmp/pixel-compare/report.json` 的报告，
产生于 **2026-09-29 的一个尚未提交的工作区状态**；那个状态直到 **2026-09-30** 才被固化成
`0674ab4`（创建于 `2026-09-30T09:03:05+08:00`，比该报告晚一天）。
即：报告产生时，那套改动**还不是 commit，只是工作区里的未提交编辑**。
在 `0674ab4` 上重跑 harness 可逐字节复现该报告（sha256 一致），
在 `00447f3` 与 `ec440bf` 上则不能。
**因此它的产生状态从来不是任何 commit**——它早于承载其内容的那个 commit 的存在。
给这份旧报告按 commit 命名不是「稍有偏差」，而是原则上不可能。

**其二：本归档对象是本次新复现的报告，不是迁移前那份旧文件。**
归档对象是本次审计在 `00447f3` 的**干净 detached worktree** 中重新复现出来的报告。
**旧文件不迁移。** 二者的逐场景关系记录在 `report.json` 的 `legacy_tmp_report_section`。

---

## 十二、逐场景复现矩阵（16 个场景全列）

对照基准 = `00447f3` 的复现结果。三个状态各自在**干净 detached worktree** 中重跑 harness，
逐场景读取该次产出的 `hashNew`；任一状态的 `hashNew` 与 `00447f3` 不同即记为分歧。
**无任何场景被跳过，无任何值留空或填推测值**；`passed=true` 的 `calendar` 与 `state`
与失败场景走同一套记录，未因已过门禁而略过。

| 场景 | 复现一致的状态 | 不一致的状态 |
|---|---|---|
| `base` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `box` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `box-detail` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `box-summary` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `calendar` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `card` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `depot` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `enemy` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `gacha` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `headhunt` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `help` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `lottery` | `00447f3`, `ec440bf` | `0674ab4` |
| `missing` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `operator` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `recruit` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |
| `state` | `00447f3`, `ec440bf`, `0674ab4` | —（无） |

**结果：16 个场景中 15 个在三态下完全一致，仅 1 个（`lottery`）随代码状态位移。**

- **不随状态位移（15 个）**：`base`, `box`, `box-detail`, `box-summary`, `calendar`, `card`, `depot`, `enemy`, `gacha`, `headhunt`, `help`, `missing`, `operator`, `recruit`, `state`
- **随状态位移（1 个）**：`lottery`

### 这是阴性结果，不是调查失败

**除 `lottery` 外没有发现第二个位移场景。** 这是一个**阴性结果**——
即「查了 16 个，没有第二个」，而不是「没查出来」。

**探测器自检：** 同一套判据在 `lottery` 上**确实报了分歧**，
说明探测器在有差异时确实会响；因此其余 15 行是**阳性结果**（查过且一致），
而不是「探测器坏了什么都没找到」的空结果。
另外 `00447f3` 与 `ec440bf` 的整份报告 sha256 相同，
这说明整份文件的哈希比对对内容是敏感的。

### 与已有结论一致，未被推翻

- `boundary` 不变：`lottery` 的分歧**只**出现在 `0674ab4`；`00447f3` 与 `ec440bf` 在报告上不可区分。
- `card_invariance` 不变：card 在三态下完全一致，
  **"the card row alone CANNOT distinguish these trees and must not be used to date a report."**

`lottery` 的位移已被完整归因于 `0674ab4` 相对其父的唯一实质改动 `func RenderLottery`；
**其余 15 个场景对该改动完全不敏感**，这从渲染侧印证了 `func RenderLottery` 的改动范围
确实只覆盖 lottery 场景。

---

## 十三、有效性与归属是两件事

矩阵回答的是**确定性**：某个场景在哪些状态下能复现出同一结果。

它**不回答有效性**：这些分数是否真的证明渲染与冻结 Playwright 基线在结构上对应。
有效性上的已知缺口（harness 不计算内容掩膜重叠率、`similarity` 单独不足以证明结构对应）
在**第一节**单独陈述，与本节无关，也**不因**本节的三态复现结果而减轻或加重。
详见第一节与 `report.json` 的 `mask_overlap_metric`。
