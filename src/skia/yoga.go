package skia

import (
	"math"
	"sync"
)

// ponytail: global lock, Yoga cgo if throughput matters
// P1 minimal flex stub — pure Go, no cgo, for depot 80x78 grid.
// Real path P2: facebook/yoga cgo YGNodeCalculateLayout + MeasureFunc via text.go.
// This stub computes depot-like flexWrap with gap 3.5, padding 10, per-row floor((w+gap)/(80+gap)).
// Windows cgo unavailable → pure Go fallback guarantees go vet ./skia passes.

var yogaMu sync.Mutex // ponytail: global lock

type MeasureMode int

const (
	MeasureUndefined MeasureMode = iota
	MeasureExactly
	MeasureAtMost
)

type Size struct{ Width, Height float32 }

type Layout struct {
	X, Y, Width, Height float32
}

type Style struct {
	Width, Height    float32
	FlexDirection    string
	FlexWrap         string
	AlignItems       string
	AlignContent     string
	JustifyContent   string
	Gap              float32
	Position         string
	Top, Right, Bottom, Left *float32
	Display          string
}

type YogaNode struct {
	Style       Style
	Children    []*YogaNode
	Layout      Layout
	MeasureFunc func(width float32, widthMode MeasureMode, height float32, heightMode MeasureMode) Size
	Text        string
	IconPath    string
	BadgeText   string
}

func NewYogaNode() *YogaNode { return &YogaNode{Style: Style{Display: "flex"}} }

func (n *YogaNode) AddChild(c *YogaNode) { n.Children = append(n.Children, c) }

func (n *YogaNode) SetMeasureFunc(fn func(width float32, widthMode MeasureMode, height float32, heightMode MeasureMode) Size) {
	n.MeasureFunc = fn
}

func f32(v float32) *float32 { return &v }

func (n *YogaNode) CalculateLayout(availableWidth, availableHeight float32) {
	yogaMu.Lock()
	defer yogaMu.Unlock()
	if n.MeasureFunc != nil && len(n.Children) == 0 {
		sz := n.MeasureFunc(availableWidth, MeasureAtMost, availableHeight, MeasureUndefined)
		n.Layout.Width = sz.Width
		n.Layout.Height = sz.Height
		return
	}
	w := availableWidth
	if n.Style.Width > 0 {
		w = n.Style.Width
		n.Layout.Width = w
	} else {
		n.Layout.Width = w
	}
	h := n.Style.Height
	if h > 0 { n.Layout.Height = h }
	if n.Style.FlexWrap == "wrap" {
		gap := n.Style.Gap
		x, y := float32(0), float32(0)
		lineH := float32(0)
		lineW := float32(0)
		if gap > 0 {
			cols := int(math.Floor(float64((w + gap) / (80 + gap))))
			if cols < 1 { cols = 1 }
			_ = cols
		}
		for _, child := range n.Children {
			cw := child.Style.Width
			ch := child.Style.Height
			if cw == 0 && child.MeasureFunc != nil {
				sz := child.MeasureFunc(w, MeasureAtMost, 0, MeasureUndefined)
				cw = sz.Width; ch = sz.Height
			}
			if cw == 0 { cw = 80 }
			if ch == 0 { ch = 78 }
			child.Layout.Width = cw
			child.Layout.Height = ch
			if lineW+cw > w && lineW > 0 {
				x = 0; y += lineH + gap; lineW = 0; lineH = 0
			}
			child.Layout.X = x; child.Layout.Y = y
			x += cw + gap; lineW += cw + gap
			if ch > lineH { lineH = ch }
			for _, gc := range child.Children {
				if gc.Style.Position == "absolute" {
					gw := gc.Style.Width; gh := gc.Style.Height
					if gc.MeasureFunc != nil {
						sz := gc.MeasureFunc(w, MeasureAtMost, 0, MeasureUndefined)
						gw = sz.Width; gh = sz.Height
						if gh == 0 { gh = 14 }
					}
					if gw == 0 { gw = 30 }
					if gh == 0 { gh = 14 }
					gc.Layout.Width = gw; gc.Layout.Height = gh
					if gc.Style.Top != nil { gc.Layout.Y = child.Layout.Y + *gc.Style.Top } else { gc.Layout.Y = child.Layout.Y }
					if gc.Style.Right != nil { gc.Layout.X = child.Layout.X + child.Layout.Width - *gc.Style.Right - gw } else if gc.Style.Left != nil { gc.Layout.X = child.Layout.X + *gc.Style.Left } else { gc.Layout.X = child.Layout.X }
				}
			}
		}
		if n.Style.Height == 0 { n.Layout.Height = y + lineH }
		return
	}
	y := float32(0)
	for _, c := range n.Children {
		c.Layout.X = 0; c.Layout.Y = y
		if c.Style.Height > 0 { c.Layout.Height = c.Style.Height }
		if c.Style.Width > 0 { c.Layout.Width = c.Style.Width }
		y += c.Layout.Height
	}
	if n.Style.Height == 0 { n.Layout.Height = y }
}

func NewNode() *YogaNode { return NewYogaNode() }
func (n *YogaNode) SetWidth(v float32)  { n.Style.Width = v }
func (n *YogaNode) SetHeight(v float32) { n.Style.Height = v }
func (n *YogaNode) SetFlexDirection(s string) { n.Style.FlexDirection = s }
func (n *YogaNode) SetGap(v float32)    { n.Style.Gap = v }
func (n *YogaNode) SetPadding(v float32) {}
func (n *YogaNode) GetComputedLayout() (x, y, w, h float32) { return n.Layout.X, n.Layout.Y, n.Layout.Width, n.Layout.Height }
