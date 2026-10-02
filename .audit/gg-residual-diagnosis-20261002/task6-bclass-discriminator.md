# Task 6 — B 类 5 行的机械判别

```
provenance:   measured
measured_at:  branch measure/align-ceiling @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47
              worktree C:/WorkSpace/Golang/arknights_bot-measure-align（复用，未新建）
              生成时间 2026-10-02T12:15:51Z
original_path: C:/WorkSpace/Golang/.lead2-reports/task6-bclass-discriminator.md
判别口径:      tmp/residual/discriminator/instrument.json（跑数前落盘，判据不在脚本里硬编码）
脚本:          tmp/residual/discriminator/discriminate.py（两 stage，P1 结构性隔离）
产物:          tmp/residual/discriminator/{controls.json, result.json, controls.log,
               measure_stdout.log(0 字节), measure_stderr.log(0 字节), gate_test.log}
输入:          tmp/residual/report.json (task5) + tmp/pixel-compare/<scene>/{old,new}.png @ 4584759
produced_by:   src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47（逐行记录在 result.json）
```

**修改权声明**：未改任何渲染代码、未改计分代码、未 commit、未 push、未新建工作树。`git diff --stat` 空，工作树仅 task4 遗留的未跟踪 `.audit/align_ceiling.py`。禁碰树未读未渲。

---

## 0. Boss 点名的三个数

| # | 项 | 值 |
|---|---|---|
| 1 | **判定前 B = 5** | `base, box-detail, enemy, lottery, operator` |
| 2 | **判定后剩 4** | `box-detail, enemy, lottery, operator` |
| 3 | **假阳性率（flag-level）= 0.20** | FP = 1/5，唯一 FP 是 **`base`** |

**剔除依据（当场写明）**：`base` 的最大内洞只有 **25.70%** 的面积两侧不同，未过 R1（≥50% 面积）判据 —— 该洞内四分之三的区域在两侧是一致的纯色，差异只是零星渗入，其残差占比 8.20% 虽过 R2 但**分量不足以外「整块没画」论处**。它的 4 个命中全部未过 R1（25.70% / 19.76% / 12.35% / 15.00%），不是单洞偶然。**故 `base` 判为检测规则产物，当场剔除。**

**⇒ B 类本轮期望收益按 4 算，不按 5 算。** 与「能算出的上限否决一整类工时」同构。

## 1. 逐行机械判别（5 行一次跑完）

判据（跑数前落盘于 `instrument.json`）：区域取 task5 的 `B_info.largest_hole.pixel_box`；`diff_px = Σ_c|O_c−N_c|`（4 通道同 `pixel_test.go` 口径）；`frac_diff` = 洞内差异像素占比；`share` = 洞内 Σ|Δ| 占全画布残差之比。
**判定 = R1(frac_diff ≥ 0.50) AND R2(share ≥ 0.05)**，两个必要条件缺一不可。

| 场景 | frac_diff | share | R1 | R2 | 判定 | Boss 的洞心单像素规则 | 是否分歧 |
|---|---|---|---|---|---|---|---|
| base | 0.2570 | 0.0820 | ✗ | ✓ | **假阳性 → 剔除** | 判「非真」（两侧逐位相同） | 否 |
| box-detail | 0.9993 | 0.2025 | ✓ | ✓ | **真目标** | 判「真」 | 否 |
| enemy | 0.9999 | 0.4492 | ✓ | ✓ | **真目标** | 判「真」 | 否 |
| lottery | 0.5341 | 0.3185 | ✓ | ✓ | **真目标** | 判「真」 | 否 |
| operator | 0.8511 | 0.2037 | ✓ | ✓ | **真目标** | 判「真」 | 否 |

**稳健性**：两种口径（预登记的面积/分量双条件 vs Boss 给定的洞心单像素规则）**5 行全部一致，零分歧**。也就是说 `base` 的剔除**不是阈值选出来的** —— 它在「洞心逐位相同」和「面积不足一半」两条互不相关的口径下都判假。

**唯一对阈值敏感的行是 `lottery`**：`frac_diff = 0.5341`，距 R1 的 0.50 只有 **0.0341** 裕度。若把 R1 抬到 0.55，lottery 会翻成假阳性，而 Boss 的洞心规则仍判它真 —— 即该行的存亡取决于我预登记的 0.50 这个数。这是本轮唯一一条「换个常数就变结论」的行，单独标出来。其它 4 行裕度充足（base 差 0.2430 判负、box-detail/enemy 的 frac 均 > 0.999）。

### 逐命中明细（不只最大洞）

| 场景 | 命中 | frac_diff | share | 判定 |
|---|---|---|---|---|
| base | hit0 / hit1 / hit2 / hit3 | 0.2570 / 0.1976 / 0.1235 / 0.1500 | 0.0820 / 0.0607 / 0.0340 / 0.0463 | 全 ✗ |
| box-detail | hit0–hit3 | 0.9993 / 0.9993 / 0.9998 / 0.9989 | 0.2025 / 0.1248 / 0.1267 / 0.2032 | 全 ✓ |
| enemy | hit0 / hit1 | 0.9999 / 1.0000 | 0.4492 / 0.0267 | ✓ / ✗（R2 不过） |
| lottery | hit0 | 0.5341 | 0.3185 | ✓ |
| operator | hit0 | 0.8511 | 0.2037 | ✓ |

`enemy` 的第二个命中虽然面积上 100% 两侧不同，但只占全画布残差的 2.67%，未过 R2 ⇒ 不单独构成目标；它很可能是 hit0 那块大洞的附属小块。**注意：场景级存活与命中级存活不是一回事**，B_after=4 是场景级计数。

## 2. 判别器自身的对照（P1：先验收，后测量）

| 对照 | 输入 | 期望 | 实测 |
|---|---|---|---|
| POS 两侧不同 | `box` 洞内两侧刷**不同**颜色 | 判真 | **true** |
| NEG 两侧逐位相同 | 同上刷**完全相同**颜色 | 判假 | **false** |
| NEG2 无洞 | 未改动的 `box` | 无区域，返回 `no_region` | 符合 |

三条全过，**判别器不是恒真也不是恒假**。

## 3. 常设流程 P1（本轮起执行，已落进 `instrument.json`）

> **验收（对照 + 阳性/阴性）先跑；通过之后才允许跑分类测量；且分类输出在对照通过之前不许被任何人或任何进程读取。**

**本轮的结构性执行证据**（不是承诺，是可复核的产物）：

| 要求 | 实现 | 证据 |
|---|---|---|
| 验收先跑 | 脚本分 `--stage controls` / `--stage measure`；measure 先 `json.load` 读 `controls.json`，`controls_pass != true` 则 `SystemExit(2)` 拒绝执行 | 我**实测过闸门有牙齿**：把 `controls.json` 移走后再跑 measure → 输出 `REFUSED: controls.json 不存在`，**exit=2**，未产生任何结果文件；随后恢复 `controls.json` 并读回 `controls_pass=True` |
| 分类输出不泄漏 | measure 阶段不向 stdout/stderr 写任何一行分类结果，只写 `result.json` | 实测 `measure_stdout.log` = **0 字节**、`measure_stderr.log` = **0 字节** |

**为什么这条比表本身值钱**：上一轮 task5 的 v4–v6 修订发生在 14 行分类已打到 stderr 之后，于是「先看数、再回头改对照」在结构上成为可能 —— 我当时靠自觉只改对照构造、没改分类规则，而自觉不是机制。P1 把那条路径从结构上堵死：**对照不通过，分类结果在物理上不存在**。

## 4. 措辞收窄（按 Boss 裁定）

- ✅ 本轮可写：「**分类器在数据前冻结**」。
- ❌ 本轮**不可**写：「全流程预登记」。task5 的 v4–v6 发生在 14 行分类已打到 stderr 之后，**该保证对整轮不再成立，只对分类器成立**。
- ⚠️ task5 报告里的原披露（污染披露一节）**保留，不删** —— 删了就成了掩盖。

**本轮（task6）的对应事实**：判别口径 `R1=0.50 / R2=0.05` 在跑任何数之前落盘于 `instrument.json`；判别器自身的 3 条对照先跑并通过（§2），之后才允许 measure 阶段存在。**因此 task6 的判别器本身满足「分类器在数据前冻结」+「验收先于测量」，不受 task5 那条污染影响。**

## 5. A 类的引用（三条必须同时写，缺一即为误导）

引用「A=0」时必须同时出现以下三条：

1. **天花板只覆盖全局常数修正**（每通道一个偏移）。**分区域/分元素偏移不在内** ⇒ 「不可行动」仅限这一种修法，不排除「分区各修各的」这条路。
2. **A=0 指「无可行动的系统性成因」，不是「不存在系统性成分」**。task5 实测：10/14 场景逐通道有符号中位数 ≥ 2.55（`help` −44/−46/−47、`enemy` +23/+20/+17、`box-summary` +19/+19/+19、`depot` +12/+32/+38）。
3. **「中位数 −44/−46/−47 却仍判 A=0」这个反差本身就是结论**：该色偏被更强的因素盖过，单独修它的上限只有 **+0.04536**，修正后离门禁仍差 **0.00490**（`lottery` 为最小缺口）。不写这条，读者会以为二者互相矛盾。

## 6. 排序规则（Boss 新增）

- **类序：B 先于 C。**
- **B 类内部按「机械判别后是否仍成立」排序，不许按集中度排序。**
  理由两条，都已被本轮数据证实：`base` 演示了**集中度高 ≠ 真**（它 share 8.20% 过 R2，却因面积只差 25.70% 被剔）；`enemy` 的 44.92% 集中在单洞确实是 B 类最强的单点信号，但**在机械判别跑完之前不许据此排第一** —— 现在判别跑完了，它以 `frac_diff=0.9999 / share=0.4492` 双条件通过，**此时**才可以据其集中度排序。

## 7. 取证附录（命令 + exit code）

| # | 命令（在 `arknights_bot-measure-align`） | 结果 | exit |
|---|---|---|---|
| 1 | `git log --oneline -1` | `4584759 merge(gg): 并入 feat/gg-render-depot …` | 0 |
| 2 | `git status --porcelain` | 仅 `?? .audit/align_ceiling.py` | 0 |
| 3 | `git diff --stat` | 无输出（未改任何被跟踪文件） | 0 |
| 4 | `python -B tmp/residual/discriminator/discriminate.py controls` | 3/3 对照通过，`controls_pass=true` | 0 |
| 5 | 闸门牙齿测试：`mv controls.json controls.json.bak` 后跑 `--stage measure` | `REFUSED: controls.json 不存在`，未产出结果 | **2** |
| 6 | 恢复 `controls.json` 并 `json.load` 读回 | `controls_pass=True` | 0 |
| 7 | `python -B tmp/residual/discriminator/discriminate.py measure` | 产出 `result.json`；stdout/stderr 均 **0 字节** | 0 |
| 8 | `wc -c` 于 measure 的 stdout / stderr 日志 | 0 / 0 字节 | 0 |
| 9 | 校验式：B_flagged_before=5、B_after=4、survivors 4 个、false_positives 1 个、5−1=4 | 全部相符（脚本写出即校验） | 0 |
| 10 | 禁碰树 `arknights_bot-satori-yoga-skia-go` | 未读、未渲染、未落产物（本任务全部命令的工作目录都是 measure-align） | — |

## 8. 出身逐项标注

| 项 | 出身 |
|---|---|
| B_flagged_before=5 | **measured**（task5 `report.json` @ 4584759） |
| frac_diff / share / R1 / R2 / 判定 | **measured**（`result.json`，判据先落盘） |
| 假阳性 `base` 与剔除依据 | **measured**（4 个命中全部 frac < 0.50） |
| flag-level 假阳性率 0.20 | **measured**（FP/B_flagged，口径定义在 `instrument.json`） |
| 场景级 FPR | **无法测量**，未举证 —— 未标记的场景没有 ground truth 可比 |
| `lottery` 是唯一阈值敏感行（裕度 0.0341） | **measured**（frac 0.5341 vs R1 0.50） |
| 「B 类期望收益按 4 算」 | **inferred**（由存活数直接推得，非测量） |
| 「集中度高 ≠ 真」 | **measured**（`base` share 8.20% 过 R2 仍被剔） |
| §5 的 A 类三条限定 | **measured**（数值来自 task5 §4）+ **asserted**（引用规则） |
