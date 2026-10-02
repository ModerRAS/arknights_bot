---
provenance: measured
measured_at: feat/satori-renderer @ b946661 (parent eb55d5a), worktree C:/WorkSpace/Golang/arknights_bot-satori, generated 2026-10-02T10:35:13Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task2-satori-audit.md
scope: read-only integrity audit. No revert, no rewrite, no commit, no push, no branch delete. Only this file was written.
---

# 诚信审计：`feat/satori-renderer` 两个未推送提交

## 0. 工作树三件套（我实际读过的每一棵）

| 路径 | 分支 | commit | 我做了什么 |
|---|---|---|---|
| `C:/WorkSpace/Golang/arknights_bot-satori` | `feat/satori-renderer` | `b946661` | 全量读：diff、文件全版本、模板/CSS、commit 历史 |
| `C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go` | — | — | **未读**（禁碰树，本任务不需要；未进入、未执行任何命令） |

其余 21 棵工作树（`git worktree list | wc -l` = **22**，含本树）**未读**。我在这些树上没有执行任何命令、没有读任何文件。
需要说明的口径：`git ls-tree` / `git log -S` 读的是**共享对象库**里 `b946661`/`eb55d5a` 这两个 commit 的树对象，不是其它工作树的**工作目录**内容。本次审计没有依赖任何一棵未列出的工作树的文件系统状态。

未推送确认（自测，非转述）：
- `git merge-base --is-ancestor eb55d5a b946661` → **exit=0** ⇒ `eb55d5a` 是父、`b946661` 是子（Lead 派工清单把两条按相反顺序列了，以 Lead 更正后的为准）。
- `git merge-base --is-ancestor b946661 origin/feat/satori-renderer` → **exit=1** ⇒ 确认未推送。
- `git status --porcelain | wc -l` → **0** ⇒ 工作树干净，审计没有留下任何改动。

## 1. 逐取值来源表

方法判据（Lead 给的三分法）：
- 该取值在**模板/CSS 声明里本来就有** → 「声明」
- 模板里**没有**，代码里是硬编码常数，提交信息/代码注释自述来自 ink 剖面 → 「量出来的」
- 与某处声明**恰好一致** → **不能**据此判「声明」，两格写「无法判定」

模板事实（`template/BoxDetail.tmpl`，最后修改 `f24e079`，本审计两个 commit **均未改动该模板** —— `git show --name-only eb55d5a` 只输出 `renderer/components/box-detail.mjs`，`b946661` 只输出 `renderer/manifest-webp.test.mjs`）：
- 模板里**没有任何坐标声明**。`tr` 只有 `box-shadow`（`:15`），`td` 只有 `vertical-align/text-align/white-space`（`:20-24`），`img` 只有 `width: 50px`（`:17-19`）。行高、行首 top、各元素垂直位置全部是浏览器 table 布局的**涌现量**，模板无从声明。
- `assets/css/common.css` 声明 `body { font-family: 'NotoSansHans', serif }`（`:6-9`），**无 `font-size`** ⇒ 16px 是浏览器默认值，非模板显式声明。

| # | 取值 | 来源声明 | 依据（函数/行号 @ commit） | 备注 |
|---|---|---|---|---|
| 1 | row pitch 117px → `ROW_PITCH = 78` | **量出来的** | `box-detail.mjs` 顶层 `const ROW_PITCH = 78` @ `eb55d5a` **第 13 行**；注释来源声明在 **第 8 行** `Element tops measured from baseline ink (device px / 1.5): ... row pitch 117px (78css)`；消费点 `render()` 内 `const y = ri * ROW_PITCH` **第 58 行** | `117/1.5 = 78` 整除，声明的换算链**算术自洽**（已核）。模板无行高声明 → 不可能来自声明。上游「ink 剖面真是 117」无独立产物可核（见 §4-N1） |
| 2 | row1 top 47px → `T.evolve = T.skill = 31.3` | **量出来的** | `const T` 表 @ `eb55d5a` **第 19 行 / 第 22 行**（`31.3 // 47px`）；消费点 `at(COLS[1], T.evolve + y, ...)` **第 64 行** | `47/1.5 = 31.33 → 31.3`，四舍五入到十分位，与 `potential 44`、`equip 43.3` 同一套舍入口径（`65/1.5=43.33`）。模板无 top 声明 → 不可能来自声明 |
| 3 | **name 的 x 坐标** → `left: 60` | **无法判定** | `render()` 内 avatar/name 那一行 @ `eb55d5a` **第 63 行**：`left: 60, top: T.name + y, fontSize: FONT`。**该行没有任何来源注释**（对比第 17-25 行 `T` 表每项都带 `// NNpx` 出处注释） | 提交信息第 6 行自述 `55.3css estimate pending; current 60css`。代码里 55.3 与 60 各自对应哪个轴**提交信息没说清**（见 §3）。模板对该 x 有间接贡献（`img{width:50px}` + 内层 `inline-flex` + `td{text-align:center}`），但**不声明 60 本身**。证据缺的是：60 的出处注释缺失 |
| 4 | potential 66px → `T.potential = 44` | **量出来的** | `const T` 表 @ `eb55d5a` **第 21 行**（`potential: 44, // 66px`）；提交信息第 7 行 `- evolve/skill icons top 47px, potential 66px, equip 65px`；消费点 **第 65 行** `at(COLS[2], T.potential + y, ...)` | `66/1.5 = 44` 整除。模板 `:41` 是裸 `<td><img/></td>`，位置纯涌现 → 不可能来自声明 |
| 5 | equip 65px → `T.equip = 43.3` | **量出来的** | `const T` 表 @ `eb55d5a` **第 24 行**（`equip: 43.3, // 65px`）；提交信息第 7 行同上；消费点 **第 67 行** | `65/1.5 = 43.33 → 43.3`。模板 `:52` 是 `inline-flex column` 包 img+文字，无 top 声明 → 不可能来自声明 |

这条线**第二次**做「measured-geometry rebuild」，第一次是 `1141506 feat(satori/box-detail): measured legacy geometry rebuild, 0.9122->0.9500`。

## 1b. 扩展枚举（Lead 第一轮复核后补做）：文件内全部坐标类取值

Lead 指出我只审了 5 个取值，同文件还有一批同样无出处的坐标常数。下面是**穷举该文件全部坐标类字面量**后的结果 —— 枚举范围 = `renderer/components/box-detail.mjs` @ `eb55d5a` 全文 73 行（C7），穷举方式是逐条 grep 出所有 `COLS` / `HEADER_H` / `left:` / `top:` / `inlineCenters(` 出现处（P3），不是「我看过的那些」。

| # | 取值 | 位置 @ `eb55d5a` | 来源声明 | 依据 | 备注 |
|---|---|---|---|---|---|
| 6 | `COLS = [51.7, 129.3, 184, 291.3, 425.7]`（5 个列中心） | `:11` | **量出来的** | **父版本有明确出处注释**：`eb55d5a~1:6` `// column centers (header th == icon centers): 51.7 / 129.3 / 184 / 291.3 / 425.7`，其上 `:5` 写 `// geometry measured off the frozen Playwright baseline`。**`eb55d5a` 删掉了这行注释、保留了值**（P2 对照） | 本 commit **未改动数值**，但**降低了对它的标注**。真正引入者是 `1141506` |
| 7 | `HEADER_H = 35` | `:12` | **无法判定** | `git log --reverse -S'HEADER_H' -- renderer/components/box-detail.mjs` 只返回 `1141506`（P4，exit=0）⇒ 引入于 `1141506`，**且 `1141506` 的提交信息里没有 35**（该信息只提 col centers / header 16px bold / rows 75 / pitch 53.4,53）。`eb55d5a~1:8` 亦无注释，现版本 `:12` 亦无注释 ⇒ **从引入至今任何版本都没有出处标注** | 模板 `<th>` 行无声明高度 ⇒ 结构上排除「声明」；但也没有任何自述来源 ⇒ 两头都不占 |
| 8 | `left: 3.3`（avatar 行） | `:62` | **量出来的** | 父版本有出处注释：`eb55d5a~1:41` `// operator: avatar 50x50 at (3.3, 7.7) + name 12px left 63.3 v-centered`（P5）。**`eb55d5a` 删掉了这行注释、保留了值** | **但父版本注释自身不可靠**：同一行注释写 `top 7.7`，而父版本代码 `:35` 写的是 `top: 8.3` —— 注释与代码差 0.6px。该行「measured」的自证精度存疑 |
| 9 | inline 组间距 `53.4`（skills） | `:59` | **无法判定** | 注释 `// inline group: n icons of 50px with ~3.4px word space, centered on cx` 在 `eb55d5a~1:12` 与 `eb55d5a:30` **逐字保留**。但该注释只说「~3.4px word space」是**描述性**的，没说是量出来的还是从字体推的 | 结构上**不需要基线**也能得到：模板 `:45/:52` 的 `<div style="display: inline-flex;...">` 是**行内级**盒子，同级之间的 HTML 源码空白会渲染成一个空格（P6 已验证模板里这些 div 之间确有换行+缩进），该空格宽度由 `assets/css/common.css:2-4` 声明的 `NotoSansHans` 的 space advance 在 16px 下决定 ⇒ **声明 + 字体度量**即可推出 |
| 10 | inline 组间距 `53`（equips） | `:60` | **无法判定** | 同 #9 | 同 #9。53 与 53.4 相差 0.4px，注释只解释了 53.4，**53 的 0.4px 差异无任何说明** |

### 1b-1. 一个可验证的、系统性的标注缺口（新发现，比 #3 更硬）

把新文件的文件头注释（`eb55d5a:5-10`）逐条拆开，它其实是一张**三栏的来源登记表**：

| 注释行 | 内容 | 覆盖的来源类别 |
|---|---|---|
| `:5` | `Geometry measured off the frozen Playwright baseline (CSS px @1.5 scale).` | **文件级「量出来的」总声明** |
| `:6` | `The legacy page is a <table> with default 16px font everywhere and` | **声明**（逐字引用模板/CSS） |
| `:7` | `img{width:50px;height:auto}.` | **声明**（逐字引用 CSS） |
| `:8-10` | `row1 top 47px, row pitch 117px (78css), avatar 64px, name ink 88px, evolve/skill box 47px, potential 66px, equip 65px, evolve LV ink 117px, skill LV ink 128px, equip LV ink 120px` | **纵向量（10 个，全部是 top / ink top）** |

**结论（P1 实测，可复核）**：`avatar 64px` 与 `name ink 88px` 都是 **top**（`T.avatar = 42.7 = 64/1.5`、`T.name = 56` ← `name ink 88px`），**整个 `:8-10` 枚举里没有任何一个 x 坐标**。于是：

- **该文件会为「声明」来源的取值做标注** —— `FONT = 16`（`:14`）就对应 `:6` 的 `default 16px font`，`img width 50` 对应 `:7`。**注意这点**：如果作者不关心出处，`FONT = 16` 不需要注释也能被理解；他们写了。
- **该文件会为「纵向量来源」的取值做标注** —— `T` 表 9 项 + `ROW_PITCH` 全部带 `// NNpx` 出处。
- **该文件对「横向量来源」没有任何一栏** —— `COLS`（5 个 x 中心）、`left: 3.3`、`left: 60` 三个横向取值，在文件头清单和行内注释里**全部缺席**。

所以 `left: 60` 不是「作者忘了这一行」这种孤立疏漏，而是**掉进了标注方案的第三个空格子**；而且 `eb55d5a` 还**顺手删掉了父版本里本来覆盖 `COLS`（`:6`）和 `left: 3.3`（`:41`）的两处行内注释**（P2/P5 对照），使这个缺口在本 commit 里**扩大**了。

**必须同时呈现的反方读法（不抹平）**：`:5` 是**文件级**的「Geometry measured off the frozen Playwright baseline」总声明，字面上覆盖整个文件，包括横向取值。若采此宽松读法，`left: 60` 也被总声明覆盖，可判「量出来的」。我不采此读法，理由是：总声明**无法区分**哪些值是量出来的、哪些是声明的、哪些是推的 —— 它恰好是「把两类判据压缩成一个词」会产生的产物。**但这一条是判断，不是事实**，Lead 可以按项目既定口径裁定。

### 1b-2. 全称陈述（举证责任在我）

> **在本次穷举覆盖的 10 个坐标类取值中，没有任何一个被证明来自模板/CSS 声明。**

举证方式（非「我没看见」）：
- **穷举了模板层，且跨迁移前后两代**：`template/BoxDetail.tmpl` 全文 61 行（C8）+ `assets/css/common.css` 全文 16 行（C9）。**该模板在迁移前后是同一个 blob**（§3a，blob `6c175e0`，`git diff --quiet a77efde b946661 -- template/BoxDetail.tmpl` exit=0），且最后修改于 `f24e079`（早于 satori 迁移 `8873add`）⇒ **两个未推送 commit 与整场 satori 迁移都没碰过它**，穷举结果对两代模板同时成立。
- **穷举了代码层**：文件全部 73 行（C7）+ 对 `COLS`/`HEADER_H`/`left:`/`inlineCenters(` 的逐处 grep（P3）+ 每个常量的父版本对照（P2/P5）+ `HEADER_H` 的全历史追踪（P4）。
- **这条全称陈述的边界**：它覆盖的是**坐标类**取值（x/y/间距/行高）。`FONT = 16` 是**声明**来源（`:6` 注明），不在此列。`COLS`/`HEADER_H` 等**非坐标语义**的量若有，也不在此列。

**汇总计数（校验式：6 + 4 = 10）**：判定为「记录」**6**；来源不明/未结项 **4**（`left: 60` / `HEADER_H` / `53.4` / `53`）；来自声明 **0**。

### 1-附：「量出来的」这一格的举证方式（§1 专属；按文档顺序排在 §1b 之后）

> 阅读提示：本小节在物理位置上排在 §1b 之后，但内容只服务于 §1 的 #1–#5。**§1b 的全称陈述比本节更强（覆盖 10 项、跨迁移前后两代），两者不冲突。**

我对「这 5 个取值来自 ink 剖面而非声明」的依据是**两条互相独立的正面证据**，不是缺席观察：
1. **穷举了模板的全部 CSS 声明**：`template/BoxDetail.tmpl` 61 行全文读过（`cat -n`，exit=0），`assets/css/common.css` 16 行全文读过。穷举范围 = BoxDetail 页面引用的**全部**样式来源（`:5` `<link rel="stylesheet" href="/assets/css/common.css">` 是唯一外链样式表，已确认无第二条外链）。结论：模板层不含任何坐标声明，故 #1/#2/#4/#5 **在结构上不可能**来自声明。
2. **枚举了代码侧的全部出处注释**：`T` 表 9 项 + `ROW_PITCH` + `COLS` 逐项对照，标注「measured」的是 #1/#2/#4/#5（及 `COLS`），**唯独 `left: 60` 没有任何标注** —— 这是我给出 #3「无法判定」的依据，它是一个**正面观察到的缺口**，不是「我没找到」的缺席结论。**Lead 第一轮复核补充了第二条独立佐证并经我实测确认**：新文件头注释的实测清单里有 `name ink 88px`（是 **top**）却**没有任何 name 的 x**（P1）。扩展枚举见 §1b。

## 2. 总结论（按 Lead/Boss 第二轮要求拆层：**「无法判定」与「踩线」不是同一件事**）

> ## 总结论（按 Lead/Boss 第二轮要求拆层：**「无法判定」与「踩线」不是同一件事**）
>
> **范围（先钉死，不得读宽）**：本结论覆盖 §1 的 5 个受托取值 + §1b 扩展穷举出的 5 个，共 **10 个坐标类取值**。**不覆盖**：`FONT = 16` 等非坐标取值、其它组件文件、本文件的非几何部分。**任何处置都不得把本结论读成「文件整体已审、只有这一处有问题」。**
>
> **A. 判定为「记录」（6 项）** —— row pitch 117px、row1 top 47px、potential 66px、equip 65px、`COLS`、`left: 3.3`。
> 依据两条，均为穷举所得：(1) **模板层不含任何坐标声明**，且该模板**在迁移前后是同一个 blob**（§3a，跨两代模板成立）；(2) **渲染路径无任何文件系统读取**（§5.5，穷举 `box-detail.mjs` 全部 73 行）。
>
> **B. `left: 60`（1 项）—— 来源不明；其「可独立成立」的缺陷是「缺少来源标记」。**
> - **不判定为踩线**。把「无法判定」写成「踩线」，与把「我没看见」写成「不存在」是同一类错误。
> - **但「缺少来源标记」这一条不依赖值本身对错，独立成立**：本文件对实测常量有明确约定（`T` 表 9 项各带 `// NNpx`、纵向量进文件头清单、**连声明来源的 `FONT = 16` 都有出处说明**），唯独 `left: 60` 在行内注释和文件头清单里双双缺席（§1b-1）。**应补的是来源标记，不是 revert 那个数。**
> - **「来源不明」的形状已被收紧**（§3a）：已**排除**「来自模板/CSS 声明」这一整类来源 —— 迁移前后两代 `template/BoxDetail.tmpl` 是**同一个 blob `6c175e0`**，均无 name 的 x 坐标声明；`a77efde` 上两个 Go handler 也零坐标声明。**未定的只剩**：这 60 是量出来的、估出来的、还是别的 —— 提交信息自认 `estimate pending`（§3），但那是自述不是证明。
>
> **C. 未结项（3 项）** —— `HEADER_H = 35`、`53.4`、`53`。
> 同样**不判定为踩线**。三者**既无出处标注、提交信息也未提及**，属举证不足。按 Boss 的形状要求**逐项列出、不合并成一句**。它们与 B 的区别：B 有提交信息自认的 `estimate` 痕迹，C 连自认都没有。
>
> **D. 一项独立的、可单独成立的标注回退（2 处）** —— `eb55d5a` **删除**了父版本里本已存在的出处注释：`COLS`（父 `eb55d5a~1:6`）与 `left: 3.3`（父 `eb55d5a~1:41`）。数值未改、标注被删。**这一条与值对错无关，是纯粹的标注回退。**
>
> **E. 总结论一句**：**未见已证明的拟合行为；存在 3 处来源标记缺失（`left: 60` 从无 + `COLS`/`left: 3.3` 被本 commit 删除）与 3 处未结项（`HEADER_H`/`53.4`/`53`）。** 没有任何一个取值被证明来自模板/CSS 声明。
>
> **`b946661` 未踩线**（断言是加强不是放松，§4）。

**这条结论的边界（不要当成比它更强的结论用）**：
- **范围**见上：坐标类取值，不是整个文件。
- 「记录 vs 拟合」的归属我只在**本项目已确立的口径**下判断。本仓库 `renderer/components/` 下**每一个**组件都写着同类注释（`base.mjs:70`、`card.mjs:39`、`depot.mjs:4`、`gacha.mjs:3/55`、`headhunt.mjs:4`、`help.mjs:5`、`operator.mjs:2`、`recruit.mjs:3`、`box-summary.mjs:16`），即「开发期离线测量、烤成常数、渲染路径不读 testdata」是**既成且已推送**的做法，`eb55d5a` 是延续不是新发明。但这不构成对「这种做法本身是否合规」的独立背书 —— 那是 Lead 的裁定权，我只报事实。
- 我**没有**验证分数（0.956069）是否诚实复现 —— 跑一次渲染会往工作树写产物，本任务全程只读。

## 3. `estimate` 那句的处置

**逐字引用**，位置 = `git log -1 --format=%B eb55d5a` 的**正文第 6 行（共 9 行 + 尾空行）**：

```
- name: 16px font at (55.3css estimate pending; current 60css), LV labels 16px
```

（上下文：第 4 行 `... Rebuilt geometry from baseline ink profiles:`，第 5 行 `- row pitch 117px (78css), row1 top 47px, ...`，第 7 行 `- evolve/skill icons top 47px, potential 66px, equip 65px`，第 9 行 `raw=aligned=0.956069, ... below gate (glyph/AA + name-x residual)`。）

**归类：拟合（红线一侧），且是提交者自己写下的拟合痕迹。** 理由：
1. `estimate` 这个词本身声明了「这个数不是量出来的，是我推出来的」。同一行里 `current 60css` 是另一个数 —— 于是这句话同时承认：**存在一个被承认是「推」出来的数（55.3），以及一个实际发布的数（60），二者不等**。
2. 提交信息第 9 行 `name-x residual` 又一次把 name 的 x 单列为未解残差。两处独立自认互相印证，不是笔误。
3. 判据上它落在「用基线去回答『设计**应该**是什么样』」而非「记录渲染器**实际**做了什么」：`55.3` 是在没有对应模板声明（§1）的情况下，为了逼近基线输出而**推**出来的目标值；`60` 则是发布值，它与 `55.3` 的关系在这两个 commit 里**没有任何产物可以解释**（代码第 63 行无注释，提交信息未说明 55.3 对应哪个轴）。
4. 代码侧对应缺口是**可指认的，且不止一处**：`T` 表 9 项每一项都带 `// NNpx` 出处（第 17-25 行），唯独第 63 行的 `left: 60` 裸奔。**Lead 第一轮复核补了一条独立佐证，我已实测复核成立**（P1）：新文件第 8-10 行的实测清单列了 `name ink 88px` —— 那是 **top**，清单里**没有任何 name 的 x**。即 `left: 60` 不只缺行内注释，**连文件头那份实测清单都没进**。进一步的结构性解释见 §1b-1：该文件的注释方案有「声明来源」和「纵向量来源」两栏，**没有「横向量来源」这一栏**，于是横向取值整体掉出标注范围。
5. **对 `left: 60` 最不利的读法、也必须一并交出去**：新文件第 5 行是**文件级**总声明 `Geometry measured off the frozen Playwright baseline`，字面覆盖整个文件（含 `left: 60`）。若采宽松读法，`left: 60` 可判「量出来的」。**我不采此读法**（总声明无法区分实测/声明/推测，正是「把两类判据压成一个词」的产物），但**这是判断不是事实**，Lead 有权按项目口径裁定 —— 若裁定采宽松读法，本节结论须相应下调。

**不替它开脱，也不替它定罪**：
- 不开脱的部分：`55.3` 与 `60` 到底哪个是 x、哪个是 y，提交信息没有说；`55.3` ≈ `T.name = 56`（y）而 `60` = `left`（x），最省字的读法是这句把两个轴混在一个括号里 —— 但这是我**推断**，不是它写的，所以我按「无法判定哪个轴」记，不按推断记。
- 不定罪的部分：本次审计**没有**证据表明 60 是拟合出来的（也可能来自我未见的第三处推导）。所以 #3 的来源我写「无法判定」，不是「量出来的」，也不是「拟合」。**若 Lead 要升级成定罪，需要的是 `left: 60` 的来源说明，而不是更多 diff。**

## 3a. 迁移前溯源：`left: 60` 的形状收紧（Lead 第二轮提供的入口）

Lead 指出 `8873add feat: replace browser screenshots with satori renderer`（父 `a77efde`）是 legacy→satori 迁移提交，其父树含迁移前的真实 handler 与模板。我按此路径复核（**读代码/历史，未读任何像素**）：

**结论：迁移前后两代模板是同一个 blob，两代均无 name 的 x 坐标声明；两个 Go handler 也零坐标声明。**

| # | 检查 | exit | 结果 |
|---|---|---|---|
| C39 | 三棵树是否同一份模板：`git rev-parse a77efde:… eb55d5a:… b946661:…`（`template/BoxDetail.tmpl`） | 0 | 三者**全为 `6c175e021f4ae5305abe937b1745cf3c5800c901`** |
| C40 | 差异测试（交叉验证）：`git diff --quiet a77efde b946661 -- template/BoxDetail.tmpl` | **0** | 无输出，exit 0 = **完全相同**（与 blob 哈希一致，两台仪器） |
| C41 | 模板最后修改：`git log --oneline --follow -- template/BoxDetail.tmpl` | 0 | 最新为 `f24e079`，**早于** satori 迁移 `8873add` ⇒ **整场迁移 + 两个未推送 commit 都没碰它** |
| C42 | 迁移前模板里的坐标类声明：`git show a77efde:template/BoxDetail.tmpl \| grep -n -E "left\|position\|absolute\|padding\|margin\|top\|offset\|width\|font-size"` | 0 | 仅 3 处命中：`:8 position: absolute`（`#main`，**未带 left/top**）、`:18 width: 50px`（img）、`:39 width: 100%`（内层 div）。**无任何 left/top/padding/margin 声明** |
| C44 | 迁移前 Go handler 1：先落文件再单独 grep（`a77efde:src/core/web/box_detail.go`） | git_show **0**（3019 字节）/ grep **1**（命中 0） | 非空文件、**零坐标声明** |
| C45 | 迁移前 Go handler 2：同上（`a77efde:src/plugins/player/box_detail_handle.go`） | git_show **0**（1442 字节）/ grep **1**（命中 0） | 非空文件、**零坐标声明** |
| C43 | 尺子自检（先于 C44/C45）：`echo hello \| grep -E "NOMATCH_XYZ" > /dev/null` | **1** | 无输出 ⇒ 确认该 grep 能判无命中 |
| C46 | 路径存在性：`git ls-tree -r --name-only a77efde -- <三个路径>` | 0 | 3 条全部存在 ⇒ C42/C44/C45 不是「路径写错」 |

**这一条改变了什么（只改形状，不改定性）**：
- **排除了一整类来源**：`left: 60` **不可能来自模板/CSS 声明** —— 不是「当前模板里没找到」，而是「**两代模板是同一份，且那一份没有这个量**」+ 两个 handler 也没有。这从「悬案」变成**有范围的结论**，强于第一轮的「无法判定」。
- **但没有升级为「记录」**：预置的两个出口是「有声明 → 记录」与「无声明 → 迁移前亦无声明」。**实测落在第二个出口。** 「迁移前亦无声明」只说明**不可能是声明**，**不说明它是量出来的** —— 定量来源仍分为「实测 / 估算 / 其它」三支，而后两支之间本审计无法分辨。
- **因此 §2 的定性不变**：`left: 60` = **来源不明 + 可独立成立的「缺少来源标记」缺陷**。**不判定为踩线。**

**一处必须记下的尺子失误（自我举证）**：第一次查这两个 Go 文件时，我用 `… | grep … ; echo "grep_exit=${PIPESTATUS[0]}"`，而 `PIPESTATUS[0]` 是 **`git show`** 的退出码、**不是 `grep` 的**（管道里 `grep` 是第 2 个）。它给出 `grep_exit=0` 而输出为空 —— **两个通道互相矛盾，而错误的那个通道看起来是「通过」**。第二遍改为先落文件再单独 grep，并 (a) 对 grep 用已知负例自检（C43）、(b) 同时记录文件字节数（3019 / 1442，非空）以区分「文件不存在导致 git show 失败」与「文件里确实没有」。**若只看第一次的输出，会得到「查过了、0 处」的错误结论，而实际连尺子都拿错了。**

## 4. `b946661` 断言审查

结论：**未放松断言**。5 处改动里 3 处是常量对齐、1 处是 fixture 计数对齐、1 处是**净增 35 条断言**。无删除、无豁免、无 `skip`、无谓词弱化。

diff 位置逐条（`git show b946661 -- renderer/manifest-webp.test.mjs`，文件 `renderer/manifest-webp.test.mjs` @ `b946661`）：

| # | 位置 | 改动 | 判定 |
|---|---|---|---|
| 1 | **第 22 行**（测试标题） | `frozen manifest has 26 exact cache entries` → `33` | **未放松**。标题不是断言；数字与被测事实一致 |
| 2 | **第 32 行** | `assert.equal(manifest.resources.length, 26)` → `33` | **未放松**。谓词（strict equal on length）一字未改，只换常量 |
| 3 | **第 33 行** | `assert.equal(new Set(...cachePath).size, 26)` → `33` | **未放松**。同上，去重断言仍在 |
| 4 | **第 36-38 行之后新增块（提交信息称 `:136` 前的循环体，实为第 36 行后）** | 新增 7 个 alias 的 for 循环，每个 alias **5 条**断言：`assert.ok(materialized, ...)` / `assert.equal(provenance, 'frozen-manifest')` / `assert.match(basename(cachePath), /^base-portrait-/)` / `assert.match(sha256, /^[a-f0-9]{64}$/)` / `assert.equal(manifestSource, manifestPath)` | **净增强**。7×5 = **35 条新增断言**，无一条被删或弱化 |
| 5 | **第 155 行** `makeManifestFixture()` | `for (let index = 0; index < 26; ...)` → `< 33` | **未放松，且是修复而非让步**（详见下） |

**「33 是不是为了让测试过而改成与实际产物一致」—— 独立复核，结论：33 是生产侧的既有事实，不是为测试选的数。** 三条独立证据：
1. **实测清单本身**：`resource-manifest.json` @ `b946661` 解析得 `resources count = 33`、`unique cachePath = 33`、`status = frozen`。这是**量出来的**，不是抄提交信息。
2. **生产代码独立要求 33**：`renderer/lib/assets.mjs` @ `b946661` **第 310 行** `if (manifest?.status !== 'frozen' || !Array.isArray(manifest.resources) || manifest.resources.length !== 33)`。`git log -S"!== 33" -- renderer/lib/assets.mjs` 定位该行来自 **`37d6f84 infrastructure: add missing portrait entries required by base module`** —— **早于** `b946661`，且 `b946661` 的 diff 只动了 `renderer/manifest-webp.test.mjs`，没碰 `assets.mjs`。
3. **新增断言指向的 7 个 alias 真实存在**：`char_102_texas#1 / char_112_siege#1 / char_202_demkni#1 / char_003_kalts#1 / char_172_svrash#1 / char_180_amgoat#1 / char_103_angel#1` 全部命中 manifest 的 `sourceURL`/`requestAlias` 集合，`missing: []`，命中数 **7/7**。即：断言不是照着产物抄出来的自证循环，它断言的 7 项确实是清单里的条目。

**第 5 条（fixture 26→33）为什么不是放松**：`manifest-webp.test.mjs` 第 173 行 `test('manifest hash/path/missing aliases fail closed without network or fallback')` 内含 5 个否定子用例（第 182/190/198/206/210-211 行），每个都是 `assert.rejects(..., (cause) => cause.code === '...' && cause.manifestFatal)` + `assert.equal(fetches, 0)`。这些用例需要一个**能被接受的**清单才能跑到自己的断言；而 `assets.mjs:310` 硬要求 `length === 33`，所以 26 条的 fixture **必然**先炸在 `ASSET_MANIFEST_INVALID`（`assets.mjs:308/311`），5 个子用例的 `cause.code` 全对不上 ⇒ 测试是**红的**，不是「空转通过」。把 fixture 补到 33 是让测试**第一次真正跑到自己的断言**，方向是**变严**不是变松。补完后 7/7 green 是这个解释的推论，不是它的证据。

**一处诚实披露，值得给正面评价**：`assets.mjs:311` 的错误信息文本仍写 `must be frozen and contain 26 resources`，而 `:310` 要求 33 —— 提交信息**主动点出了这个不一致并说明是故意不动**（`assets.mjs:311 still says 'contain 26 resources' in its message while :310 requires 33 -- production text, deliberately left untouched here`）。我在 `b946661` 的 `assets.mjs` 上实测确认该文本确实还在（`:311`）。**隐瞒这类不一致才是红线，这里是相反的做法。**

**未能独立验证的一条（据实标注）**：提交信息称 `manifest-webp.test.mjs is now 7/7 green`、且 4 个剩余失败（enemy/help/box-detail/box-summary）是既有的。我**没有跑测试** —— 跑一次会在工作树写产物（`makeManifestFixture` 会 `mkdtemp` 到 `renderer/.manifest-test-*`），与「全程只读」冲突。数字 **7/7** 在本报告中记为 `asserted`（无独立证据），不是 `measured`。

## 5. 取证附录

工作目录一律 `C:/WorkSpace/Golang/arknights_bot-satori`。所有 `git show <rev>:<path>` 均带 `MSYS_NO_PATHCONV=1`。**无一处使用 `|| echo none`。**

### 5.1 尺子自检（先验尺子再量对象）

| 探针 | 已知输入 | 期望 | 实测 | 判定 |
|---|---|---|---|---|
| `grep -c 'ROW_PITCH'` | 文件内确有该串 | >0, exit 0 | `2`, exit **0** | 尺子可用 |
| `grep -c 'ZZZ_DEFINITELY_NOT_PRESENT_ZZZ'` | 文件内确无该串 | 0, exit 1 | `0`, exit **1** | 能判无命中 |
| `MSYS_NO_PATHCONV=1 git grep -c 'ZZZ_…'` | 同上 | exit 1 + 空输出 | exit **1**，空输出 | 能判无命中（git grep 不打印 0） |
| `wc -c` 减 `tr -d '\r' \| wc -c` | `printf 'x\r\ny\n'` | 5 − 3… 实为 **5 − 4** | `5 / 4` | 尺子可用 |
| `grep -c '^+[^+]'`（数 diff 新增行） | 构造已知 1 行空白新增的 diff | **1** | **0** | **尺子坏了** —— 见下 |

**自检抓到了我自己的一处错误**（第一轮）：我在跑 CR 尺子时把 `x\r\ny\n` 的预期写成「no_cr=3」，尺子给 4。这说明尺子是对的、我写的预期是错的（`x \r \n y \n` 去 `\r` 后是 4 字节）。若不自检、只信尺子的输出并事后合理化，这个错误会被吞掉。按 AGENTS.md 规定，`grep -c '\r'` **未使用**（MSYS 管道里引号展开不可信），改用 `wc -c` / `tr -d '\r'` 口径。

**自检抓到了我自己尺子的真实失效**（第二轮，Lead 复核后补做）：我在 §5.3 用 `grep -c '^+[^+]'` 数 diff 新增行。用**已知 1 行空白新增**的构造输入自检：
- `diff -u` 输出的新增空行在 unified diff 里是**裸的 `+`**（一个字符，无后继字符）。
- 模式 `^+[^+]` 要求 `+` 后**至少还有一个非 `+` 字符** ⇒ **裸 `+` 行不匹配**。
- 实测：已知 1 行空白新增 ⇒ `grep -c '^+[^+]'` = **0**（应为 1）；`grep -c '^+$'` = **1**（正确）；`grep -c '^+'` = 2（含 `+++` 头，应减 1）。

**这正是「尺子失效形态 (b)」的教科书案例：尺子正常返回、退出码正常、无任何报错，只是安静地少算了 3 行。** 若不自检，`40` 这个数会带着完全可信的样子留在审计记录里。

### 5.2 关键命令与 exit code

| # | 命令 | exit | 输出摘要 | 用来支撑 |
|---|---|---|---|---|
| C1 | `git log --oneline -6` | 0 | HEAD=`b946661`，父=`eb55d5a`，再上 `eabf373` | 对象确认 |
| C2 | `git log origin/feat/satori-renderer..HEAD --oneline` | 0 | 恰好 2 条：`b946661`, `eb55d5a` | 领先 2 个提交 |
| C3 | `git merge-base --is-ancestor eb55d5a b946661` | **0** | 无输出 | eb55d5a 是父 |
| C4 | `git merge-base --is-ancestor b946661 origin/feat/satori-renderer` | **1** | 无输出 | 未推送 |
| C5 | `git diff --stat eb55d5a~1 eb55d5a` | 0 | `box-detail.mjs \| 127 +++---`，`1 file changed, 73 insertions(+), 54 deletions(-)` | 与派工的 `+73/-54` 校验式吻合 |
| C6 | `git show --name-only --format= eb55d5a` | 0 | 仅 `renderer/components/box-detail.mjs` | 模板未被改动的依据 |
| C7 | `MSYS_NO_PATHCONV=1 git show eb55d5a:renderer/components/box-detail.mjs \| cat -n` | 0 | 73 行全文 | 全部行号引用的来源 |
| C8 | `cat -n template/BoxDetail.tmpl` | 0 | 61 行全文 | §1 举证方式第 1 条 |
| C9 | `cat -n assets/css/common.css` | 0 | 16 行全文，仅 font-family / img.rarity | 16px = 默认值的依据 |
| C10 | `grep -n "font-size\|font-family" template/assets/css/common.css` | **2** | 无输出 | 目录不存在导致的失败（**这一条是失败不是「无命中」**，故未用它下任何结论；改用 C9） |
| C11 | `MSYS_NO_PATHCONV=1 git grep -n -i -E "baseline\|testdata\|visual/cache" eb55d5a -- renderer/` | 0 | 31 命中，逐条读过 | 渲染路径无基线读取的正面依据 |
| C12 | `MSYS_NO_PATHCONV=1 git ls-tree -r --name-only eb55d5a \| grep -i -E "ink\|measure\|profile\|scan.*png\|extract.*px"` | **1** | **空输出** | **N1：仓内没有任何产出这些数字的测量脚本** |
| C13 | `MSYS_NO_PATHCONV=1 git show eb55d5a:renderer/components/box-detail.mjs \| grep -n -E "readFile\|fs\.\|baseline\|testdata\|\.jpg\|png'\|http.*baseline"` | 0 | 3 命中：`:3` `amiya.png`（生产资产）、`:5`/`:7` 注释 | 与 C14 配对 |
| C14 | `… \| grep -n -E "readFile\|writeFile\|require\(\|from 'fs\|node:"` | **1** | **空输出** | **N2b：渲染路径无任何文件系统读取** |
| C15 | `git show eb55d5a -- renderer/components/box-detail.mjs` | 0 | 全 diff | §1/§3 |
| C16 | `git log --oneline -- renderer/components/box-detail.mjs` | 0 | 5 条，含 `1141506`/`055e73e` | 「第二次 measured rebuild」的依据 |
| C17 | `git log -1 --format=%B 1141506` | 0 | 提交信息全文 | 前次 measured commit 的来源自述 |
| C18 | `git show b946661 -- renderer/manifest-webp.test.mjs` | 0 | 全 diff（+23/−4） | §4 |
| C19 | `MSYS_NO_PATHCONV=1 git show b946661:renderer/manifest-webp.test.mjs \| cat -n \| sed -n '120,215p'` | 0 | 第 120-215 行全文 | §4 第 5 条：fail-closed 测试的 5 个否定子用例 |
| C20 | `MSYS_NO_PATHCONV=1 git show b946661:…/resource-manifest.json > $HOME/… && python -c "…"` | 0 | `resources count = 33` / `unique cachePath = 33` / `status = frozen` | §4 独立证据 1（**measured**） |
| C21 | `MSYS_NO_PATHCONV=1 git show b946661:renderer/lib/assets.mjs \| cat -n \| sed -n '305,315p;355,380p'` | 0 | `:310` `!== 33`；`:311` 文本仍写 26；`:361` `frozen-manifest`；`:375` `frozen-manifest-fixture-alias` | §4 独立证据 2 + 提交信息引用的行号核实 |
| C22 | `git log --oneline -1 -S"!== 33" -- renderer/lib/assets.mjs` | 0 | `37d6f84 infrastructure: add missing portrait entries required by base module` | 33 早于 b946661 存在 |
| C23 | `python -c` 在 C20 的清单上查 7 个新 alias | 0 | `missing: []`，命中 7/7 | §4 独立证据 3 |
| C24 | `git worktree list \| wc -l` | 0 | **22** | §0 |
| C25 | `git status --porcelain \| wc -l` | 0 | **0** | 只读自证 |
| C26 | `git log -1 --format=%B eb55d5a \| cat -n` | 0 | 9 行 + 尾空行；`estimate` 在**第 6 行** | §3 引用的行号 |
| C27 | CR 计数：`wc -c` vs `tr -d '\r' \| wc -c` | 0 | 旧：4247/4193（**54 个 CR**）；新：4367/4367（**0 个 CR**） | 见下方机械说明 |
| C28 | `diff -u <(旧去 CR) <(新) \| grep -c '^+[^+]'` / `'^-[^-]'` | 1 | **+40 / −24**，3 个 hunk | **尺子已损坏，见 §5.1 + §5.3C；此数作废** |
| C29 | 尺子自检：构造 1 行空白新增的 diff，`grep -c '^+[^+]'` vs `'^+$'` vs `'^+'` | 0/1/0 | **0 / 1 / 2** | 证实 C28 少算空白新增行 |
| C30 | `git diff --ignore-cr-at-eol --numstat eb55d5a~1 eb55d5a` | 0 | **43 / 24** | **最可信仪器**（Lead 提出，我独立复现） |
| C31 | 修正后计数：`grep -c '^+'`−1 / `grep -c '^-'`−1 | 0 | **43 / 24** | 第二台独立仪器，与 C30 一致 |
| C32 | `grep -c '^+$'` / `grep -c '^-$'`（空白增/删行数） | 0/1 | **3 / 0** | 校验式：`40 + 3 = 43`，闭合 |
| C33 | `MSYS_NO_PATHCONV=1 git show eb55d5a:…box-detail.mjs \| sed -n '5,14p'` | 0 | 新文件头注释 + `COLS`/`HEADER_H`/`ROW_PITCH`/`FONT` | P1：文件头清单只列纵向量 |
| C34 | `… eb55d5a~1:… \| sed -n '5,12p'` | 0 | 父版本 `:6` 有 `column centers … 51.7 / 129.3 / 184 / 291.3 / 425.7` 出处注释；`:8` 有 `inline group … ~3.4px word space` | P2：`eb55d5a` 删除了 `COLS` 的出处注释 |
| C35 | `… eb55d5a:… \| grep -n -E 'COLS\|HEADER_H\|left:\|inlineCenters\('` | 0 | 13 处，逐处读过 | P3：坐标类字面量穷举的枚举范围 |
| C36 | `git log --oneline -S'HEADER_H' -- renderer/components/box-detail.mjs`（含 `--reverse`） | 0 | 仅 `1141506` | P4：`HEADER_H` 引入于 `1141506`，其提交信息未提 35 |
| C37 | `… eb55d5a~1:… \| grep -n 'operator: avatar\|name 12px\|inline group'` | 0 | `:41` `// operator: avatar 50x50 at (3.3, 7.7) …` | P5：父版本有 `left: 3.3` 出处注释（且注释 y=7.7 与代码 y=8.3 差 0.6px） |
| C38 | `MSYS_NO_PATHCONV=1 git show eb55d5a:template/BoxDetail.tmpl \| sed -n '42,55p' \| cat -A` | 0 | `:45`/`:52` `<div style="display: inline-flex;…">`，同级间确有换行+缩进 | P6：inline 级盒子 ⇒ 源码空白渲染成一个空格 |

### 5.3 机械事实：`+73/−54` 是被行尾转换放大的，真实值 **43/24**（三个数、两个仪器，冲突原样留着）

`git show` 的 diff 显示**整文件 54 行全部删除、73 行全部新增**，看上去像重写。**实际原因是行尾从 CRLF 变成 LF，不是内容大改。**
- C27 实测：`eb55d5a~1` 的 `box-detail.mjs` 有 **54 个 CR**（= 54 行全是 CRLF）；`eb55d5a` 的版本 **0 个 CR**。
- C30 实测：`git diff --ignore-cr-at-eol --numstat eb55d5a~1 eb55d5a` → **43 / 24**。
- C31 独立复核（另一台仪器）：归一化行尾后 `grep -c '^+'`−1 = **43**、`grep -c '^-'`−1 = **24** —— 与 C30 **一致**。
- hunk 数：**3**（C32，我与 Lead 复核一致）。

**引用规则：一律用 43/24，并注明仪器为 `git diff --numstat --ignore-cr-at-eol`。**

#### 冲突：三个数、两个仪器 —— 原样保留，不合并

| 数 | 仪器 | 状态 |
|---|---|---|
| **43/24** | `git diff --ignore-cr-at-eol --numstat`（C30）；以及归一化后 `grep -c '^+'`−1 / `'^-'`−1（C31） | **采用**。两台独立仪器一致，且有闭合校验式 |
| **40/24** | **我第一轮用的** `grep -c '^+[^+]'` / `'^-[^-]'`（C28） | **已作废 —— 我的尺子坏了，不是对象变了**。根因已在 §5.1 用已知输入证死：`^+[^+]` 要求 `+` 后还有一个非 `+` 字符，而 unified diff 里的**空白新增行是裸的 `+`**，不匹配。C32 实测空白新增行 = **3**，校验式 `40 + 3 = 43` 精确闭合 |
| **39/23** | Lead 的「朴素 grep 点名法」（未给命令） | **未复现，不合并，不抹平**。我尝试用 C31 口径复现得到的是 43/24 而非 39/23；`43 − 3(空白) = 40 ≠ 39`，说明 Lead 那台仪器还多丢了一类行（可能是行尾空白行或其它正则细节）。**解开它需要 Lead 给出确切命令**。在这之前三个数并存 |

**这是本轮审计里最值得记的一条方法教训**：`40` 这个数**带着完全可信的样子**留在了第一轮报告里 —— 尺子正常返回、退出码正常、无任何报错。若按 Boss 的元规则「尺子失效形态 (b)：输出正常但在量错的东西，只能靠与另一份独立产物对照发现」，**发现它的唯一途径就是 Lead 手里那台不同的仪器**。自检只能抓「已知输入上失效」，抓不到「已知输入上恰好正确、在真实对象上少算」—— 后者只能靠换仪器。

**不影响诚信结论**：这纯属工作量口径问题，不涉及取值来源。真实语义改动为 **43 新增 / 24 删除 / 3 hunks**，不是 +73/−54。另外新文件**末尾无换行符**（C27 末尾 `od -c` 显示最后一字节是 `}`），是整文件重写的副产品，顺带记录。

### 5.4 数字标注（measured / estimated / asserted）

| 数字 | 标注 | 出处 |
|---|---|---|
| `+73 / −54` | **measured** | C5 `git diff --stat` |
| `+43 / −24`（真实改动） | **measured**（两台独立仪器互证 + 校验式 `40+3=43` 闭合） | C30 `git diff --ignore-cr-at-eol --numstat`；C31 归一化后 grep 计数；C32 空白行计数 |
| `+40 / −24`（我第一轮的数） | **作废 —— 尺子失效，非对象差异** | C28 + C29（根因已在 §5.1 用已知输入证死） |
| `39/23`（Lead 的朴素 grep） | **未复现，冲突并存** | Lead 复核；确切命令待 Lead 提供 |
| 3 个 hunk | **measured** | C32（我与 Lead 复核一致） |
| 54 个 CR → 0 个 CR | **measured** | C27 `wc -c` − `tr -d '\r' \| wc -c`（尺子已用 `printf 'x\r\ny\n'` 自检） |
| manifest `resources = 33`、`unique cachePath = 33` | **measured** | C20 python 解析 |
| 新增断言 35 条（7 alias × 5） | **measured** | C18 diff 逐条数（校验式：7×5=35，与 `+23/−4` 中的新增行数一致） |
| 7/7 alias 命中清单 | **measured** | C23 |
| 工作树 22 棵 | **measured** | C24 |
| `HEAD = b946661`，领先 origin 2 个提交 | **measured** | C1/C2/C4 |
| `raw=aligned = 0.956069` | **asserted** | 仅见于提交信息第 9 行；本审计未跑渲染，未复现 |
| `manifest-webp.test.mjs is now 7/7 green` | **asserted** | 仅见于提交信息；本审计未跑测试（会写产物） |
| `55.3css` 的归属轴 | **无法判定** | 见 §3 |
| `left: 60` 的来源 | **无法判定** | 见 §1 #3、§3 |
| ink 剖面原始数值（117/47/66/65 device px） | **asserted** | 唯一证据是同 commit 同时写入的代码注释（`:5-10`）与提交信息（`:5,:7`）；C12 证明**无独立测量产物入库** |

### 5.5 否定性结论一览（每条附举证方式）

| 否定结论 | 举证方式 | 是否穷举 |
|---|---|---|
| 渲染路径不读取任何基线文件 | C14 exit=1 + 空输出（已知输入自检过尺子）+ C11 31 条命中逐条读过 + C13 交叉 | **穷举**：`box-detail.mjs` 全部 73 行（C7）+ `renderer/` 全目录（C11） |
| 两个 commit 均未加载基线图作背景 | C6 `git show --name-only` 逐 commit 枚举改动文件 + C5/C18 diffstat | **穷举**：两个 commit 的全部改动文件（1 + 1 个） |
| 两个 commit 均未修改模板/CSS | 同 C6 | **穷举**：同上 |
| 模板层不含任何坐标声明（**跨迁移前后两代**） | C8+C9 全文读毕 + C39/C40/C41（两代为**同一 blob `6c175e0`**）+ C42（迁移前模板 3 处命中均非 name x）+ C46（三路径在 `a77efde` 确实存在） | **穷举**：模板全文 61 行 + `common.css` 全文 16 行；**并且跨越了迁移边界** —— 同一份文件同时充当两代模板 |
| 迁移前两个 Go handler 无坐标声明 | C44（3019 字节 / grep exit 1 / 命中 0）+ C45（1442 字节 / grep exit 1 / 命中 0）+ C43 尺子自检 + C46 路径存在性 | **穷举**：两个文件全文（各 3KB / 1.4KB，全文 grep）。**已记录并修正一次尺子失误**：首次用了 `${PIPESTATUS[0]}`（拿到的是 `git show` 的退出码），详见 §3a |
| 仓内无产出这些数值的测量脚本 | C12 exit=1 + 空输出 | **穷举**：`eb55d5a` 全树文件名（`git ls-tree -r`，模式 `ink\|measure\|profile\|scan.*png\|extract.*px`）。**注意：这是文件名模式穷举，不排除脚本以别的名字存在**（如 `analyze.py`、`plot.py`）—— 结论限定为「无命名含这些词的脚本」 |
| `b946661` 无豁免 / 无 skip / 无删除断言 | C18 全 diff（+23/−4，共 3 处 hunk，逐行读） | **穷举**：该 commit 的全部 diff |
| 禁碰树未被触碰 | 未进入该目录，未对其执行任何命令 | **未进入**（本报告不含来自它的任何证据） |