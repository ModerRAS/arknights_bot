# 常备检查：冻结基线是否被当作渲染输入

**这是一个常备件（standing check），不是某次测量的归档。**
它管的是整棵 `src/ggrender/` 渲染路径，与任何单个 commit 无关。

本目录**只归档脚本与它的可信度依据**，**不归档任何一次运行的输出**。
原因：`0 violations` / `VERDICT: CLEAN` 是**某一刻对某一个 ref 的观测**，
把它写进 git 等于把一个会过期的快照焊死，下一个人会以为那个 CLEAN 是永久结论。
**脚本是检查能力的载体，它不随代码变；输出会。**

---

## 它回答什么问题

> `src/ggrender/` 下**任何渲染路径**，有没有把冻结的 Playwright 基线目录
> 当作**渲染输入**读入（即：把基线图 load 进来、画到画布上、当背景/当元素）？

核心难点是**必须区分两种角色**：

| 角色 | 谁消费 | 判定 |
|---|---|---|
| **比较基准** | 差异计算 / 相似度 / 尺寸检查 | **合法** |
| **渲染输入** | 任何 `DrawImage` / `Scale*` / 编码路径 | **作弊** |

因此这个检查**不能**简单地把 `pixel_test.go` 排除掉——
那会漏掉"有人往测试里加了渲染输入"这种情况；
也**不能**简单地 grep `baseline` 就算作弊——那会把合法的比较侧全部误报
（`scene_state.go` 里 `baseline := labelTop + 21` 是**排版基线 y 坐标变量**，与冻结基线毫无关系）。

做法是：枚举每个读入原语的每个调用点，判定路径参数形态，
再判定**消费方**是谁。

---

## 怎么用

```bash
cd C:/WorkSpace/Golang/arknights_bot-gg-card-atomic
python .audit/standing-checks/baseline-as-render-input/check_baseline_as_render_input.py --ref HEAD
```

可选：`--json <path>` `--md <path>` 输出机器可读 / 人类可读两份。
`--ref <sha|ref>` 可对任意 ref 跑（本脚本只读 git 对象，不 import 渲染包）。

**运行约定：**

- **从仓库外运行，或加 `python -B`**，避免在归档目录里生成 `__pycache__`。
- 脚本不写任何文件（除非显式给 `--json`/`--md`），不 import 渲染包。
- 因此它**必须**放在 `src/` 之外——放进去会污染渲染包，而它不是渲染代码。

### 退出码

| 码 | 含义 |
|---|---|
| `0` | **干净**：无渲染路径把冻结基线当渲染输入 |
| `1` | **发现违规**：有渲染路径读了基线当渲染输入 |
| `2` | **尺子自检失败**：本次结果**不可信**，**不可当作干净** |

> `2` 是这个检查里最重要的一条纪律：**"尺子坏了"与"没查到"在输出上完全一样。**
> 自检不通过时脚本拒绝输出任何判定，而不是给一个"看起来干净"的结果。

---

## 自检：怎么判断这个脚本可不可信

**这一节永远有效**——它描述的是"凭什么能相信这个脚本"，而不是某次跑出了什么。

### 硬对照（任何 ref 都必然成立）

| 对照 | 探针 |
|---|---|
| `git show` + `grep` 能读到该包 | `package ggrender` |
| 读入原语的正则能匹配真实代码 | 全部 9 个原语合计命中数 > 0 |

**硬对照不成立 → 立即中止并退 `2`**，不输出任何判定。

> 硬对照最初被写成语义性的（硬编码 `ScaleToManifest` 是否存在）。
> 拿已知作假的 `641988f` 做阴性对照时它**失败了**（退 2）——因为
> `ScaleToManifest` 在那个 commit 里还不存在。**尺子反过来挡住了它本该放行的测试。**
> 改成结构性强对照后才可用。教训：对照本身也要能在旧 ref 上成立。

### 软对照（语义性，旧 ref 上可能本就不存在，只记录不致命）

- `pixel_test.go` 中出现 `baseline` 字样
- `helpers.go` 中出现 `ScaleToManifest`

---

## 阴性对照：对已知作假的 `641988f` 必须报违规

**这是把"0 违规"从"没查"变成"查过且没有"的唯一依据。**

`641988f`（`card 0.99939 atomic dedup`）是已知的伪达标提交，它在 `scene_card.go` 里做了：

```go
if bg, err := LoadImage(filepath.Join("C:/WorkSpace/Golang/arknights_bot-card-overflow2/src/ggrender/testdata/visual/baseline/images/card.jpg")); err == nil {
    dc.DrawImage(ScaleCover(bg, mainW, mainH), 0, 0)
```

—— 把**另一个工作树**里的冻结基线图当背景画上去。

**同一个脚本对 `641988f` 必须报 `VIOLATION`（exit 1）**，并准确抓到那两行
（`RenderCard` 的 `card.jpg`、`RenderDepot` 的 `depot.jpg`）。

**只要这个阴性对照还能报出违规，当前 `CLEAN` 就是有信息量的结论；
一旦它报 `CLEAN`，说明尺子坏了，此时所有 `CLEAN` 都必须作废。**

---

## 何时必须重跑

这是常备件的价值所在。以下任一情况发生，**必须重跑**并确认退出码为 `0`：

1. **任何改动 `src/ggrender/` 下任何文件的提交**——渲染代码与测试代码**都算**。
   脚本刻意**不排除** `pixel_test.go`，正是为了抓住"有人往测试里加了渲染输入"。
2. **任何新增 asset 加载点**（`LoadImage` / `tryLocal` / `cardAsset` / `AssetPath` /
   `os.Open` / `os.ReadFile` / `FetchImage` / `fetch` / `LoadFontFace` 的调用）。
3. **任何改动 `AssetRoot` 解析**的提交（helpers.go 里 `AssetRoot` 的求值）。
4. **任何合并 / rebase / 分支推进之后。**
5. **任何一次 `card` 分数被引用之前。**
   （该分数的有效性此前未被检验：harness 不计算内容掩膜重叠率，
   `similarity` 单独不足以证明与冻结基线结构对应。）

建议挂成 pre-commit 或 CI job，`exit 1` 即阻断。

---

## 维护注意

- 本脚本**不做** `content_mask_overlap` 一类的门禁判定，那是另一件事。
- 它的输出**不要归档**。要看结果就跑一次，结果只对当次那个 ref 有效。
- 改这个脚本时，**先确认阴性对照仍然报违规**，再谈其它改动。
