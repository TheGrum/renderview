// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package mandelbrot

import (
	"math"

	rv "github.com/TheGrum/renderview"
)

type MandelModel rv.BasicRenderModel

func getInnerRenderFunc(m *MandelModel) func() {
	return func() {
		innerRender(m)
	}
}

func innerRender(m *MandelModel) {
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

func NewMandelModel() *rv.BasicRenderModel {
	m := rv.NewBasicRenderModel()
	m.InnerRender = getInnerRenderFunc((*MandelModel)(m))
	m.AddParameters(
		rv.NewFloat64RP("left", -2),
		rv.NewFloat64RP("top", -1),
		rv.NewFloat64RP("right", 0.5),
		rv.NewFloat64RP("bottom", 1),
		rv.NewIntRP("maxEsc", 100),
		rv.NewIntRP("width", 100),
		rv.NewIntRP("height", 100),
		rv.NewIntRP("red", 230),
		rv.NewIntRP("green", 235),
		rv.NewIntRP("blue", 255),
		rv.NewFloat64RP("mouseX", 0),
		rv.NewFloat64RP("mouseY", 0),
		NewZoomRP("zoom", 1, (*MandelModel)(m)),
		rv.NewIntRP("options", rv.OPT_NONE))
	go m.GoRender()
	return m
}

// Many applications can simply use OPT_AUTO_ZOOM
// but since the Mandelbrot algorithm we are using ignores the height
// and produces a square image, we use a custom parameter
// to calculate the zoom ourselves, also taking the opportunity to
// dynamically adjust the Escape parameter.
type ZoomRenderParameter struct {
	rv.EmptyParameter

	Value int
	Model *MandelModel
}

func (e *ZoomRenderParameter) GetValueInt() int {
	return e.Value
}

func (e *ZoomRenderParameter) SetValueInt(v int) int {
	dz := float64(v - e.Value)
	if dz < 0 {
		dz = 1.1
	} else if dz > 0 {
		dz = 0.9
	} else {
		return v
	}

	e.Value = v

	rMin := e.Model.Params[0].GetValueFloat64()
	iMin := e.Model.Params[1].GetValueFloat64()
	rMax := e.Model.Params[2].GetValueFloat64()
	//iMax := e.Model.Params[3].GetValueFloat64()
	width := e.Model.Params[5].GetValueInt()
	//height := e.Model.Params[6].GetValueInt()
	mouseX := e.Model.Params[10].GetValueFloat64()
	mouseY := e.Model.Params[11].GetValueFloat64()

	zwidth := rMax - rMin
	//zheight := iMax - iMin
	nzwidth := zwidth * dz
	//nzheight := zheight * dz

	cx := mouseX / float64(width)
	cy := mouseY / float64(width)

	nleft := rMin - ((nzwidth - zwidth) * cx)
	nright := nleft + nzwidth
	ntop := iMin - ((nzwidth - zwidth) * cy)
	nbottom := ntop + nzwidth
	e.Model.Params[0].SetValueFloat64(nleft)
	e.Model.Params[1].SetValueFloat64(ntop)
	e.Model.Params[2].SetValueFloat64(nright)
	e.Model.Params[3].SetValueFloat64(nbottom)
	e.Model.Params[4].SetValueInt(100 + int(math.Pow(1.1, float64(v))))

	e.Model.RequestPaint()

	return e.Value
}

func NewZoomRP(name string, value int, m *MandelModel) *ZoomRenderParameter {
	return &ZoomRenderParameter{
		EmptyParameter: rv.EmptyParameter{
			Name: name,
			Type: "int",
		},
		Value: value,
		Model: m,
	}
}
