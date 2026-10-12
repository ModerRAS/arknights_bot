# GG 渲染线 16 场景 · 资源加载率实测报告

- **工作树**：`C:\WorkSpace\Golang\arknights_bot-upstream-check`（本次新建，独占）
- **分支**：`check/upstream-gg`（基线 `origin/feat/gg-render-card-atomic` @ `7773d62`）
- **复现脚本**：`src/cmd/ggassetcheck/main.go`（`go build ./cmd/ggassetcheck/` 后运行；`runtime` 参数跳过 10 分钟消融扫描）
- **原始产物**：`out/`（已 gitignore，不提交）

## 0. 诚实声明

| 红线 | 本报告如何满足 |
|---|---|
| 禁把冻结基线图当渲染输入 | `src/ggrender/testdata/visual/baseline/images/*` **从未**被当作渲染输入。消融扫描只读 `assets/`。基线图仅在 §5 溯源比对中按字节读取做哈希，不进任何渲染路径。 |
| 禁改评分/相似度代码提分 | 未修改 `pixel_test.go`。`similarityNormalized` 公式逐行未动；rebase 前后 `git diff -- src/ggrender/` 为空。 |
| 禁跨工作树引用资源 | 所有资源读取经 `ggrender.AssetRoot`，脚本启动时断言其等于本工作树 `assets/`，否则 `os.Exit(1)`。从未读取 `arknights_bot-gg-card-atomic` 或任何其他工作树。 |
| 数字只能来自真实存在性检查/真实渲染 | 全部数字来自 (a) 逐资源消融 + 真跑 `RenderGG` + 像素比对，(b) 真实 `go test` harness 输出。 |

## 1. 方法（两条方法论标准）

### 1.1 资源覆盖率 = 逐资源消融实测（不用"数文件树"）
把 `assets/` 下每个文件 `rename` 成 `<name>.ablate`（移出磁盘）→ 真跑一次 `RenderGG(scene)` → 与基线渲染的 PNG 哈希比对。哈希变了 = 该资源**确实参与了该场景的像素**。

- 规模：**102 assets × 16 scenes = 1632 次真实渲染，耗时 9m42s**
- 之所以用原地 rename 而不是把 `AssetRoot` 指向影子目录：包在 init 时就把 `amiyaPath` 固化了，改 `AssetRoot` 会让 fallback 路径仍然活着，正好会掩盖要查的失败。脚本结束时检查无残留 `.ablate` 文件。
- 这是**下界**：只统计"像素可见"的使用。加载了但被完全遮挡的资源不计入（对本报告结论无影响）。

### 1.2 rebase 溯源 = `git patch-id --stable`（不用 `merge-base --is-ancestor`）
rebase 重写 sha，`is-ancestor` 对旧 sha 必然 `exit=1`。改用 patch-id：
```
git show 8d6b325 | git patch-id --stable  -> 61f4a07f46c2e35b45baacd345585a5c04a738ea
git show 905cb8e | git patch-id --stable  -> 61f4a07f46c2e35b45baacd345585a5c04a738ea
```

## 2. 核心数字：L1 磁盘层

**16 场景全部渲染成功，代码实际加载的本地资源无一缺失 → 52/52 = 100.00%**

| 场景 | 期望资源数 | 磁盘实存 | 加载率 | 实际加载的资源 |
|---|---|---|---|---|
| base | 1 | 1 | 100.00% | `font/NotoSansHans-Regular.ttf` |
| box | 12 | 12 | 100.00% | `box/{CASTER,MEDIC,PIONEER,SNIPER,SPECIAL,SUPPORT,TANK,WARRIOR}.png`, `box/Evolve_1.png`, `box/Evolve_2.png`, `common/amiya.png`, font |
| box-detail | 5 | 5 | 100.00% | `box/Evolve_2.png`, `box/Potential_4.png`, `box/Potential_5.png`, `common/amiya.png`, font |
| box-summary | 6 | 6 | 100.00% | `box/{CASTER,MEDIC,PIONEER,WARRIOR}.png`, `common/amiya.png`, font |
| calendar | 1 | 1 | 100.00% | font |
| card | 2 | 2 | 100.00% | `card/bg.png`, font |
| depot | 1 | 1 | 100.00% | font |
| enemy | 1 | 1 | 100.00% | font |
| gacha | 1 | 1 | 100.00% | font |
| headhunt | 3 | 3 | 100.00% | `box/WARRIOR.png`, `common/amiya.png`, font |
| help | 1 | 1 | 100.00% | font |
| lottery | 1 | 1 | 100.00% | font |
| missing | 3 | 3 | 100.00% | `box/WARRIOR.png`, `common/amiya.png`, font |
| operator | **1** | 1 | 100.00% | **仅 font** |
| recruit | 3 | 3 | 100.00% | `box/WARRIOR.png`, `common/amiya.png`, font |
| state | 10 | 10 | 100.00% | `state/{bg,ap,campaign,avatar-amiya,avatar-trainee,manufactures,recruit,tired_chars,tradings}.png`, font |
| **合计** | **52** | **52** | **100.00%** | |

### 2.1 资源按类型归类（`assets/` 102 个文件，全部存在）

| 类型 | 文件数 | 被 GG 实际加载 |
|---|---|---|
| 职业图标 `box/*.{PROF}.png` | 8 | 8 |
| 精英化 `box/Evolve_*.png` | 3 | 2 |
| 潜能 `box/Potential_*.png` | 6 | 2 |
| 星级 `box/Rarity_*.png` | 6 | **0** |
| `card/*` 卡片装饰 | 39 | **1**（`bg.png`） |
| `headhunt/*` | 19 | **0** |
| `state/*` | 9 | 9 |
| `help/*` | 4 | **0**（含 `label.png`） |
| `operator/*` | 1 | **0** |
| `calendar/bg.png` | 1 | **0** |
| `gacha/{header,footer}.png` | 2 | **0** |
| `common/{amiya.png,FrostNova.jpg}` | 2 | 1 |
| `css/common.css` | 1 | 0（Go 渲染器不用 CSS） |
| 字体 | 1 | 1 |

**缺口分类结论：磁盘层零缺口。** 50/102 个素材是"代码从来没要过"，不是"要了但没有"。这两者必须分开看 —— 前者不需要补资源，需要改代码。

### 2.2 字体（`LoadDefaultFont` 依次尝试）

| # | 候选 | 存在 |
|---|---|---|
| 0 | `assets/font/NotoSansHans-Regular.ttf`（本工作树） | ✅ |
| 1 | `C:/Windows/Fonts/msyh.ttc` | ✅ |
| 2 | `C:/Windows/Fonts/simhei.ttf` | ✅ |
| 3 | `C:/Windows/Fonts/msyh.ttf` | ❌ |

字体不是硬依赖（3 个系统回退），但**第 4 个候选是死代码**。

## 3. 核心数字：L2 运行层（真实网络）—— 这才是真缺口

**36 次远程请求，成功 2 次 → 5.56%**

| 场景 | 请求 URL 数 | 成功 | 结果 |
|---|---|---|---|
| box-summary | 10 | 0 | **10/10 静默降级** → amiya |
| missing | 24 | 0 | **24/24 静默降级** → amiya |
| state | 2 | 2 | ✅ 全成功（180×180，65206B / 73173B） |
| 其余 13 场景 | 0 | — | 不发网络请求 |
| **合计** | **36** | **2** | **5.56%** |

19 个 distinct URL 里 **17 个失败**，错误完全一致：
```
FETCH-FAIL unsupported protocol scheme ""   char_2001_x_1
FETCH-FAIL unsupported protocol scheme ""   char_2001_x_1 ... (共 17 个，全是裸 skin id)
```

### 3.1 根因：代码 bug，不是缺文件、也不是网络问题

`src/ggrender/render.go:556`（`RenderMissing`）与 `render.go:336/354`（`SampleBoxSummary`）把 **SkinId 当成 portraitURL** 传给 `DrawPortraitTile`：

```go
DrawPortraitTile(dc, x, y, tileW, tileH, c.SkinId, c.Profession, c.Rarity, 0, c.Name)
//                                                      ^^^^^^^ "char_2001_x_1" 是 skin id，不是 URL
```

链路：`FetchImage("char_2001_x_1", ...)` → `client.Get()` → `unsupported protocol scheme ""` → 回落 `amiya.png` → 再失败回落 1×1 像素。**全程无任何报错输出。**

后果：box-summary 的 5 个、missing 的 12 个头像位**全部画成同一张 amiya**。`SampleRecruit` 传 `Avatar: ""` 所以不发网络请求（0 降级），但同样是全 amiya。

唯一成功的是 `scene_state.go:39-40` 里两个硬编码的完整 hycdn URL。

## 4. 缺口分类

| 缺口 | 数量 | 属于哪一类 |
|---|---|---|
| 磁盘缺失的素材 | **0** | — |
| 远程 URL 构造错误导致静默降级 | **17 个 URL / 34 次调用** | **代码改动**（改 `portraitURL` 传参或 skin id→URL 拼接） |
| `resource-manifest.json` 声明但 GG 从不读 | **26 个** | **代码改动**（GG 零引用 `cache/` 目录） |
| GG 从不请求的既有素材 | 50 个 | **代码改动**（不是补资源） |
| 依赖 CDN 存活的 PASS | 1 个（state） | 稳健性隐患，非当前缺口 |

### 4.1 `resource-manifest.json` 的 26 个资源：全部在磁盘上，但 GG 一个都不读

`src/ggrender/testdata/visual/baseline/resource-manifest.json` 记录了 26 个 Playwright 抓取的源素材（`status: "frozen"`, `missing: []`, `fallbacks: []`），全部存在于 `baseline/cache/`。但：

```
grep -rn cache src/ggrender/*.go   ->  空
```

`src/ggrender/` **零引用** `cache/` 目录。GG 走的是另一套 `assets/`。所以这 26 个是"声明完整但消费为 0"：

| 资源 | targets | `assets/` 下的对应物 |
|---|---|---|
| `cache/operator-painting-1024.png` (1024×1024) | Operator painting | **无** |
| `cache/operator-skill-128.png` / `operator-building-36.png` | Operator skill / building | **无** |
| `cache/enemy-originium-slug-158.png` | Enemy pic | **无** |
| `cache/depot-lmd.png` | Depot | **无** |
| `cache/base-avatar-*.png` ×5, `base-skill-amiya-128.png` | Base 5 个干员头像 + 技能 | **无** |
| `cache/boxdetail-{avatar,skill,equip}-*` | BoxDetail 头像/技能/装备 | **无** |
| `cache/gacha-avatar-{amiya,exusiai}.webp` | Gacha | **无** |
| `cache/state-{avatar,training}-amiya-180.png` | State avatar/training | ✅ 有等价物 `assets/state/avatar-{amiya,trainee}.png` |
| `cache/amiya-{avatar,half}.webp`, `box-portrait-amiya.png`, `card-*-*.png` | 其余 | 部分有等价物 |

## 5. 溯源判定：`assets/operator/{card_bg,ring}.png` 是否来自冻结基线

**判定：溯源 PASS（不是从基线抠出来的）**

| 图 | 字节 | 尺寸 | sha256[:16] |
|---|---|---|---|
| `card_bg.png` | 7020 | 1200×675 RGB | `6d7157cd89ffb7c6` |
| `ring.png` | 60210 | 1024×1024 RGB | `d6c830411c14defa` |
| `bg.png` | 33523 | 1024×576 P | `b2b3a2de31146854` |

与 `src/ggrender/testdata/visual/baseline/**` 下全部 **43 张**图（26 cache + 16 images + card-legacy）逐张比对：
- **无字节相同**（sha256 全不匹配）
- **无像素相同**（同尺寸图逐像素 RGBA 全等比对，无命中；`ring.png` 与 `card-secretary-1024.png`/`operator-painting-1024.png` 同为 1024×1024 但像素不同）

## 6. 分数对照（`go test ./ggrender/ -run TestGGPixelParity`）

`rebase` 前后分数**逐个完全相同**，因为 `git diff ec440bf 71e7875 -- src/ggrender/` 为空（渲染代码逐字节相同 → 数学上不可能改变任何像素）。

| 场景 | 分数 | PASS |
|---|---|---|
| calendar | 0.99432 | ✅ |
| state | 0.99093 | ✅ |
| lottery | 0.98358 | |
| card | 0.96020 | |
| base | 0.95900 | |
| enemy | 0.91951 | |
| depot | 0.91728 | |
| gacha | 0.90251 | |
| box-summary | 0.88098 | |
| box-detail | 0.87844 | |
| missing | 0.87614 | |
| help | 0.83943 | |
| box | 0.82504 | |
| headhunt | 0.79983 | |
| operator | 0.69063 | |
| recruit | 0.61351 | |
| **PASS** | **2/16** | |

### PASS 的资源可复现性

| 场景 | 可复现性 | 依据 |
|---|---|---|
| state 0.99093 | 可复现，**但依赖 CDN 存活** | 2 个硬编码 hycdn URL 实测取到 180×180；离线立即掉分 |
| calendar 0.99432 | 可复现，**且不依赖网络** | 只用 `{calendar/bg.png, font}` 两个本地资源，零网络请求 |

两个 PASS 都不依赖缓存资源或基线图。calendar 的可复现性来源更稳固。

## 7. 结论：最小可行下一步（按工作量排序）

1. **修 `portraitURL` 传参**（`render.go:336/354/556`）。把裸 skin id 拼成完整 hycdn URL，或改传真实 URL 字段。影响 box-summary + missing 共 17 个头像位、34 次静默降级。**纯代码改动，零新资源，预计提分最直接。**
2. **给 `RenderOperator` 接上 `assets/operator/bg.png`**。该文件一直存在于仓库、GG 从未引用。实测 **0.69063 → 0.84554（+0.15491）**。一行代码。
3. **不要接 `card_bg.png` / `ring.png`**。实测 `ring.png` **降低**分数（−0.012，叠加 bg 后 −0.052），`card_bg.png` 仅 +0.012。它们服务 upstream 新版 `Operator.tmpl`，冻结基线里既没有环也没有深蓝网格。
4. **把 `resource-manifest.json` 的 26 个资源接进 `assets/`**。缺的是 `operator-painting-1024`（立绘，占 operator 画布约 21%）、`enemy-originium-slug-158`、`depot-lmd`、`gacha-avatar-*`、`boxdetail-*`、`base-avatar-*`。属**网络抓取**（URL 已在 `resource-manifest.json` 里现成）—— 但要先决定是否允许从 `cache/` 搬到 `assets/`。
5. **operator 版式按基线实测坐标重排**。现状 800×700 被 `ScaleToManifest` 非等比拉伸到 1800×1200（宽高比 1.14→1.5），几何本身就歪；且缺 9 格属性表 / 天赋技能面板 / 潜能面板 / 职业图标。属最大工作量，建议按 measured-geometry 路线单独立项。

### 7.1 operator 剩余 0.15 差距的重新定位（measured-geometry 工作量估算）

背景已由 V3 修掉（+0.155），剩下的 0.84554 → 0.99 差距来自**结构与主体**，三块：

| 差距来源 | 性质 | 说明 |
|---|---|---|
| **中央立绘缺失** | 资源 + 代码 | `cache/operator-painting-1024.png`（1024×1024）只在 `cache/`，`assets/` 无对应文件，GG 零代码。占画布约 **21%**（≈465k / 2.16M 像素）。**属最大单项**，且需先决定是否允许把 `cache/` 素材搬进 `assets/`。 |
| **布局结构不同** | 代码 | 基线 6 个独立面板：①满幅碎片背景 ②中央立绘 ③左上 3×3 属性表（9 格）④右上天赋/技能面板（含技能图标 + 技力胶囊）⑤左中潜能提升面板 ⑥左下职业图标 + 6 星 + 定位标签 + 底部姓名/编号。GG 画的是「顶部 name bar + 描述卡 + 3 列属性网格」，无一对应。 |
| **画布非等比拉伸** | 代码（前置） | 800×700 → 1800×1200，宽高比 1.14→1.5。**必须先修**，否则所有坐标测量都建立在错误变换上。 |

**工作量估算（对比法，非实测工时）**

以本仓库已有同类重建为标尺：`calendar` 0.46492 → 0.99432（1 个提交 `fd31d2e`）、`state` 0.85 → 0.99093（2 个提交 `bc477f4` + `a35d1d9`）。operator 需要 6 个面板 + 1 个前置修复 + 1 项资源决策，**量级约等于 calendar + state 之和，估 3–5 个聚焦提交**。

建议顺序：① 先改原生 1800×1200 画布（去掉拉伸）② 接立绘 ③ 逐面板实测坐标对齐 ④ 迭代收敛。每轮校准成本极低 —— 完整 harness 跑一次仅 6.5s，消融扫描 10min 只在需要复核资源时跑。

> 注：此估算基于同仓库同类场景的提交粒度对比，**不是实测工时**，需人工判断是否接受。

## 8. 取证证据附录（所有否定性结论的 exit code）

规则：否定性结论必须 `exit≠0` 且输出为空，或有自校验；否则一律标注为**视觉判读**。禁止 `|| echo none` 兜底（它会把「查询失败」伪装成「未命中」）。

| # | 否定性结论 | 命令 | exit | 输出 | 判定 |
|---|---|---|---|---|---|
| N1 | ggrender 不引用 skland | `grep -rn "skland" src/ggrender/*.go` | **1** | 空 | ✅ 真无命中 |
| N2 | ggrender 不 import core/web | `grep -rn "core/web" src/ggrender/*.go` | **1** | 空 | ✅ 真无命中 |
| N3 | 无任何生产代码调用 RenderGG | `grep -rn RenderGG --include=*.go src/ \| grep -v src/ggrender/ \| grep -v _test \| grep -v ggassetcheck` | **1** | 空 | ✅ 真无命中 |
| N4 | rebase 前后 src/ggrender 逐字节相同 | `git diff --quiet ec440bf 71e7875 -- src/ggrender/` | **0** | — | ✅ 0 = 无差异（此处 0 才是「成立」） |
| N5 | ggrender 不引用 card_bg / ring.png | `git grep -n -iE "card_bg\|ring\.png" 71e7875 -- src/ggrender` | **1** | 空 | ✅ 真无命中 |
| N6 | ggrender 代码不读 baseline/cache/ | 6 个 pattern 逐一试（`baseline/cache`、`"cache"`、`cache/`、`CachePath`、`resource-manifest` 等） | **全部 1** | 全空 | ✅ 真无命中 |
| N6c | manifest/cache 仅被**测试**引用 | `git grep -ln 'testdata/visual' -- src/ggrender/*.go` | 0 | 仅 `pixel_test.go` | ✅ 渲染代码零引用 |
| N7 | rebase 后无未合并路径 | `git diff --name-only --diff-filter=U` | 0 | 空 | ✅ |
| N8 | .gitignore 无冲突标记 | `grep -c '<<<<<<<' .gitignore` | **1** | `0` | ✅ |
| N9 | 43 张基线图枚举完整 | Python `assert len(imgs)==43` | 0 | `enumeration self-check: PASS (43)` | ✅ 枚举本身可证 |
| N10 | card_bg/ring 非基线抠图 | Python 逐张 sha256 + 同尺寸逐像素 RGBA 全等比对 | 0 | 0 字节匹配 / 0 像素匹配 | ✅ 循环确实执行（ring 实际比了 2 张同尺寸） |

### 8.1 行尾纠正（我先前报错过）
先前我说「仓库是 CRLF」，**是错的**。权威字节证据：
```
$ head -c 40 src/ggrender/render.go | od -c
0000000   p   a   c   k   a   g   e       g   g   r   e   n   d   e   r
0000020  \n  \n   i   m   p   o   r   t       (  \n
```
仓库是 **LF**。之前的误判来自 `git show ... | grep -c $'\r'` —— 管道经过 MSYS 时引号展开不可靠，**不可用作行尾取证**。我新建的两个文件（`main.go`、`UPSTREAM_GG_CHECK.md`）同样是 LF，与仓库一致。所有提交均为**纯新增、0 删除**（`git diff --numstat` 已核），无整文件重写、无行尾膨胀。

### 8.2 哪些是视觉判读而非 grep

以下**来自我实际打开 `src/ggrender/testdata/visual/baseline/images/operator.jpg`（1800×1200）用眼睛看**，不是 grep 结论，特此标明：

- 「基线里**没有头像环**」
- 「基线里**没有模组（module）图标**」

判读方式与局限：用图像读取工具直接打开该 JPEG 缩略查看。局限：缩略后极小的图标可能漏看；这两条属于「未见」，强度弱于像素级证据。若需升级为硬证据，应对头像区域做像素级模板比对。**不影响 R6 结论方向** —— 因为 `ring.png` 的 A/B 分数是**实测下降 0.012**（硬数字），比「基线无环」的目视判读更强。

其余所有元素（碎片背景、属性表、天赋/技能面板、潜能面板、职业图标 + 6 星、姓名/编号）均为同一张图的视觉判读，同此声明。

## 9. 复现

```bash
cd C:/WorkSpace/Golang/arknights_bot-upstream-check/src
go build -o ggac.exe ./cmd/ggassetcheck/
./ggac.exe            # 全量：消融扫描（约 10 分钟）+ 真实网络运行层
./ggac.exe runtime    # 只跑运行层（秒级）
```

## 10. 产出物定位

| 分支 | 内容 | 定位 |
|---|---|---|
| `check/upstream-gg` | `src/cmd/ggassetcheck/main.go` + 本报告 | **1632 次消融实测的可复现凭据**。不推远端。 |
| `feat/gg-rebase-upstream` (`71e7875`) | 纯 rebase，12 提交，`src/ggrender/` 零改动 | 已推远端（单条新分支，不 force） |
| `check/operator-img-ab` (`846cea0`) | R6 的 env 驱动探针（`GGOPBG` / `GGOPRING`） | **实验用，不推远端**。未设环境变量时输出与实验前**逐位一致**（V0 复现 0.69063 已验证），可随时 revert。 |
