# Task 3 — gg 线可从 yoga 线 / satori 线合法复用哪些渲染能力（只读调查）

- `provenance:` **measured**（本文所有「存在 / 不存在 / 干净 / 无来源」类结论均为本次实测，见取证附录；推断项逐条标注 `inferred`）
- `measured_at:` 分支 `feat/yoga-skia-go-renderer` @ `3704262`；对照分支 `feat/gg-render-card-atomic` @ `d939c38`、`feat/satori-renderer` @ `b946661`；测量时间 `2026-10-02T18:43:23+08:00`（本机 `date -Iseconds`）
- `original_path:` `C:/WorkSpace/Golang/.lead2-reports/task3-render-reuse.md`
- 本任务**全程只读**。唯一写入 = 本文件。禁碰树 `arknights_bot-satori-yoga-skia-go` 未被写入（证据见附录 A-0）。

---

## 0. 先纠正两条范围性事实（我自己的枚举曾出错，现已修正）

我第一次枚举工作树时**目测**数成 21 棵，并漏掉 `arknights_bot-skland`。lead-2 的更正是对的。现用机械计数重做：

| 量 | 值 | 证据 |
|---|---|---|
| `git worktree list` 总行数 | **22** | `total_lines=22`，附录 A-1 |
| 表头行数 | **0** | `header_lines=0`，探针自检见 A-1 |
| 含 `src/ggrender/` 的工作树 | **18** | `[ -d "$d/src/ggrender" ]` 逐棵判定，见 A-2 |
| 不含 `src/ggrender/` 的工作树 | **4**（satori 家族） | 同上 |

**satori 家族 4 棵没有 `src/ggrender`**：`arknights_bot-satori`、`-satori-depot`、`-satori-enemy`、`-satori-evidence-pr-3-playwright-vs-satori-current`。即 **yoga 线有 `src/ggrender`，satori 线没有** —— 「yoga 线 vs satori 线」不是同一种东西的两种写法。

> 因此本文下文凡写「穷举」，其枚举范围 = 上表 22 棵工作树的 `[ -d .../src/ggrender ]` 判定结果，**外加** §2 里逐条列出的 git 对象级查询（按 blob sha 枚举，不依赖目录）。凡本文没给枚举范围的句子，不构成全称结论。

---

## 1. 语言栈（先确认，再谈「可复用」是移植还是直接引用）

| 线 | 渲染代码位置 | 语言 / 渲染后端 | 证据 |
|---|---|---|---|
| gg 线 | `src/ggrender/*.go` | **Go + `github.com/fogleman/gg`** | `src/go.mod` 含 `fogleman/gg v1.3.0` |
| yoga 线（禁碰） | `src/ggrender/*.go` **+** `src/skia/*.go` | **Go + `gg`**（ggrender）/ **Go + 自写 skia façade**（`src/skia`） | `wc -l src/skia/*.go` = 3040（附录 A-3） |
| satori 线 | `renderer/**/*.mjs` | **JavaScript (ESM) + satori/sharp** | `renderer/` 下 30 个 `.mjs`，`find renderer -type f \| sed 's/.*\.//' \| sort \| uniq -c` → `30 mjs / 2 json / 1 py`（排除 node_modules） |

**结论（measured）**：

1. **yoga 线的 `src/ggrender` 与 gg 线是同一个包名、同一条渲染技术栈（Go + gg）**。「抄 yoga 线」对这一半**是直接复制 Go 源码**，不是移植。
2. **satori 线是 JS**。它对 gg 线**只能作为规格 / 算法参考**，不能作为代码来源。
3. yoga 线内部有两套东西：`src/ggrender`（gg，与 gg 线同源）和 `src/skia`（另一套，见 §3）。

---

## 2. 候选清单表

### 2.1 「yoga 独有的 ggrender 场景文件」

对照口径：`git ls-tree -r --name-only <branch> -- src/ggrender | grep '^scene_.*\.go$' | wc -l`

| 分支 | 工作树 | 场景数 |
|---|---|---|
| `feat/gg-render-card-atomic` | `arknights_bot-gg-card-atomic` | 9 |
| `feat/gg-render-lists` | `arknights_bot-gg-lists` | 12 |
| `feat/gg-render-depot-state` | `arknights_bot-gg-depot-state` | 10 |
| `feat/gg-render-card` | `arknights_bot-gg-card-revival` | 10 |
| (detached `0d77efe`) | `arknights_bot-reviewer` | 13 |
| `feat/yoga-skia-go-renderer`（禁碰） | `arknights_bot-satori-yoga-skia-go` | **16** |

gg 侧缺的 7 个：`box` / `box_detail` / `box_summary` / `missing` / `enemy` / `headhunt` / `recruit`。

**关键取证 —— 这些场景文件的 blob 身份（measured，附录 A-4）**

`git rev-parse <rev>:src/ggrender/scene_X.go`，blob sha 完全相同 = 同一份字节：

| 场景 | yoga `3704262` | `0d77efe` (reviewer) | `feat/gg-render-lists` | `feat/gg-render-depot-state` / `-card` | `feat/gg-render-card-atomic` / `consolidate/gg-mainline` |
|---|---|---|---|---|---|
| `scene_box.go` | `eabf8926` | **`eabf8926` 同** | 缺 | 缺 | 缺 |
| `scene_box_detail.go` | `c6f450da` | **`c6f450da` 同** | 缺 | 缺 | 缺 |
| `scene_box_summary.go` | `deab5b76` | **`deab5b76` 同** | 缺 | 缺 | 缺 |
| `scene_missing.go` | `4a21ed76` | **`4a21ed76` 同** | 缺 | 缺 | 缺 |
| `scene_enemy.go` | `035cd2e4` | 缺 | `6857b5f3` **不同** | **`035cd2e4` 同** | 缺 |
| `scene_headhunt.go` | `5e42acd8` | 缺 | `05c6dd7f` **不同** | 缺 | 缺 |
| `scene_recruit.go` | `8f12808b` | 缺 | `aa367047` **不同** | 缺 | 缺 |

两条判读：

- **box 四件套在 `0d77efe`（`arknights_bot-reviewer`）里是 yoga 的逐字节副本，不是独立实现。** 且 `0d77efe` 的祖先判定：`8cbffef` exit=**0**（含作弊提交），`641988f` exit=**1**（不含）。→ **它不能当「干净来源」。**
- **`enemy`/`headhunt`/`recruit` 在 `feat/gg-render-lists` 里是独立实现**：其新增提交 `75cb1a2` / `6fc0843` / `200c414` / `12955b7` 对 `feat/yoga-skia-go-renderer` 的祖先判定全部 exit=**1**（不是 yoga 的祖先，也不在 yoga 线上）；对 `feat/gg-render-lists` exit=**0**。→ **gg 线自己写的，不是抄 yoga。** 但 `feat/gg-render-lists` **含 `8cbffef`**（exit=0）。

### 2.2 候选表

> 「干净来源是否存在」一栏：判据 = 该能力在**不含 `8cbffef` 且不含 `641988f`** 的分支上是否有（功能等价的）实现。**名字不同也算存在** —— 见 §2.3 关于「找不到名字≠不存在」的处理。

| # | 能力 | 语言栈 | 干净来源是否存在 | (a) 基线图作渲染输入 | (b) 硬编码 / 跨树路径 | 出处（函数名 + 工作树 + 分支 + commit） | 追过的变量取值链 |
|---|---|---|---|---|---|---|---|
| C1 | `box` / `box_detail` / `box_summary` / `missing` 四个场景 | Go（gg） | **否**。唯一非禁碰来源 `0d77efe` 是 yoga 的逐字节副本且含 `8cbffef` | **否** | **否** | `RenderBox`/`RenderBoxDetail`/`RenderBoxSummary`/`RenderMissing`；工作树 `arknights_bot-satori-yoga-skia-go`；分支 `feat/yoga-skia-go-renderer`；commit `3704262` | `RenderBox(d *BoxInfo)` → `BuildBox(info)` → `NewLoader(findRepoRoot())` → `ld.root` → `LoadImage(filepath.Join(ld.root, rel))`。**注：`src/ggrender` 这条链里没有 skia 的 `cacheRoot` 分支**（详见 §4） |
| C2 | `enemy` 场景 | Go（gg） | **部分**。`feat/gg-render-lists` 有独立实现，但该分支含 `8cbffef`（exit=0）；`feat/gg-render-depot-state` / `-card` 的是 yoga 逐字节副本（blob `035cd2e4`） | **否** | **否** | 同上，`src/ggrender/scene_enemy.go::RenderEnemy` / `drawStringBoldW`；工作树/分支/commit 同 C1 | `RenderEnemy` → `tryLocal(c.Avatar)` → `LoadImage(filepath.Join(AssetRoot, rel))`；`AssetRoot` ← `ggRepoRoot()` ← `runtime.Caller(0)` → **本工作树根**。无 `C:/WorkSpace` |
| C3 | `headhunt` 场景 | Go（gg） | **部分**，同 C2（`feat/gg-render-lists` 独立，含 `8cbffef`） | **否** | **否** | `src/ggrender/scene_headhunt.go`；工作树/分支/commit 同 C1 | 同 C2 |
| C4 | `recruit` 场景 | Go（gg） | **部分**，同 C2 | **否** | **否** | `src/ggrender/scene_recruit.go::RenderRecruit`；工作树/分支/commit 同 C1 | 同 C2 |
| C5 | 双线性 / cover 缩放族（`scaleSmooth`、`scaleSmoothCR`、`smoothCover`、`drawImageReal`） | Go（gg） | **是**。gg 线已有等价物，只是改了名：`ScaleContain` / `ScaleCover` / `ScaleExact`（三者内部都调 `draw.BiLinear.Scale`）。实测：card-atomic / consolidate-gg-mainline / gg-lists / skland 各自的 `src/ggrender/helpers.go` 各有 3 处 `BiLinear`；仅 `feat/gg-render` 为 0 | **否** | **否** | yoga 侧 `src/ggrender/scene_base.go::scaleSmooth` / `scene_operator.go::scaleSmoothCR` / `scene_box.go::smoothCover`；gg 侧 `src/ggrender/helpers.go::ScaleExact` 等；工作树 `arknights_bot-gg-card-atomic`；分支 `feat/gg-render-card-atomic`；commit `d939c38` | `ScaleExact(img,w,h)` → `draw.BiLinear.Scale(dst,…,nil)`，纯像素运算，无路径变量 |
| C6 | 文本度量族（`measureString`、`drawStringBold`、`drawStringBoldW`、`drawStringMixed`、`drawStringMixedBold`） | Go（gg） | **部分**。gg 侧有 `measure` / `measureW` / `drawString` / `drawStringAnchored`（`helpers.go` 函数表实测）；yoga 独有的是 **bold / mixed（半粗混排）变体**，gg 侧未见 | **否** | **否** | yoga：`scene_base.go::measureString` / `drawStringBold`、`scene_enemy.go::drawStringBoldW`、`scene_gacha.go::drawStringMixed{,Bold}`；gg：`src/ggrender/helpers.go::measure`（工作树 `arknights_bot-gg-card-atomic` / `feat/gg-render-card-atomic` / `d939c38`） | `measure(dc, s)` → `dc.MeasureString(s)`（gg 原生）；yoga `measureString` 同源。**粗体合成**一路（`LoadDefaultFont` → `FontCandidates`）gg 侧亦存在 |
| C7 | card 合成族（`cardBold`、`cardFadeBottom`、`cardPanelLayer`、`cardTintedByMask`、`cardWithOpacity`、`cardAvatarURL`） | Go（gg） | **部分**。gg 侧有 `cardWithOpacity` 之外的等价物：`progressBar`/`ProgressBar`、`DrawPortraitTile`、`card-gate.json` 等；但 **mask 染色 / 底部渐隐 / 面板分层** 三个函数名在 5 条干净分支上均未命中（枚举范围见 §2.4） | **否** | **否** | yoga `src/ggrender/scene_card.go`（工作树 `arknights_bot-satori-yoga-skia-go` / `feat/yoga-skia-go-renderer` / `3704262`） | `cardTintedByMask` 等为纯 `gg.Context` 绘制操作，函数体内不触及 `AssetRoot` / `ld.root` |
| C8 | base 家具族（`drawBaseChar`、`moodColor`、`drawMoodIcon`、`drawAPBar`、`drawNoteIcon`） | Go（gg） | **部分**。`RenderBase`/`SampleBase` gg 侧存在（9 场景之一），但这 6 个绘制原语名在干净分支未命中 | **否** | **否** | yoga `src/ggrender/scene_base.go`；工作树/分支/commit 同 C7 | `drawBaseChar` → `tryLocal(c.Avatar)` → `LoadImage(filepath.Join(AssetRoot, rel))`；`AssetRoot` ← `ggRepoRoot()` ← `runtime.Caller(0)` → **本工作树根**；实测 `grep -rn 'WorkSpace' src/ggrender/*.go` 无命中（附录 A-6） |
| C9 | 图标 / 元件原语（`drawBoxFamilyHeader`、`drawBoxTile`、`renderDetailIcon`、`drawStar`、`drawPersonIcon`、`drawHelpChips`、`fillSector`、`monthGrid`） | Go（gg） | **部分**。`drawPersonIcon` / `drawHelpChips` / `fillSector` 在 `feat/gg-render-lists` 有（但该分支含 `8cbffef`）；其余 5 个在干净分支未命中 | **否** | **否** | yoga：`scene_box.go::drawBoxFamilyHeader/drawBoxTile`、`scene_box_detail.go::renderDetailIcon`、`scene_operator.go::drawStar`、`scene_help.go::drawPersonIcon/drawHelpChips`、`scene_gacha.go::fillSector`、`scene_calendar.go::monthGrid`；gg-lists 侧：`src/ggrender/scene_help.go::drawPersonIcon`、`src/ggrender/scene_gacha.go::fillSector` | 纯绘制；`drawBoxTile` 内部走 `tryLocal`/`AssetPath`（同 C8 链） |
| C10 | depot 图标取回缓存（`fetchCached`） | Go（gg） | **部分**。gg 侧 `helpers.go` 有 `FetchImage`（无缓存），无 URL→image 的 LRU 缓存 | **否** | **否** | yoga `src/ggrender/scene_depot.go::fetchCached`（`depotIconCache map[string]image.Image` + `depotIconMu`）；工作树/分支/commit 同 C7 | `fetchCached(url)` → `FetchImage(url, fallback)` → ①`fetch(url)` 走 `http.Client{Timeout:6s}` 远程 ②失败则 `LoadImage(AssetPath(rel))`；`fallback` ← `AssetPath("common/amiya.png")` 或 `AssetPath("depot/lmd-full.png")` ← `AssetRoot` ← `ggRepoRoot()` ← `runtime.Caller(0)` |
| C11 | `src/skia` 整包（backend / canvas / font / text / yoga / image / renderer / depot） | Go（自写 façade） | **否**。`src/skia/renderer.go` 只存在于 6 条分支，**与含 `8cbffef` 的分支集合完全重合**（附录 A-7） | **是 —— 且是硬违规**，见 §3 与 §4 | **是 —— 硬编码 `C:/WorkSpace`**，见 §3 | `findRepoRoot` / `Loader.loadUncached` / `Loader.NewLoader`；工作树 `arknights_bot-satori-yoga-skia-go`；分支 `feat/yoga-skia-go-renderer`；commit `3704262` | 见 §4 完整链 |
| C12 | satori 线 16 个场景组件 + 3 个 lib | **JS（.mjs）** | **不适用**（跨语言）。可作**规格参考**，不可作代码来源 | **否**（生产路径）。但 `renderer/lib/assets.mjs` 内含 `FROZEN_FIXTURE_CACHE_ALIASES`，形态与 skia 的 manifest→cache 相同 —— 见 §4.3 | **否**。`runner.mjs:18` `REPO_ROOT = path.resolve(fileURLToPath(new URL('..', import.meta.url)))`，无绝对路径 | `renderer/runner.mjs`、`renderer/components/*.mjs`、`renderer/lib/assets.mjs::createAssetLoader`；工作树 `arknights_bot-satori`；分支 `feat/satori-renderer`；commit `b946661` | `createAssetLoader({repoRoot: REPO_ROOT})` → `root = path.resolve(options.repoRoot ?? process.cwd())` → `manifestSource = options.manifestPath ?? process.env.SATORI_ASSET_MANIFEST ?? null` → **`runner.mjs` 未传 `manifestPath`，故 `manifestPromise` 解析为 `null`，frozen-cache 别名分支在生产渲染中不启用** |

### 2.3 「干净来源是否存在」的计数

| 口径 | 行数 |
|---|---|
| 候选表总行数 | **12**（C1–C12） |
| 其中「干净来源存在」 | **0 行**（C1 无；C2/C3/C4/C10 干净分支无独立实现，`feat/gg-render-lists` 有但含 `8cbffef`；C5/C6/C7/C8/C9 部分存在，缺口见备注；C11 无；C12 跨语言不适用） |
| 其中「干净分支上已存在等价物、无需从 yoga 取」 | **1 行明确成立（C5）**，另 C6/C7/C8/C9 为「部分存在，缺口需自行实现或另找来源」 |

### 2.4 「未命中」类结论的枚举范围（举证栏）

我写的任何「在干净分支上未命中」都限定在下面这个枚举内，**不是全称结论**：

- 分支集合（5 条，均 `8cbffef` exit=1 且 `641988f` exit=1 或已 revert）：
  `consolidate/gg-mainline`、`feat/gg-render-card-atomic`、`feat/gg-render`、`feat/skland-integration`、`feat/gg-render-base`
- 文件 glob：`<branch>:src/ggrender/**/*.go`（由 `git grep -hoE '^func …'` 覆盖）
- 逐分支符号表大小（`wc -l`，measured）：105 / 101 / 69 / 109 / 103
- yoga 独有符号集（34 个，`comm -23` 结果）与上述 5 表的交集：**5/5 均为 0**

**我不写「任何干净分支都没有 X」** —— 上面的枚举只覆盖 5 条分支与 `src/ggrender` 一个包。`src/utils/media`、`template/` 等目录未被这次符号枚举覆盖。

---

## 3. 禁碰树独有清单（**不建议直接取**）

| 能力 | 出处 | 标注 | 一句话理由 |
|---|---|---|---|
| `src/skia` 整包（11 文件 / `wc -l` = 3040 行，含 `skia_test.go` 302 行） | `arknights_bot-satori-yoga-skia-go` / `feat/yoga-skia-go-renderer` / `3704262` | **不建议直接取** | 它**读冻结基线缓存目录当渲染输入**（§4.1），且把 `C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go` 硬编码成默认值 |
| `findRepoRoot()` 的多候选回退 + `SKIA_REPO_ROOT` 环境变量 | 同上，`src/skia/renderer.go` | **不建议直接取** | 该函数**本身就是**那条组合式违规的一半；抄它等于抄违规 |
| Yoga flex 布局引擎（`YogaNode` / `Style` / `FlexDirection` / `AddChild`） | 同上，`src/skia/yoga.go`（313 行） | **不建议直接取** | 它与 `src/skia` 同包、同一次提交（`8cbffef`）引入，无法只取布局不取资产加载；而 gg 线用 `gg.Context` 绝对定位，不需要布局引擎 |
| 真 Skia cgo 后端（`backend_skia.go`，`-tags skia`） | 同上 | **不建议直接取** | 文件内自述为「placeholder-free build」的注释态 cgo 路径（`backend_skia.go:19,23`），本机无 skia SDK，不构成可运行能力 |
| satori 线 16 个场景组件的**具体几何数值** | `arknights_bot-satori` / `feat/satori-renderer` / `b946661`，`renderer/components/*.mjs` | **可读，但数值不得当验收目标** | 组件内注释自陈几何是「measured off the frozen Playwright baseline」（例：`box-detail.mjs:5`、`box-summary.mjs:16`、`depot.mjs:4`）。**红线：不得从冻结基线反推取值。** 这些数值只能当「同源模板的独立测量记录」，且属 satori 线（JS），对 gg 线（Go+gg）无代码复用价值 |

---

## 4. 组合式违规排查（追数据流，不 grep 单个字符串）

### 4.1 `src/skia` 的资产加载链 —— 确认违规，并给出完整取值链

**违规形态与已知样本一致**，我把每一跳都钉死：

```
src/skia/renderer.go:115,227,385,511,634,861
    ld, _ := NewLoader(findRepoRoot())
      │  ↑ 参数值
      ↓
src/skia/renderer.go:66-78  func findRepoRoot() string
    cands := []string{"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go", "."}
    if r := os.Getenv("SKIA_REPO_ROOT"); r != "" { cands = prepend(r) }
    … 逐个 os.Stat(c+"/assets/font/NotoSansHans-Regular.ttf") …
    return "C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go"   // ← 末尾兜底
      │  ↑ 返回值
      ↓
src/skia/image.go:90-93  func NewLoader(repoRoot string) → ld.root = repoRoot
      │  ↑ 结构体字段
      ↓
src/skia/image.go:241  cacheRoot := filepath.Join(ld.root, "src/utils/media/testdata/visual/baseline")
      │  ↑ 拼接
      ↓
src/skia/image.go:326-347  os.ReadFile(abs) → validateMagic → sha256 比对 entry.SHA256 → maybeNormalizeWebP
```

**判定**：
- 含 `C:/WorkSpace` 的行（`renderer.go:66,78`）**不含** `baseline`；含 `baseline` 的行（`image.go:241`）**不含** `C:/WorkSpace`。→ 正是「两处单独看都干净，合起来才是违规」的形态，逐行审读与单串 grep 都看不到。
- 该缓存目录是**冻结 Playwright 基线的资源缓存**（对照：`arknights_bot-satori` 树内同路径下确有 `card-secretary-1024.png` / `depot-lmd.png` 等实体文件，附录 A-8）。→ **把基线资源当渲染输入，命中硬红线。**
- 与已知样本的一处**更正**：AGENTS.md 记 `ld.root` 由 `renderer.go:66,78` 硬编码；本次实测确认行号仍为 66/78，但形态是**函数 `findRepoRoot()` 的候选列表首项 + 末尾兜底**，不是裸常量。结论不变。

**可达性（measured，附录 A-9）**：`src/skia` 在 yoga 树内被 import 的文件**只有一个** —— `src/utils/media/yoga_skia_test.go`（`_test.go`）。
- 枚举范围：`grep -rln '"arknights_bot/skia"' --include='*.go' src/` 于 `arknights_bot-satori-yoga-skia-go` @ `3704262`，exit=0，输出 1 行。
- 交叉验证：`grep -rnoE '\bskia\.[A-Za-z0-9_]+' --include='*.go' src/` 的 28 处命中**全部**在该 `_test.go` 内。
- → 与已知结论一致：**不进生产二进制，但会在 `go test` 下被编译并执行**。
- **未做的事**：我没有运行 `go test`（那会在禁碰树产生 `tmp/` 等产物）。

### 4.2 `src/ggrender` 的资产加载链 —— **未**发现违规（与 skia 相反）

```
src/ggrender/helpers.go:31-40  func ggRepoRoot() string
    runtime.Caller(0) → thisFile = <本文件绝对路径> → filepath.Dir ×3 → 本工作树根
    兜底：filepath.Abs("../..")；再兜底 "."（相对路径）
      ↓
src/ggrender/helpers.go:44-50  AssetRoot = filepath.Join(ggRepoRoot(), "assets")；不存在则 filepath.Join("..","..","assets")
      ↓
src/ggrender/helpers.go:258    func AssetPath(rel) = filepath.Join(AssetRoot, rel)
src/ggrender/helpers.go:110-117 func tryLocal(rel) → LoadImage(filepath.Join(AssetRoot, rel))
src/ggrender/scene_depot.go:44 func fetchCached(url) → FetchImage(url, AssetPath("common/amiya.png") | AssetPath("depot/lmd-full.png"))
```

**判定：(a) 否、(b) 否。**
- (b) 依据：取值来源是 `runtime.Caller(0)`（编译器注入的本文件路径），**没有任何字面量路径**。`grep -rn 'WorkSpace' src/ggrender/*.go` 实测 exit=1、输出为空（附录 A-6，探针自检见同处）。
- (a) 依据：`src/ggrender` 的非测试文件中，`baseline` 一词的 15 处命中**全是注释或局部变量名**，无一处是文件路径 —— 例如 `scene_state.go:225` 的 `baseline := labelTop + 21` 是文字基线（typographic baseline），不是图片基线。这是「同一个词两种含义」的陷阱，我逐条看过 15 行（附录 A-6）。
- 唯一读冻结基线的是 `src/ggrender/pixel_test.go`（`_test.go`），且它读的是 `manifest.json` 指定的 `old baseline (JPEG)` 去做**比对**（`pixel_test.go:202-214`），不是渲染输入 —— 这是门禁 harness 的正当职责。

**这一节的实践含义**：`src/ggrender`（16 场景）与 `src/skia`（11 文件）是**两个独立包**，资产加载路径完全不同。**污染只在 `src/skia` 里**。但 —— 按 AGENTS.md 的规则 —— 「只扫一个包就下全称结论」是方法缺陷；我的结论因此**限定在 `src/ggrender` 与 `src/skia` 两个包**，不外推到 `src/utils/media` 等其它目录。

### 4.3 satori 线 `assets.mjs` 的 frozen-cache 别名 —— **存在同形态设计，但生产路径未启用**

```
renderer/lib/assets.mjs:37   const FROZEN_FIXTURE_CACHE_ALIASES = { 'https://fixture-cache.invalid/card/secretary-painting.png': 'card-secretary-1024.png', … }
renderer/lib/assets.mjs:390  const root = path.resolve(options.repoRoot ?? process.cwd())
renderer/lib/assets.mjs:393  const manifestSource = options.manifestPath ?? process.env.SATORI_ASSET_MANIFEST ?? null
renderer/lib/assets.mjs:394  const manifestPromise = manifestSource == null ? Promise.resolve(null) : loadManifest(manifestSource, root)
renderer/lib/assets.mjs:315  const cacheRoot = path.dirname(manifestPath)        // ← 仅在 loadManifest 内可达
renderer/lib/assets.mjs:326-336  isInside(cacheRoot, …) && isInside(root, …) 双重越界守卫 → readFile
renderer/lib/assets.mjs:372-375  for (alias, cacheName) of FROZEN_FIXTURE_CACHE_ALIASES → aliases.set(alias, {…, provenance:'frozen-manifest-fixture-alias'})
```

**判定**：
- 与 skia 的**形态相同**（manifest/fixture 别名 → 冻结缓存目录 → `readFile` 当素材）。
- **但**：`renderer/runner.mjs:28-30` 的唯一生产调用是 `createAssetLoader({ repoRoot: REPO_ROOT })`，**未传 `manifestPath`** → `manifestPromise = Promise.resolve(null)` → `loadManifest` 与 frozen-cache 分支**在生产渲染中不可达**。`FROZEN_FIXTURE_CACHE_ALIASES` 仅被 `assets.mjs` 自身与 `renderer/manifest-webp.test.mjs` 引用。
- 且 satori 版有 skia 版**没有**的守卫：`isInside(cacheRoot, ·) && isInside(root, ·)` 双重校验（`assets.mjs:327,336`）；skia 的 `image.go:241` 那个 join **没有任何守卫**。
- (b) 判定为**否**：`REPO_ROOT` 来自 `new URL('..', import.meta.url)`（`runner.mjs:18`），无绝对路径字面量。

> 这条要写清楚，避免下一个人误以为「satori 也作弊」：**它没作弊，但它的设计里放着一模一样的机制，只是被 manifest 开关关着。** 谁要是把 `SATORI_ASSET_MANIFEST` 环境变量设上，机制就会打开。

---

## 5. 作弊提交排除结果

### 5.1 现查输出（快照，**分支列表会腐坏，勿引用条数**）

受影响集合 = 现查命令的输出，见取证附录 A-10 / A-11：

- `git branch -a --contains 8cbffef` → exit=0
- `git branch -a --contains 641988f` → exit=0

**本报告的可复用判据一律用现查命令，不在正文枚举受影响分支的条数或名单。** 上文引用具体分支名处，均已附该分支自己的 `--is-ancestor` exit code。

### 5.2 逐分支祖先判定（`git merge-base --is-ancestor <sha> <branch>`；**exit 0 = 是祖先，exit 1 = 不是祖先**）

| 分支 | `8cbffef` | `641988f` |
|---|---|---|
| `feat/gg-render` | 1 | 0 |
| `feat/gg-render-base` | 1 | 0 |
| `feat/gg-render-base-layout` | 1 | 0 |
| `feat/gg-render-card-atomic` | 1 | 0 |
| `feat/gg-render-card-layout` | 1 | 0 |
| `feat/gg-render-depot` | 1 | 0 |
| `feat/gg-render-headhunt-layout` | 1 | 0 |
| `feat/gg-render-recruit` | 1 | 0 |
| `feat/gg-render-recruit-headhunt` | 1 | 0 |
| `consolidate/gg-mainline` | 1 | 0 |
| `fix/operatorinfo-fields` | 1 | 0 |
| `feat/skland-integration` | 1 | 0 |
| **`feat/gg-render-canvas`** | **0** | 1 |
| **`feat/gg-render-card`** | **0** | 1 |
| **`feat/gg-render-depot-state`** | **0** | 1 |
| **`feat/gg-render-lists`** | **0** | 1 |
| **`feat/yoga-skia-go-renderer`**（禁碰） | **0** | 1 |
| `feat/satori-renderer` | 1 | 1 |
| `exp-satori-depot-next` | 1 | 1 |
| `exp-satori-enemy-next` | 1 | 1 |
| `evidence/pr-3-playwright-vs-satori-current` | 1 | 1 |
| `main` | 1 | 1 |
| `0d77efe`（detached，`arknights_bot-reviewer`） | **0** | 1 |

**要点（三条，都是「否定结论必须带定位」）**：

1. **`641988f`（card 0.99939 假达标）在 `feat/gg-render-card-atomic` 上 exit=0（是祖先）** —— 与「某分支不含某作弊提交」的说法方向相反。该分支的 mainline 侧已有 revert，但**祖先关系不等于内容存在**，引用时必须同时给分支与 commit。
2. **同一对提交在两条 gg 线分支上结论相反**：`feat/gg-render-card`（exit=1/0）与 `feat/gg-render-card-atomic`（exit=1/0 顺序相反，见上表）。这印证了「某作弊提交在某分支上」是**逐分支属性**。
3. **`feat/gg-render-lists` 含 `8cbffef`（exit=0）** —— 所以它虽然是 §2.1 里 enemy/headhunt/recruit 的**独立**实现来源，仍**不能**直接当「干净来源」使用。这是本次调查最容易被忽略、也最容易被下一个人踩的一条。

---

## 6. 取证附录

> 凡「判无命中」的条目，均给出 **exit code + 空输出**；禁用 `|| echo none`（它把「查询失败」与「结果为空」压成同一信号）。

### A-0 禁碰树未被写入
```
cd /c/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go && git status --porcelain
```
输出 3 行，全部是 **untracked（`??`）**、mtime 为 `2026-08-25`：`?? .github/workflows/visual-yoga-skia.yml`、`?? satori-yoga-skia-go/`、`?? tmp/`。**本次调查期间（2026-10-02）无任何新条目。**
对照组：`arknights_bot-gg-card-atomic` 与 `arknights_bot-satori` 的 `git status --porcelain` 均为 0 行。
自检：同一命令在 `arknights_bot` 裸库上返回 2 行（`?? fallback`、`?? tmp/`），exit=0 → 探针本身能出信号，不是恒零。

### A-1 工作树计数 + 表头探针自检
```
git worktree list > /tmp/wt.txt
total_lines=$(wc -l < /tmp/wt.txt)          # 22
path_lines=$(grep -c 'WorkSpace' …)          # 22
header_lines=$(grep -c '^worktree ' …)       # 0
```
**自检（已知答案）**：造 `worktree /fake/a` + 2 行路径的合成输入 → `fake_total=3 fake_header=1 fake_path=0`。即 `^worktree ` 探针能识出已知表头，故真实输出 `header_lines=0` 是可信读数而非尺子失灵。

### A-2 `src/ggrender` 逐棵判定
```
for p in $(awk '{print $1}' /tmp/wt.txt); do d=$(basename "$p"); [ -d "$d/src/ggrender" ] && … ; done
→ has_ggrender=18  none=4  total=22
```
**探针自检**：初次运行我用 `dirname` 取目录名，得到「22 行全部 = `C:/WorkSpace/Golang`」的荒谬输出 —— 那是 `dirname` 把完整路径截成父目录所致，属**尺子失效形态 (a)（输出异常）**。换 `basename` 后先做已知答案自检：`basename "C:/WorkSpace/Golang/arknights_bot-skland"` → `arknights_bot-skland`（`SELFCHECK_OK`），再跑真实对象。

### A-3 skia 包规模
```
wc -l src/skia/*.go   # 3040（含 skia_test.go 302）
```
逐文件：backend 58 / backend_skia 78 / canvas 417 / depot 107 / doc 14 / font 178 / image 507 / renderer 947 / skia_test 302 / text 119 / yoga 313。校验式 `58+78+417+107+14+178+507+947+302+119+313 = 3040` ✓

### A-4 blob 身份矩阵
```
MSYS_NO_PATHCONV=1 git rev-parse -q --verify "<rev>:src/ggrender/scene_X.go"
```
`MSYS_NO_PATHCONV=1` 是必须的：不加时 MSYS 会把 `:` 转成 `;`，**stderr 印 `fatal:` 而退出码返回 0**。
输出见 §2.1 表。

### A-5 yoga-only 符号集（`comm`）
```
git grep -hoE '^func (\([^)]*\) )?[A-Za-z0-9_]+' <branch> -- src/ggrender | sed 's/^func //' | sort -u
```
`yoga=119`、`card-atomic=101`、`lists=83`、`mainline=105` 符号；`comm -23 yoga card-atomic` = 34 个（清单见 §2.2 C7–C10 与附录 A-4 之外的逐符号-文件映射）。
**符号-文件映射**（`git grep -l "func <name>"`）已逐条落盘：`cardBold`/`cardFadeBottom`/`cardPanelLayer`/`cardTintedByMask`/`cardWithOpacity`/`cardAvatarURL` → `scene_card.go`；`measureString`/`drawStringBold`/`drawImageReal`/`scaleSmooth`/`drawBaseChar`/`moodColor`/`drawMoodIcon`/`drawAPBar`/`drawNoteIcon` → `scene_base.go`；`drawStringBoldW` → `scene_enemy.go`；`drawStringMixed{,Bold}`/`fillSector`/`sin`/`cos` → `scene_gacha.go`；`drawBoxFamilyHeader`/`drawBoxTile`/`smoothCover` → `scene_box.go`；`renderDetailIcon` → `scene_box_detail.go`；`ivals` → `scene_box_summary.go`；`drawStar`/`scaleSmoothCR` → `scene_operator.go`；`drawPersonIcon`/`drawHelpChips` → `scene_help.go`；`fetchCached` → `scene_depot.go`；`monthGrid` → `scene_calendar.go`。

### A-6 `src/ggrender` 的路径类扫描（判无命中）
```
cd arknights_bot-satori-yoga-skia-go
grep -rn 'WorkSpace' src/ggrender/*.go            → exit=1, 输出为空
```
**探针自检（已知阳性）**：`grep -rn 'skia' src/ggrender/` → exit=1 空输出；同一探针在 `grep -rln 'skia' src/skia/` 上 → exit=0 且列出 3 个文件。→ 探针能出信号。
`baseline|WorkSpace` 逐文件扫描（非测试 .go）：`hit_count=15`，全部为注释或局部变量名，逐行内容见 §4.2。
另：`grep -rn 'skia' src/ggrender/` → **exit=1，输出为空** → **`src/ggrender` 不 import `src/skia`**（这是「两包独立」的判定依据）。

### A-7 `src/skia` 的分支分布（穷举）
```
for b in $(git branch -a --format='%(refname:short)' | grep -v '^origin$'); do
  git cat-file -e "$b:src/skia/renderer.go" 2>/dev/null && echo "HAS src/skia: $b"; done
```
枚举范围 = 全部 67 条 `git branch -a --format='%(refname:short)'` 输出（本地 + remote，去掉 `origin` / `upstream` 两个 remote 别名行）。命中集合与 §5.2 中 `8cbffef` exit=0 的分支集合**完全重合**。

### A-8 冻结基线缓存目录确有实体文件（佐证 `cacheRoot` 指向的确实是基线资源）
`ls arknights_bot-satori/src/utils/media/testdata/visual/baseline/cache/` → 含 `card-secretary-1024.png`、`depot-lmd.png`、`gacha-avatar-amiya.webp` 等（工作树 `arknights_bot-satori` / `feat/satori-renderer` / `b946661`）。**只列了文件名，没有打开任何像素。**

### A-9 `src/skia` 可达性
```
grep -rln '"arknights_bot/skia"' --include='*.go' src/     → exit=0, 1 行: src/utils/media/yoga_skia_test.go
grep -rnoE '\bskia\.[A-Za-z0-9_]+' --include='*.go' src/  → exit=0, 28 处, 全部落在上一行那个文件里
```
枚举范围：工作树 `arknights_bot-satori-yoga-skia-go` @ `3704262`，`--include='*.go'`，glob `src/`（含全部子目录）。

### A-10 / A-11 作弊提交现查
```
git branch -a --contains 8cbffef     → exit=0
git branch -a --contains 641988f    → exit=0
git log --oneline -1 8cbffef         → "feat(yoga): P1 depot Yoga+ freetype 0.9xxx→0.99"
git show --stat --format='%h %s' 8cbffef --name-only | grep -c 'ggrender'  → 0  (exit=1 = 零命中)
```
`8cbffef` 的 diffstat：12 个文件全部在 `src/skia/*` + `src/utils/media/testdata/visual/final-yoga-skia/report.json`，**`+2648` 行，`ggrender` 命中数 0**。→ 该提交不触碰 `src/ggrender`。

### A-12 祖先判定的正确读法 + 我踩过的坑（**尺子失效形态 (b)**，记录在案）

我第一版祖先矩阵写成了：

```bash
git merge-base --is-ancestor $s "$b"; printf "%s=%s " "$(echo $b|sed …)" "$?"
```

**这是错的**：`$(…)` 命令替换在 `printf` 的参数求值时先执行，`$?` 拿到的是 `echo|sed` 的退出码（恒 0），**不是 git 的**。结果整张矩阵全打 `0`，看起来完全合理（"所有提交都是所有分支的祖先"）——这正是**尺子正常返回、却在量错的东西**形态，只能靠已知答案对照发现。

**自检与修正**：
```
git merge-base --is-ancestor 7130338 feat/gg-render-card-atomic ; echo $?   # 1  (已知应为 1)
git merge-base --is-ancestor a77efde   feat/gg-render-card-atomic ; echo $?   # 0  (已知应为 0)
```
探针在两个已知答案上给出预期读数后，才重跑全矩阵；修正后结果与 `git branch -a --contains` 的输出**互相独立地一致**，这才采信。
教训（写给下一个人）：**`$?` 必须紧跟被测命令赋值给变量，中间不得插入任何命令替换。**

### A-13 未做的事（避免被误读为已完成）
- 未运行 `go test` / `go build`（会在禁碰树产生产物）。
- 未打开、未测量、未比较任何像素（红线：只读代码不读像素）。
- 未读取 `.audit/` 下任何归档产物，因此本报告**不涉及**「归档产物缺 `provenance: retroactive-migration` 标记」一类条目。
- 未对 §2.4 枚举范围之外的目录（`src/utils/media`、`template/`、`assets/`）做符号级穷举。

---

## 7. 总结论

**用户的「gg 与 yoga 线很像，直接抄」这个想法：**

- **成立的范围**：yoga 线的 `src/ggrender` 与 gg 线**是同一个包名、同一条技术栈（Go + `fogleman/gg`）**，且已实测其资产加载链（`ggRepoRoot()` → `runtime.Caller(0)` → 本工作树根）**既不读冻结基线、也不含任何绝对路径字面量**。就「代码可移植性」而言，这一半确实是同构的，**逐文件复制在技术上可行**。
- **不成立的范围**（三条，各自独立成立）：
  1. **禁碰边界**：`src/skia` 整包把冻结基线缓存目录当渲染输入（完整取值链见 §4.1），且硬编码 `C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go` —— 这一半**不可取**，与「能不能抄」无关，是红线。
  2. **没有干净来源可抄**：gg 真正缺的 7 个场景 + 34 个 ggrender 独有符号，与 5 条干净分支的符号表**交集为 0/34**；而看似存在的两个非禁碰来源各有硬伤 —— `0d77efe`（reviewer）的 box 四件套是 yoga 的**逐字节副本**且含 `8cbffef`；`feat/gg-render-lists` 的 enemy/headhunt/recruit 虽是**独立实现**（其提交不是 yoga 的祖先），但该分支**含 `8cbffef`**。**换句话说：yoga 独有的东西，几乎都没有第二条干净的腿。**
  3. **「缺」有一部分是假缺**：换名字找过之后，`image 缩放族`（`scaleSmooth`/`smoothCover` vs gg 的 `ScaleContain`/`ScaleCover`/`ScaleExact`，内部同样用 `draw.BiLinear`）与 `文本度量族`（`measureString` vs gg 的 `measure`/`measureW`）在干净分支上**已有等价物**。按符号名差集投工时，会把已经有的东西再做一遍 —— 这正是「找不到某个名字≠不存在」那条规则的代价。
- **净结论**：可合法复用的**只有「读 yoga 的 ggrender 代码作为参考实现」这一件事**（读可以，取要逐个重新验证），**不存在可直接落地的干净代码来源**。真正省事的做法是：把 `feat/gg-render-lists` 的 enemy/headhunt/recruit 视为 gg 线自己的成果（它本来就是），并把 `feat/gg-render-lists` 上的 `8cbffef` 当作**该分支不可直接引用**的理由，而不是去 yoga 树取一份更早的副本。