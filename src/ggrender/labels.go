package ggrender

import "github.com/fogleman/gg"

// B1 顶部属性面板的渲染契约：标签字面量 + 行序 + CSS 几何。
//
// 标签来源：template/Operator.tmpl 的 #attr 表，行 118/120/122/126/128/130/134/136/138。
// 该模板就是冻结基线 images/operator.jpg 的渲染来源（manifest entry: id=operator,
// template=Operator.tmpl, route=/operator, caller=plugins/operator/operator_handle.go,
// rootType=operator.Operator, scale=1.5, bbox 1200x800, pixelWidth 1800）。
//
// 下面每个几何数字都能追到模板 CSS / assets/css/common.css / manifest，
// 无一项是从基线图像素反推出来的。
const (
	// 第 1 行
	labelMaxHP = "最大生命值"
	labelATK   = "攻击力"
	labelDEF   = "防御力"
	// 第 2 行
	labelRES      = "法抗"
	labelInterval = "攻击间隔"
	labelReDeploy = "再部署时间"
	// 第 3 行
	labelBlock = "阻挡数"
	labelCost  = "部署费用"
	labelLogo  = "所属"
)

// #attr 表的 CSS 几何，全部照 template/Operator.tmpl 内 <style> 与 assets/css/common.css。
const (
	cssPanelW   = 1200.0 // #main { width: 1200px }，与 manifest bbox.width 一致
	cssPanelH   = 800.0  // #main { min-height: 800px }，与 manifest bbox.height 一致
	cssFontPx   = 16.0   // common.css 的 body 未设 font-size -> 浏览器默认 16px
	cssLabelW   = 100.0  // .b { width: 100px }
	cssValueW   = 70.0   // .w { width: 70px }
	cssPanelTop = 20.0   // #attr { margin-top: 20px }
	cssCellPad  = 1.0    // 浏览器默认 td 内边距 1px，模板未覆盖
	cellAlpha   = 204    // .b/.w 的 opacity: 0.8 -> round(0.8*255)
)

// operatorStat 是模板里相邻的一对 td.b（标签）/ td.w（值）。
type operatorStat struct {
	label string
	value string
}

// operatorStats 按模板行序展开 B1 的 9 项：行内 3 对、行间 3 行。
// 用具名字段而不是 map[string]string：行序是模板契约，map 无序；
// 且具名字段杜绝了 key 拼错静默取到空串。
func (d *OperatorInfo) operatorStats() []operatorStat {
	return []operatorStat{
		{labelMaxHP, d.MaxHP}, {labelATK, d.ATK}, {labelDEF, d.DEF},
		{labelRES, d.RES}, {labelInterval, d.Interval}, {labelReDeploy, d.ReDeploy},
		{labelBlock, d.Block}, {labelCost, d.Cost}, {labelLogo, d.Logo},
	}
}

// cssLineHeight 返回给定 CSS px 字号下字体的真实行高（px）。
//
// 刻意不用 dc.FontHeight()：gg 的 LoadFontFace 把它固定成 points*72/96
// （gg@v1.3.0 context.go:688），那是 DPI 约定值，不是字体度量，拿它当行高是错的。
// 这里直接问 truetype face 要 Metrics().Height，与浏览器 line-height:normal 同源。
// 全部候选字体都不可用时回落到 1.4 倍字号（Chromium 对多数 CJK 字体的 normal 行高）。
func cssLineHeight(px float64) float64 {
	for _, p := range FontCandidates {
		if face, err := gg.LoadFontFace(p, px); err == nil {
			return float64(face.Metrics().Height) / 64
		}
	}
	return px * 1.4
}
