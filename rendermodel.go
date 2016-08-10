package renderview

import (
	"image"
	"sync"
)

type RenderModel interface {
	Lock()
	Unlock()

	GetParameter(name string) RenderParameter
	Render() image.Image
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

func (e *EmptyRenderModel) Render() image.Image {
	return nil
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

func (m *BasicRenderModel) Render() image.Image {
	m.Lock()
	defer m.Unlock()
	rendering := m.Rendering
	if !rendering {
		m.RequestRender <- true
		m.NeedsRender = false
	} else {
		m.NeedsRender = true
	}
	return m.Img
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
