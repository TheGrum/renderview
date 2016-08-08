package mandelbrot

import (
	"image"
	"image/draw"
	rv "renderview"
)

type MandelView struct {
	rv.EmptyRenderModel

	requestRender chan interface{}
	needsRender   bool
	rendering     bool
	img           image.Image
}

func (m *MandelView) Render(img *image.RGBA) {
	m.Lock()
	rendering := m.rendering
	if !(m.img == nil) {
		r := m.img.Bounds()
		draw.Draw(img, r, m.img, image.ZP, draw.Src)
	}
	m.Unlock()
	if !rendering {
		m.requestRender <- true
		m.needsRender = false
	} else {
		m.needsRender = true
	}
}

func (m *MandelView) GoRender() {
	for {
		select {
		case <-m.requestRender:
			m.innerRender()
			if m.needsRender {
				m.needsRender = false
				m.innerRender()
			}
		}
	}
}

func (m *MandelView) innerRender() {
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
	m.rendering = true
	m.Unlock()

	i2 := generateMandelbrot(rMin, iMin, rMax, iMax, width, red, green, blue, maxEsc)

	m.Lock()
	m.img = i2
	m.rendering = false
	m.Unlock()
	if !(m.RequestPaint == nil) {
		m.RequestPaint()
	}
}

func NewMandelView() *MandelView {
	m := MandelView{
		EmptyRenderModel: rv.EmptyRenderModel{
			Params: make([]rv.RenderParameter, 0, 10),
		},
		requestRender: make(chan interface{}, 10),
	}
	m.AddParameters(rv.NewFloat64RP("left", -2), rv.NewFloat64RP("top", -1), rv.NewFloat64RP("right", 0.5), rv.NewFloat64RP("bottom", 1), rv.NewIntRP("maxEsc", 100), rv.NewIntRP("width", 100), rv.NewIntRP("height", 100), rv.NewIntRP("red", 230), rv.NewIntRP("green", 235), rv.NewIntRP("blue", 255))
	go m.GoRender()
	return &m
}
