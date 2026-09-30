// Command carditer renders only the card scene for fast iteration.
// It writes tmp/pixel-compare/card/new.png (same layout the parity test uses);
// scoring stays in the official harness / offline analysis scripts.
package main

import (
	"os"
	"path/filepath"

	"arknights_bot/ggrender"
)

func main() {
	dc, err := ggrender.RenderGGContext("card", nil)
	if err != nil {
		panic(err)
	}
	wd, _ := os.Getwd()
	root := wd
	for {
		if st, err := os.Stat(filepath.Join(root, "go.mod")); err == nil && !st.IsDir() {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			panic("go.mod not found")
		}
		root = parent
	}
	out := filepath.Join(filepath.Dir(root), "tmp", "pixel-compare", "card", "new.png")
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := dc.EncodePNG(f); err != nil {
		panic(err)
	}
	_ = filepath.Join
}
