# .audit/gg-residual-diagnosis-20261002

本目录归档 **gg 残差诊断线（tasks 1–12）** 的全部证据与尺子。

## 归档规则（Boss 裁定）

1. **本目录全部条目一律 `provenance: retroactive-migration`** —— 这批东西**全部在仓库外产生**，没有例外。
2. **一律不改写归档内容物的内部字段。** 每个内容物旁有 `<原文件名>.meta.json` **边车**，承载归档层的三项声明。
3. **原始文件逐字节不动。**

### 为什么用边车、而不是改写内部 `provenance` 字段
两个问题，两个都真，**答案分属不同载体**：

| 问题 | 答案 | 载体 |
|---|---|---|
| 这个仪器**写出来时**是什么来源？ | 在量测树里当场量的 | **文件内部字段**（原样保留） |
| 这个文件**被搬进 `.audit/`** 是什么性质？ | 仓库外产生的事后追认 | **边车 + 本文件** |

改写内部字段 = 为满足第二个去毁掉第一个，**且会破坏已逐字节核验过的产物**。
边车与文件同名同目录、机器可读、不会与文件错位（而 README 里一张 30 行的表会因 `measured_at` 逐份不同而必然漂移）。

## ⚠️ 引用本目录任何数字前必读：`measured_at` **逐份不同**

**不要假设本目录的产物都出自同一个基底。它们不是。**

| 内容物 | 基底 |
|---|---|
| `task1-instrument-compare.md` | **跨三棵树**：`arknights_bot-satori@b946661` + `arknights_bot-satori-yoga-skia-go@3704262` + `arknights_bot-gg-card-atomic@d939c38`（无单一基底；多树只读调查） |
| `task2-satori-audit.md` | `arknights_bot-satori@b946661`（父 `eb55d5a`）—— **不在量测树上** |
| `task3-render-reuse.md` | **跨树**：`arknights_bot-gg-card-atomic@d939c38` + `arknights_bot-satori-yoga-skia-go@3704262` |
| `task4` … `task12` 及全部 `instrument.json` / `controls.json` / `*.py` | `arknights_bot-measure-align@measure/align-ceiling` `4584759` |
| `SYNTHESIS.md` | 综合稿，跨上述全部 |

⚠️ **一个规整但伪造的出处，比留空更坏。** 逐份填的理由见各边车。

### `single_base` 字段：机器可查，勿只读散文

每个边车除 `measured_at`（人类可读的自由文本）外，**另有一个结构化字段 `single_base`**：

| 值 | 含义 | 命中 |
|---|---|---|
| `"none"` | **该轮跨多棵树、无单一基底** | `SYNTHESIS.md` / `task1` / `task3` |
| `"b946661"` | 单树单基底（`arknights_bot-satori`） | `task2` |
| `"4584759"` | 量测树单基底 | task4–task12 及全部仪器 |

```bash
# 「哪些产物没有单一基底」——一条 grep，不解释
grep -rl '"single_base": "none"' .audit/
```
⚠️ **不要只 grep `measured_at` 里的中文「无单一基底」**：那是自由文本，换个措辞就 grep 不到。**机器消费一律读 `single_base`。**

## ⚠️ 本目录出现过两个已知作弊提交的 SHA —— **它们是引用，不是采纳**

`task1` / `task2` / `task3` 的文本里出现 `8cbffef` 与 `641988f`，**那是审计结论里对它们的讨论与判定，不是本目录采用了它们**。
**不写这句，下一个人读到两个作弊提交出现在归档里，会以为这份归档带毒。**

## ⚠️ 本目录内的脚本：**只被写入，不被执行**

**不要在本目录内执行任何 `.py`。** 项目规矩：归档目录内执行脚本会就地生成派生产物（`__pycache__` 等），**而「产生事故的那条命令还在文档里，下一个人照着跑会复发」**。
这些脚本是**尺子的定义**，不是待运行的东西。要跑请回原路径 `C:/WorkSpace/Golang/arknights_bot-measure-align/tmp/residual/<sub>/`。

## 零二进制（红线）

只收 `.md` / `.json` / `.py`。**基线图（`.jpg` / `.png`）一律不进归档。**
`tmp/residual/` 下的图像仅用于对照，**不是证据的一部分**。

## 被推翻的结论：**原样保留，不删**

以下版本**已被作废或更正**，但**必须保留** —— 它们是「我们一度以为」，**删掉就等于伪造一致**：

- `SYNTHESIS.md` §7.10 —— **像素分类轴关闭预登记，已作废**。作废理由不是「不该关」，而是**绑定条件绑错了对象**（绑在 `PC0` 失败这个**预测的失败原因**上，而实际死因是 `PC1`；**真正在乎的量是「这根轴能不能产出可用的分解」**）。见 P24。
- `SYNTHESIS.md` §7.11 —— **B 类机械判别整条链条（`5→4`、「域外率 0.20」）已撤回**，因分类依据「R1R2 通过 = 真目标」不成立（元素级映射证明 5 个洞中漏画元素为 0）。**`0.20` 不得再被引用。**
- `SYNTHESIS.md` §7.8 —— **旧头条措辞「地板下界比最小缺口低一个数量级以上（0.0042 vs 0.0064）」已废止**：那是拿 `headhunt` 的亏损除 `lottery` 的缺口，**不描述任何场景**。逐场景约束值见该节表格。
- `task8` / `task10` / `task12` —— **阳性对照未通过 ⇒ 分层为「未测」，不交表**。未交表不是失败，是规则生效。
- `common.css` —— **字节数 417（真值）与 391（字符数，被误报为字节数）并存**，见 P26。

## 一处已知分歧：归档时未修（留给 Boss 裁定）

`task12-residual-by-declared-element.md` 写：`teeth8c` → **`elements.json` 未产出（ls exit=2）`**。
**Lead 实测（时间戳）**：`teeth8c.log` 22:18:00 → `elements.json` 22:18:22。
⇒ **闸门确实拒绝了（两次拒绝均已验证、内容不同），但该文件由**随后对照全过的正当运行**产出，而非在拒绝状态下漏出。**
**性质**：把一次瞬时观测写成了机制断言 —— 「读不到 ≠ 不存在」的翻版。**不是造假。**
**归档时不修改该报告原文**（原始文件逐字节不动）；分歧记录于此。**待 Boss 裁定是否要求 Worker 更正。**

## 目录结构

```
SYNTHESIS.md          ← 全部结论 + 常设流程 P1…P27
task1..task12-*.md    ← 各轮报告
NOTE-card-cheat-*.md  ← **不收**（属另一条线的处置记录，见 SYNTHESIS §归档规格）
<name>.meta.json      ← 边车：provenance / original_path / measured_at
byelem/               ← task12 声明元素分解
discriminator/        ← task6  B 类机械判别
jpegfloor/            ← task9  JPEG 编解码噪声地板
mapping/              ← task7  元素级映射
strata/ strata2/ strata3/ ← task8 / task10 / task11 分层（均「未测」）
```

## 这个分支的职责

**保存，不是参与。** 不合并、不改写。

---

## 路径映射表（2026-10-02 追加，供归档报告中的引用定位）

归档报告里的引用写的是**量测树（`arkskadi_bot-measure-align`）中的原始路径**。对应关系：

| 引用中的原始路径 | 归档内相对路径 |
|---|---|
| `.audit/align_ceiling.py` | **`align/align_ceiling.py`** |
| `tmp/align/{instrument,report,controls}.json` | `align/*.json` |
| `tmp/align-coarse/*` | `align-coarse/*` |
| `tmp/align-smoke/*` | `align-smoke/*` |
| `tmp/align-teeth/*` | `align-teeth/*` |
| `tmp/residual/{instrument,ink_regions,controls}.json` | `residual/*.json` |
| `tmp/residual/<sub>/…`（早前已归档） | `<sub>/…` |

⚠️ **`align/align_ceiling.py` 是 `SYNTHESIS.md:827` 那条 `field_semantics_boundary_vs_probe_boundary` 定义的权威原件**（对应引用 `align_ceiling.py:327-328`）。此前该脚本**只被引用、从未归档**，共 11 处引用 / 4 个文件悬空；本次补入后闭合。

---

## 路径映射表 · 补（2026-10-02，52 项批次）

本批按**公式**产生（非手工枚举）：
`{ 源侧 tmp/** 与 .audit/** 的文本文件 ∧ 内容 sha256 不在归档 ∧ 该路径在 git 中未被跟踪 }`

| 引用中的原始路径 | 归档内相对路径 |
|---|---|
| `tmp/resolved/{rects,rects_t2}.json` | `resolved/*.json` ← **task14 的 T1/T2 原始数据** |
| `tmp/resolved/t2/overlay.json` | `resolved/t2/overlay.json` |
| `tmp/residual/report.json` | `residual/report.json` ← **task5 分层主产物** |
| `tmp/residual/residual_probe.py` | `residual/residual_probe.py` |
| `tmp/residual/byelem/derivable_boxes.json` | `byelem/derivable_boxes.json` |
| `tmp/pixel-compare/report.{json,md}` | `pixel-compare/report.*` ← **go harness 主产物** |
| `tmp/<其余子目录>/*.log` | `<子目录>/*.log` |
| `.audit/<name>`（本轮新归档部分） | `audit-src/<name>` |

⚠️ **0 字节文件是本轮最重要的证据之一，已全部归档并配边车。**
`mapping/measure_stdout.log`、`discriminator/measure_stderr.log` 等的空文件，
就是「**闸门在对照未过时没有偷偷产出任何输出**」这句话的物证。
**若按「空文件没用」的直觉跳过，归档就失去了这个判据的存在性证明。**

⚠️ **一项已识别但未归档，等待裁定**：`tmp/resolved/boxdetail.html`（3664 B）——
本批公式选中 53 项，其中此项**未包含在授权的 52 项内**，故暂扣。
