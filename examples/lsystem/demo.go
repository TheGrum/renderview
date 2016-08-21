package main

import (
	rv "renderview"
	"renderview/driver"
)

func main() {
	m := rv.NewBasicRenderModel()
	m.AddParameters(
		rv.SetHints(rv.HINT_HIDE,
			rv.NewFloat64RP("left", -2),
			rv.NewFloat64RP("top", -1),
			rv.NewFloat64RP("right", 0.5),
			rv.NewFloat64RP("bottom", 1),
			rv.NewIntRP("width", 0),
			rv.NewIntRP("height", 0),
		)...)
	m.AddParameters(
		rv.SetHints(rv.HINT_FULLTEXT,
			rv.NewString("lsystem", "FX\nX=X+YF+\nY=-FX-Y\n"))...)
	m.AddParameters(
		rv.SetHints(rv.HINT_FOOTER,
			rv.NewFloat64RP("angle", 90),
			rv.NewIntRP("depth", 5))...)
	m.InnerRender = func() {
		m.Img = RenderLSystem(m)
	}
	driver.Main(m)
}
