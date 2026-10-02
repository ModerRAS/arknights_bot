# Task 8 — C 类分层残差：阳性对照未通过，**不交表**

```
provenance:   measured（本轮结论是「未测」，不是「已测得为否」——这两者的区别是全文主线）
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align（复用，未新建）
              生成时间 2026-10-02T12:53:18Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task8-cclass-strata.md
分层口径:      tmp/residual/strata/instrument.json（跑数前落盘；v1→v3 修订全部登记）
脚本/产物:    tmp/residual/strata/{strata.py, controls.json, controls.log}
              tmp/residual/mapping/{instrument.json(前提修复), premise_survey.json}
输入:          tmp/pixel-compare/<scene>/{old,new}.png @ 4584759
produced_by:   src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
```

**修改权声明**：未改任何渲染代码、未改任何计分代码、未 commit、未 push、未新建工作树。`git diff --stat` 空。禁碰树未读未渲。

---

## 0. 一句话

**阳性对照未通过 ⇒ 按 Boss 指令不交表。** 我对分层口径做了两次有据可查的修正（v2 换特征、v3 修对照构造），第三次失败后按自己在 instrument 里写下的上限停止，**没有为了让它通过而调第四次阈值**。

**⇒ 第 3 项（主导类别、跨场景一致主导）与第 4 项（预登记三选一）本轮都不成立。**

⚠️ **必须说清：「三选一未给出」不等于「选 C」。** 「未测」与「已测得三类分散」是两件事 —— 把未测的路径记成已否决，正是 Boss 上一轮专门警告的那类错误（「否则下一个人会把一条未测的路径当成已否决的路径而永远不再看它」）。C 的成立条件是「**测过了**、三类分散、无一致主导」，而本轮是「**没测成**」。

## 1. 阳性对照：三轮，逐轮都有取证

| 轮 | 对照结果 | 诊断（数据） | 处置 |
|---|---|---|---|
| v1 | **失败**：`glyph_edge = 0.0`，合成文本残差 **100%** 落进 `sep_decor` | 我数的是**边缘带像素**的密度；边缘带天生细（22px 字号下 3×3 范围带在 9×9 窗口只占 ~3/81=0.037），永远 < 0.15 | **换特征**：`density9` → `solid9`（9×9 窗口内**非平坦**像素占比）。阈值带 0.15/0.60 **未动**（它们本来就是按物体几何推的：1px 线 9/81=0.111、2–3px 笔画 0.22–0.33、连续照片 >0.6） |
| v2 | **失败**：`glyph_edge = 0.0`，仍 100% `sep_decor` | 诊断到**对照构造**问题：我只扰动 `range3>=T_edge` 的像素 —— 那是抗锯齿梯度上对比度最高的**单个像素**，不是「字形边缘带」。实测该处 3×3 非平坦窗口的 9×9 邻域有 50/81 是背景 | **修对照**：扰动 `range3 > T_flat` 的全部非平坦像素（字形 + AA 带），这是「残差全在字形上」的忠实实现。同时修图像对照的期望（见下） |
| v3 | **失败**：`glyph_edge = 0.0`，96.91% `sep_decor`、3.09% `other` | 口径在 9×9 窗口尺度上**测不出字形笔画与 1px 分隔线的差别** —— 两者都很薄，都落在 `<0.15` | **停止**。按 instrument 里写下的「只做这一次修正；再失败就交失败」 |

**修正后用全新对照实例复验**（不同文本、不同字号 26px、不同素材裁切、不同缩放比），**没有对着失败的那份调参**。

### 1.1 两条对照**通过了**，必须一起交

| 对照 | 结果 | 读法 |
|---|---|---|
| `NEG_flat_colour_bias`（纯色偏） | `flat_bg = 1.000` | 分层口径**不会**把颜色差异误报成文本 —— 方向正确 |
| `NEG_image_resampling`（素材重采样） | 主导类 = `other`（`glyph_edge = 0.0`） | 分层口径**不会**把纹理误报成文本 |

⇒ **该口径是系统性地「少报文本」，不是「多报文本」。** 这条方向性很重要：**它无法用来排除「文本是主因」这个假设** —— 一个只会把字形判成分隔线的仪器，对文本残差会给出 0。

### 1.2 这轮失败**证明**了什么（比份额表更值钱）

在 9×9 窗口的局部厚度尺度上，**「字形笔画」与「1px 分隔线/边框」不是可分的类别**。可分的判据需要另一族特征：笔画宽度变换（stroke-width transform）、连通域的长宽比与方向一致性（structure tensor）、或文本行的共线排列 —— 这些都超出本轮预算，且每一族都需要自己的阳性对照。本轮不猜、不估、不许诺。

## 2. 第 3 / 第 4 项：本轮不成立

- **每个场景的主导类别**：不报。没有通过验证的仪器，主导类别就是无证据的数字。
- **跨场景一致主导的那几类**：不报。同上。
- **预登记三选一（文本 / 图片素材 / 三类分散）**：**不给出**，且**不得记为 C**。
  - Boss 的指令是「必须三选一，不许写视情况而定」，本轮的执行是：**「未测」这个状态被显式记录，而不是被伪装成三个选项之一。** 在阳性对照未过的数据上选 A/B/C，无论选哪个都是在拟合。
  - 三选一的判定口径（一致性阈值 ≥6 个场景、主导规则、平局规则）**已落盘在 instrument 里**，仪器修好后可直接复用，无需重新讨论口径。

**C 类的两个已知不同质事实仍然成立，与本轮失败无关**（来自 task5/task6 的既有实测）：`box-summary` 有 13 个内洞但都小于 `A_min`、`depot` 中位数高但集中度 `G=1.10`。C 类不可当同质清单排优先级。

## 3. 第 5 项：检测器前提修复（已交付）

**「平坦 ⇒ 空」→「平坦 ⇒ 沉默」。**

| 项 | 结果 |
|---|---|
| 代码落点 | `tmp/residual/mapping/premise_survey.py` 的内洞判定 |
| 语义改动 | 内洞不再输出任何真值字段；恒为 `bucket="silent_not_painted"` + `premise="flat => silence (detector has no standing here)"` |
| 验证（落盘复核） | `premise_survey.json` 顶层新增 `premise: "flat => silence"`；每个 hole 条目的键为 `pixel_box / area_frac / min_dim / meets_Amin / bucket / premise`，**已确认不再存在 `verdict` / `R1` / `R2` 任何真值语义字段** |
| 是否重跑 B | **否**。B 类已闭合，本轮只修仪器，没有新的洞假设 |

### 闸门验牙齿 —— 第 4 次：**通过**

| 项 | 结果 |
|---|---|
| 抽掉闸门输入（移走 `controls.json`）→ `premise_survey.py measure` | `REFUSED: controls_pass != true` |
| 退出码 | **2** |
| 结果文件 | 未被覆盖/产出（修复前的旧产物已另存 `premise_survey_v2_backup.json` 对照） |
| 闸门恢复后重跑 | exit=0，**stdout = 0 字节、stderr = 0 字节** |

累计 4 次闸门验牙齿（task6 一次、task7 两次、本轮一次），全部 `REFUSED` / `exit=2` / 无结果文件。

## 4. P7″（聚合前先断言非空并打印大小）—— 已执行

每次 `stratify()` 开头即取 `n_diff = int(diff.sum())`、打印 `PRE-AGG n_diff=<N>`，随后 `assert n_diff > 0`。实测三轮对照的打印值：v1 `9551 / 36000 / 25479`，v3 `14288 / 36000 / 23166` —— **没有一个空集合上的聚合被写进产物**。若某场景差异像素为 0，`assert` 会直接让它崩，而不是产出一行「份额 0」。

## 5. 「其它」这一栏的处理（未产生，但口径已登记）

instrument 里 `other` 的定义是 `T_flat < range3 < T_edge`：既非平坦、也够不上边缘阈值。**它被显式登记为「本项最大的陷阱」**，口径原文写着：若它份额很大，**那是未测量，不是已归类**。

本轮 `other` 只出现在合成对照里（图像重采样 51.87%、文本对照 3.09%），**没有进入任何真实场景的统计**，因为真实场景的统计本轮不产出。

## 6. 取证附录（命令 + exit code）

| # | 命令 | 结果 | exit |
|---|---|---|---|
| 1 | `git log --oneline -1` | `4584759` | 0 |
| 2 | `git diff --stat` | 无输出（未改被跟踪文件） | 0 |
| 3 | `strata.py controls`（v1） | POS 失败，`glyph_edge=0.0` | 1 |
| 4 | `strata.py controls`（v2） | POS 失败，`glyph_edge=0.0` | 1 |
| 5 | `strata.py controls`（v3） | POS 失败，`glyph_edge=0.0`，96.91% `sep_decor` | 1 |
| 6 | 闸门验牙齿 #4：移走 `controls.json` → `premise_survey.py measure` | `REFUSED: controls_pass != true` | **2** |
| 7 | 恢复 `controls.json` → `json.load` 读回 | `controls_pass=True` | 0 |
| 8 | `premise_survey.py measure`（修复后） | 16 行；`stdout=0B / stderr=0B` | 0 |
| 9 | 写 JSON 后 `json.load` 读回 | 全部成功 | 0 |
| 10 | 禁碰树 | 未读、未渲染、未落产物 | — |

**本轮踩到并显式崩掉的两个代码缺陷**（都是崩掉，不是返回 0）：① `synth_text` 里 `enumerate([..., "line %d" % i])` 引用尚未绑定的 `i` → `UnboundLocalError`；② `features()` 增参数后 `solid` 仍引用旧局部名 `t` → `NameError`。**显式失败是安全的，安静的 0 是危险的。**

## 7. 出身逐项标注

| 项 | 出身 |
|---|---|
| 三轮阳性对照失败与其诊断数据 | **measured**（`controls.json` + 每轮 `controls.log`） |
| 两次修正的性质（换特征 / 修对照构造）与依据 | **measured**（instrument 的 `amendment_v2` / `amendment_v3` 原文） |
| 「停止、第三次不调」 | **按 instrument 预登记上限执行**（`amendment_v3.capped`） |
| 「三选一不给出、且不等于 C」 | **measured + assert**：口径成立的必要条件（通过验证的仪器）未满足 |
| 两条阴性对照通过（flat 色偏 → flat_bg 1.000；重采样 → 非 glyph_edge） | **measured** |
| 「该口径系统性少报文本，无法据此排除文本假设」 | **measured**（三轮 `glyph_edge` 恒为 0） |
| 前提修复「平坦 ⇒ 沉默」 | **measured**（产物字段已复核，真值字段已消失） |
| 闸门验牙齿第 4 次 | **measured**（`REFUSED` / exit=2 / stdout=stderr=0B） |
| 「9×9 尺度上字形与分隔线不可分，需换特征族」 | **inferred**（由三次对照失败推出，未做进一步实验；已声明不猜不估） |
| 真实场景的分层份额、主导类别、跨场景一致性 | **未测**（无证据资格） |
