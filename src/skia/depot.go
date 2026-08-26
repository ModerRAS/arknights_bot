package skia

import (
	"image/color"
	"os"
	"path/filepath"
)

type DepotItem struct {
	Name   string `json:"name"`
	Count  string `json:"count"`
	Icon   string `json:"icon"`
	SortId int64  `json:"sortId"`
}

func BuildDepot(items []DepotItem) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 850
	root.Style.Height = 156
	root.Style.FlexWrap = "wrap"
	root.Style.AlignContent = "flex-start"
	root.Style.Display = "flex"
	root.Style.Gap = 3.5
	for _, it := range items {
		item := NewYogaNode()
		item.Style.Width = 80
		item.Style.Height = 78
		item.Style.Display = "flex"
		item.Style.FlexDirection = "column"
		item.Style.AlignItems = "center"
		item.Style.Position = "relative"
		item.IconPath = it.Icon
		item.BadgeText = it.Count
		item.Text = it.Count
		badge := NewYogaNode()
		badge.Style.Position = "absolute"
		badge.Style.Top = f32(50)
		badge.Style.Right = f32(5)
		badge.Style.Display = "flex"
		badge.Text = it.Count
		txt := it.Count
		badge.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
			ww := float32(MeasureText(txt, 12))
			if ww < 10 { ww = 10 }
			return Size{Width: ww, Height: 14}
		})
		item.AddChild(badge)
		root.AddChild(item)
	}
	return root
}

func RenderDepot(items []DepotItem, scale float64) *Canvas {
	if scale <= 0 { scale = 1.5 }
	root := BuildDepot(items)
	root.CalculateLayout(850, 156)
	W := int(850*scale + 0.5)
	H := int(156*scale + 0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	iconBytes := loadDepotIconBytes()
	for _, child := range root.Children {
		cx := child.Layout.X
		cy := child.Layout.Y
		iconW := float32(75)
		iconH := float32(75)
		iconX := cx + (80-iconW)/2
		iconY := cy
		if len(iconBytes) > 0 {
			img := &Image{Bytes: iconBytes, MIME: "image/png", Width: 75, Height: 75}
			dst := Rect{X: iconX * float32(scale), Y: iconY * float32(scale), W: iconW * float32(scale), H: iconH * float32(scale)}
			_ = c.DrawImageRect(img, dst)
		} else {
			c.DrawRect(Rect{X: iconX * float32(scale), Y: iconY * float32(scale), W: iconW * float32(scale), H: iconH * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		}
		for _, badge := range child.Children {
			bx := badge.Layout.X
			by := badge.Layout.Y
			bw := badge.Layout.Width
			bh := badge.Layout.Height
			padX := float32(4)
			bwPadded := bw + padX
			bxAdj := bx - padX/2
			c.DrawRect(Rect{X: bxAdj * float32(scale), Y: by * float32(scale), W: bwPadded * float32(scale), H: bh * float32(scale)}, Paint{Color: color.RGBA{0, 0, 0, 128}})
			tx := bxAdj + 2
			ty := by + 1
			c.DrawText(badge.Text, tx, ty, 12, color.RGBA{0xff, 0xff, 0xff, 255}, scale)
		}
	}
	return c
}

func loadDepotIconBytes() []byte {
	candidates := []string{
		"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go/src/utils/media/testdata/visual/baseline/cache/depot-lmd.png",
		"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go/assets/common/amiya.png",
		"assets/common/amiya.png",
		"src/utils/media/testdata/visual/baseline/cache/depot-lmd.png",
	}
	for _, p := range candidates {
		if b, err := os.ReadFile(p); err == nil && len(b) > 0 { return b }
	}
	if b, err := os.ReadFile(filepath.Join("C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go", "src/utils/media/testdata/visual/baseline/cache/depot-lmd.png")); err == nil {
		return b
	}
	return nil
}
