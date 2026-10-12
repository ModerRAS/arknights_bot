# pr-images — 出处标注（Boss 硬约束：无出处的图不许进 PR）

分支 `gg-comparison-images-20261002` · 基底 `consolidate/gg-mainline` @ `4584759`
本目录只放图与本文件；**本分支不参与渲染、不改渲染或计分代码**。

## 数据来源（三个来源，互不混用）

| 来源 | 路径 | 说明 |
|---|---|---|
| **我们的真实渲染** | `arkskadi_bot-measure-align/tmp/pixel-compare/state/new.png` | 由 gg harness 产出，1092×510 |
| **冻结基线（只读比对）** | `arkskadi_bot-measure-align/tmp/pixel-compare/state/old.png` | **仅作并排比对用，绝不作为渲染输入或背景**（`641988f` 作弊形态）。1092×510 |
| **门禁分数** | `tmp/pixel-compare/report.json` | 16 场景 PASS 2 / FAIL 14，2+14==16 |

> **红线声明**：本轮渲染过程**从未读取基线图**；图中的并排是「两张已存在产物的比对」，
> 不是把基线喂进渲染管线。**基线的角色是靶子，不是输入。**

## 每张图

### `01-state-campaign_recover_bg.png`
- 场景 `state` · 局部放大 ×3 · 裁剪区 (912,198)-(1050,251)
- 红框 = legacy 页面 resolve 出的 rect `(926,212,110,25)`
- 门禁状态：`state` **已过门禁**（sim 0.99093）
- ⚠️ **这处不属于「像素差 1–4px」那一类**：CSS 声明 `110×25` · gg 代码 `112×21` · 浏览器解析 `110×25`
  ⇒ **代码与声明、与解析值都不一致，出处不明**，属 `hhCardH` 那一族失效形态。

### `02-state-remain_secs_bg.png`
- 场景 `state` · 局部放大 ×3 · 裁剪区 (909,402)-(1070,455)
- 红框 = resolve rect `(923,416,133,25)`
- ⚠️ **尺寸与声明逐字一致**（CSS `133×25` · 代码 `133×25`），差异只在**位置** Δ`[+1,+4,0,0]`。

### `03-state-full-scene.png`
- 场景 `state` 整幅并排，缩放 1/2。

### `04-gate-overview-16-scenes.png`
- 16 场景门禁分数条形图，标注 0.99 门禁线，PASS 2 / FAIL 14（2+14==16）。
- 数据来自 `report.json`，**非图像派生**。

## 未做成图的一项（性质不能丢）
`measure_std*.log` 等 **0 字节**闸门证据**刻意不做图** ——
**「0 字节」本身就是证据**（闸门在对照未过时未产出任何输出），做成图反而丢掉性质。

## 复算路径
`png/`(legacy 捕获) → `goenc/main.go`(Go 编码) → `goj/`(98 张)；聚合 json 与原始产物互相可指。

## failing 14 场景整幅并排（2026-10-12，代码 consolidate/gg-mainline @ 1c4e82b）
- 每行左 = 冻结基线 old.png（只读比对），右 = 本分支 gg 渲染 new.png；行标签 sim 来自 harness `report.json`，生成时读入、非手抄。
- 过门禁的 `calendar` / `state` 不在此列；14 张全是 `passed=false`，14+2==16。

### `failed-1.png`
- `lottery` 1473x1667 sim 0.98358 fail
- `base` 1665x918 sim 0.97073 fail
- `headhunt` 1049x576 sim 0.97021 fail
- `card` 1280x720 sim 0.96015 fail

### `failed-2.png`
- `depot` 1275x234 sim 0.92352 fail
- `enemy` 984x477 sim 0.91951 fail
- `box-summary` 1350x723 sim 0.88098 fail
- `box-detail` 722x279 sim 0.87844 fail

### `failed-3.png`
- `missing` 1050x536 sim 0.87614 fail
- `help` 990x2049 sim 0.83943 fail
- `box` 1050x536 sim 0.82504 fail
- `gacha` 1500x1323 sim 0.90251 fail

### `failed-4.png`
- `operator` 1800x1200 sim 0.76866 fail
- `recruit` 1350x534 sim 0.88603 fail
