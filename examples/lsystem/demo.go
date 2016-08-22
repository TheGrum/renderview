package main

import (
	rv "renderview"
	"renderview/driver"
)

func main() {
	m := rv.NewBasicRenderModel()
	m.AddParameters(
		rv.SetHints(rv.HINT_SIDEBAR,
			rv.NewFloat64RP("left", -10),
			rv.NewFloat64RP("top", -10),
			rv.NewFloat64RP("right", 10),
			rv.NewFloat64RP("bottom", 10),
			rv.NewIntRP("width", 100),
			rv.NewIntRP("height", 100),
			rv.NewIntRP("options", rv.OPT_AUTO_ZOOM),
		)...)
	m.AddParameters(
		rv.SetHints(rv.HINT_FULLTEXT,
			rv.NewStringRP("lsystem", "FX\nX=X+YF+\nY=-FX-Y\n"))...)
	m.AddParameters(
		rv.SetHints(rv.HINT_FOOTER,
			rv.NewFloat64RP("angle", 90),
			rv.NewIntRP("depth", 5))...)
	m.InnerRender = func() {
		m.Img = RenderLSystem(m)
	}
	m.NeedsRender = true
	driver.Main(m)
}
