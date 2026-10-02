# Task 4 — alignment 配准天花板表

```
provenance:   measured（本任务全部数字为实测；逐项标注见 §6）
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align
              生成时间 2026-10-02T11:36:55Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task4-alignment-ceiling.md
产物:          tmp/pixel-compare/{report.json,<scene>/{old,new,diff}.png}   （go harness，@4584759）
              tmp/align/{report.json,instrument.json,controls.json,run.log,<scene>/reg-diff.png}
              tmp/align-coarse/{report.json,...}                              （补充宽扫）
仪器脚本:      .audit/align_ceiling.py（未提交；工作树除该未跟踪文件外 0 改动）
```

**红线遵守声明**：本任务算出的最佳位移 (dx,dy) **只出现在本报告的测量栏**。未写入任何渲染常量、geometry 表、模板或 CSS 值；`git diff --stat` 输出为空，工作树除未跟踪的分析脚本外无任何改动（§5 命令与 exit code）。用基线回答「这个渲染器当初实际做了什么」= 合法；用它决定「设计应该是什么样」= 红线，本报告不做后者。

---

## 0. 一句话结论

**14 个未过线场景，无一例外，配准后的最优相似度仍够不到 0.99。** 配准最多解释掉 **6.05%** 的误差（`recruit`），其中 8 个场景的最优位移就是 **(0,0)**（配准什么也解释不掉）。误差是**弥散**的（差异像素占全画布 30–100%，误差最高的前 1% 像素只占 2.78%–20.0%），不是「整体没对齐」。

## 1. 天花板表（未过线场景 = manifest 16 − 过线 2 = 14，逐行机械枚举）

「当前 sim」来源：`tmp/pixel-compare/report.json` + `go test` stdout，本工作树 @ `4584759` 实跑（**不是**引用别人树里的未跟踪产物）。
「配准后最优 sim」= 零填充口径在 `[-24,24]²` 上的穷举最优（见 §3 仪器定义）。
「保守上限」= `max(零填充最优, band-free 探针)`——**只取大的那一侧**，防止本表低估、误杀可行方法。

| 场景 id | 当前 sim | 配准后最优 sim | 提升量 | 保守上限 | 距 0.99 还差 | 判定 |
|---|---|---|---|---|---|---|
| base | 0.97073 | 0.97073 @ (0,0) | +0.00000 | 0.97073 | 0.01927 | 此路不通 |
| box | 0.82504 | 0.82640 @ (0,5) | +0.00136 | 0.82640 | 0.16360 | 此路不通 |
| box-detail | 0.87844 | 0.87844 @ (0,0) | +0.00000 | 0.87844 | 0.11156 | 此路不通 |
| box-summary | 0.88098 | 0.88098 @ (0,0) | +0.00000 | 0.88098 | 0.10902 | 此路不通 |
| card | 0.96015 | 0.96015 @ (0,0) | +0.00000 | 0.96015 | 0.02985 | 此路不通 |
| depot | 0.92352 | 0.92472 @ (2,0) | +0.00119 | 0.92533 | 0.06467 | 此路不通 |
| enemy | 0.91951 | 0.91951 @ (0,0) | +0.00000 | 0.91951 | 0.07049 | 此路不通 |
| gacha | 0.90251 | 0.90251 @ (0,0) | +0.00000 | 0.90251 | 0.08749 | 此路不通 |
| headhunt | 0.97021 | 0.97021 @ (0,0) | +0.00000 | 0.97021 | 0.01979 | 此路不通 |
| help | 0.83943 | 0.83943 @ (0,0) | +0.00000 | 0.83943 | 0.15057 | 此路不通 |
| lottery | 0.98358 | 0.98358 @ (0,0) | +0.00000 | 0.98423 | 0.00577 | 此路不通 |
| missing | 0.87614 | 0.87697 @ (2,4) | +0.00083 | 0.88213 | 0.10787 | 此路不通 |
| operator | 0.76866 | 0.76866 @ (0,0) | +0.00000 | 0.76866 | 0.22134 | 此路不通 |
| recruit | 0.88603 | 0.89293 @ (-9,0) | +0.00689 | 0.90104 | 0.08896 | 此路不通 |

**「此路不通」14 行 / 「够得到」0 行**（校验式，非手数）：

```
manifest entries           = 16        （src/ggrender/testdata/visual/baseline/manifest.json）
本表行数                    = 16
id 集合相等（manifest vs 行）= True
过线数(go passed=true)      = 2         （calendar 0.99432, state 0.99093）
未过线数                    = 14
校验式: 2 + 14 == 16        = True
未过线且保守上限 ≥ 0.99     = 0  行
未过线且保守上限 < 0.99     = 14 行
最优解落在搜索范围边界       = 0  行（[-24,24] 与补充粗扫 [-96,96] 都是 0 行）
```

过线的 2 个场景同样被扫过，配准最优解都是 (0,0)、分数不变 —— 说明配准没有偷偷「解释掉」任何已达标场景的误差（0.99432 / 0.99093 原样）。

## 2. 阳性对照结果 —— **通过**

| 对照 | 内容 | 结果 |
|---|---|---|
| C0 尺子自检 | 16 个场景 `S(0,0)`（本脚本 numpy）vs go `similarityNormalized` | **pass**，max_delta = **0.000e+00**（逐位相同，非近似） |
| C1 真实数据 | 把 `base` / `lottery` / `enemy` 的 `new.png` 内容按已知位移边缘复制平移，再用同一搜索找回 | **15/15 exact**，例：`known=(2,22) → recovered=(-2,-22)`；`known=(-23,-22) → recovered=(23,22)` |
| C2 合成噪声 | 220×300×4 随机噪声图按已知位移平移后找回 | **2/2 exact**：`known=(11,11) → (-11,-11)`；`known=(-11,7) → (11,-7)` |
| 阴性对照（对照本身有牙齿） | `--inject-fault invert`（故意把搜索结果取反） | **controls pass=False**，7 用例里 6 个判错 → 阳性对照能识别坏掉的搜索，不是恒通过 |

C0 的 0.000e+00 是本报告最关键的一条：Python 侧与 Go 侧是**同一把尺子**，否则下面的差值全是两把尺子的差。

**对照设计上的一个坑（已修，记下来）**：初版对照用的已知位移（7,-3）/(-11,5)/(2,13)/(-5,-9) 超出了当时的搜索范围 R=6，对照报 **False**。这不是搜索坏了，是**对照测的是范围不是搜索**。改为真值严格落在范围内部（|shift| ≤ R-1）后才通过。教训：阳性对照失败时先查对照本身，别急着判搜索坏 —— 但也不能因为「结论碰巧对」就放过，本次是它失败后才动的手。

## 3. 搜索范围与仪器定义（完整落盘）

脚本 `.audit/align_ceiling.py`，同内容另存 `tmp/align/instrument.json`：

| 项 | 定义 |
|---|---|
| 度量 | `S(dx,dy) = 1 - Σ_{x,y,c} \|O[x,y,c] - Nshift[y,x,c]\| / (W*H*4*255)`；**4 通道分母**、逐像素 abs 差，与 go `similarityNormalized` 同式；`S(0,0)` 已验证逐位等于 go 输出 |
| 位移约定 | `Nshift[y,x,c] = N[y-dy, x-dx, c]`，有效域 `[max(0,dy), min(H,H+dy)) × [max(0,dx), min(W,W+dx))`，**域外按 0 计**（零填充）。故报告里的 (dx,dy) 读作「把 new 的内容往 (dx,dy) 方向挪，能对齐 old」 |
| 搜索 | 整数 (dx,dy) ∈ `[-24,24]²`，**步长 1**，全画布，穷举 2401 点；不降采样、不做金字塔、不做相关/互信息 |
| 边界规则 | argmin 落在 `\|dx\|=24` 或 `\|dy\|=24` 上 → 自动以 `R'=48` 重搜；仍在边界则该行 `boundary=true`，**不进「此路不通」判定，只标「需扩大搜索范围」**。实测：0 行触发 |
| 补充宽扫 | 独立一轮 `[-96,96]²` 步长 4（2401 点），用来证明 ±24 的边界不是人为限制。实测：0 行触发边界，且没有任何场景在 ±96 内取得超过 §1 的分数（见 §4） |
| tie-break | `(err, \|dx\|+\|dy\|, \|dx\|, \|dy\|, dx, dy)` 字典序取最小，保证确定性 |
| 门禁 | 只**复用** go 的 0.99 判据做「够不够得到」。`pixel_test.go` 的 `similarityNormalized` / `gatePassed` / `zeroDiffBBox` **一字未改** |
| band-free 探针 | 同一遍扫描里额外追 `err_noband = err_zero - band_sum(O)`（用积分图像 O(1) 取被零填充带带进来的 O 自身质量），用于检验零填充带有没有在驱动排名。**它不是天花板口径**：按构造，去掉带惩罚后「把内容移出画布」不花钱，它的 argmax 会单调跑到范围边缘（实测 `depot` 在 ±96/step4 下冲到 **0.98688**，纯属仪器假象）。因此只在 argmax **不贴边**时才被采纳进「保守上限」列 |

## 4. 逐场景判定 + 举证方式

全部 14 行的判定都是「此路不通」。举证方式对每一行相同，逐条列在表后：

**举证方式（每行同一套，共三步）**
1. **穷举搜索**：在该场景全画布上穷举 `[-24,24]²` 全部 2401 个整数位移，取最小误差；该最小误差对应的 sim 仍 < 0.99。这不是抽样。
2. **宽范围佐证**：独立一轮 `[-96,96]²` 步长 4，2401 点，覆盖 4 倍范围，分数**未超过**第 1 步（`recruit` 0.89293@step1 vs 0.89256@step4、`box` 0.82640 vs 0.82623，差值只来自 4px 粒度，方向都是 step1 更高）。
3. **保守上限不抬结论**：即便取「零填充最优」与「band-free 探针（不贴边时）」中**较大**的一个，14 行仍全部 < 0.99，最小缺口 0.00577（`lottery`）。

**这三条证明的是什么、不证明什么**
- 证明：在「整图刚性整数平移」这一族配准下，**没有**任何未过线场景能到 0.99。误差不是整体错位造成的。
- **不**证明：旋转、缩放、非刚性形变、分块/分区域配准能到 0.99。本表只覆盖整数平移，标题即限定。
- **不**证明：这些场景没有别的可修之处。**「配准此路不通」≠「场景不值得投工」** —— 例如 §5 的残差分布显示误差弥散、差异像素占比 30–100%，那指向的是别的病因（元素缺失/字号缩放/颜色系统偏差），需要另外的诊断去定。

**残差分布（配准最优位下）** —— 与 AGENTS.md 的「误差集中度检查」同族：

| 场景 | 配准解释掉的误差 | 差异像素占比 | 前 1% 像素占误差 | 最优位移 |
|---|---|---|---|---|
| base | 0.00% | 30.96% | 6.69% | (0,0) |
| box | 0.78% | 90.91% | 3.64% | (0,5) |
| box-detail | 0.00% | 99.94% | 5.29% | (0,0) |
| box-summary | 0.00% | 100.00% | 5.59% | (0,0) |
| card | 0.00% | 73.03% | 11.33% | (0,0) |
| depot | 1.56% | 42.22% | 3.38% | (2,0) |
| enemy | 0.00% | 99.99% | 7.79% | (0,0) |
| gacha | 0.00% | 99.99% | 6.51% | (0,0) |
| headhunt | 0.00% | 77.08% | 15.07% | (0,0) |
| help | 0.00% | 99.93% | 3.80% | (0,0) |
| lottery | 0.00% | 49.54% | 14.81% | (0,0) |
| missing | 0.67% | 70.69% | 3.93% | (2,4) |
| operator | 0.00% | 89.87% | 2.78% | (0,0) |
| recruit | **6.05%** | 53.43% | 4.96% | (-9,0) |

无单点热区（前 1% 像素只占 2.78%–20.0% 误差），差异像素铺满全画布 —— 与「没有整块没画、也没有单一错位源」一致。

## 5. 取证附录（命令 + exit code）

| # | 命令（在 `C:/WorkSpace/Golang/arknights_bot-measure-align`） | 结果 | exit |
|---|---|---|---|
| 1 | `git worktree add -b measure/align-ceiling C:/WorkSpace/Golang/arknights_bot-measure-align consolidate/gg-mainline` | HEAD now at 4584759 | 0 |
| 2 | `git log --oneline -1 && git status --porcelain` | 4584759 / 无输出 | 0 |
| 3 | `cd src && go test ./ggrender/ -run TestGGPixelParity -count=1` | 16 场景渲染完成，FAIL（14 未过线，**这是预期的诚实红灯**） | 1 |
| 4 | `python -B .audit/align_ceiling.py --range 24 --control-pairs 3 --out tmp/align` | 16/16 行，C0+C1+C2 全过，elapsed 1205.6s | 0 |
| 5 | `python -B .audit/align_ceiling.py --range 8 --scenes box-detail --controls-only --control-pairs 1 --inject-fault invert --out tmp/align-teeth` | 对照 **pass=False**，7 个用例里 6 个判错（证明对照有牙齿） | 0 |
| 6 | `python -B .audit/align_ceiling.py --range 96 --step 4 --skip-controls --out tmp/align-coarse` | 宽扫 16/16 行，0 行边界 | 0 |
| 7 | `git merge-base --is-ancestor 8cbffef HEAD` | `8cbffef`（skia 基线偷读）**不在**基底历史 | **1** |
| 8 | `git merge-base --is-ancestor 641988f HEAD` | **在**（card 假达标作弊提交） | 0 |
| 9 | `git merge-base --is-ancestor 8d6b325 HEAD` | **也在** —— 即 `641988f` 的 revert 同样在基底历史，净效果已抵消 | 0 |
| 10 | `git merge-base --is-ancestor ec440bf HEAD` | 在 ⇒ 零面积 bbox 守卫（`gatePassed` / `zeroDiffBBox`）在位 | 0 |
| 11 | `grep -c "func gatePassed" src/ggrender/pixel_test.go` | 1 | 0 |
| 12 | `grep -rn "C:/WorkSpace" src/ggrender/pixel_test.go src/ggrender/helpers.go` | **0 字节输出**，无跨工作树引用 | **1** |
| 13 | `grep -n "skia\|8cbffef\|baseline/cache" src/ggrender/pixel_test.go` | **0 字节输出**，harness 不含 skia 线任何内容 | **1** |
| 14 | `ls -d src/skia` | 不存在（本基底无 skia 包，结构上不可能偷读基线） | 2 |
| 15 | `git diff --stat` | 无输出（**未改任何被跟踪文件**） | 0 |
| 16 | `git status --porcelain` | 仅 `?? .audit/align_ceiling.py`（未跟踪的分析脚本） | 0 |
| 17 | `git worktree list \| wc -l` | 23（原 22 + 本专用树；无既有树被占用） | 0 |
| 18 | 比对 `tmp/pixel-compare/report.json` 的 `hashNew`（本树）vs `arknights_bot-consolidate` 里那批未跟踪产物 | 16/16 **完全相同** → 渲染确定性可复现（**仅作旁证**；本表数字全部来自本树自己那一份，未引用那批产物） | 0 |

**必须被记录的基底历史事实**：第 8/9 行合起来说明 —— 基底历史里**含** `641988f`（已知作弊提交），但**同时含它的 revert `8d6b325`**，净效果抵消。不写这一条，下一个人查 `641988f` 在不在基底上会得到 exit=0，然后误判基底被污染。

**禁碰树**：`arknights_bot-satori-yoga-skia-go` 全程**未读代码、未渲染、未产生任何文件**（命令 7/13 是在本树跑的 grep 与 git，不是对该树的写操作）。天花板表的每一行都由 **gg 自己的 harness**（`src/ggrender/pixel_test.go` → `RenderGGContext`）产出，行内 `produced_by` 字段记 `src/ggrender/pixel_test.go @ 4584759`。

## 6. 出身逐项标注

| 项 | 出身 |
|---|---|
| 16 个场景的 id 清单 | **mechanically enumerated** from `manifest.json` `entries`（`asserted` 的部分：零 —— 无一项从 AGENTS.md 散文抄录） |
| 各场景 `current_sim` / `passed` | **measured**（本树 @4584759 的 go 输出，`tmp/pixel-compare/report.json`） |
| `best_sim` / `best_shift` / 提升量 / 残差统计 | **measured**（`tmp/align/report.json`） |
| 「此路不通」14 行判定 | **measured + 穷举举证**（§4 三步），非「我没看见」 |
| 天花板表的适用边界（仅整数刚性平移） | **asserted**（由 §3 的搜索空间定义决定，非测得） |
| `calendar` / `state` 已过线 | **measured**（go `passed=true`），但它们**不**进「此路不通」统计（校验式用 2 + 14 == 16） |
| 「误差弥散指向别的病因」 | **inferred**（由残差分布推出，非本任务测量对象；未据此下任何实现结论） |
| 「instrument 一栏的 `probe` 不可作天花板」 | **measured**（`depot` 在 ±96/step4 下探针冲到 0.98688，是仪器假象的直接证据） |

**与 AGENTS.md 已有分数表的关系**：本表**不使用** `@1b433bc` 或 `@ec440bf` 任一版分数。`current_sim` 全部取自本树 @ `4584759` 实跑（例：`card` 此处 0.96015，与 AGENTS.md 记的 0.96020 不同 —— 不是谁错，是不同 commit 上的数，正是「引用任何数字都要带 commit」那条规则要防的事）。
