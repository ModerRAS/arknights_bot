# Task 9 — JPEG 编解码噪声地板

```
provenance:   measured（数字） + 未测（判定）—— 两者在本报告里逐句分开，不合并
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align（复用，未新建）
              生成时间 2026-10-02T13:15:18Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task9-jpeg-floor.md
量具定义:      tmp/residual/jpegfloor/instrument.json（跑数前落盘，v1→v8 全部修订登记）
量具实现:      tmp/residual/jpegfloor/goenc/main.go（Go image/jpeg，与 fixtures 的 source 一致）
产物:          tmp/residual/jpegfloor/{controls.json, floor.json, controls.log,
               measure_stdout.log(0B), teeth5.log, ctrl/*.png}
输入:          tmp/pixel-compare/<scene>/{new,old}.png @ 4584759
produced_by:   src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47（逐行落盘）
```

**修改权声明**：未改任何渲染代码、未改任何计分代码、未 commit、未 push、未新建工作树。`git diff --stat` 空。禁碰树未读未渲。

---

## 0. Boss 点名的两个数

- **地板下界 ≥ 缺口的场景数 = 0**
- **地板下界 << 缺口的场景数 = 14**（算术分类）
- 另 2 个（`calendar` / `state`）**已过门禁**，不适用本判据，单独一档
- **校验式：0 + 14 + 2 = 16 ✓**（三值互斥且穷尽，脚本 assert）

⚠️ **但这 14 个判定本轮不具备证据资格 —— 标为「未测」。** 原因见 §2：量具对编解码噪声的**敏感性充分性未证明**。**数字已测，判定未测，两句分开写。**「未测」≠「已测得编解码不是主因」。

## 1. 地板下界：最高的几档（数字，measured）

| 场景 | 门禁缺口 | 往返亏损下界（最优 q） | 交叉核对自重编码亏损 | 推翻该判定所需低报倍数 |
|---|---|---|---|---|
| `headhunt` | 0.01979 | **0.004222** | 0.000754 | **4.7×** |
| `recruit` | 0.10397 | 0.002100 | 0.000481 | 49.5× |
| `box` | 0.16496 | 0.002035 | 0.000600 | 81.1× |
| `missing` | 0.11386 | 0.001379 | 0.000447 | 82.6× |
| `box-summary` | 0.10902 | 0.001233 | 0.000547 | 88.4× |
| `gacha` | 0.08749 | 0.001175 | 0.000148 | 74.5× |

全部 16 行的最佳质量都是 **q=100**（`fixture.jpegQualitySweep.qualities` 里质量最高的一档）。最低一档是 `operator` 0.00018（缺口 0.22134，需低报 1222.9× 才能推翻）。

**「推翻该判定所需低报倍数」这一列是把「未测」变成「需要多少证据」**：要否掉「编解码不是主因」，量具必须在这些场景上低报 4.7×–1222.9×。**它不构成通过，只标出需要多少证据。**

## 2. 敏感性：选 **(b)** —— 未证明（Lead 要求三选一，照实选）

**Lead 的判断正确，我同意 (b)。** 逐条对照：

| 性质 | 状态 | 证据 |
|---|---|---|
| **口径同一 / 可复现** | **measured ✓** | 量具 cmp 模式算 old.png vs new.png，与 task4 `align/report.json` 里 Go harness 自己算的 `go_sim` 比较：**16 场景 max delta = 0.000e+00**（逐位相同）。期望值来自已落盘产物，不是挑出来的强度 |
| **对编解码噪声有响应（敏感性）** | **未测 ✗** | 三条敏感性对照全部未达预登记门槛（见下） |

**三条敏感性对照的实测与各自失效原因**（都是**我的推导错**，不是尺子瞎）：

| 对照 | 期望（`estimated`） | 实测 deficit(q=70) | 我的推导错在哪 |
|---|---|---|---|
| 白噪声 | > 0.5 | **0.139233** | 我假设主导机制是 AC 量化。实测 deficit 几乎不随 q 变（0.1392→0.1291，q=100 时仍有 0.129）⇒ 主导机制是**色度子采样（4:2:0）**，不是量化。这也解释了量级只有我估的 ~1/4：子采样每 4 个横向样本只保留 1 个 |
| 8px 周期块对齐硬边 | > 0.05 | **0.001471**（q≥75 时**恰好 0.0**） | 我写「周期 8 = 8×8 块的最高 DCT 频率」。**错：周期 8 恰好是频率 bin 1，是最低的非 DC 频率**，量化得最细，重建几乎无损 |
| 纯色 + 1px 细线（局部） | > 0.001 | **0.000118** | 我把「误差上界 ≈0.023」当成了期望值，再拍一个 0.001 当门槛。上界不是期望；实测 1.18e-4，q≥95 时归 0 |

**关键区分（Lead 强调的那句）**：`go_sim` 校准证明的是**口径同一**，不是**灵敏度**；纯色对照 `S_rt=1.0` 只证明尺子**不会凭空造出残差**，同样不是灵敏度。三次失败使 (a) 不成立 ⇒ 落 **(b)**。

**因此**：
- 「地板 ≥ 往返残差」这个**下界方向仍然成立** —— 敏感性偏低不会让下界方向出错（偏低只会让真实地板更高）。
- 「地板下界 << 缺口 ⇒ 编解码不是主因」**本轮不具备证据资格**，标 **未测**。
- **编码器差异：未量化**（Chromium 与 Go `image/jpeg` 的振铃与量化伪影形态不同，那一项是往返测不出来的加项）。

## 3. 三次口径修订（全部在出表前，逐条登记）

| 版本 | 改了什么 | 为什么 |
|---|---|---|
| **v2** | 量纲修正 | v1 拿 **similarity** 去比 **deficit**（`floor_bound < gap`）。冒烟测试（card q=70/100）暴露后，在碰任何真实场景前改掉。阈值与方向未变 |
| **v3** | 对照断言：严格逐档单调 → 趋势断言 | 实测 q=80 比 q=75 差。**JPEG 量化表不是嵌套的**，逐档单调本就不保证 —— 我把编解码器的一个不成立的性质当成了量具验收条件 |
| **v4** | 停止条款 | 动态范围断言 0.0069 < 0.01，1px 棋盘加强后 0.009725 仍 < 0.01。**不降阈值、不第三次改输入** |
| **v5** | 控制集替换（透明记录，见 §4） | 用「复现 `go_sim`」替换幅度断言 —— 但**失去了一条证明力**，见 §4 |
| **v6** | 判定二值 → 三值 | 首次 measure 后发现：对**已过线**场景 gap 为负（calendar −0.00432、state −0.00093），`deficit ≥ gap` 平凡成立 ⇒ 把已过线场景叫「结构性不可达」是**范畴错误**。已改为 `(1) gap≤0 → already_passes_gate (2) gap>0 且 deficit≥gap → 结构性不可达 (3) gap>0 且 deficit<gap → 编解码非主因`，16 行互斥穷尽 |
| **v7 / v8** | 敏感性对照按 Lead 顺序补入 | 见 §2。三条全部未达门槛 ⇒ 落 (b) |

**v4 条款（Lead 精确化后，写入常设流程）**：
> **禁止在既有失败输入上调阈值；禁止按失败样本挑输入；允许更换为性质不同的新输入实例来建立仪器属性。**
> 「不许重试」与「不许换输入」是两件事，混成一条会禁止掉唯一正确的做法。

## 4. 控制集替换：失去了哪一条证明力（**本节不许省略**）

**原合成对照想证明的**：这条尺子**对 JPEG 往返噪声有响应**（敏感性）。
**替换成的 `go_sim` 校准对照证明的**：这条尺子**与 harness 的度量逐位一致**（口径同一 / 可复现）。
**替换后失去的证明力：敏感性。** 二者不是替代关系 —— 校准无法回答「尺子看不看得见编解码伪影」。

**这个替换本身是有代价的、我在替换时并未完全意识到的一步。** 校准更强的地方是它对着已落盘产物校验口径；它弱的地方是它对敏感性**零覆盖**。我一度把 `controls_pass=True` 当成了「量具已验证」，而实际上当时只验证了口径同一。**Lead 的追问把这个错误抓了出来。**

## 5. 口径与被 Boss 点名的两项（均照做，未改写）

- **`jpegQualitySweep.rule` 全文**（读完再动手，不照截断句猜）：
  > **"A quality choice never waives a failed uncompressed PNG layout comparison."**

  这是一条**策略约束**：选 JPEG quality 永远不能豁免一次未通过的未压缩 PNG 布局比较。**因此本轮测出的 quality-floor 只能用于判断可达性，绝不可被用来在门禁里换格式换分数；本报告不提出任何此类用法。**

- **`candidateFormats`**：**不在顶层键里**（顶层为 version / fixtureService / viewport / browserDeterminism / font / remoteResourceCache / jpegQualitySweep / scenes / newRendererAcceptance）。按 Lead 要求**去搜而不是假定不存在**：`grep -rn "candidateFormats" src/ggrender/` 命中 `fixtures.json` 每个 scene 条目，**16 个场景全是 `["png","jpeg"]`** ⇒ JPEG 往返对全部 16 场景都是合法对照。「输出用哪个格式」本身是一个尚未被检验的选择 —— 本轮把它量化了，**不替它做决定**。

## 6. 口径三条的遵守

1. **主测量不许把基线当输入**：主测量全部用**我们自己的渲染** `new.png` 做 encode→decode→与原图比。基线仅出现在**交叉核对**项（解基线 → Go 重编码 → 解码 → 自比），产物里字段名与 `crosscheck_label` 显式标注 `instrument noise floor (self re-encode); NOT a render input`。红线是「基线不得作为渲染输入或背景」，**自比不构成渲染输入**。
2. **报成下界**：全文用「地板 **≥** 往返残差」，并单列「**编码器差异：未量化**」。
3. **与门禁缺口对比**：三值互斥穷尽判定，**校验式 assert 通过（0+14+2=16）**。

## 7. 闸门验牙齿 —— 第 5 次：**通过**

| 项 | 结果 |
|---|---|
| 抽掉闸门输入（移走 `controls.json`）→ `floor.py measure` | `REFUSED: controls.json missing` |
| 退出码 | **2** |
| `floor.json` | 未产出（`ls` 退出码 2 = 不存在） |
| 恢复后 `json.load` 读回 | `controls_pass=True` |
| measure 阶段 | **stdout = 0 字节、stderr = 0 字节**（PRE-AGG 与逐场景行进 stderr 属 instrumentation 行；正式结果只写 `floor.json`） |

**另附一次额外证据**：在最终控制集（敏感性三条未过）下再跑 measure → `REFUSED: controls_pass != true`，**exit=2**。即闸门不只挡「输入缺失」，也挡「对照未过」——**没有在对照失败的状态下偷偷产出数字**。

## 8. 取证附录（命令 + exit code）

| # | 命令 | 结果 | exit |
|---|---|---|---|
| 1 | `find . -name fixtures.json` | 确认只有一份，路径 `./src/ggrender/testdata/visual/fixtures.json` | 0 |
| 2 | `python` dump `list(d.keys())` | 顶层 9 键，**无 `candidateFormats`** | 0 |
| 3 | 读 `jpegQualitySweep.rule` 全文 | `A quality choice never waives a failed uncompressed PNG layout comparison.` | 0 |
| 4 | `grep -rn "candidateFormats" src/ggrender/` | 命中 `fixtures.json`，16 场景全 `["png","jpeg"]` | 0 |
| 5 | `go build`（量具） | 成功，`jpegrt.exe` | 0 |
| 6 | 冒烟：card q=70/100 | 0.99456 / 0.99922（触发 v2 量纲修正） | 0 |
| 7 | `floor.py controls`（v1 断言） | 严格单调对照失败 → v3 | 1 |
| 8 | `floor.py controls`（1px 棋盘加强） | 动态范围 0.009725 未达 → v4 停手 | 1 |
| 9 | 校准：16 场景 `cmp` vs `go_sim` | max delta **0.000e+00** | 0 |
| 10 | 闸门牙齿 #5：移走 `controls.json` → measure | `REFUSED`，无 `floor.json` | **2** |
| 11 | `floor.py measure`（当时阻塞集通过） | 16 行，checksum True | 0 |
| 12 | 加入三条敏感性对照后 `floor.py controls` | 三条全未达门槛 → 落 (b) | 1 |
| 13 | 最终控制集下再跑 measure | `REFUSED: controls_pass != true` | **2** |
| 14 | `go_sim` 字段名 | `KeyError: 'similarity'` ⇒ 实为 `go_sim`（**P7′ 起作用**） | 1 |
| 15 | 写 JSON 后 `json.load` 读回 | 全部成功 | 0 |
| 16 | 禁碰树 | 未读、未渲染、未落产物 | — |

## 9. 出身逐项标注

| 项 | 出身 |
|---|---|
| 各场景 `codec_deficit_best` / `gap_to_gate` / 交叉核对自重编码亏损 / 推翻所需倍数 | **measured**（`floor.json`，Go `image/jpeg` 实跑，qualities 照抄 fixtures） |
| `go_sim` 校准 max delta = 0.000e+00 | **measured** |
| 三条敏感性对照的实测值与各自失效原因 | **measured**（值）+ **inferred**（失效归因于我的推导错误，未另做实验证伪） |
| 「地板下界 << 缺口 ⇒ 编解码不是主因」 | **未测**（敏感性充分性未证明，Boss 裁定 (b)） |
| 「没有任何场景是结构性不可达」 | **未测**（同上；算术上为 0，但推断资格不足） |
| 期望值 0.5 / 0.05 / 0.001（白噪声 / 块对齐 / 局部） | **estimated**，推导链写在 instrument；**三条全部被实测推翻**，推导错误已逐条指认 |
| 「编码器差异：未量化」 | **asserted**（口径要求，非测量） |
| 「quality 最佳档恒为 q=100」 | **measured** |
| 「白噪声 deficit 几乎不随 q 变 ⇒ 主导机制是色度子采样」 | **inferred**（由 deficit(q) 曲线形状推出，未单独验证 4:2:0） |
