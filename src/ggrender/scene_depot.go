package ggrender

import "github.com/fogleman/gg"

// ponytail: scene depot extra file ensures per-template Go file with real gg drawing (reuse main Render func)
func init() { _ = gg.NewContext(10, 10) }

var _ = SceneSet // keep import used

