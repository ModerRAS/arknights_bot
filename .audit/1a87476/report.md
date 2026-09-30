# 像素门禁报告 —— operator 线事后追认归档

**本文件是一份事后追认（retroactive migration）：它在原始运行发生之后才被创建，所归档的证据产于一个从未被提交的工作区状态。**
原始证据位于仓库之外的 `C:/WorkSpace/Golang/_operator_measure/`，从未被 git 跟踪（该目录甚至不是 git 仓库，`rev-parse --show-toplevel` 退出 128），因此它在产生当时的来源无法被独立复核。
**归档这个动作本身不构成对来源的任何证明。它保存文本，不认证文本。**

本文件**没有重跑任何东西**。下面的数字是转录的，不是复现的。

---

## 一、有效性警告（强制字段缺失，不可删、不可降级为脚注）

`content_mask_overlap` 在本文件所有 16 个条目上都是 `null`，并且**不是省略**。

原因（本次自行复核，非照抄 card 线）：`src/ggrender/pixel_test.go` 只导出 `similarityNormalized`、`gatePassed`、`TestGGPixelParity`、`TestGGPixelParity_Negative`、`TestGGPixelParity_Negative_EmptyBBoxGate` 五个函数。
实测 `grep -niE 'overlap|mask|iou|structur' src/ggrender/pixel_test.go` **退出 1 且输出为空**；阳性对照 `printf 'x overlap y\n' | grep -c overlap` 返回 1，证明尺子本身可用。
要算出它就必须改评分代码，**禁止，且不在范围内**。

后果：本文件中的 similarity 分数**本身并不足以**证明任何渲染在结构上对应冻结的 Playwright 基线。在这一 UI 家族里内容掩膜重叠率低于 50% 即"结构无关"，而高相似度可能只是大面积深色背景带来的假象。

---

## 二、归属（归属与有效性分开陈述，互不暗示）

| 项 | 值 |
|---|---|
| 被审计提交 | `1a87476eaacdb98b341c2ca90fbdb69c014f59ab` |
| tree | `82cd0db81978c8ec959ad67eaf29dc6e6393ac63` |
| 同 tree 前身 | `2be9c2cd78c7934871f0bf4e2449cb4d129bd991`（与上者**逐字节相同**，仅 message 不同） |
| 空提交更正 | `b96bf89f5778c71d672ec02ba1eeb8bad15b0463`（tree 不变，记录本文件所述矛盾） |
| harness blob | `src/ggrender/pixel_test.go` = `33a55e418e5d993e1233fbb40995641463c00dcc` |
| 字体 blob | `assets/font/NotoSansHans-Regular.ttf` = `0c4e8e94d2472d9386990aa2824820811056c9f1`（**已跟踪**） |

`1a87476` 与 `b96bf89` 必须与本文件一起读；本仓库关于该矛盾**只有一种说法**，即 `b96bf89` 的 message 与本文件，二者一致。

**harness 版本已核对且一致**：`pixel_test.go` 的 blob 在 `2be9c2c`、`1a87476`、`ec440bf` 三者完全相同，因此归档日志里引用的每一个行号都精确落在这一版 harness 上。

**目录名是指针，不是证据。** `.audit/1a87476/` 这个名字**不主张**这些测量由 `1a87476` 产出，归属问题也不在这个名字里解决。真正承载归属的是会被读到的那几处：`provenance: retroactive-migration`、`original_path`、本文件第一句。

**产生状态无从归属（纯归属陈述，两条）：**

1. operator 这份证据产于**仓库之外的未提交状态**——`_operator_measure/` 连被 gitignore 覆盖都算不上，它根本不是 git 仓库（`git rev-parse --show-toplevel` 退出码 128）。证据文件写于 2026-09-29 21:40–22:11，而 `2be9c2c` 提交于 22:12:04。因此它**不是 `1a87476` 的同期产物**，`provenance: retroactive-migration` 说的正是这件事。
2. 同理，card 线那份旧报告产于 9 月 29 日的未提交工作区、9 月 30 日才固化为 `0674ab4`，**无法归属到任何 commit**。两条线在这里是同一个模式。

> ⚠️ **上面两句不许顺带读成"所以分数不可信"。**
> 那是**有效性**，独立的一件事。归属（谁的状态产出了它）与有效性（这些数字有无意义）分开陈述，任何一方都不许暗示另一方。

---

## 三、失败的自检（证据自带，非事后评判）

`step4-region-breakdown.json` 自带一条断言：

> *"recomputed here from the PNGs; must equal the harness numbers in the reports"*

**这条自检未通过。** breakdown 报 `0.6996612917574437` → `0.6989360180646332`，harness 报 `0.6996090831517792` → `0.6988968532135076`，在第 5~6 位小数就分岔。这是证据内部的一条失败自检，原样保留在 `diagnostics` 里，且它直接决定区域表该怎么读。

---

## 四、两套测量与各自权威

| 来源 | before | after | delta |
|---|---|---|---|
| harness（`step4-before-report.json` / `step4-after-report.json`） | `0.6996090831517792` | `0.6988968532135076` | `-0.0007122299382715802` |
| breakdown（`step4-region-breakdown.json`） | `0.6996612917574437` | `0.6989360180646332` | `-0.0007252736928105019` |

**哪一套对什么权威：**
- 标题行的 `operator similarity 0.69961 -> 0.69890` —— **harness 那一套**。
- 区域表各行与 `total` 行 —— **breakdown 那一套**。

绝对误差口径上两套相差 **28,738**（harness 数字反推出的 total delta 为 1,569,185；breakdown 记的是 1,597,923）。

> ⚠️ **total 行本身没有算错，方向不要搞反。**
> 区域表三行相加 1,592,666，与 breakdown 的 `attributed_delta` **完全一致**；残差 5,257 正是该文件自己命名的 `unattributed_delta`。
> 矛盾在于 message **把两套测量混用**，而不在于 total 算错。读者若得出"total 也有问题"的结论，是推理方向反了。

---

## 五、operator 分数轨迹（过程记录，未整理）

证据里出现过的每一个不同的 operator 相似度，按 mtime 排列，**未折叠、未整理**：

| 来源文件 | operator similarity |
|---|---|
| `before.json` | `0.6906304752178649` |
| `after.json` | `0.6996090831517792` |
| `step2_after.json` | `0.697063611111111` |
| `step4-before-report.json` | `0.6996090831517792` |
| `step4-after-report.json` | `0.6988968532135076` |
| `step4-region-breakdown.json`（breakdown 口径） | `0.6996612917574437` → `0.6989360180646332` |

记录这一节是因为：把它压成单一 before/after 会掩盖分数在本次会话中**经历了若干中间状态**，而 `1a87476` 引用的只有最后一对。

**文件名有误导性，未做整理**：`after.json`（21:41）并不是 B1 改动的"after"，它是 step 4 的 before 状态。名字按原样保留。`before.json` 的 `0.6906304752178649` 是 B1 之前的 operator 分数，其它工作树未跟踪的 `tmp/pixel-compare/report.json` 里也有这个值，可交叉核对。

---

## 六、未验证假设（**不是结论**）

**假设：** `similarityNormalized` 的等效分母既不等于朴素的 `w*h*4*255`，在 before/after 两图之间也不相等，提示它跳过了某些像素。

用 breakdown 的 `total_abs_error` 反推 harness 相似度，得到两图各自约 2.20282e9 的等效分母，朴素值为 2.2032e9，且两者不相等。

**状态：未验证。本归档没有检验它。** 记录它是因为它是 28,738 缺口的一条活线索，不是因为它解释了这个缺口。**不得把它当作解释来引用。**

---

## 七、跨机可复现性缺口（已知缺口，非场景缺陷，不进分数表）

operator 场景的字体经由 `src/ggrender/helpers.go` 的 `FontCandidates` 回退链加载：先试仓库内已跟踪的 `assets/font/NotoSansHans-Regular.ttf`，失败则退到 `C:/Windows/Fonts/msyh.ttc`、`simhei.ttf`、`msyh.ttf`。

在跟踪字体加载失败的机器上会**静默换用不同度量的字体**，而本场景是文字密集的 9 项中文面板，跨此类机器分数不可比。

**本归档不确立**：这些数字在任何其它机器上可复现；也**不确立**在本机可复现。什么都没有重跑。

---

## 八、复现命令（写死）

复现 **harness** 那一套：

```bash
git worktree add --detach C:/WorkSpace/Golang/_audit-repro-1a87476 1a87476eaacdb98b341c2ca90fbdb69c014f59ab
cd C:/WorkSpace/Golang/_audit-repro-1a87476/src
go test ./ggrender/ -run TestGGPixelParity -v
# 负对照（同时归档为 negative-run.log）:
go test ./ggrender/ -run TestGGPixelParity_Negative -v
```

复现 **breakdown** 那一套：**做不到**。`step4-region-breakdown.json` 由一次临时的 python 运行从 PNG 重算绝对误差得出，**没有任何已提交的脚本**，因此无法从一条已提交的命令复现——这正是它的自检能够静默失败的原因之一。

清理：`git worktree remove C:/WorkSpace/Golang/_audit-repro-1a87476`

预期退出码 1：**FAIL 是诚实的红灯**（16 个场景中 14 个低于 0.99 门禁），不是 harness 错误。

> **警告**：重跑**不会**复现 breakdown 那一套，也不应期待会复现——breakdown 用的是不同分母（见第六节）。上面的命令只复现 harness 那一套。

---

## 九、门禁来源

`src/ggrender/pixel_test.go` 的 `gatePassed(sim float64, bbox [4]int) (bool, string)`，阈值 0.99。被审计提交未触碰它。**引用请用函数名，不要用行号**（行号每次改动都会漂）。该函数同时拒绝零面积 bbox，因此 harness 在结构上已无法给出 1.0 分。

本归档未修改任何评分代码。

---

## 十、分数表（全精度，未四舍五入；状态 = after，即被审计提交对应的状态）

16 个场景，2 过 14 不过，`content_mask_overlap` 全部为 `null`（原因见第一节）。逐条全精度值见 `report.json` 的 `scenes` 数组。

operator = `0.6988968532135076`，`passed=false`，`bbox=[0,0,1799,1199]`。**门禁未通过。**

---

## 十一、归档清单（已收 / 未收）

**已收 15 / 15，无一排除。** 预筛**未发现**任何 PNG、图片或二进制：全目录 15 个文件、0 个子目录、0 个符号链接，逐文件 NUL 字节扫描 + `file(1)` + 扩展名扫描三重确认全为文本。4 个 `.log` 作为文本过程记录收下（它们是 `go test -v` 的原始 stdout）。

预筛方法：`find` 全目录枚举（不按文件名猜）→ 逐文件 NUL 字节扫描（python 实现，locale 无关；阳性对照已自检）→ `file(1)` → 图片/归档扩展名扫描。

**未整理。** 过程记录逐字节嵌入，包含其中的错误假设与自我纠正——那正是它有价值的部分。没有摘要化、没有为美观重排、没有删减。

---

## 十二、operator 分数全表（六个值，逐个列出；本节为权威表）

本节是六个值的**权威表**。第五节列的是同一批数据，只少「时间戳 / 所属测量套 / 是否被 `1a87476` 引用」三列；两处若有出入，以本节为准。

| # | 数值 | 源文件名 | 时间戳 | 所属测量套 | `1a87476` 是否引用 |
|---|---|---|---|---|---|
| 1 | `0.6906304752178649` | `before.json` | 2026-09-29 21:40:18 | harness | 否 |
| 2 | `0.6996090831517792` | `after.json` | 2026-09-29 21:41:23 | harness | 数值相同，但 message **未指明**取自本文件 |
| 3 | `0.697063611111111` | `step2_after.json` | 2026-09-29 21:47:23 | harness | 否 |
| 4 | `0.6996090831517792` | `step4-before-report.json` | 2026-09-29 22:04:35 | harness | **是**，作为 before |
| 5 | `0.6988968532135076` | `step4-after-report.json` | 2026-09-29 22:08:08 | harness | **是**，作为 after |
| 6 | `0.6996612917574437` → `0.6989360180646332` | `step4-region-breakdown.json` | 2026-09-29 22:10:36 | **breakdown** | 不作为 similarity 引用；但区域表与 total 行出自此文件 |

**5 个不同的数值，6 次出现。** 六个必须逐个列：任何顺着 `1a87476` 那条 message 去查的人，只能看到六个答案里的一个，而且看不出另外五个为什么存在。哪两个属于哪一套测量，已在上表逐行标明。

### 第 2 行与第 4 行：同值不同来源（已查清，不得合并成一行）

两行的 `similarity` 逐位相同，但**不是同一测量被重复记录两次，而是同一次测量被两种格式的快照各记了一遍**：

- `hashNew` 两者**完全相同**：`e9cda49d2ac630d2ae7472c4343389b35e7abe53152857ccfc5dcbdb0b8e3f2a`，即同一张渲染图
- 文件不同：`sha256` 分别为 `07dbbced…` 与 `8282da9b…`，相隔 23 分 12 秒
- schema 不同：`after.json` 的 operator 条目缺 9 个字段（`bbox` / `diffPath` / `format` / `hash` / `hashOld` / `newPath` / `oldPath` / `scale` / `score`），`step4-before-report.json` 是完整 harness 格式；两者共有且相同的仅 `hashNew` / `height` / `passed` / `scene` / `similarity` / `width` 六个

**因此这个重复不是两次独立测量，互相不构成佐证。**

由此得到一条一般性结论：**一个裸数值承载不了来源语义。** `0.6996090831517792` 这串字符无法告诉读者它出自哪个文件、哪种 schema、哪个时刻。实测两两组合后，`after.json → step4-after-report.json` 与 `step4-before-report.json → step4-after-report.json` **都**给出 `-0.0007122299382715802`，所以 `1a87476` 那条 `0.69961 -> 0.69890` **无法区分**自己引用的是哪一份。

### 命名误导（与 `b96bf89` 同一条线索，不是两套说法）

`after.json`（21:41:23）**不是** B1 改动之后的分数——它比 `step2_after.json`（21:47:23）还早 6 分钟，实际是 **step 4 的 before 状态**。文件名与内容相悖。

文件名**一律未改**：改名会毁掉与产生时点的对应关系，而那正是本节要说明的东西。归属由本文件第一句、`provenance: retroactive-migration` 与 `original_path` 承载，不由文件名承载。

---

## 十三、通用规矩：裸数值承载不了来源（第十二节的泛化）

这是一条**可复用的通用规矩**，不是 operator 线的私事。

> **裸数值承载不了来源。任何引用数值的地方，必须同时给出来源文件 + 时间戳 + `hashNew`。**

**理由（实测，非推演）：** 两份不同文件——`after.json` 与 `step4-before-report.json`——**都能**产生 `-0.0007122299382715802`，因此 `1a87476` 那条 message **无法区分自己引的是哪一份**。

所以：**看到同一个数出现两次，不能推断“测了两次都一样”。** 它可能只测了一次，被记了两遍。

**判别方法：比 `hashNew`，不是比数值。** 数值相同只说明结论相同；`hashNew` 相同才是同一次测量。

**推论：全精度值不比短数字更权威。** `-0.0007122299382715802` 看起来比任何短数字都更可信，**它只是恰好长得一样**。小数位多不等于证据强，只等于它恰好没在这两次测量之间变过。

### 它属于哪一族：三层是同一句话

**指针不是证据。** 本报告里同一个毛病以三种形态各出现过一次：

| 层 | 指针是什么 | 它承载不了什么 | 本报告对应处 |
|---|---|---|---|
| 一 | **目录名** `.audit/1a87476/` | 承载不了**归属**——名字不主张这些测量由 `1a87476` 产出 | 第二节 |
| 二 | **文件名** `after.json` | 承载不了**时点**——它比 `step2_after.json` 还早 6 分钟，实为 step 4 的 before | 第十二节「命名误导」 |
| 三 | **裸数值** `0.6996090831517792` | 承载不了**来源**——出自哪个文件、哪种 schema、哪个时刻，都读不出来 | 本节 |

三层是同一句话：**指针不是证据。** 每一次都得靠会被读到的东西承载——分别是 `provenance` / `original_path` / 报告首句、`mtime` / `hashNew`、以及本节的"来源 + 时间戳 + `hashNew`"三件套。

**与第十二节的关系：** 第十二节是这条规则的一个具体实例（6 个值那张表，逐行给了来源文件、时间戳、所属测量套、是否被引用）；本节是它的泛化。两者是同一件事的两种粒度，**不构成两个说法**——第十二节每一行都满足本节的要求。

---

## 声明

本归档**没有**修改 `fix/operatorinfo-fields`、`1a87476`、`b96bf89` 中的任何一个，也没有修改 `src/ggrender/pixel_test.go` 或任何评分代码。它是一个独立分支 `archive/operatorinfo-audit-20260930`（基于 `b96bf89`）上的纯新增文件提交。它**没有** force、**没有**合并、**没有**删除任何东西，也没有把任何图片或二进制放进仓库。
