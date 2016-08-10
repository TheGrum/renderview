package renderview

import (
	"image"
	"image/draw"
	"sync"
)

type RenderModel interface {
	Lock()
	Unlock()

	GetParameter(name string) RenderParameter
	Render(*image.RGBA)
	SetRequestPaintFunc(func())
}

type EmptyRenderModel struct {
	sync.Mutex

	Params       []RenderParameter
	RequestPaint func()
}

func (e *EmptyRenderModel) GetParameter(name string) RenderParameter {
	for _, p := range e.Params {
		if name == p.GetName() {
			return p
		}
	}

	return &EmptyParameter{}
}

func (e *EmptyRenderModel) Render(i *image.RGBA) {
	return
}

func (e *EmptyRenderModel) AddParameters(Params ...RenderParameter) {
	e.Params = append(e.Params, Params...)
}

func (e *EmptyRenderModel) SetRequestPaintFunc(f func()) {
	e.RequestPaint = f
}

type BasicRenderModel struct {
	EmptyRenderModel

	RequestRender chan interface{}
	NeedsRender   bool
	Rendering     bool
	Img           image.Image

	InnerRender func()
}

func (m *BasicRenderModel) Render(img *image.RGBA) {
	m.Lock()
	rendering := m.Rendering
	if !(m.Img == nil) && !(img == nil) {
		r := m.Img.Bounds()
		r2 := img.Bounds()
		mx := r.Dx()
		my := r.Dy()
		if r2.Dx() < mx {
			mx = r2.Dx()
		}
		if r2.Dy() < my {
			my = r2.Dy()
		}
		r3 := image.Rect(0, 0, mx, my)
		draw.Draw(img, r3, m.Img, image.ZP, draw.Src)
	}
	m.Unlock()
	if !rendering {
		m.RequestRender <- true
		m.NeedsRender = false
	} else {
		m.NeedsRender = true
	}
}

func (m *BasicRenderModel) GoRender() {
	for {
		select {
		case <-m.RequestRender:
			if !(m.InnerRender == nil) {
				m.InnerRender()
				if m.NeedsRender {
					m.NeedsRender = false
					m.InnerRender()
				}
			}
		}
	}
}

func NewBasicRenderModel() *BasicRenderModel {
	m := BasicRenderModel{
		EmptyRenderModel: EmptyRenderModel{
			Params: make([]RenderParameter, 0, 10),
		},
		RequestRender: make(chan interface{}, 10),
	}
	go m.GoRender()
	return &m
}
