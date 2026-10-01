# 验收证据 — fix/gate-hardening 门禁加固

本目录是 `fix/gate-hardening` 的**验收证据落盘**，对应 Boss 规范「改评分门禁必须提交验收证据」。
除 `README.md` 与 `pristine.sha` 外全部为**原始命令输出**，未经手工编辑。

| 文件 | 是什么 |
|---|---|
| `before.scenes` | 16 场景分数行，**改动前**（HEAD=`0ea71c6` 干净树）。原始抽取，**带** `pixel_test.go:311:` 行号前缀 |
| `after_task2.txt` | 改动后整份 `-v` 输出（`aa88012` 代码状态）。Task ② 落地后的最终一次 |
| `after_task3.txt` | Task ③ 落地后整份 `-v` 输出 |
| `after_task1.txt` | **被回退的钉常量方案**那一次运行（保留作为该方案的实测记录） |
| `mutant.scenes` | 变异后（0.99→0.5）16 场景分数行，`passed` 大面积翻转 |
| `pixel_test.go.pristine` | 变异前的对照副本（`aa88012:src/ggrender/pixel_test.go`，工作树 CRLF 形态） |
| `pristine.sha` | 上述副本的 sha256 + 对应 git blob，含恢复正确性判据 |
| `vr_main.go` | 复现 `treeSHA256()` 的独立可执行程序（见下） |
| `rejected-anchor-approach.patch` | **被否决方案原文**（见下） |

## 1. Task ①：manifest 树哈希判定「不可实施」

理由三条 + 一条固定附加，**第一条是决定性理由**；完整表述见
`src/ggrender/pixel_test.go` 顶部注释块「为什么这里没有 manifest 树哈希校验（Task 1 结论：不可实施）」
与审计文档 `_audit_20260929/verifier-review-20260929.md` 文末修订块（两处同一句原文）。

1. **决定性理由：量错了对象。** 根是 `template/` 与 `assets/`，**不覆盖 `baseline/images/*.jpg`**。
   即便完美可复现也检测不到 #6。
2. **覆盖范围两树对不齐。** `assets/` gg 102 vs satori 100。
3. **值不可复现。** 473 commit 穷举 0 命中（`tOK=473 aOK=473`、`templateTree` 69 种、`assetTree` 21 种、MATCH=0）。
   正控制：两树 `template/` 哈希**完全相同**（`1fccd181f7b38565ad4a2eda36a96c031ac9bd47c8554b346b6fde5ff0680e7c`），
   证明算法复现无误 —— 否则「0 命中」会被误读成「算法写错了」。

`vr_main.go` 的 `treeSHA256()` / `fileSHA256()` 逐行抄自 `4bda363:src/cmd/visual-regression/main.go`
（commit `8873add` 引入），只额外加了 `main()`：

```
go run .audit/vr_main.go <repoRoot>/template <repoRoot>/assets
```

正控制与反例都在该程序的实际输出里（satori 4bda363 与 gg 0ea71c6 的 `template/` 哈希相同；
`assets/` 两侧各 102/100 个文件，哈希互不相同且都对不上 manifest）。

## 2. `rejected-anchor-approach.patch` 的定位（**最容易读错的一处**）

> 该 patch 解决的是**另一个问题** —— 16 张 baseline images 的独立 sha256 常量锚，**与树哈希无关**。
> 它**不需要**树哈希可复现性，且实测能真正拦下缺陷 #6。
> 被否决的理由是「多一处需手工同步 + 不是更强的信任锚」，**不是**「它复现不了」。
> 理由若写错，将来复活它的人会以为「只要解决复现性就能用上」。

**「我们有一个实测有效但因工程理由否决的方案」和「我们没有方案」是两回事。** 前者让后来者知道
路存在，值不值得走由他们判断；后者会让人重新发明一遍。

该 patch 已在收到 Boss 裁定后从 `pixel_test.go` 回退，**未合入**。复活它是 Boss 的权限。
方案自身仍有局限：常量可被同改者一并改掉，它提高的是「改动必须出现在 git diff 里」的可见性与
成本，**不是不可伪造性**。

## 3. Task ②：0.99 阈值负向测试的变异检测

变异方式：`sed -i '177s/if sim < 0\.99 {/if sim < 0.5 {/'` —— 定点第 177 行，即 `gatePassed` 本体。
第 486 行是测试内前置条件断言，**未**被改动（首次用无锚点 `sed` 时误改过两处，已改为定点行号）。

| 阶段 | 证据 | exit code |
|---|---|---|
| 变异后 | `mutant.scenes` —— `TestGGPixelParity_Negative_ThresholdGate` **FAIL**；14 个场景 `passed=false`→`true` | `MUTANT_NEG_EXIT=1` |
| 恢复后 | `after_task2.txt` 全部 **PASS**（`-count=1` 禁缓存） | `RESTORED_NEG_EXIT=0` |
| 恢复正确性 | `pristine.sha` 指纹 `e22a303d…`；`cmp -s` 逐字节一致 | `CMP_EXIT=0` |

复算指纹：

```
sha256sum .audit/pixel_test.go.pristine   # 应为 e22a303d…
git rev-parse aa88012:src/ggrender/pixel_test.go
```

## 4. Task ③：零差异 vs 单像素差异

`after_task2.txt` 末尾 `TestGGPixelParity_Negative_EmptyBBoxGate` 的两行 `t.Logf` 即哨兵生效的实测输出：

- 零差异 → `bbox=[-1 -1 -1 -1]`，报 `zero-diff`
- 单像素差异（差异点取 `(37,41)`，刻意避开旧哨兵撞值位置）→ `bbox=[37 41 37 41]`，报 `degenerate`，
  `sim=0.99992` 仍被拒（fail-closed 方向未动）

## 5. 16 场景分数逐个不变

`before.scenes` 与 `after_task2.txt` **两侧都带** `pixel_test.go:NNN:` 行号前缀（改动前 311、改动后 373）。
复算时**必须两侧都剥**，只剥一侧会得到 16/16 全「变了」的假阳性：

```
sed -E 's/^ *pixel_test\.go:[0-9]+: //' before.scenes > /tmp/b
grep -E "scene [a-z-]+ +[0-9]+x[0-9]+" after_task2.txt | sed -E 's/^ *pixel_test\.go:[0-9]+: //' > /tmp/a
diff /tmp/b /tmp/a; echo "DIFF_EXIT=$?"     # 实测 DIFF_EXIT=0
```

行号是历史追溯用的，不是稳定引用 —— 文档引用代码位置一律用函数名/字段名。

## 6. 已知未修复（**不要读成「已安全」**）

- **缺陷 #6（信任锚自证）未修复**。
- **缺陷 #7（基线作底 + 约 1.33% 画布扰动）未修复**。
- **单像素差异被误杀未修**（`sim=0.99992` 被拒）—— 经 Boss 裁决不修：修它要把零跨度判定
  `<= 0` 放宽为 `< 0`，那是放宽门禁。

详见 `src/ggrender/testdata/visual/pixel-parity-scores.txt` 头部 `change:` 段与审计文档文末修订块。

## 7. 一个取证纪律的正面案例

诊断 CRLF 时，`git show <rev>:<path> | grep -c $'\r$'` 返回 **0**，看起来像「blob 是 LF」。
那个数字是**假的**：MSYS 文本模式在管道里吃掉了 CR。改用 `sha256` 指纹才发现
「blob 存的是 CRLF，每行多一个 `\r`，多出 542 字节」。

`grep -c` 给了一个看起来合理、但测的不是被测对象的数字；`sha256` 不会说谎。
**声明任何结论前，先确认测量真的测到了目标。**

*证据：worker-17 ｜ 2026-09-29 ｜ commit aa88012 + 理由重排 commit*
