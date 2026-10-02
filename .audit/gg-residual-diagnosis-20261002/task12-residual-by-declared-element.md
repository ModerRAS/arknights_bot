# Task 12 — 残差按声明元素分解

```
provenance:   measured（元素表、声明出处、闸门） + 未测（元素级残差归因，见 §4）
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align（复用，未新建）
              生成时间 2026-10-02T14:18:22Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task12-residual-by-declared-element.md
量具定义:      tmp/residual/byelem/instrument.json（跑数前落盘）
脚本/产物:    tmp/residual/byelem/{byelem.py, tmpl_geometry.json, derivable_boxes.json,
              controls.json, elements.json, teeth8b.log, teeth8c.log, measure.log}
produced_by:  src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47（elements.json 逐行）
```

**修改权声明**：未改任何渲染代码、未改任何计分代码、未 commit、未 push、未新建工作树。`git diff --stat` 空。禁碰树未读未渲。

> 本表为**像素类构成**，用作**声明轴分解的交叉核对**；**它不单独构成实现依据**，因为像素类不是可指名的目标。

---

## 0. 先认一条纪律问题

上一轮 Boss 点名「补 `teeth8c`，**不许只记录**」，并给了失败时的处置（照实报「该分支没走到」）。**我既没补测，也没申报它被跳过 —— 两半都丢了，而且丢得静默。**

「忘记」是能力问题，「静默跳过已点名的动作」是纪律问题，两者补救与评价不同。本轮已补测（§5），并写进常设流程：**做不了就当场报，不要静默跳过。**

## 1. 权威措辞（替换 Boss 原句）

Boss 原句「『哪个元素最差』只允许由『我们画的 vs 声明的』得出」**按字面读会禁止使用基线幅度，而那在物理上不可能** —— 缺口本身就是以基线为参照的量。**原句作废**，理由记此。权威版三句：

> **元素归属来自声明**（模板 + 布局状态导出元素框）；
> **残差幅度来自「我们 vs 基线」**（把比较限制在该元素框内）—— **这一项不可回避**；
> **不得用从基线读到的任何数值，去决定任何元素「应该是什么」**（位置、内容、颜色、尺寸一律不许）。

**前两句是分工，第三句才是红线。红线不是「不许碰基线」，是「不许让靶子变成答案」。** 我们全程拿基线当测量靶子，合法；让靶子变成答案，那是拟合。

**自查**：元素框全部由模板 CSS 块导出（见 §3 的行号追溯），**没有任何一个框来自基线差异像素的聚类**；残差只用来看「框内有多少误差」，**没有用基线的任何数值去决定任何元素的位置、内容、颜色或尺寸**。

## 2. 声明来源确认（我的探针，不采信转述）

| 候选 | 结论 | 依据 | 路径 |
|---|---|---|---|
| **legacy 模板**（`template/*.tmpl` + `assets/css/common.css`） | **声明来源成立**；元素框**部分**可导出 | 16 个 `.tmpl` 共 **181** 条几何声明；其中带 `position:absolute` 且同时含位置量与尺寸量的 **13** 条 | `template/*.tmpl`、`assets/css/common.css` |
| gg 代码里的绘制坐标 | **否决** | 那是「我们画的」，与「我们画的」相比是空的 | — |

**措辞校正（按裁定，不写成「模板不存在」）**：
> `ggrender` 没有 `ParseHTML`、各场景手绘 —— 这只蕴含「声明不在这个包里」，**不蕴含「没有声明」**。模板在仓库里，只是不在 `src/ggrender` 包内。

**共享样式表（我的自查）**：`assets/css/common.css` = **417 字节 / 16 行 / 7 条规则**，`position:` / `left` / `top` **各 0 次**（三次 grep 均 exit=1、输出为空）。⇒ **补扫它不改变任何计数。**

## 3. 元素表（过滤后，已确认 —— 取消「暂定」标注）

**过滤链**：181 条几何声明 → 含位置量者 111 → **同时含尺寸量者 97** → **且位于 `position:absolute` 下者 13**（过滤在脚本内 inline 完成，故 **13 是过滤后的数**；**27 是过滤前的上界**）。

| # | 场景 | 选择器 | 模板行号 | 声明的属性 | 设备坐标框 | 框内残差份额 |
|---|---|---|---|---|---|---|
| 1 | `box` | `.profession` | `Box.tmpl:31` | width 15, top 5 | (0, 7.5, 22.5, **0**) | 0.00000 |
| 2 | `box` | `.rarity` | `Box.tmpl:37` | width 40, top 5 | (0, 7.5, 60, **0**) | 0.00000 |
| 3 | `box` | `.evolve` | `Box.tmpl:43` | width 40, top 80 | (0, 120, 60, **0**) | 0.00000 |
| 4 | `box` | `.level` | `Box.tmpl:49` | width 20, left 5 | (7.5, 0, 30, **0**) | 0.00000 |
| 5 | `box` | `.potential` | `Box.tmpl:59` | width 30, top 100 | (0, 150, 45, **0**) | 0.00000 |
| 6 | `card` | `.level` | `Card.tmpl:106` | width 25, left 15 | (15, 0, 25, **0**) | 0.00000 |
| 7 | `headhunt` | `#main` | `Headhunt.tmpl:7` | width 1024, height 576, left 25 | (25, 0, 1024, 576) | **0.99659** |
| 8 | `missing` | `.profession` | `Missing.tmpl:31` | width 15, top 5 | (0, 7.5, 22.5, **0**) | 0.00000 |
| 9 | `missing` | `.rarity` | `Missing.tmpl:37` | width 40, top 5 | (0, 7.5, 60, **0**) | 0.00000 |
| 10 | `recruit` | `.rarity` | `Recruit.tmpl:34` | height 20, left 30 | (45, 0, **0**, 30) | 0.00000 |
| 11 | `state` | `#campaign_recover_bg` | `State.tmpl:105` | width 110, height 25, left 60 | (60, 0, 110, 25) | 0.00000 |
| 12 | `state` | `#trainee` | `State.tmpl:184` | width 130, left 922 | (922, 0, 130, **0**) | 0.00000 |
| 13 | `state` | `#remain_secs_bg` | `State.tmpl:190` | width 133, height 25, left 923 | (923, 0, 133, 25) | 0.00000 |

**框的来源可追溯性自查**：每一行的框都由「模板文件 + 行号 + 选择器 + 声明属性」四项确定，写在 `derivable_boxes.json` 里；**没有任何一行的框来自「我看了图觉得那儿是个框」**。

### ⚠️ 12 个 0 是尺子侧的现象，不是场景侧 —— 已先查尺子

按「任一类别给 0 ⇒ 先查尺子再查场景」：12 个框的**高度为 0**，因为这些元素只声明了 `width`（或只声明 `height`），**另一个维度由内容决定，静态不可导出**。所以份额 0 来自**框退化**，不是「这些元素没问题」。

**其中只有 3 个元素四量齐全**（同时声明 width 与 height）：`Headhunt.tmpl:7 #main`、`State.tmpl:105 #campaign_recover_bg`、`State.tmpl:190 #remain_secs_bg`。

**而唯一拿到非零份额的 `#main` 是容器**：1024×576 覆盖了 1049×576 画布的 99.66%，**它没有指名任何具体元素**。

⇒ **声明轴当前产出的元素级目标数为 0。** 不是「没找到」，是**当前声明里没有可指名到细粒度、且框四量齐全的元素**。

## 4. 未覆盖的 10 个模板：**声明轴结构性不适用**

**这不是解析器的局限，是那 10 个模板的属性**（元素不带绝对位置声明，布局是流式 + 运行时 JS 算出来的）。

> ⚠️ **不许写成「尚未找到」** —— 那会被下一个人当成待办；写成属性才是终局判断。

`Base` / `BoxDetail` / `BoxSummary` / `Calendar` / `Depot` / `Enemy` / `Gacha` / `Help` / `Lottery` / `Operator` —— 共 **10** 个。

**举证方式**：`tmpl_geometry.json` 逐模板扫描结果 + `common.css` 的 0 次 `position:`/`left`/`top`（三次 grep exit=1、输出为空）。非「我没看见」。

## 5. 闸门验牙齿 —— 第 8 次：**两种模式均已验证**

| 模式 | 操作 | 实测输出 | exit | 产物 |
|---|---|---|---|---|
| 对照失败 | 强制 `controls_pass=false` → `byelem.py measure` | `REFUSED: controls_pass != true` | **2** | `elements.json` **未产出**（`ls` exit=2） |
| **输入缺失（teeth8c，我上一轮静默跳过的那条）** | 移走 `controls.json` → `byelem.py measure` | **`REFUSED: controls.json missing`** | **2** | `elements.json` **未产出**（`ls` exit=2） |

**两种模式均已验证**，`elements.json` 只在 `controls_pass=true` 时产出。

**对照结果（3 条 P19′ 序关系，无阈值）**：`ORD1_containment` ✓、`ORD2_frame_leak` ✓（框扩大后框外份额严格下降 ⇒ P23 机理陈述未被证伪）、`ORD3_empty_box_boundary` ✓。

## 6. C 类重叠与两个永远回答不了的场景

**C 类 9 个场景中，5 个被声明轴覆盖**（缺口**都不是小缺口**）：

| 场景 | sim | gap | 是否被覆盖 |
|---|---|---|---|
| `box` | 0.82504 | **0.16496** | ✅（5 个元素，框均退化） |
| `missing` | 0.87614 | **0.11386** | ✅（2 个，框均退化） |
| `recruit` | 0.88603 | **0.10397** | ✅（1 个，框退化） |
| `card` | 0.96015 | **0.02985** | ✅（1 个，框退化） |
| `headhunt` | 0.97021 | **0.01979** | ✅（`#main`，唯一非退化，但是容器） |

（gap 取自 `tmp/align/report.json` @ `4584759`。**转述里 `missing` 的 gap 写作 0.10397，与我实测的 0.11386 不符 —— 以我的为准**；0.10397 是 `recruit` 的值。）

**4 个未覆盖**：`box-summary`（0.10902）、`depot`（0.06648）、`gacha`（0.08749）、`help`（0.15057）。
**`state`** 已过门禁（0.99093），不在本题范围。

### `enemy` 与 `box-detail` —— 单独结论

> **它们是 B 类里残差最集中的两个（`enemy` 单洞占其残差 44.92%、`box-detail` 4 个命中全过 R1/R2），而声明轴永远回答不了它们 —— 两者都在那 10 个结构性不适用的模板里。**
> **这不是排期问题，是这条路对它们结构性无效。** 写清楚，免得有人以为再测一轮就能覆盖。

**举证方式**：`tmpl_geometry.json` 中 `Enemy.tmpl`（4 条尺寸声明、**0 条位置声明**）与 `BoxDetail.tmpl`（1 条尺寸声明、**0 条位置声明**）的扫描结果 + `common.css` 的 0 次位置声明。

## 7. 「哪个元素贡献最多」—— 本轮不出结论

**未测。** 两个原因，均已查证，不是回避：

1. **12/13 的框退化**（缺 height 或 width）⇒ 份额必然为 0，不能读作「这些元素没问题」；
2. 唯一非退化的 `#main` 是**容器**，覆盖 99.66% ⇒ 它回答的是「headhunt 的残差基本都在容器里」，**不是「哪个元素最差」**。

**跨场景一致性：不出结论**（不是「不一致」）。可比较的非退化样本量为 **1**。

## 8. 我那次尺子失效的诊断（保留一句）

第一版扫描器用**单行正则**解析 CSS，而模板是**多行块**（`.lh {` 与属性分行）⇒ 它**干净地报告 `Headhunt` 有 0 条几何规则**，而 `render.go:548` 明确引用 `.lh width:95px (Headhunt.tmpl:25)`。

**真相是我的尺子坏了，不是模板没有声明。** 改写为多行块解析后：**181** 条几何声明，`.lh { width:95px; height:190px }` 定位在 **`Headhunt.tmpl:26`**。

⚠️ **行号差异保留**：我的探针给出 **`:26`**（属性行），转述的表写 `:25`（选择器行）。**差一行，以我的为准，不抹平。**

**这条与 P18 的跨行少匹配同族、方向相反：两个方向都把「我没找到」伪装成「它不存在」。** 已进常设流程。

## 9. 常设流程执行情况

| 流程 | 执行 |
|---|---|
| P1 闸门 + 验牙齿第 8 次（**两种模式**） | ✅ §5（`teeth8b` / `teeth8c`，均 exit=2、无产物） |
| P6 没被测过的分支不算验证过 | ✅ 输入缺失分支本轮才真正测到 |
| P7′ `d["k"]` 禁 `.get()` | ✅ 读 `report.json` / `controls.json` 前先 dump 键名；`scene_of` 的 `KeyError` 是它生效的结果 |
| P7″ `PRE-AGG` + 断言非空 | ✅ `assert dm.sum() > 0` |
| P10 对照未过 ⇒ 未测 | ✅ §7 |
| P19′ 序关系对照（无阈值、前置） | ✅ 3 条，§5 |
| P20 先归因「机理读错了」 | ✅ §3 的 12 个 0 先查尺子（框退化）再谈场景 |
| P21「已改 ≠ 已验证」 | ✅ 13 标为**过滤后已确认**（过滤 inline 完成），取消暂定 |
| P23 机理陈述落盘 | ✅ `instrument.json` 的 `P23_mechanism_statement` + `ORD2_frame_leak` 反证控制 |
| P24 预登记绑「我在乎的量」 | ✅ 绑的是「元素框能否由模板逐行追溯」与「框是否非退化」 |
| P25 产物存在 ⇒ 报告必须存在 | ✅ 本报告与产物同批交付 |

## 10. 取证附录（命令 + exit code）

| # | 命令 | 结果 | exit |
|---|---|---|---|
| 1 | `git log --oneline -1` / `git diff --stat` | `4584759` / 无输出 | 0 |
| 2 | `find template -type f \| wc -l` | **20**（16 `.tmpl` + 4 个 js） | 0 |
| 3 | `grep -c "position:" assets/css/common.css` | 0，输出为空 | **1** |
| 4 | `grep -c "left" / "top" assets/css/common.css` | 各 0，输出为空 | **1** |
| 5 | `wc -c -l assets/css/common.css` | 417 字节 / 16 行 | 0 |
| 6 | 多行块解析重扫 | 181 条几何声明 | 0 |
| 7 | `position:absolute` 过滤 | **13**（过滤后已确认）；27 = 过滤前上界 | 0 |
| 8 | `byelem.py controls` | 3/3 序关系对照通过 | 0 |
| 9 | `teeth8b`：`controls_pass=false` → measure | `REFUSED: controls_pass != true` | **2** |
| 10 | `teeth8c`：移走 `controls.json` → measure | **`REFUSED: controls.json missing`** | **2** |
| 11 | `byelem.py measure`（对照通过后） | `elements.json` 产出，13 行 | 0 |
| 12 | 写 JSON 后 `json.load` 读回 | 全部成功 | 0 |
| 13 | 禁碰树 | 未读、未渲染、未落产物 | — |

**显式崩掉的缺陷**：`scene_of` 键名错（`KeyError: 'Box'`）—— P7′ 让它崩而不是返回 0。

## 11. 出身逐项标注

| 项 | 出身 |
|---|---|
| 13 个元素的选择器、模板行号、声明属性、设备框 | **measured**（`derivable_boxes.json`，逐行可追溯到模板） |
| 181 / 97 / 13 三级过滤计数 | **measured**（脚本 inline 过滤，`27` 为过滤前上界） |
| 12 个 0 源于框退化（缺 height/width） | **measured + derived**（先查尺子的结论） |
| `#main` 份额 0.99659 | **measured** |
| 10 个模板「结构性不适用」 | **measured**（`tmpl_geometry.json` + `common.css` 0 次位置声明，三次 grep exit=1） |
| `enemy` / `box-detail` 声明轴永远无效 | **measured**（两模板 0 条位置声明） |
| 各场景 gap | **measured**（`align/report.json` @ `4584759`）；**转述里 `missing`=0.10397 与实测 0.11386 不符，以我的为准** |
| 「哪个元素贡献最多」/ 跨场景一致性 | **未测**（非退化样本量 1，结构上不可评估） |
| `common.css` 417 字节 | **measured**（我的探针；与转述一致） |
| `template/` 20 个文件 | **measured**（`find`；转述称 17，以我的为准） |
| `.lh` 在 `:26` | **measured**；与转述的 `:25` **差一行，保留不抹平** |