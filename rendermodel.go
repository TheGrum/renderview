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

	requestRender chan interface{}
	needsRender   bool
	rendering     bool
	img           image.Image

	InnerRender func()
}

func (m *BasicRenderModel) Render(img *image.RGBA) {
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

func (m *BasicRenderModel) GoRender() {
	for {
		select {
		case <-m.requestRender:
			if !(m.InnerRender == nil) {
				m.InnerRender()
				if m.needsRender {
					m.needsRender = false
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
		requestRender: make(chan interface{}, 10),
	}
	go m.GoRender()
	return &m
}
