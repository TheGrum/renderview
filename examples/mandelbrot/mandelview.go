package mandelbrot

import (
	"fmt"
	"math"
	rv "renderview"
)

type MandelView rv.BasicRenderModel

func getInnerRenderFunc(m *MandelView) func() {
	return func() {
		innerRender(m)
	}
}

func innerRender(m *MandelView) {
	var rMin, iMin, rMax, iMax float64
	var width, red, green, blue int
	var maxEsc int

	m.Lock()
	maxEsc = m.Params[4].GetValueInt()
	rMin = m.Params[0].GetValueFloat64()
	iMin = m.Params[1].GetValueFloat64()
	rMax = m.Params[2].GetValueFloat64()
	iMax = m.Params[3].GetValueFloat64()
	width = m.Params[5].GetValueInt()
	red = m.Params[7].GetValueInt()
	green = m.Params[8].GetValueInt()
	blue = m.Params[9].GetValueInt()
	m.Rendering = true
	m.Unlock()

	i2 := generateMandelbrot(rMin, iMin, rMax, iMax, width, red, green, blue, maxEsc)

	m.Lock()
	m.Img = i2
	m.Rendering = false
	m.Unlock()
	if !(m.RequestPaint == nil) {
		m.RequestPaint()
	}
}

func NewMandelView() *rv.BasicRenderModel {
	m := rv.NewBasicRenderModel()
	m.InnerRender = getInnerRenderFunc((*MandelView)(m))
	m.AddParameters(rv.NewFloat64RP("left", -2), rv.NewFloat64RP("top", -1), rv.NewFloat64RP("right", 0.5), rv.NewFloat64RP("bottom", 1), rv.NewIntRP("maxEsc", 100), rv.NewIntRP("width", 100), rv.NewIntRP("height", 100), rv.NewIntRP("red", 230), rv.NewIntRP("green", 235), rv.NewIntRP("blue", 255), rv.NewFloat64RP("mouseX", 0), rv.NewFloat64RP("mouseY", 0), rv.NewIntRP("zoom", 1), rv.NewIntRP("options", rv.OPT_AUTO_ZOOM))
	go m.GoRender()
	return m
}

type ZoomRenderParameter struct {
	rv.EmptyParameter

	Value int
	Model *MandelView
}

func (e *ZoomRenderParameter) GetValueInt() int {
	return e.Value
}

func (e *ZoomRenderParameter) SetValueInt(v int) int {
	dz := float64(v-e.Value) * 0.1
	if dz < 0 {
		dz = 1 / (1 + math.Abs(dz))
	} else if dz > 0 {
		dz = dz + 1
	} else {
		return v
	}

	e.Value = v

	rMin := e.Model.Params[0].GetValueFloat64()
	iMin := e.Model.Params[1].GetValueFloat64()
	rMax := e.Model.Params[2].GetValueFloat64()
	iMax := e.Model.Params[3].GetValueFloat64()
	width := e.Model.Params[5].GetValueInt()
	height := e.Model.Params[6].GetValueInt()
	mouseX := e.Model.Params[11].GetValueFloat64()
	mouseY := e.Model.Params[12].GetValueFloat64()

	dx := math.Abs(rMax - rMin)
	dy := math.Abs(iMax - iMin)
	rdx := dx / float64(width)
	rdy := dy / float64(height)
	ndx := dx * dz
	ndy := dy * dz

	mx := (mouseX * rdx) + rMin
	my := (mouseY * rdy) + iMin

	nleft := rMin - (mx * (ndx - dx))
	fmt.Printf("%v, %v, %v, %v, %v, %v\n", dz, rMin, mx, ndx, dx, nleft)
	nright := rMin + ndx
	ntop := iMin - (my * (ndy - dy))
	nbottom := iMax + ndy
	e.Model.Params[0].SetValueFloat64(nleft)
	e.Model.Params[1].SetValueFloat64(ntop)
	e.Model.Params[2].SetValueFloat64(nright)
	e.Model.Params[3].SetValueFloat64(nbottom)

	e.Model.RequestPaint()

	return e.Value
}

func NewZoomRP(name string, value int, m *MandelView) *ZoomRenderParameter {
	return &ZoomRenderParameter{
		EmptyParameter: rv.EmptyParameter{
			Name: name,
			Type: "int",
		},
		Value: value,
		Model: m,
	}
}
