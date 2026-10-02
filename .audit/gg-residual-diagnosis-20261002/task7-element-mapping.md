# Task 7 — 元素级映射 + 跨轮一致性答复（三桶制）

```
provenance:   measured
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align（复用，未新建）
              生成时间 2026-10-02T13:40:00Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task7-element-mapping.md
判别口径:      tmp/residual/mapping/instrument.json（跑数前落盘，v2 修订已登记）
脚本:          tmp/residual/mapping/{map_elements.py, premise_survey.py}
产物:          tmp/residual/mapping/{controls.json, result.json, premise_survey.json,
               ceiling_tiles.json, measure_stdout.log(0B), measure_stderr.log(0B),
               survey_stdout.log(0B), survey_stderr.log(0B), gate_test.log, teeth2.log}
输入:          tmp/residual/discriminator/result.json (task6) + tmp/align/report.json (task4)
produced_by:   src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47（逐行落盘）
```

**修改权声明**：未改任何渲染代码、未改计分代码、未 commit、未 push、未新建工作树。禁碰树未读未渲。`git diff --stat` 空。

---

## 0. Boss 的三个数 + 三桶制

| 桶 | 判定语句（逐字照用） | 行数 |
|---|---|---|
| 真目标 | 「检测器在工作，它说有洞」 | **0** |
| 假阳性 | 「检测器在工作，它说有洞，但那是产物」 | **0** |
| **域外** | **「检测器在这里没有发言权」** | **5**（base + 其余 4 行） |

- **适用分母：0 行适用、0 假阳性**（域外行不进入分母）
- **原始分母：5 行入选、5 行事后判为域外**

### ⚠️ 必须订正的一句话

> 我先前写的「该内洞检测器的假阳性率 = 0.20」——**这句话是错的，请勿再引用。**

错在两处，且两处都是硬的：
1. **分母混了两种错误代价完全不同的失败模式**。「判错」= 按它说的做了无用功；「不适用 / 域外」= 按它说的做了**根本不该做的事**。0.20 把两者压成一个数，下一个人只会记住「它有 20% 不准」，不会记住「其中一部分场景它压根不该发言」。
2. **它是在错误的「真」的定义下测的**。0.20 的分母把「R1R2 通过」当成「真目标」。本轮映射证明：**5 个洞里 0 个是「声明里有、渲染里没画」**。按 Boss 真正在意的定义（漏画元素），这 5 个洞**全部**不是目标。

**⇒ 若非要给一个数：「在 5 个被检测器点名的洞里，漏画元素 0 个」；「检测器把这 5 个洞说成整块没画，5 次全错」。不要再用 0.20 这个数。**

### ⚠️ 与 Boss 预期形状的分歧（必须报，不迎合）

Boss 的裁定写的是「适用分母：4 行适用、0 假阳性」。**我的证据不支持「4 行适用」**，见 §2/§3：4 行的洞内都测到了被渲染的内容（ink 0.0035–0.120），且 4 行的「旧侧众数色命中声明色、新侧未命中」或「两侧同色」都指向**填色来源与声明不符**，而不是漏画。前提在这 4 行同样不成立 ⇒ 全部进域外桶。

我没有为了对上 Boss 的预期而改判。分歧的证据在 §2、§3 逐行给出，可复核。

## 1. 逐行映射结果（4 行一次跑完，跑序按 Boss 指定）

| 序 | 场景 | 占残差份额 | ink_fraction | 洞内众数色 new / old | 声明侧颜色判定 | 结局 |
|---|---|---|---|---|---|---|
| 1 | `enemy` | 0.4492 | **0.10247** | (40,42,43) / (51,51,51) | old 命中 `#323332`（Enemy.tmpl:10 `#main background-color`），new 未命中任何声明色 ⇒ **renderer_used_non_declared_colour** | 域外 |
| 2 | `operator` | 0.2037 | **0.11992** | (40,40,40) / (40,40,40) | 两侧同色且都不命中声明色（该背景色不在声明里） | 域外 |
| 3 | `box-detail` | 0.2025 | **0.00354** | (37,39,40) / (46,48,49) | old 命中声明色，new 未命中 ⇒ **renderer_used_non_declared_colour** | 域外 |
| 4 | `lottery` | 0.3185 | **0.10880** | (26,26,26) / (34,34,34) | new 命中 `#1A1A1A`（Lottery.tmpl:9 `--container-bg`）⇒ 声明色在场 | 域外 |

**逐行三件套**

**① `enemy`**
- 声明中的元素标识：`#main { background-color: #323332; width: 656px }`（`template/Enemy.tmpl:8-11`）、`td { background-color: #323332 !important }`（:16-19）、`.level { width: 656px }`（:30-31）、通用 `img { width: 158px }`（**`Enemy.tmpl:44` —— 通用标签选择器，不是 `.avatar` 类名，正是「搜不到名字≠不存在」的样本**）、`img src="{{.Pic}}" onerror="this.src='assets/common/amiya.png'"`（:65）。共享样式表 `/assets/css/common.css`（:5，仅登记）。
- 渲染中的存在性判定：**内容在场**（ink 0.10247，裁图可见 HP/ATK/DEF/RES 与能力/技能文本行）。洞 = 文本行之间的平坦背景连通域。
- 判定依据：洞内众数色 old (51,51,51) 命中声明色 #323332，new (40,42,43) 未命中任何声明色 ⇒ **基线按声明填色、渲染用了非声明色**。这是「画错了」，不是「漏画」。

**② `operator`**
- 声明标识：`width: 1200px`（Operator.tmpl:9）、`width: 300px`（:46）、`width: 600px`×3（:63/:77/:87）、`img {{.Painting}}`（:114）、`img class="potential" /assets/box/Potential_{{.Rank}}.png`（:148）、`img /assets/box/{{.OP.Profession}}.png`（:158）、`img /assets/box/Rarity_{{.OP.Rarity}}.png`（:160）、`img class="skill" onerror=…amiya.png`（:235）。通用 `img {}` 规则同样存在（已按多种叫法枚举：标签名 / class / id / 模板自定义块）。
- 存在性判定：**内容在场**（ink 0.11992，裁图可见一条斜切的立绘亮块 + 暗背景）。
- 判定依据：洞内两侧众数色**完全相同** (40,40,40)，但 `frac_diff = 0.85111` ⇒ 差异来自结构/纹理而非众数色；该颜色不在任何声明里 ⇒ 背景来源与声明不符。

**③ `box-detail`**
- 声明标识：`img onerror=…amiya.png`（:39，远端 hycdn URL）、`img /assets/box/Evolve_{{.EvolvePhase}}.png`（:40）、`img /assets/box/Potential_{{.PotentialRank}}.png`（:41）、另两条远端 `img`（:45、:52）。**通用 `img { width: … }` 规则，无 `.avatar` 类名**。
- 存在性判定：**结构近乎为空**（ink 0.00354，全画布 0.1353 结构占比下的局部空区），但众数色 old 命中声明色、new 未命中 ⇒ 不是「什么都没画」，而是「画了另一种颜色」。
- 判定依据：4 个命中全部 `frac_diff ≈ 0.999` —— 即整块区域**逐像素都在两侧不同**，这是颜色差异的特征，不是缺内容的特征（缺内容会给出大片 `frac_diff=0`）。

**④ `lottery`**
- 声明标识：`--container-bg: #1a1a1a`（`template/Lottery.tmpl:9`）。该模板**不引用共享样式表**、**不声明任何 `img/video/canvas`**。
- 存在性判定：**内容在场**（ink 0.10880，裁图可见完整 9×11 抽奖卡网格、第 41 号为选中态并带 ID 与「参与用户」标签）。
- 判定依据：众数色 new (26,26,26) **命中**声明色 #1A1A1A ⇒ 声明的底色在场；洞 = 卡片内部平坦区。

**三种结局计数（校验式）**：真目标 0 + 假阳性 0 + 域外 4 = **4** ✓（域外是 task6 的 4 行；连同 `base` 共 5，见 §2）

**⇒ 直接回答 Boss 那一句：「声明里有、渲染里没画」= 0 个。**
⚠️ 按 Boss 指令，域外桶的行先回报、不往下走 —— 本轮不给实现方向。

## 2. `base`：结论保留，理由改写（按 Boss 裁定）

> **结论：`base` 没有可投工的「整块没画」目标。**（维持）
> **理由改为**：在以大面积纯色为底的画布上，「结构量 / 非平坦区域」前提不成立，洞的定义不适用；这 4 个命中不是与其它场景可比的洞对象。**检测器在这里是沉默的，不是投了反对票。**

可复核的量（`premise_survey.json`）：`nonflat_frac_scene = 0.0805`（全 16 场景第 3 低）、`flat_tile_frac = 0.6094`、`shared_flat_frac = 0.4774`（近一半 tile 在 old 与 new 中**都**平坦）、洞内两侧众数色**逐位相同** (33,38,47) 且都命中声明色、4 个命中 `frac_diff` 全部不过 0.50（0.2570/0.1976/0.1235/0.1500）。

**为什么必须改写理由而不只是改数字**：旧写法「未过 R1」读起来像**检测器正常工作后给出了否定**（下一个人会放心复用）；新写法「前提不成立」读起来像**检测器在这里是沉默的**（下一个人知道它没有投票权）。沉默与否定是两件事。

## 3. 另外 4 行：逐行给前提证据与**两个方向**的影响

**不写「应该不受影响」——「应该」是背景知识，不是测量。**

| 场景 | 前提是否满足（可复核量） | 缺陷可能把真目标误判为**假/域外**的方向 | 缺陷可能把真目标误判为**真**的方向 |
|---|---|---|---|
| `enemy` | **不满足**。`nonflat_frac_scene = 0.0912`（第 2 低）、`flat_tile_frac = 0.6667`（全 16 最高）、`shared_flat_frac = 0.3283` | 若一个真实缺失元素被**同色填充盖住**，该区域会呈现为「平坦 + 两侧同色/近同色」，检测器沉默 ⇒ **真目标被判成域外而永不进 B 类** | 若把「平坦 = 空」当成真，则任何大面积纯色底都会被读成洞 ⇒ **域外被读成真目标**，按它去画不该画的东西 |
| `operator` | **不满足**。`nonflat_frac = 0.1544`、`flat_tile_frac = 0.4583`、`shared_flat_frac = 0.0747`（低，但众数色两侧相同） | 同上：同色掩盖 ⇒ 真目标被吞 | 同上 |
| `box-detail` | **不满足**。`nonflat_frac = 0.1353`、`flat_tile_frac = 0.5967`；洞内 ink 仅 0.00354，4 个命中 `frac_diff ≈ 0.999` | 近乎全平的区域里若藏一个缺失元素，结构量不足以把它顶出来 ⇒ 真目标被判成「纯色背景」 | 逐像素全不同的整块会被读成「洞里该有东西」⇒ 把**颜色差异**误当成**漏画元素** |
| `lottery` | **部分满足但仍不成立**。`flat_tile_frac = 0.1042`（低，画布本身结构够），但洞内众数色**命中声明色**、ink 0.10880 ⇒ 「平坦 = 空」在该区域直接为假 | 若某张卡片整块没画且底色与容器底色相同，则该卡变成一个「平坦且与声明底色一致」的块 ⇒ 被沉默 | 把「卡片内部的平坦底色」读成洞 ⇒ 把**样式差异**误当成漏画 |

**结论：4 行全部不满足前提，全部进域外桶。** 这不是「检测器坏了」，而是**检测器的前提（平坦 ⇒ 该区域本该有内容却空着）在这些画布上系统性地不成立**。

## 4. 问题 3：修正仪器前后，两个结论都报

| 行 | v1 口径（只看结构：`ink < 0.02` 即「空」） | v2 口径（结构 ∧ 颜色不命中声明色） | 是否翻转 |
|---|---|---|---|
| `enemy` | 域外（ink 0.10247 → 非空） | 域外（同） | **否** |
| `operator` | 域外（ink 0.11992 → 非空） | 域外（同） | **否** |
| `box-detail` | 域外（ink 0.00354 → 空，但无几何候选 → 落 O2「声明中无对应元素」） | 域外（同，众数色未命中声明色） | **否** |
| `lottery` | 域外（ink 0.10880 → 非空） | 域外（同） | **否** |

**修正后没有任何一行结论改变** —— 但修正**不是多余的**：v1 的缺陷是「**被正确渲染的纯色元素结构量为零，会被判成未绘制**」，它由对照 `POS2_correctly_painted_flat_element_NOT_empty` 抓出来（该对照在 v1 下失败）。若某行的洞恰好是一个**画对了的纯色元素**（例如声明了 `background-color` 的 div），v1 会误报 O1。本轮 4 行都没踩到，但下一轮会。**这段恰恰最值钱 —— 它告诉下一个人这套东西在哪种条件下会翻转。**

## 5. 全 16 场景的前提普查（**沉默不是证据**）

⚠️ Boss 指出的陷阱，我用数据正面回应：`base` 当初被判假阳性，是因为它**有 4 个命中却都不过 R1**。**一个画布若根本没有命中，它不会进 B 类，也就永远不会被发现有这个问题。**

| 场景 | nonflat | flatFrac | sharedFlat | 内洞数 | 过 A_min 的命中 | R1R2 | 当前 sim |
|---|---|---|---|---|---|---|---|
| base | 0.0805 | 0.6094 | 0.4774 | 5 | 4 | ✗ | 0.97073 |
| box | 0.5260 | 0.1667 | 0.1146 | 0 | 0 | — | 0.82504 |
| box-detail | 0.1353 | 0.5967 | 0.1067 | 12 | 4 | ✓ | 0.87844 |
| box-summary | 0.1501 | 0.4219 | 0.1198 | 13 | **0** | — | 0.88098 |
| calendar | 0.0427 | 0.6059 | 0.5556 | 2 | 2 | ✗ | **0.99432（已过线）** |
| card | 0.5908 | 0.0851 | 0.0590 | 1 | 0 | — | 0.96015 |
| depot | 0.0670 | 0.9022 | 0.4615 | 0 | 0 | — | 0.92352 |
| enemy | 0.0912 | 0.6667 | 0.3283 | 3 | 2 | ✓ | 0.91951 |
| gacha | 0.0702 | 0.4462 | 0.0816 | 0 | 0 | — | 0.90251 |
| headhunt | 0.5464 | 0.0521 | 0.0330 | 1 | 0 | — | 0.97021 |
| help | 0.0810 | 0.2569 | 0.0312 | 0 | 0 | — | 0.83943 |
| lottery | 0.1264 | 0.1042 | 0.0312 | 2 | 1 | ✓ | 0.98358 |
| missing | 0.3531 | 0.4236 | 0.3385 | 0 | 0 | — | 0.87614 |
| operator | 0.1544 | 0.4583 | 0.0747 | 7 | 1 | ✓ | 0.76866 |
| recruit | 0.3647 | 0.4497 | 0.3003 | 0 | 0 | — | 0.88603 |
| **state** | 0.1220 | 0.5139 | 0.4635 | 5 | **4** | **✓** | **0.99093（已过线）** |

**画布层面「平坦即空」前提可疑的名单（`flat_tile_frac ≥ 0.40` 或 `shared_flat_frac ≥ 0.30`）**：
`depot`(0.9022/0.4615)、`enemy`(0.6667/0.3283)、`calendar`(0.6059/0.5556)、`state`(0.5139/0.4635)、`box-detail`(0.5967)、`box-summary`(0.4219)、`base`(0.6094/0.4774)、`recruit`(0.4497/0.3003)、`operator`(0.4583)、`gacha`(0.4462)、`missing`(0.4236)。

**两个必须一起读的事实：**
1. **`state` 有 5 个内洞、4 个过 A_min、且 R1R2 全部通过 —— 而它的 sim 是 0.99093，已经过线。** 也就是说「有命中 + 通过判别」**不等于**「需要投工」。这条反过来印证 Boss 的陷阱：沉默（没被报过）不能证明没问题；**同样，通过了也不能证明需要修**。
2. **`box-summary` 有 13 个内洞但 0 个过 A_min**，它从未被判成真目标。若 `A_min` 口径改动，它可能翻转。**我没有把它列入任何桶，因为「没进 B 类」不是证据。**

**不许说的话**：「名单里其它场景没问题」——目前**没有任何证据**支持这句话。已过线的 `calendar`/`state` 恰恰各自有命中，说明该检测器在已过线场景上照样响。

## 6. 三件事必须分开说（Boss §6）

| # | 事项 | 状态 | 量 |
|---|---|---|---|
| 1 | **全局常数色偏**（每通道一个偏移） | **已测死** | 天花板最高 `lottery` 0.98510 / 0.98501（task5 口径），最小缺口 **0.00490**；最大单场景收益 `help` +0.04536。**这条成立。** |
| 2 | **逐元素 / 分区域错色** | **本轮第一次被测** | 修正族 = 每个 24×24 tile 一个独立每通道常数（576 参数/场景）。结果：**没有任何未过线场景达到 0.99**。最高 `lottery` **0.98572**（缺口 0.00428），`operator` 0.91834、`box-detail` 0.93008、`enemy` 0.96414、`help` 0.90478。 |
| 3 | 4 个洞多数落在「可见但非纯色 ⇒ 错色」 | 成立（§1） | 它们的期望收益 = **逐区域错色的收益**，与第 1 条**性质完全不同**。 |

**第 2 条的关键发现（此前不可见）**：`operator` 的分区域收益 **+0.14968**，而它的全局常数收益只有 **+0.00804** —— **相差 18 倍**。原因：该场景的色偏在空间上不均匀，全局常数族**根本看不见它**。这证明「全局色偏已测死」**不能**推广成「色偏问题已测死」。

**必须同时说清的限制**：第 2 条是**诊断上界**，不是可达目标 —— 576 个自由参数/场景是拟合出来的量，它衡量的是「残差里有多少是空间变化的颜色」，不是一个合法的修法。**不许把它当成「分区域修色能到 0.98572」的承诺。**

**结论口径**：逐区域错色这条路**不再是「上限未知」**（本轮已测），也**不是「零」**（收益真实存在，最大 +0.14968），而是**「已测得不足」**——上限 0.98572，仍差 0.00428 到门禁。

## 7. 两件便宜的事

**`onerror` 回退触发率：检出 0 处。**

| 场景 | 模板中带 onerror 回退的元素数 | 渲染侧检出数 | 最高归一化相关 |
|---|---|---|---|
| enemy | 1（`Enemy.tmpl:65`） | **0** | 未取得（模板尺寸超画布，未检） |
| operator | 1（`Operator.tmpl:235`，`class="skill"`） | **0** | 未取得（同上） |
| box-detail | 1（`BoxDetail.tmpl:39`） | **0** | 0.2727（阈值 0.90） |
| lottery | **0**（该模板不声明任何媒体元素） | **0** | 0.0060 |

口径与限制：检出为**下界**（相关 > 0.90 才记检出，缩放/抗锯齿会削弱相关性）；enemy/operator 两行因探测尺寸超出画布而**未取得相关值**，标「未测」而不是「0」。**结论只能说到「在可检的两行里没有回退痕迹」，不能说「回退从未发生」。**

**`enemy` hit 级拆分**：
- **主洞 = hit0**，pixel box `[41,171,942,417]`，占残差 **0.44922**，`frac_diff 0.99988`，R1R2 通过。
- 附属小块 = hit1，pixel box `[41,19,163,132]`，`frac_diff 1.0`（面积上 100% 两侧不同）但只占残差 **0.02666**，**未过 R2**。位置在画面**顶部左侧**、远小于主洞 —— 大概率是主洞上方的 `#name` 行附近的背景块。
- ⇒ 主洞承载该场景残差的 **94.4%**（0.44922 / (0.44922+0.02666)），排序与后续动作应以主洞为准。

## 8. 闸门必须验牙齿（Boss 常设流程 A，本轮做了 3 次）

| 次 | 脚本 | 抽掉的输入 | 结果 | exit |
|---|---|---|---|---|
| 1 | `map_elements.py` | `controls.json` | `REFUSED: controls.json missing`，**未产出 result.json**（`ls` 退出码 2 = 不存在） | **2** |
| 2 | `premise_survey.py` | `controls.json` | `REFUSED: controls_pass != true`，**未产出 premise_survey.json** | **2** |
| 3（前置） | `discriminate.py`（task6） | `controls.json` | `REFUSED`，无结果文件 | **2** |

另有两条结构性证据：三个 measure 阶段的 **stdout 与 stderr 均为 0 字节**（分类结果只写文件）；`--stage measure` 在 `controls.json` 不存在时 `SystemExit(2)`，是硬失败而非警告。

**对照结果（映射器 4 条 + 前提普查 2 条，全过）**：
`POS1_truly_blank_region_is_empty` ✓、`POS2_correctly_painted_flat_element_NOT_empty` ✓（这条在 v1 下失败，抓出了仪器缺陷）、`POS3_real_content_region_NOT_empty` ✓、`POS_declared_colour_lookup` ✓；`POS_shared_flat_fraction_measurable` ✓、`POS_flat_on_both_sides_but_colours_differ` ✓（证明「两侧都平坦」推不出「两侧相同」，也推不出「未绘制」）。

## 9. 过程 B（先 dump 键名）：本轮被自己抓到 2 次

| 次 | 我猜错的字段 | 实际 | 后果 |
|---|---|---|---|
| 1 | `result.json` 的 `class` / `kept`（Boss 先前猜错的两处） | 先 `print(list(d.keys()))` 读到真实结构：`B_flagged_before / B_after / survivors / false_positives / flag_level_false_positive_rate / rows[].largest_hole.{pixel_box, frac_diff, share_of_canvas_residual, R1_area, R2_materiality, verdict_real_target}` | 避免了「判定后剩 0 行」的假象 |
| 2 | task5 `report.json` rows 的 `go_passed` | rows 实际键里**没有** `go_passed`，且只有 14 行（不含 calendar/state）；`current_sim` 在 task4 的 `align/report.json`（16 行） | 脚本 `KeyError` 崩在 measure 阶段 —— **崩掉比返回 0 好**：崩掉是显式的，0 是安静的 |

**结论写进常设流程**：读任何未读过的 JSON，第一步 `print(list(obj.keys()))`；键名猜错的后果不是报错，而是**一个看起来干净的 0**。

## 10. 取证附录（命令 + exit code）

| # | 命令（在 `arknights_bot-measure-align`） | 结果 | exit |
|---|---|---|---|
| 1 | `git log --oneline -1` | `4584759 merge(gg): 并入 feat/gg-render-depot …` | 0 |
| 2 | `git diff --stat` | 无输出（未改任何被跟踪文件） | 0 |
| 3 | `map_elements.py controls` | 4/4 通过 | 0 |
| 4 | 闸门牙齿 1：移走 `controls.json` → `map_elements.py measure` | `REFUSED`；`ls result.json` 不存在 | **2** |
| 5 | 恢复 `controls.json` → `json.load` 读回 | `controls_pass=True` | 0 |
| 6 | `map_elements.py measure` | `result.json` 落盘；stdout=0B / stderr=0B | 0 |
| 7 | `premise_survey.py controls` | 2/2 通过 | 0 |
| 8 | 闸门牙齿 2：移走 `controls.json` → `premise_survey.py measure` | `REFUSED`；`premise_survey.json` 不存在 | **2** |
| 9 | `premise_survey.py measure` | 16 行全场景普查；stdout=0B / stderr=0B | 0 |
| 10 | `premise_survey.py ceilings` | 16 行分区域天花板；stdout=0B / stderr=0B | 0 |
| 11 | 写 JSON 后 `json.load` 读回 | 全部成功 | 0 |
| 12 | 禁碰树 `arknights_bot-satori-yoga-skia-go` | 未读、未渲染、未落产物 | — |

**两处代码缺陷的自查**（本轮踩到并修掉）：① `modal_color` 用 `reshape(-1,3)` 处理 4 通道数组 → `ValueError`，显式崩掉；② `current_sim` 取自只有 14 行的 task5 报告 → `KeyError: 'calendar'`，显式崩掉。**两次都是崩掉而不是返回 0** —— 这正是「尺子失效时输出仍可能长得像通过」的反面教材：显式失败是安全的。

## 11. 出身逐项标注

| 项 | 出身 |
|---|---|
| 4 行的 ink_fraction / 众数色 / 声明色命中 / 几何候选 | **measured**（`result.json`，判据 v2 先落盘） |
| 5 洞全是域外、真目标 0 | **measured + 穷举举证**（每行给出可复核量与两个方向的影响） |
| 16 场景前提普查（含 `state` 有 4 个命中且 R1R2 通过却已过线） | **measured**（`premise_survey.json`） |
| `base` 结论保留、理由改写为「前提不成立 / 检测器沉默」 | **measured**（各项量）+ Boss 裁定 |
| 「0.20 必须订正、勿再引用」 | **measured**（本轮映射把 5 个洞全部判为非漏画） |
| 分区域色偏天花板（最高 0.98572，`operator` 收益 +0.14968） | **measured**（`ceiling_tiles.json`）；**限制**：576 参数/场景的拟合上界，非可达目标 |
| 「逐区域错色的期望收益 ≠ 0，也 ≠ 未知，而是已测得不足」 | **measured** |
| `onerror` 检出 0 | **measured（可检的 2 行）**；enemy/operator 标 **未测**（探测尺寸超画布） |
| `enemy` 主洞 = hit0（占其残差 94.4%） | **measured** |
| 「Lottery 排最后」的两条独立理由 | 理由 1 **measured**（裕度 0.0341）；理由 2（legacy 抓取非确定 12 次 6 个 sha）为**外部既有事实**，且**不影响 Go 侧门禁测量** —— 门禁比的是 gg 输出 vs 磁盘冻结基线，与 legacy 抓取链路无关。**两条分开写，未因此把 lottery 踢出排期。** |
| 「域外桶的行先回报、不往下走」 | 已遵守：本轮不给任何实现方向 |
