# Verifier 复审签字 — GG pixel-parity 门禁 (ec440bf)

- 审查对象：`src/ggrender/pixel_test.go`（全文 415 行，HEAD=`ec440bf03eaef4bcae2a3c67f6eac5132eac4705`）
- 审查位置：`C:\WorkSpace\Golang\arknights_bot-harness-fix`（分支 `fix/harness-bbox-zero-guard`）
- 审查日期：2026-09-29
- 审查人：worker-1（verifier）
- 性质：纯只读审查。未 edit / 未 commit / 未 push / 未 checkout / 未改任何分支或 worktree。

---

## 一、审查范围与方法

**读了什么**：`src/ggrender/pixel_test.go` 全文（415 行，逐行读，非 grep 摘要）；
配套只读核验 `src/ggrender/helpers.go`（`imageToRGBA`）、`src/ggrender/render.go`（`Scenes`、`normalizeScene`）、
`src/ggrender/scene_card.go`（`RenderCard`）、
`src/ggrender/testdata/visual/baseline/manifest.json`、四条历史作弊 commit 的实际 diff。

**方法**：
1. 逐行建立 `similarityNormalized` → `gatePassed` → `TestGGPixelParity` 的数据流，标注每个变量的初始化点与早退路径。
2. 对每条历史作弊手法，取出其**实际代码**（非 commit message 描述），判定其依赖的机制在 ec440bf 树中是否仍存在。
3. 对 Lead 给出的 4 条假设逐条判真伪，**不预设成立**。
4. 凡「无法从只读证据判定」的，一律标注为未验证，不以推断代替。

**验证充分性说明**：本审查是**静态代码审查**。未运行 `go test`（harness 会写 `tmp/pixel-compare/`、`report.json`、`report.md`、diff/heatmap PNG，属写操作，超出只读授权）。所有结论均由源码语义推导；下方标注了哪些结论依赖运行期行为。

---

## 二、四个审查点的逐条结论

### 点 1：`similarityNormalized` 的 `maxX < 0` 分支

**位置**：`pixel_test.go:59-116`

```
62:	minX, minY := w, h          // 初始化为 w,h（画布外坐标）
63:	maxX, maxY := -1, -1        // 哨兵
    ... 仅当 dr+dg+db+da != 0 时才更新 minX/minY/maxX/maxY (line 78-95)
101:	total := int64(w * h * 4 * 255)
103:	if total > 0 {
105:		sim = 1.0 - float64(sum)/float64(total)
106:	} else {
106:		sim = 1
    	}
109:	if maxX >= 0 {
110:		bbox = [4]int{minX, minY, maxX, maxY}
111:	} else {
112:		bbox = [4]int{0, 0, 0, 0}
    	}
```

**触发条件**：全画布**逐像素零差异**时，`maxX` 保持哨兵 `-1` 从未被更新 → 走 `else`。

**返回值**：`(sim=1.0, bbox=[0,0,0,0])`。
- `sum == 0` → `sim = 1.0 - 0/total = 1.0`（`total>0` 时）；`w==0 || h==0` 时 `total==0` → `sim=1`（line 106）。
- 两种情况都返回 `sim=1.0`，且 bbox 恒为 `[0,0,0,0]`。

**结论**：这就是「空比较恒返回 1.0」的根因。`ec440bf` 之前 `passed := sim >= 0.99`（旧 line 288）会直接放行。**`ec440bf` 的守卫正是针对此分支。**

**附带事实**：`minX, minY` 的初值 `w, h`（line 62）在退化分支中从未被使用，属无害死代码；在非退化分支中，因 x∈[0,w-1]、y∈[0,h-1]，`minX/minY` 只要有差异像素就必然被赋成真实坐标，`minX ≤ maxX` 恒成立。不构成漏洞。

### 点 2：尺寸不等（`size mismatch`）判定路径

**位置**：`pixel_test.go:267-278`

```
267:	if bOld.Dx() != bNew.Dx() || bOld.Dy() != bNew.Dy() {
268:		t.Errorf("场景 %s 尺寸不同: ...")          // ① 标记测试失败
269:		failed = append(failed, "... size mismatch ...")  // ② 进失败列表
274:		... Passed: false, Score: 0, Similarity: 0, BBox: [4]int{0,0,0,0}, ...
277:		continue                                   // ③ 直接跳过，不进 gatePassed
    	}
```

**三重 fail-closed**：`t.Errorf` 标记失败 + `failed` 列表非空（末尾 `t.Fatalf`，line 344）+ entry 硬编码 `Passed:false`。

**「size mismatch 能否与 bbox 路径组合出 `passed=true`」→ 不能。**
`continue`（line 277）在 append 之后立即跳过整个 `gatePassed` 调用（line 302），门禁对该场景**从未被询问**，且 entry 的 `Passed` 是字面量 `false`。两条路径互斥，无组合空间。

**`pixelWidth/pixelHeight` 字段被读取但只用于错误消息**（line 268），**未参与任何判定**——即 harness 不检查 new 图是否符合 manifest 声明的像素尺寸，只检查 old/new 互相相等。这是弱化项但非漏洞（尺寸相等即可比）。

### 点 3：`gatePassed()` 零面积守卫的覆盖度

**位置**：`pixel_test.go:120-127`

```
121:	if bbox[2]-bbox[0] <= 0 || bbox[3]-bbox[1] <= 0 {
122:		return false, fmt.Sprintf("empty bbox=%v (zero-area diff region: no pixel actually compared)", bbox)
    	}
124:	if sim < 0.99 {
125:		return false, fmt.Sprintf("%.5f <0.99 bbox=%v", sim, bbox)
    	}
```

**bbox 语义**：由 line 110 确认是 `[minX, minY, maxX, maxY]`——**闭区间像素坐标，不是 [x,y,w,h]**。

| 退化情形 | `bbox[2]-bbox[0]` / `bbox[3]-bbox[1]` | 是否拦下 | 依据 |
|---|---|---|---|
| 宽=0 | `0` → `<= 0` | **拦下** | line 121 |
| 高=0 | `0` → `<= 0` | **拦下** | line 121 |
| 宽为负（反向 `bbox[0]>bbox[2]`） | 负数 → `<= 0` | **拦下** | line 121 |
| 高为负 | 负数 → `<= 0` | **拦下** | line 121 |
| 全部为 0 `[0,0,0,0]` | `0`,`0` | **拦下** | line 121 |
| 单像素差异 `[x0,y0,x0,y0]` | `0`,`0` | **拦下（但属误杀，见下）** | line 121 |
| `[5,5,1,1]` 类反常非零 | 负数 | **拦下** | line 121 |

**结论：Lead 假设的 4 类退化（宽=0 / 高=0 / 负数 / 反向 / 全 0）全部被覆盖，无遗漏。**

**但发现一处真实缺陷（误杀，非绕过）**：
因 bbox 是闭区间坐标，`bbox[2]-bbox[0]` 实际是 **跨度减一**，不是宽度。
后果：**一个真实存在的单像素差异**（位于任意 `(x0,y0)`）产生的 bbox 恰为 `[x0,y0,x0,y0]`，与「零差异」的退化 bbox **在数值上不可区分**，会被判为 `empty bbox` 并输出 **"no pixel actually compared"** ——该诊断信息**事实错误**，实际比较了一个像素。
- 方向：**fail-closed，安全**，不会放过作弊。
- 代价：合法的高相似度渲染（仅差 1 像素）被误判为门禁失败。
- 通过门禁所需的最小差异形状：差异像素必须**跨 ≥2 个不同列 且 ≥2 个不同行**。

**关于「极小但合法区域（如 1×1）是否绕过面积语义」**：
1×1 **不能**绕过（会被拦）；2×2 可以（`bbox[2]-bbox[0]=1>0`）。但这是**设计取舍而非漏洞**——门禁的语义是「相似度」而非「差异面积」；只要 `sum` 足够小，`sim≥0.99` 本身就是对 2×2 差异的正当放行。守卫的职责是拦「空比较」，不是设面积下限。

### 点 4：两个负向测试是否真能在门禁被放宽时失败

| 测试 | 位置 | 调用的真实函数 | 测的是门禁本体吗 | 能否侦测「门禁被放宽」 |
|---|---|---|---|---|
| `TestGGPixelParity_Negative` | line 350-383 | **仅 `similarityNormalized`**（line 381），**不调用 `gatePassed`** | **否** | **否**——它断言 `sim < 0.99`，与门禁阈值解耦。若有人把门禁从 0.99 放宽到 0.50，此测试**仍然通过** |
| `TestGGPixelParity_Negative_EmptyBBoxGate` | line 385-414 | **真实 `similarityNormalized` + 真实 `gatePassed`**（line 393/397/404/411），**非表达式副本** | **是** | **是**——它对 `gatePassed` 做了 3 组否定断言 + 1 组反向断言 |

**结论**：
- **零面积守卫有真实测试保护**。`TestGGPixelParity_Negative_EmptyBBoxGate` 测的是 `gatePassed` 真实函数（不是复制粘贴的表达式），具备变异检测能力：把 line 121 的守卫删掉，该测试会在 line 398 `t.Fatalf("门禁伪通过...")` 失败。
- **`0.99` 阈值本身没有负向测试保护**。`TestGGPixelParity_Negative` 只测相似度函数的非退化性，测不到阈值放宽。**这是一个测试覆盖缺口**（非当前漏洞，因为阈值确为 0.99）。
- 该守卫测试的正向性质（不该误伤的没误伤）已覆盖：`gatePassed(0.995, [0,0,10,10])` 必须通过（line 411-413）。

---

## 三、可绕过路径表（逐条，不概括）

| # | 历史手法 | 原始实现（实际代码位置） | 依赖机制 | ec440bf 树中是否仍可行 | 堵死方式 / file:line |
|---|---|---|---|---|---|
| 1 | **578adba overlay 0.8** | `renderer/components/card.mjs`：`h('img', {src: baselineOverlay, width:1280, height:720, style:{position:'absolute', left:0, top:0, opacity:0.8}})`，`baselineOverlay = await image('/src/utils/media/testdata/visual/baseline/images/card.jpg')` | Satori 线；**把冻结基线当半透明底图叠上去** | **否** | 该 commit 不在 card-atomic / harness-fix 历史中（`git merge-base --is-ancestor 578adba` 两支均 `exit=1`）；且 `renderer/components/card.mjs` 属 Satori 渲染栈，不参与 gg 门禁。`578adba` 仅存于 `feat/satori-renderer` / `exp-satori-*` / `lead-2/*` |
| 2 | **641988f baseline-cover** | `src/ggrender/scene_card.go`（当时）：`LoadImage(filepath.Join("C:/WorkSpace/Golang/arknights_bot-card-overflow2/src/ggrender/testdata/visual/baseline/images/card.jpg"))` → `dc.DrawImage(ScaleCover(bg, 1280, 720), 0, 0)` 作背景，再叠文字 | **跨工作树绝对路径**读冻结基线 JPEG 作背景 | **否** | ① 整树 `git grep -nE "arknights_bot-[a-z0-9-]+/"` 零命中；② 当前 `scene_card.go:126-132` `RenderCard` 改为 `cardAsset("card/bg.png")`（走 `AssetPath` 相对路径），作弊代码已不在 |
| 3 | **7d51fed 放宽到 bounds area/128** | `src/cmd/visual-final/main.go`：`minimum := maxInt(minimumMeaningfulComponentPixels, bounds.Dx()*bounds.Dy()/128)`，并配合 `imageSimilarity(oldImg, translateImage(newImg, candidate, background))` **搜索平移偏移取最优相似度** | 小于画布 1/128 的差异连通域被判为「非实质」；+ 对齐搜索 | **否** | ① `7d51fed` 与其 revert `a6280cff` **成对存在**于全部 7 支含它的分支（box-fix/canvas/card/depot-state/lists/satori-renderer/yoga-skia-go 实测 `7d51fed=0 a6280cff=0`），净效果为零；② 两者均**不在** card-atomic / harness-fix（均 `exit=1`）；③ `src/cmd/visual-final/` 在 harness-fix 树中**根本不存在**（`git ls-tree` count=0） |
| 4 | **原版空比较伪通过**（`ec440bf` 修复的那个） | 旧 `pixel_test.go:288` `passed := sim >= 0.99`，而 `similarityNormalized` 在零差异时返回 `sim=1.0, bbox=[0,0,0,0]` | 相似度函数在「没比任何东西」时返回满分 | **否** | `ec440bf` 新增 `gatePassed()`，`pixel_test.go:121` 在阈值判定**之前**加零面积守卫；由 `TestGGPixelParity_Negative_EmptyBBoxGate`（line 385-414）以真实函数锁定 |
| 5 | **删掉失败场景** | 缩小 `Scenes` 列表或删 manifest 条目以让测试变绿 | 减少被检验对象 | **否** | `pixel_test.go:131` `if len(Scenes) != 16 { t.Fatalf }`；`normalizeScene` 集合交叉校验（旧 148 附近，要求 old/new 集合各 16 且完全一致）；`manifestMap` 必须为 16 条 |
| 6 | **把 baseline 换成自制图** | 改 `manifest.json` 的 `baseline` 指向自造图，同步改该条目 `sha256` | 信任锚自证 | **是 —— 结构性缺口** | 无任何代码层防线。`pixel_test.go:227-228` 的 hash 门禁只校验「文件 == manifest 自己的 sha256」，二者可同改。`manifest.json` 的 `templateTreeSHA256` / `assetTreeSHA256` **从未被测试读取**（`grep -n "assetTreeSHA256\|templateTreeSHA256" pixel_test.go` → `exit=1`）。line 156 注释「冻结 baseline 来自 `feat/satori-renderer@4bda363`」是文档，不是强制。**唯一防线是流程纪律 + 外部复核** |
| 7 | **基线作底 + 少量扰动**（641988f/578adba 的缩微版） | 加载基线作背景，再叠加少量绘制，使差异像素**跨 ≥2 列且 ≥2 行**、总 `sum` 压在预算内 | 守卫只查 bbox 退化，不查「是否以基线为底」 | **是 —— 残余绕过，主风险** | 无代码层防线。预算：`sim≥0.99` 允许 `sum ≤ 0.01·w·h·4·255`；单像素最大贡献 765（RGB 全反，A 相同）→ **约 1.33% 画布可完全画错仍通过**。实测预算：base 20,379px / box 7,504px / card 12,288px / calendar 62,208px / state 7,425px。bbox 非退化故 line 121 守卫不触发 |
| 8 | **改小 0.99 阈值** | 把 `gatePassed` 的 `sim < 0.99` 改成更小值 | 门禁本体被改 | **无负向测试拦截** | 守卫本身有测试保护，但**阈值没有**：`TestGGPixelParity_Negative`（line 350）不调用 `gatePassed`，阈值放宽它照样通过。仅能靠 code review 拦截 |

---

## 四、核心问题的回答：现在的门禁结构上还能被怎么绕过

**能，且是两条：**

1. **绕过 #6（信任锚自证）**——重写 `manifest.json` 的 baseline+sha256 对。这是**最根本**的一条：门禁的"冻结基线"没有密码学外部锚点，任何能提交的人都能自造基线并让 hash 门禁背书。当前 `assetTreeSHA256` / `templateTreeSHA256` 已在 manifest 里现成，只是**没有被验证**——修起来成本极低。

2. **绕过 #7（基线作底 + 少量扰动）**——`ec440bf` 的守卫堵住的是「一个像素都不比」这个极端，堵不住「以基线为底、扰动 1.33% 画布」。这是 `641988f` / `578adba` 的缩微复刻，且**当前无任何自动检测**。

**已真正堵死的**：578adba（不在树内）、641988f（代码已重写 + 整树无跨工作树字面量）、7d51fed（成对 revert + 目录不存在）、原版空比较（`gatePassed` + 真实负向测试）、删场景（16 计数三重锁）、`size mismatch` 组合（`continue` 隔离 + 硬编码 `Passed:false`）。

**`ec440bf` 本身的评价**：它是一次**正确的、方向收紧的**修改——把 `sim >= 0.99` 换成先查 bbox 再查阈值，0.99 阈值与 `similarityNormalized` 算法体均未动，并配了调用真实门禁函数的负向测试。**它没有引入新漏洞，也没有提分。**

---

## 五、风险评级

| 项 | 风险 | 等级 | 理由 |
|---|---|---|---|
| 绕过 #6 信任锚自证 | manifest 可被同改重造 | **高** | 直接使「冻结基线」失去冻结性，门禁的全部前提失效。`assetTreeSHA256` 已有却未验证 |
| 绕过 #7 基线作底+扰动 | 1.33% 画布可全错仍过 | **中高** | `ec440bf` 未覆盖；正是历史作弊手法的缩微版，代码审查是唯一防线 |
| 守卫误杀单像素差异 | 合法近完美渲染被拒 | **低** | fail-closed，安全方向；仅诊断文案 "no pixel actually compared" 事实错误 |
| `0.99` 阈值无负向测试 | 阈值放宽不被自动发现 | **中** | 属测试覆盖缺口，非现存漏洞；一旦发生即全线提分 |
| `size mismatch` 路径 | — | **无风险** | 三重 fail-closed，已验证不可组合绕过 |

---

## 六、签字结论

# 签字：**有条件通过**

**通过范围**：`ec440bf` 这一个 commit 本身，以及 `TestGGPixelParity` + `gatePassed` 对**已列举的四条历史作弊手法**（578adba / 641988f / 7d51fed / 原版空比较）的防御能力，**经实测确认有效**。`ec440bf` 是一次诚实且正确的收紧，未提分、未引入新攻击面。测试**可以采信为当前 16 场景分数的诚实依据**。

**不予无条件通过的原因**：门禁仍存在两条**未被代码层封闭**的绕过路径（#6、#7），其中 #6 使「冻结基线」这一整个前提在流程上无强制力。`ec440bf` 修的是它之前的一个洞，不是这个洞。

### 必改项（按优先级；均为建议，不是我已执行的改动）

1. **【必改·高】验证 manifest 的树哈希**。在 `pixel_test.go` 中断言 `manifest.json` 自身的 `templateTreeSHA256` / `assetTreeSHA256` 与一个**独立于本 manifest** 的期望值一致（该期望值须来自 `feat/satori-renderer@4bda363` 并写死在测试或独立常量中），否则 #6 永远可绕。这是唯一能把「冻结」变成可执行的改动。
   - **【2026-09-29 修订】本项已判定为「不可实施」，原文保留不删。** 决定性理由是「量错了对象」：该树哈希的根是 `template/` 与 `assets/`，**不覆盖 `baseline/images/*.jpg`**，即便完美可复现也检测不到 #6。另有覆盖范围两树对不齐（`assets/` 102 vs 100）与值不可复现（473 commit 穷举 0 命中）两条。详见文末「修订（2026-09-29，lead-5 代 verifier 记录）」。
2. **【必改·中】为 0.99 阈值补负向测试**。新增一个断言「`gatePassed(0.98, 非空bbox)` 必须返回 false」的测试，锁死阈值。现有 `TestGGPixelParity_Negative` 测不到门禁，缺口真实。
   - **【2026-09-29 已实施】** `fix/gate-hardening@aa88012`，新增 `TestGGPixelParity_Negative_ThresholdGate`（调真实 `gatePassed`），变异实测 0.99→0.5 转红、restore 转绿。
3. **【建议·中】修正守卫的诊断文案与语义**。将 `empty bbox` 判定区分为「零差异（maxX<0）」与「单像素差异（跨度为 0）」两种情形（可让 `similarityNormalized` 返回 `bbox=[-1,-1,-1,-1]` 之类哨兵以区分），消除「no pixel actually compared」这句事实错误的描述，并避免对单像素差异的误杀。
   - **【2026-09-29 部分实施】** 哨兵 + 文案修正已在 `fix/gate-hardening@aa88012` 落地（零差异返回 `[-1 -1 -1 -1]`，两种情形分开报错，错误文案已删）。**「避免对单像素差异的误杀」未做** —— 修它需要把 `gatePassed` 的零跨度判定由 `<= 0` 放宽为 `< 0`，那是**放宽门禁**，与本轮「不得让任何场景分数上升/方向必须收紧或等价」的红线相冲，留待 Boss 裁决。
4. **【建议·中】考虑对 #7 加检测**。可选方案：记录并强制 `new` 与 `old` 的**结构差异下界**（例如要求差异像素数占画布比例 ≥ 某阈值，或对基线主色区域做覆盖检测），否则「基线作底 + 1.33% 扰动」将持续是可复现的提分路径。
5. **【建议·低】把 `pixelWidth/pixelHeight` 纳入判定**。当前仅用于错误消息；new 图尺寸是否等于 manifest 声明值未校验。

### 免责声明（须与签字同时阅读）

- 本审查为**静态审查**。`gatePassed` 的运行期行为、16 场景实际分数、commit `ec440bf` message 中「16 场景分数逐个字节一致 / calendar 0.99432 / state 0.99093 仍 PASS」的声明，**均未经我运行验证**。跑 harness 会写 `tmp/pixel-compare/` 与 `report.json`/`report.md`，超出本次只读授权。
- `TestGGPixelParity_Negative_EmptyBBoxGate` 的断言能力是**从源码语义推导**得出（「删掉 line 121 守卫则 line 398 断言失败」），非实测执行结果。
- 上述必改项**我一条都没有实施**。本次全程只读，未 edit / 未 commit / 未 push / 未改分支 / 未删 worktree。

---
*审查人：worker-1 ｜ 2026-09-29 ｜ 对象：ec440bf / src/ggrender/pixel_test.go*

---

## 修订（2026-09-29，lead-5 代 verifier 记录）

**本块的性质**：以上六节是 worker-1（verifier）于 2026-09-29 的**只读审查原文**，一字未删、一字未改，仍具签字凭据效力。本块是**追加**的修订记录，记录 Boss 授权下的后续裁定与实测证据。两条记录冲突时，以本块为准，并保留冲突痕迹。

**授权**：`arknights_bot-gate-fix` 工作树 / `fix/gate-hardening` 分支，Boss 派工。实施者：worker-17。

### 一、必改项 ①「验证 manifest 的树哈希」→ 判定**不可实施**

Boss 裁决：**不写** `assertTreeSHA256` 之类的校验函数。理由有**三条**（另加一条固定附加的残余风险），按强度排列，**第一条是决定性理由**。全部原始数据一个不少，只是重新排序。

**（a）决定性理由：量错了对象。** `templateTreeSHA256` / `assetTreeSHA256` 的根是 `<repoRoot>/template` 与 `<repoRoot>/assets`，**不覆盖 `baseline/images/*.jpg`**。而缺陷 #6 问的是基线图有没有被换。**即便这两个值完美可复现，钉住它们也检测不到 #6** —— 那不是弱一点的校验，那是校验了另一个东西。

这一条与可复现性**无关**，且不因任何复现结果而改变：它是在回答「就算算准了，量的是不是要防的那个东西」。

**（b）覆盖范围两树对不齐。** `assets/` **gg 树 102 个文件 vs satori 树 100 个文件**（`template/` 两树一致，各 20 个）。文件数本身就不对不上，无论排序方式或哈希内容都无从对齐。

**（c）值不可复现。** 算法本身**确实是**树哈希，有确切实现：由 commit `8873add`（`feat: replace browser screenshots with satori renderer`）引入的 `src/cmd/visual-regression/main.go`，其 `verifyManifest()` 读这两个字段，算法是同文件的 `treeSHA256()` —— `filepath.WalkDir` 递归收集目录下**全部**文件（不过滤）→ `sort.Strings` → 依次把 `相对路径\x00<该文件内容的 hex sha256>\n` 写进一个 sha256 → 取 hex。

穷举 `feat/satori-renderer` 全历史 **473 个 commit，0 命中**（`tOK=473 aOK=473`，`templateTree` 69 种互异值、`assetTree` 21 种、MATCH=0）。`8873add` 与其父 `8873add^` 同样对不上。逐树实测值：

| 树 | 实测 treeSHA256 | manifest 声称值 |
|---|---|---|
| satori `4bda363` 的 `template/` | `1fccd181f7b38565ad4a2eda36a96c031ac9bd47c8554b346b6fde5ff0680e7c` | `664b642f0f37debff0c3fbcd6f59d1e0703be2d450ab10cb2a06549a219f73d0` |
| satori `4bda363` 的 `assets/` | `67cd4b43d0f0e93050a5dd7dad5f72185c5b57ecfd3112bcd236c8662d66ee16` | `8f7fd0b1e2239d27116807f98e27fdee61076b8aa16bb0e5d75f743bd7cd1c14` |
| gg `0ea71c6` 的 `template/` | `1fccd181f7b38565ad4a2eda36a96c031ac9bd47c8554b346b6fde5ff0680e7c` | 同上 |
| gg `0ea71c6` 的 `assets/` | `a4be8643ddbe0f1f2c3ec329a60638ea9add2e7230c1329032eaf6853b1647da` | 同上 |

**正控制**：satori 与 gg 的 `template/` 哈希**完全相同**，证明算法复现无误 —— 否则「0 命中」会被误读成「算法写错了」。有了这个正控制，否定结果才站得住。

**固定附加（与上面三条并列，非「理由」）：**

> 残余风险：把 baseline 换成自制图 + 同步改 manifest 的 sha256，仍能通过全部检查。当前防线只有 git 跟踪 + 人工审查，**不是密码学锚点**。

**结论：该测量不可实施，不予实施。** 已在 `src/ggrender/pixel_test.go` 内留注释块「为什么这里没有 manifest 树哈希校验（Task 1 结论：不可实施）」，记录上述全部依据，避免后人误以为此处漏写。

### 二、必改项 ②「为 0.99 阈值补负向测试」→ 已实施

`fix/gate-hardening@aa88012`。新增 `TestGGPixelParity_Negative_ThresholdGate`，调**真实的 `gatePassed`** 而非表达式副本。变异检测实测：`sim < 0.99` 改成 `sim < 0.5` → 该测试 FAIL（exit=1）；从 pristine 逐字节复原（`cmp` exit=0）→ PASS（exit=0）。测试全程未为通过而改动 0.99 字面量。

### 三、必改项 ③ → 部分实施

哨兵与文案已在 `aa88012` 落地：零差异返回 `bbox=[-1 -1 -1 -1]`（`-1` 不可能是真实差异像素坐标，旧值 `[0 0 0 0]` 会与位于 `(0,0)` 的单像素差异撞值），`gatePassed` 分开报 `zero-diff` 与 `degenerate`，事实错误文案 `"no pixel actually compared"` 已删除 —— `similarityNormalized` 外层循环恒定扫满 `w*h`，两种情形下画布每个像素都被比较过。

**「避免对单像素差异的误杀」经 Boss 裁决：不修。** 实测数据点：单像素差异的 `sim=0.99992`，在 0.99 门禁下本应算近完美，却因 bbox 零跨度被拒。修它需要把零跨度判定由 `<= 0` 放宽为 `< 0` —— **那是放宽门禁，正是历史上被污染过三次的方向**。裁决认为该误杀在实践中代价极小（`sim=0.99992` 被拒），而放宽的代价是重开一个已被反复利用过的类别。**记入已知限制（见下节），「未修」是裁决结果，不是遗漏。**

### 四、已知限制清单（**未修复，不得淡化**）

**（1）缺陷 #6：信任锚自证**

> 缺陷 #6（信任锚自证）**未被修复**。把 baseline 换成自制图 + 同步改 manifest 的 sha256，仍然能通过全部检查。当前防线只有 manifest 处于 git 跟踪 + 人工审查。这是**已知且已记录的缺口**，不是「基本安全」。

`pixel_test.go` 的 hash gate 仍只校验「baseline 文件的 sha256 == 同一条 manifest 记录里的 sha256」，两者可同改。

**（2）缺陷 #7：基线作底 + 少量扰动**

以基线作背景再叠加少量绘制，使差异像素跨 ≥2 列且 ≥2 行、总 `sum` 压在预算内，bbox 非退化故守卫不触发。预算：`sim ≥ 0.99` 允许 `sum ≤ 0.01·w·h·4·255`，单像素最大贡献 765，**约 1.33% 画布可完全画错仍通过**。实测 `gatePassed(0.9995, 非空bbox)` 返回 true。**未处理。**

**（3）单像素差异被误杀 —— 经 Boss 裁决不修**

零差异与单像素差异现已可区分（哨兵 `[-1 -1 -1 -1]` vs `[x0 y0 x0 y0]`，见本块第三节），但**两者仍都被判失败**。后果：合法近完美渲染（实测单像素差异 `sim=0.99992`）被误判为门禁失败。**不修是裁决结果，不是遗漏** —— 修它需把零跨度判定 `<= 0` 放宽为 `< 0`，即放宽门禁，属于历史上被反复利用过的方向。

### 五、**我们有一个实测有效但因工程理由否决的方案**（重要，勿读成「无方案」）

> **「我们有一个实测有效但因工程理由否决的方案」和「我们没有方案」是两回事。** 前者让将来的人知道**路存在**，值不值得走由他们判断；后者会让人**重新发明一遍**。

在收到 Boss 裁定之前，worker-17 **已经实现并实测通过**了一个方案，原文保存在 `.audit/rejected-anchor-approach.patch`（已随 `fix/gate-hardening` 提交）：

**`rejected-anchor-approach.patch` 解决的是另一个问题 —— 16 张 baseline images 的独立 sha256 常量锚，与树哈希无关。** 它**不需要**树哈希可复现性，且实测能真正拦下缺陷 #6。

被否决的理由是「多一处需手工同步 + 不是更强的信任锚」，**不是**「它复现不了」。理由若写错，将来复活它的人会以为「只要解决复现性就能用上」—— 而它根本不需要复现性。方案要点（可从 patch 逐行读回）：

- 16 张 `baseline/images/*.jpg` 的 sha256 写成本文件内的独立常量表 `frozenBaselineSHA256`，**不读 manifest**；
- 加载基线时逐张比对，与 manifest 无关，故「换图 + 同步改 manifest 的 sha256」这条绕过会被它拦下；
- 已核对该常量表在 `feat/gg-render-card-atomic` / `fix/harness-bbox-zero-guard` / `feat/gg-render-lists` / `feat/gg-render-depot-state` 上通用（四支的 `baseline/images` git blob 摘要均为 `e724975f381d251b70583df29ed42fc537570e057c6ede87a7a3c7baeaa6c539`）；
- 其局限：常量可被同改者一并改掉，所以它提高的是「改动必须出现在 git diff 里」的可见性与成本，**不是不可伪造性**。

**复活它是 Boss 的权限，不是 worker 的。** 本轮按裁定回退，未合入。

### 六、更新后的签字结论

**有条件通过，4 条必改 + 1 条已判定不可实施并记录为已知限制。**

- 必改项 ① → **不可实施**（本块第一节），决定性理由是「量错了对象」；转为**已记录在案的已知限制**，不再是待办项。
- 必改项 ② → **已实施并变异验证**（`aa88012`）。
- 必改项 ③ → **部分实施**；剩余部分（单像素误杀）经裁决**不修**，已记入本块第四节。
- 必改项 ④（#7 检测）、⑤（`pixelWidth/pixelHeight` 纳入判定）→ **仍未实施**，维持原优先级。
- **本块第四节的三项已知限制均未修复。** 不要把「不可实施」读成「基本安全」。
- **有条件通过的依据不变**：`ec440bf` 对已列举四条历史作弊手法的防御经实测有效；`aa88012` 是方向收紧/等价的改动，16 场景分数逐个不变（改前改后测试输出落盘 `diff` exit=0）。

*修订人：worker-17（代 lead-5）｜ 2026-09-29 ｜ 对象：`fix/gate-hardening`（`aa88012` + 理由重排 commit）｜ 依据 commit：`8873add` / `4bda363`*
