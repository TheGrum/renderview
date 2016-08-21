package renderview

import (
	"image"
	"sync"
)

type RenderModel interface {
	Lock()
	Unlock()

	GetParameterNames() []string
	GetHintedParameterNames(hint int) []string
	GetHintedParameterNamesWithFallback(hint int) []string
	GetParameter(name string) RenderParameter
	Render() image.Image
	SetRequestPaintFunc(func())
	GetRequestPaintFunc() func()
}

type EmptyRenderModel struct {
	sync.Mutex

	Params       []RenderParameter
	RequestPaint func()
}

func (e *EmptyRenderModel) GetParameterNames() []string {
	s := make([]string, len(e.Params))
	for i := 0; i < len(e.Params); i++ {
		s[i] = e.Params[i].GetName()
	}
	return s
}

// GetHintedParameterNames returns a list of parameters with one of the passed hints
func (e *EmptyRenderModel) GetHintedParameterNames(hints int) []string {
	s := make([]string, 0, len(e.Params))
	for i := 0; i < len(e.Params); i++ {
	    if e.Params[i].GetHint()&hints > 0 {
		    s = append(s, e.Params[i].GetName())
		}
	}
	return s
}

// GetHintedParameterNamesWithFallback retrieves the names of parameters matching hints,
// if that is the empty set, it retrieves the names of parameters with no hints
func (e *EmptyRenderModel) GetHintedParameterNamesWithFallback(hints int) []string {
	s := make([]string, 0, len(e.Params))
	for i := 0; i < len(e.Params); i++ {
	    if e.Params[i].GetHint()&hints > 0 {
		    s = append(s, e.Params[i].GetName())
		}
	}
	if len(s) == 0 {
	    for i := 0; i < len(e.Params); i++ {
	        if e.Params[i].GetHint() == 0 {
		        s = append(s, e.Params[i].GetName())
		    }
	    }
	}
	return s
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

func (e *EmptyRenderModel) GetRequestPaintFunc() func() {
    return e.RequestPaint 
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
