// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package renderview

import (
	"image"
	"math"
	"sync"
)

// RenderModel is the interface you will implement to stand between your visualization code
// and the RenderView. You are primarily responsible for providing a set of RenderParameters
// and providing an Image upon request via Render.
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

// EmptyRenderModel concretizes the most important elements of the RenderModel, the bag of Parameters (Params)
// and the RequestPaint function (which the view sets) - call this latter function to inform the view
// that you have provided a new image or set of information needing a render. It is not usable as a RenderModel
// by itself, as the implementation of Render simply returns nil. Embed it in your RenderModel struct.
type EmptyRenderModel struct {
	sync.Mutex

	Params       []RenderParameter
	RequestPaint func()
}

// GetParameterNames returns a list of valid parameter names
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

// GetParameter returns a named parameter. If you implement your own RenderModel from scratch,
// without using the EmptyRenderModel as a basis, you must either include ALL the default
// parameters, or duplicate the behavior of EmptyRenderModel in returning an EmptyParameter
// when a non-existent parameter is requested.
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

// AddParameters accepts any number of parameters and adds them to the Params bag.
// It does not do ANY checking for duplication!
func (e *EmptyRenderModel) AddParameters(Params ...RenderParameter) {
	e.Params = append(e.Params, Params...)
}

// Included for completeness. In general, there is no need for your code to use
// the RenderModel interface instead of a concrete form, so you can simply
// access e.RequestPaint directly.
func (e *EmptyRenderModel) GetRequestPaintFunc() func() {
	return e.RequestPaint
}

// Used by the RenderView to supply a function you can call to inform the view
// that it should perform a repaint.
func (e *EmptyRenderModel) SetRequestPaintFunc(f func()) {
	e.RequestPaint = f
}

/*
// EmptyRenderModel is not functional by itself
func NewEmptyRenderModel() *EmptyRenderModel {
	return &EmptyRenderModel{
		Params: make([]RenderParameter, 0, 10),
	}
}*/

// InitializeEmptyRenderModel should be called to initialize the EmptyRenderModel
// when embedded in your own struct.
func InitializeEmptyRenderModel(m *EmptyRenderModel) {
	m.Params = make([]RenderParameter, 0, 10)
}

// BasicRenderModel should suffice for many users, and can be embedded to provide its
// functionality to your own models. It provides an easy way to attach your own
// rendering implementation that will be called in a separate goroutine.
type BasicRenderModel struct {
	EmptyRenderModel

	RequestRender chan interface{}
	NeedsRender   bool
	Rendering     bool
	Img           image.Image

	started bool

	InnerRender func()
}

// Called by RenderView
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

// GoRender is called by Start and calls your provided InnerRender function when needed.
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

// Start only needs to be called if you have embedded BasicRenderModel in your own struct.
func (m *BasicRenderModel) Start() {
	if !m.started {
		m.started = true
		go m.GoRender()
	}
}

func NewBasicRenderModel() *BasicRenderModel {
	m := BasicRenderModel{
		EmptyRenderModel: EmptyRenderModel{
			Params: make([]RenderParameter, 0, 10),
		},
		RequestRender: make(chan interface{}, 10),
	}
	m.started = true
	go m.GoRender()
	return &m
}

// Use Initialize to set up a BasicRenderModel when you have embedded it in
// your own model
// Remember to add a go m.GoRender() or call Start()
func InitializeBasicRenderModel(m *BasicRenderModel) {
	m.Params = make([]RenderParameter, 0, 10)
	m.RequestRender = make(chan interface{}, 10)
}

func DefaultParameters(useint bool, hint int, options int, left float64, top float64, right float64, bottom float64) []RenderParameter {
	if useint {
		return SetHints(hint,
			NewIntRP("left", int(math.Floor(left))),
			NewIntRP("top", int(math.Floor(top))),
			NewIntRP("right", int(math.Floor(right))),
			NewIntRP("bottom", int(math.Floor(bottom))),
			NewIntRP("width", 100),
			NewIntRP("height", 100),
			NewIntRP("options", options))
	} else {
		return SetHints(hint,
			NewFloat64RP("left", left),
			NewFloat64RP("top", top),
			NewFloat64RP("right", right),
			NewFloat64RP("bottom", bottom),
			NewIntRP("width", 100),
			NewIntRP("height", 100),
			NewIntRP("options", options))
	}
}
