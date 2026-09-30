# 六场景 `*.tmpl` 骨架 vs ggrender 实现 · 声明事实对照表

取证时间：2026-09-30。纯只读，未改任何代码、未 commit、未 push。

- 模板侧：`C:\WorkSpace\Golang\arknights_bot\template\`
- 共享样式表：`C:\WorkSpace\Golang\arknights_bot\assets\css\common.css`
- 实现侧：`C:\WorkSpace\Golang\arknights_bot-gg-card-revival\src\ggrender\`

## 判「未声明」的口径

必须 `*.tmpl` 与 `assets/css/common.css` **两侧都查过**才写「未声明」。
`common.css` 全文 16 行，已完整读过，其全部声明为：

```
@font-face { font-family: 'NotoSansHans'; src: url('/assets/font/NotoSansHans-Regular.ttf'); }
body { margin: 0; font-family: 'NotoSansHans', serif; }
img.rarity[data="4"] { width: 34px; }
img.rarity[data="3"] { width: 30px; }
img.rarity[data="2"] { width: 25px; }
img.rarity[data="1"] { width: 20px; }
img.rarity[data="0"] { width: 15px; }
```

即 `common.css` **没有** `color`、`font-size`、`line-height` 任何一条。

---

## 场景定位

| 场景 | 模板 | 模板行数 | 实现函数 | 实现位置 | 有无实现 |
|---|---|---|---|---|---|
| base | `template/Base.tmpl` | 509 | `RenderBase` | `src/ggrender/render.go:609` | 有 |
| gacha | `template/Gacha.tmpl` | 370 | `RenderGacha` | `src/ggrender/render.go:792` | 有 |
| enemy | `template/Enemy.tmpl` | 175 | `RenderEnemy` | `src/ggrender/scene_enemy.go:36` | 有 |
| box-detail | `template/BoxDetail.tmpl` | 61 | `RenderBoxDetail` | `src/ggrender/render.go:256` | 有（已排除复用） |
| help | `template/Help.tmpl` | 149 | `RenderHelp` | `src/ggrender/render.go:892` | 有 |
| depot | `template/Depot.tmpl` | 40 | `RenderDepot` | `src/ggrender/scene_depot.go:36` | 有（最低档） |

「实现在哪个文件」一列填的是 `render.go` 里的真实函数，不是文件名。

## 实现有无判定（六个场景逐一，不跳过）

| 场景 | 判定 | 实现函数 | 文件:行 | 备注 |
|---|---|---|---|---|
| base | **有实现** | `RenderBase` | `src/ggrender/render.go:609` | `scene_base.go` 是空桩 |
| gacha | **有实现** | `RenderGacha` | `src/ggrender/render.go:792` | `scene_gacha.go` 是空桩 |
| enemy | **有实现** | `RenderEnemy` | `src/ggrender/scene_enemy.go:36` | 独立 scene 文件 |
| box-detail | **有实现** | `RenderBoxDetail` | `src/ggrender/render.go:256` | 无 `scene_box_detail.go`；已排除复用（见下节） |
| help | **有实现** | `RenderHelp` | `src/ggrender/render.go:892` | `scene_help.go` 是空桩 |
| depot | **有实现（最低档）** | `RenderDepot` | `src/ggrender/scene_depot.go:36` | 独立 scene 文件；函数自述 `crude layout pending a real parity pass` |

**六个场景无一「未找到实现」。** 五个是完整实现函数，depot 是存在但自述为待重做的最低档实现
（其自述原文见 `scene_depot.go` 头部 `ponytail:` 注释块）。

### depot 的实际情况（更正「depot 属于未实现」的说法）

`scene_depot.go:36 RenderDepot` 确实存在且确实在绘制，不是空桩。其自述为：

```
// ponytail: honest minimal depot renderer (pure GG, no baseline images).
// Ceiling: crude layout pending a real parity pass; manifest depot entry is
// 1275x234 px (850x156 CSS @1.5x device scale).
```

自述天花板之外的可核对事实：底色画成 `dc.SetRGB(0.96, 0.96, 0.95)`（浅色 `#f5f5f2`），
而模板声明为 `background-color: #2e3031;`（深色）；图标位画的是 `DrawRectangle` 纯色块
而非模板的 `.icon { width: 75px; }` 真实图片。这些是声明事实差异，不是印象判断。

### box-detail 复用排查（模板仅 61 行，六个里最小）

按「模板越小越可能是复用通用卡片片段」的怀疑逐项排查，结论：**不复用，是独立布局**。

| 排查项 | 命令 | 结果 |
|---|---|---|
| BoxDetail.tmpl 是否有 include/片段指令 | `grep -nE '\{\{ *(template\|include\|block\|define)' template/BoxDetail.tmpl` | 退出码 1、输出为空 |
| 模板引擎是否有 include 机制 | `grep -rnE 'include\|ParseFiles\|ParseGlob\|template\.New\|Funcs\(' src/utils/media/ --include=*.go` | 退出码 1、输出为空 |
| box_detail 路由加载哪些模板 | `src/core/web/box_detail.go` | `r.LoadHTMLFiles("./template/BoxDetail.tmpl")` + `c.HTML(..., "BoxDetail.tmpl", ...)` |
| box 路由加载哪些模板 | `src/core/web/box.go:34,104` | `r.LoadHTMLFiles("./template/Box.tmpl")` + `c.HTML(..., "Box.tmpl", ...)` |
| 两模板 class 词表是否重合 | 见下 | **零重合** |

gin 的 `LoadHTMLFiles` 在此只加载单个文件，模板上下文中不存在其它模板，
`{{template "..."}}` 无从解析，因此片段复用在此结构下不可能发生。

class 词表对照（`grep -nE '^\s*[^ ].*\{'`）：

- `Box.tmpl`：`#main` `#label` `#title` `.operator` `.operator_bg` `.profession` `.rarity` `.evolve` `.level` `.potential` `.name`
- `BoxDetail.tmpl`：`#main` `table` `tr` `img` `td`

除 `#main`（每模板自有的 id）外无任何共用 class；Box 用 class 选择器，
BoxDetail 用元素选择器，词表不交叉。

实现侧同样独立：`RenderBoxDetail`（`render.go:256`）体内不调用 `RenderBox`，
只调用通用基础设施 `fillRoundedCard` / `DrawRoundedRectangle` / `tryLocal` / `ScaleExact` /
`setFont` / `drawString` / `drawStringAnchored` / `itoa`。

**故 box-detail 的判定是「有独立实现」，不是「实现在通用片段里」，也不是「未找到实现」。**

### 关于 `scene_*.go` 的重要事实

`src/ggrender/` 下 `scene_base.go` / `scene_gacha.go` / `scene_help.go`（以及 `scene_operator.go` / `scene_state.go`）
**不是实现文件**。三者全文各 7 行，内容为：

```go
package ggrender

import "github.com/fogleman/gg"

// ponytail: scene base extra file ensures per-template Go file with real gg drawing (reuse main Render func)
func init() { _ = gg.NewContext(10, 10) }

var _ = SceneSet // keep import used
```

即空桩（各 242~246 字节）。base / gacha / help / box-detail 四个场景的真实实现都在 `render.go`。
box-detail 确实已实现，位置是 `render.go:256 RenderBoxDetail`，**不存在** `scene_box_detail.go`。

---

## base

| 项 | 模板声明（`Base.tmpl` 行号 + 原文） | 我们现在画的（`render.go`） |
|---|---|---|
| 容器宽 | `9: width: 1110px;` | `610: const mainW = 1100` |
| 容器高 | 未声明（`#main` 无 height；`common.css` 亦无） | `613: h := 60 + 50 + 100 + len(data.Tradings)*110 + len(data.Manufactures)*110 + len(data.Powers)*80 + 90 + 80 + 90 + len(data.Dorms)*90 + 40` |
| 容器底色 | `8: background-color: #2b333d;` | `615: FillBackground(dc, 30, 32, 33)` = `#1e2021` |
| display | `.base` 段 `17: display: inline-flex;`；`.title` 段 `42: display: flex;` | 无 flex 概念；逐块 `fillRoundedCard` 绝对定位 |
| flex-direction | `18: flex-direction: column;` | 无 |
| 卡片宽 | `13: width: 550px;`；另 `85:` 与 `128:` 两处内联覆盖 `<div class="base" style="width: 1105px;">` | `mainW-2*pad` = 1100−32 = 1068（`625` 等处 `fillRoundedCard(..., float64(mainW-2*pad), ...)`） |
| 卡片高 | `14: height: 110px;` | labor 50 / control 100 / tradings 100 / manufactures 100 / powers 80 / meeting 80 / hire 70 / training 80 / dorms 80 |
| 网格列数 | 模板无 grid 声明。`.base` 为 `inline-flex` 550px 宽，容器 1110px → 每行 2 个（550×2=1100 ≤ 1110）；`85`/`128` 两块内联 1105px → 每行 1 个 | 单列纵向堆叠，无并排 |
| 间距 | `19: margin-top: 5px;`；`.title_icon` `48: margin-right: 20px;`；`h3` `26: margin-left: 10px;`；`.chars` `51: margin-left: 10px;` | `611: const pad = 16`；块间距 70/120/110/110/90/90/80/90 |
| 头像尺寸 | `23: width: 40px;`（`.avatar`） | `640: dc.DrawCircle(cx+22, cy, 22)` 圆形 r=22（直径 44）；其余处 `657/675/693/710/726/742/759: DrawCircle(cx+16, cy, 16)` 圆形 r=16 |
| 进度条 | `progress` 段 `29: width: 150px;` `30: height: 3px;` `31: border-radius: 1px;`；`#labor` 段 `37: width: 100px;` `38: height: 3px;` `39: border-radius: 1px;` | `629: ProgressBar(dc, float64(pad+250), float64(y+20), 300, 14, ...)` 宽 300 高 14 |
| 字号 | 未声明（`grep -nE 'font-size\|line-height' template/Base.tmpl` 退出码 1、输出为空；`common.css` 亦无） | `620: setFont(dc, 26)`；`626/633: setFont(dc, 16)`；`650/668/686/703/719/735: setFont(dc, 14)`；`642: setFont(dc, 11)`；`659/677/695/712/728: setFont(dc, 10)` |
| 行高 | 未声明（同上，两侧均无） | 未显式设置 |

`.board` 段 `56: width: 20px;` `58: border-radius: 5px;`、`.skill` 段 `61: width: 30px;` 在实现侧无对应物。

---

## gacha

| 项 | 模板声明（`Gacha.tmpl` 行号 + 原文） | 我们现在画的（`render.go`） |
|---|---|---|
| 容器宽 | `9: width: 1000px;` | `793: const mainW=900` |
| 容器高 | 未声明（`#main` 无 height；`common.css` 亦无） | `797: mainH:=headerH+statsH+charsH+40`，其中 `794: headerH:=90`、`795: statsH:=120`、`796: charsH:= 20*74` |
| display | `.item` 段 `24: display: inline-block;`；`#article table` 段 `51: display: inline-block;`；`.chars` 段 `65: display: inline-table;` | 无 |
| flex-direction | 未声明（模板无 `flex-direction`；`common.css` 亦无） | 无 |
| 头部高 | `44: height: 400px;`（`#header`） | `794: headerH:=90`；`802: dc.DrawRectangle(0,0,float64(mainW),float64(headerH))` |
| 尾部高 | `77: height: 269px;`（`#footer`） | 未绘制 |
| 网格列数 | `.item` 宽 300 + `29: margin-left: 20px;` 容器 1000px → 每行 3 个（320×3=960 ≤ 1000）。`#article table` 宽 230 + `margin-top: 20px` 声明于 `50/52` | 统计卡 `RoundRect(..., 200, 80, 8)`（`818`）`x+=220` → 每行 4 个（220×4=880 ≤ 900） |
| 间距 | `28: margin-top: 150px;`（`.item`）；`37: margin: 15px 0;`（`#avg tr`）；`33: margin-left: 30px;`（`#avg`）；`.chars` `63: margin-left: 20px;` `64: margin-top: 40px;` | `811: y:=headerH+16`；条目步进 74（`yy:=y+i*74`） |
| 头像尺寸 | `56: width: 100px;`（`#article img`） | `842: dc.DrawCircle(50,float64(yy+32),22)` 圆形 r=22（直径 44） |
| 星级卡 | `22: width: 300px;` `23: height: 250px;` `20px` 圆角见 `26: border-radius: 20px;` | `818: RoundRect(dc,float64(x),float64(y),200,80,8)` 圆角 8 |
| 干员列表容器 | `59: width: 465px;` | `832` 附近 `fillRoundedCard(dc,20,float64(yy),float64(mainW-40),64,8,10)` 宽 860 |
| 字号 | `19: font-size: 20px;`（`.title`）；`68: font-size: 15px;`（`.t`）；`71: font-size: 10px;`（`.new`）；`90:` 内联 `style="font-size: 23px;"` | `804: setFont(dc,24)`；`807/819/832/844: setFont(dc,14)`；`822: setFont(dc,22)`；`825/859: setFont(dc,11)`；`847/855: setFont(dc,12)` |
| 行高 | 未声明（模板无 `line-height`；`common.css` 亦无） | 未显式设置 |

---

## enemy

| 项 | 模板声明（`Enemy.tmpl` 行号 + 原文） | 我们现在画的（`scene_enemy.go`） |
|---|---|---|
| 容器宽 | `8: width: 656px;`（`#main`） | `37: const W, H = 984, 477`（= 656×1.5、318×1.5） |
| 容器高 | `9: height: auto;` | `37: const W, H = 984, 477` 固定 477 |
| 底色 | `7: background-color: #323332;` | `38: FillBackground(dc, 50, 51, 50)` = `#323332` ✓ |
| display | 未声明（模板为 `<table>` 流；无 `display` / `flex-direction` 声明） | 绝对坐标逐行绘制 |
| flex-direction | 未声明（`common.css` 亦无） | 无 |
| 表格宽 | `17: width: 656px;`（`#base`）；`33: width: 656px;`（`.level`） | `W` = 984 |
| 网格列数 | `#base` 首个信息行用 `colspan="8"` 共 8 列；`种类/地位级别/攻击方式/行动方式` 行用 `colspan="2"`×4 = 8 列；`.level` 表用 `colspan="6"` 共 6 列 | 竖线在 `425`、`611.5`、`797.5`、`981.5`（`scene_enemy.go` border 段），即 8 等分；`.level` 表未绘制 |
| 间距 | `border-spacing: 0;`（`18` 与 `34`）；`border: solid 1px #595858;`（`21` 与 `37`） | 边框色 `dc.SetRGB255(89, 88, 88)` = `#595858` ✓，线宽 2px |
| 头像尺寸 | `44: img { width: 158px; }` | `83: dc.DrawImage(ScaleExact(pic, 237, 237), 95, 102)`（237 = 158×1.5） |
| 字号 | `26: font-size: 25px;`（`#name`）；`38: font-size: 22px;`（`.title`）；`81/146/165:` 内联 `style="font-size: 20px;font-weight: 600;"` | `77: setFont(dc, 37.5)`（= 25×1.5）；`103: setFont(dc, 30)`（= 20×1.5）；`41` 与 `centerText` 调用处用 24（= 16×1.5，模板未声明 16px） |
| 行高 | 未声明（模板无 `line-height`；`common.css` 亦无） | 未显式设置 |
| 文本色 | `13-15: td { color: white; background-color: #323332 !important; }`；`41-43: a { color: white; }` | `dc.SetRGB255(255, 255, 255)` |

`.title` 与 `#name` 均有 `font-weight: 600;`（`27`、`39`）；实现侧以 `drawStringBoldW(..., 1.4)` / `(..., 1.2)` 合成加粗（`79`、`105`），非真实字重。

---

## box-detail

| 项 | 模板声明（`BoxDetail.tmpl` 行号 + 原文） | 我们现在画的（`render.go`） |
|---|---|---|
| 容器宽 | 未声明（`#main` 段 `7-11` 只有 `position: absolute;` 与 `background-color: #2e3031;`；`common.css` 亦无 width） | `257: const mainW = 900` |
| 容器高 | 未声明（同上） | `260: mainH := pad + len(data)*cardH + pad`，`258: const cardH = 155`、`259: const pad = 10` |
| 底色 | `9: background-color: #2e3031;` | `261: FillBackground(dc, 27, 29, 30)` = `#1b1d1e` |
| display | 内联 `style="display: inline-flex;align-items: center;width: 100%;"`（干员单元格）；`style="display: inline-flex;align-items: center;flex-direction: column"`（等级/潜能/技能/模组单元格） | 绝对定位 |
| flex-direction | 内联 `flex-direction: column`（同上三处） | 无 |
| 表格对齐 | `24-28: td { vertical-align: middle; text-align: center; white-space: nowrap; }` | `DrawRoundedRectangle` 卡片 + 左对齐 `drawString` |
| 网格列数 | 5 列（`28-34` 表头 `干员/等级/潜能/技能/模组` 五个 `<th>`） | 5 个语义区：头像+名字 / 等级 / 潜能 / 技能 / 模组（`265-306`） |
| 间距 | 未声明（模板无 margin / gap；`common.css` 亦无） | `259: const pad = 10`；卡片间距 155−145=10；技能步进 `291: skx += 42`；模组步进 `305: ekx += 42` |
| 头像尺寸 | `19-21: img { width: 50px; }` | `268: dc.DrawRoundedRectangle(float64(pad+10), float64(y+10), 90, 90, 6)` 占位 90×90 |
| 星级/潜能图标 | 模板内 `img` 统一 `width: 50px;` | `274: dc.DrawImage(ScaleExact(ev, 22, 22), pad+115, y+40)`；`276: ... pot, 22, 22 ...` 实际 22×22 |
| 技能/模组图标 | 模板内 `img` 统一 `width: 50px;` | `286` 与 `300: dc.DrawRoundedRectangle(..., 30, 30, 4)` 占位 30×30 |
| 字号 | 未声明（`grep -nE 'font-size\|line-height' template/BoxDetail.tmpl` 退出码 1、输出为空；`common.css` 亦无） | `270: setFont(dc, 20)`；`280/295: setFont(dc, 13)`；`288/302: setFont(dc, 11)` |
| 行高 | 未声明（同上，两侧均无） | 未显式设置 |
| 行阴影 | `16-18: tr { box-shadow: 0 3px 1px -2px rgba(0,0,0,.2),0 2px 2px rgba(0,0,0,.14),0 1px 5px rgba(0,0,0,.12); }` | 未实现 |

---

## help

| 项 | 模板声明（`Help.tmpl` 行号 + 原文） | 我们现在画的（`render.go`） |
|---|---|---|
| 容器宽 | `8: width: 660px;`（`#main`） | `893: const mainW=990` |
| 容器高 | 未声明（`#main` 无 height；`common.css` 亦无） | `897-900`：`mainH:=200+privH+pubH+adminH+60`，其中 `privH/pubH/adminH := 40+len(...)*32` |
| 底图 | `9-10: background-image: url("/assets/help/bg.jpg"); background-size: cover;` | `894: FillBackground(dc,46,48,49)` = `#2e3031` 纯色 |
| display | `.cmd` `44: float: left;`；`.label` `36: float: left;`；`.cmdType` `40: float: left;`；`.banner img` `14: float: left;`；`.banner h1` `18: float: left;`；`.banner p` `25: float: left;` | 无 |
| flex-direction | 未声明（模板无 `flex-direction`；`common.css` 亦无） | 无 |
| 网格列数 | `.cmd` 段 `42: width: 150px;` + `45: margin-left: 10px;`，容器 660px → 每行 4 个（160×4=640 ≤ 660） | 指令行单列纵向堆叠 |
| 间距 | `.cmd` `45: margin-top: 15px;` `46: margin-left: 10px;`；`.cmd p` `50-52: margin-bottom: 0; margin-top: 3px; margin-left: 5px;`；`.label` `37: margin-top: 10px;`；`.cmdType` `41: margin-top: -32px; margin-left: 25px;`；`.banner img` `15: margin-top: 10px;`；`.banner h1` `19: margin-top: -110px; margin-left: 20px;`；`.banner p` `26-27: margin-left: 20px; margin-top: -45px;` | `1017` 附近步进 32（`yy+=32`），标题步进 20 |
| 横幅图高 | `13-15: .banner img { width: 100%; margin-top: 10px; float: left; }` | `1003: dc.DrawRectangle(0,0,float64(mainW),140)` 纯色块 |
| 标签图高 | `32-35: .label img { height: 40px; width: 100%; float: left; }` | 未绘制 |
| 指令块 | `41-48: .cmd { width: 150px; float: left; margin-top: 15px; margin-left: 10px; border: solid 1px; font-size: 15px; border-radius: 10px; color: white; font-weight: 600; }` | `1019: RoundRect(dc,20,float64(yy),float64(mainW-40),28,6)` 宽 950 高 28 圆角 6 |
| 字号 | `54: font-size: 15px;`（`.cmd`） | `1005: setFont(dc,28)`；`1008/1013: setFont(dc,14/16)`；`1020: setFont(dc,13)` |
| 行高 | 未声明（模板无 `line-height`；`common.css` 亦无） | 未显式设置 |

---

## depot

| 项 | 模板声明（`Depot.tmpl` 行号 + 原文） | 我们现在画的（`scene_depot.go`） |
|---|---|---|
| 容器宽 | `8: width: 850px;`（`#main`） | `37: const w, h = 1275, 234`（= 850×1.5、156×1.5） |
| 容器高 | 未声明（`#main` 无 height；`common.css` 亦无） | `37: const w, h = 1275, 234` 固定 234 |
| 底色 | `9: background-color: #2e3031;` | `39: dc.SetRGB(0.96, 0.96, 0.95)` = `#f5f5f2` 浅色 |
| display | `.item` `12: display: inline-flex;` | 无 flex，逐格绝对定位 |
| flex-direction | `13: flex-direction: column;` | 无 |
| 格子宽 | `15: width: 80px;` | `45: const cellW, cellH = 152, 102`；绘制 `cellW-12` = 140 宽 |
| 格子高 | 未声明（`.item` 无 height；`common.css` 亦无） | `45: cellH = 102`；绘制 `cellH-14` = 88 高 |
| 网格列数 | 模板无 grid / 无固定列数声明；`.item` 为 `inline-flex` 80px 宽，容器 850px → 每行最多 10 个（80×10=800 ≤ 850） | `44: const cols = 8`，即固定 8 列 |
| 间距 | 未声明（模板无 margin / gap；`common.css` 亦无） | `45: cellW=152, cellH=102`；起点 `x = 20 + (i%cols)*cellW`、`y = 16 + (i/cols)*cellH` |
| 图标尺寸 | `17-19: .icon { width: 75px; }` | `dc.DrawRectangle(float64(x+8), float64(y+8), 52, 52)` 纯色块 52×52 |
| 计数块 | `20-27: .count { position: absolute; color: white; background-color: rgba(0, 0 ,0 ,0.5); font-size: 12px; margin-top: 50px; margin-right: -30px; }` | `dc.DrawStringAnchored(it.Count, float64(x+cellW-20), float64(y+18), 1, 0)`，无背景块 |
| 字号 | `24: font-size: 12px;`（`.count`） | `41: LoadDefaultFont(dc, 20)`，全场景统一 20 |
| 行高 | 未声明（模板无 `line-height`；`common.css` 亦无） | 未显式设置 |
| 条数上限 | 无上限声明（`{{range .}}` 全量输出） | `47-49: if i >= cols*2 { break }` 即最多 16 条 |

---

## 本轮只读取证中值得单独记一笔的事实

1. `scene_base.go` / `scene_gacha.go` / `scene_help.go` 是空桩，真实实现在 `render.go`。
   按文件名判断「某场景未实现」会得出错误结论。
2. `RenderBoxDetail` 存在于 `render.go:256`；仓库中不存在 `scene_box_detail.go` 这样的文件。
3. `common.css` 只有 16 行，且不含 `color` / `font-size` / `line-height`。
   本轮 6 个模板里，`Base.tmpl` 与 `BoxDetail.tmpl` 的 `font-size`/`line-height` 在模板内 grep 退出码 1、输出为空，
   加上 `common.css` 也没有，才判定为「未声明」。
4. 本轮未发现任何模板侧 `line-height` 声明（6 个模板全部）。
5. **管道吃掉退出码的自查**：排查 box-detail 复用时，`grep ... | head -10` 之后取的 `$?`
   是 `head` 的退出码而非 `grep` 的，会把「查询失败」与「结果为空」压成同一个信号。
   发现后已把这两条（模板引擎 include 探测、`Box.tmpl` 选择器枚举）改为
   `out=$(grep ...); rc=$?` 不接管道重做，上表所载为重做后的真实退出码。
