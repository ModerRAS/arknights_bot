# `archive/*` tag 类型混合 —— 远端 ref 条数判据的使用前提

## 1. 这份文档解决什么

`archive/` 命名空间下的 tag **类型不统一**：annotated（`git cat-file -t` 返回 `tag`）
与 lightweight（返回 `commit`）并存。

这不是审美问题。2026-10-02 实测踩到过一次：**远端 ref 条数判据是按 annotated 标定的，
对 lightweight 会给出假警报。**

当时用两条判据核验一次 tag push：

1. `git ls-remote --tags origin <name>*` 应当返回**两条**命中（`<ref>` 与 `<ref>^{}`）；
2. `git ls-remote --tags origin | wc -l` 应当**恰好增加 2**。

实际返回**一条**命中、总数**增加 1**。push 本身完全正确——tag 确实到了 origin，
且 `rev-parse ^{commit}` 指向预期的改前提交。差异的根因是**对象类型**：

| tag 类型 | `git cat-file -t` | 远端 ref 条数 |
|---|---|---|
| annotated | `tag` | **2**（tag object + `^{}` peeled ref） |
| lightweight | `commit` | **1**（只有 tag ref 本身） |

## 2. 出处与 provenance

| 项 | 值 |
|---|---|
| `provenance` | `measured-in-repo` |
| `original_path` | `refs/tags/archive/operatorinfo-rarity-ocr-derived-20261002` |
| 为什么不是 `retroactive-migration` | 本目录归档的是**一次当场实测的仓库事实**，没有任何仓库外产生的定义、代码或数值被搬进来。因此不适用「事后追认」标记。 |
| 测量日期 | 2026-10-02 |
| 测量工作树 | `C:/WorkSpace/Golang/arknights_bot-operatorinfo` |
| 测量分支 / 提交 | `fix/operatorinfo-fields` @ `e0154fb` |
| 远端 | `origin` = `ModerRAS/arknights_bot` |
| 本目录不含任何图像 | 冻结基线图一律不入库 |

## 3. 测量配方（仪器定义）

计数全部由下列命令机械生成，未手数：

```bash
ann=0; lw=0; pushd=0
for t in $(git tag --list 'archive/*'); do
  ty=$(git cat-file -t "refs/tags/$t" 2>/dev/null)
  rc=$(git ls-remote --tags origin "refs/tags/$t*" 2>/dev/null | wc -l)
  if [ "$ty" = "tag" ]; then ann=$((ann+1)); else lw=$((lw+1)); fi
  [ "$rc" -gt 0 ] && pushd=$((pushd+1))
done
echo "total=$(git tag --list 'archive/*' | wc -l) annotated=$ann lightweight=$lw on_origin=$pushd"
```

逐条 tag 的类型与远端 ref 条数：

| tag | type | 远端 refs |
|---|---|---|
| `archive/2be9c2c-pre-caveat-message` | `commit` | 1 |
| `archive/card-cheat-641988f-unreverted-on-card-atomic-20261002` | `tag` | 2 |
| `archive/gg-card-atomic-00447f3` | `commit` | — |
| `archive/gg-generality-dfb4e1e` | `tag` | — |
| `archive/gg-headhunt-20260929` | `tag` | — |
| **`archive/operatorinfo-rarity-ocr-derived-20261002`** | **`commit`** | **1** |
| `archive/reclaim-check-upstream-gg-20260929` | `tag` | — |
| `archive/reclaim-fix-card-text-color-20260929` | `tag` | — |
| `archive/reclaim-fix-font-host-independence-20260929` | `tag` | — |
| `archive/reclaim-fix-gate-hardening-20260929` | `tag` | — |
| `archive/reclaim-fix-operator-art-and-portrait-url-20260929` | `tag` | — |
| `archive/reclaim-fix-pixel-test-gofmt-20260929` | `tag` | — |
| `archive/reclaim-gg-rebase-upstream-20260929` | `tag` | — |
| `archive/reclaim-gitattributes-lf-20260929` | `tag` | — |
| `archive/reclaim-harness-bbox-20260929` | `tag` | — |
| `archive/reclaim-lead-2-help-boxes-20260929` | `tag` | — |
| `archive/reclaim-lead-2-neargate-polish-20260929` | `tag` | — |
| `archive/reviewer-detached-0d77efe-20260929` | `commit` | — |
| `archive/satori-salvage-20260929` | `commit` | — |
| `archive/satori-w3-help-boxes-stash-20260929` | `tag` | — |
| `archive/satori-w4-neargate-stash-20260929` | `tag` | — |
| `archive/skia-baseline-ref-codecheat-20260929` | `tag` | 2 |
| `archive/skia-baseline-ref-codecheat-scope-correction-20260930` | `tag` | 2 |
| `archive/skia-cheat-8cbffef-20260929` | `tag` | 2 |

**合计 24；annotated 19、lightweight 5（19+5=24 自洽）；其中 6 条在 origin 上。**

### 3.1 两处与口述不符的实测更正

1. **「`archive/*` 里只有我们这条是 lightweight」不成立。** 实测 lightweight 共 **5** 条，
   我们这条只是其中之一：`2be9c2c-pre-caveat-message`、`gg-card-atomic-00447f3`、
   `reviewer-detached-0d77efe-20260929`、`satori-salvage-20260929` 同样是 lightweight。
   判别式应当是「`archive/*` **类型混合**」，而不是「我们这条特殊」。
2. **wiki 记录 `obs-2026-09-30-skia-tag-origin`（「两个既有 skia 注解 tag 从未推到 origin」）已过时。**
   2026-10-02 实测两条**都在 origin 上**且各带 2 条远端 ref。
   即：它们是在该条记录之后被推上去的。**引用那条 wiki 时必须带这个时间差。**

## 4. 规则（适用范围比 tag 宽得多）

> **任何判据若隐含假设了对象的某个属性（类型 / 数量 / 命名规则 / 存在性），
> 必须先断言那个属性，再套用判据。**

本项目已在同一个方向上栽过两次，形状不同：

| 次序 | 判据假设了什么 | 实测落差 |
|---|---|---|
| 第一次 | 假设远端 ref 集合里**存在**一条 `^{}` 伪 ref | 用精确 pattern 去匹配它，而 `^{}` 是客户端合成的、不在远端真实 ref 集合里，精确匹配必然落空 |
| 第二次（本文件） | 假设对象是 **annotated 类型** | lightweight 只有 1 条远端 ref，两条判据同时给出假警报 |

两次的共同点：**判据假设了一个没被验证的属性**。差别只在于假设的对象是「存在性」还是「类型」。

### 4.1 由此得到的两条操作纪律

1. **先 `git cat-file -t refs/tags/<name>`，再套用 ref 条数判据。**
2. **核验值与预期不符时，任何方向都不构成下一步写操作的授权——无论是多还是少。**
   当时止损条件只写了「若总数增加更多，立刻停下报我」，只覆盖了「多」；
   而本次实测恰好是「少」。按正确的原则应当停下并上报；按字面条件本可以继续。
   **只有完全落在预期内才继续。**

## 5. 为什么 `archive/operatorinfo-rarity-ocr-derived-20261002` 保留 lightweight

它记录的是**改前状态** `d5c019b`（`fix/operatorinfo-fields` 上的处置前提交，
对应 commit `e0154fb` 里的 `Rarity: 6 → 5` 与出处订正）。

把它转成 annotated 的唯一路径是**先删除远端 ref 再重推**，属破坏性远端操作；
而它要携带的说明**已经在别处逐字存在**：`e0154fb` 的提交信息里写着
`Pre-change state is preserved at tag archive/operatorinfo-rarity-ocr-derived-20261002`。
tag 名本身也已自解释（operatorinfo / rarity / ocr-derived / 日期）。

**在「把改前状态与丢数据解耦」这件事上，lightweight 与 annotated 没有区别**——
两者都把 `d5c019b` 钉在 origin 上，差别只是 tagger/日期/说明这段元信息。
因此保留 lightweight，不删任何远端 ref。