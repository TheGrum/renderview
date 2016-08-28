// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver"

	"github.com/llgcode/draw2d/draw2dimg"
)

type ff float64

func main() {
	sig := ""
	rand.Seed(time.Now().UnixNano())
	m := rv.NewBasicRenderModel()
	m.AddParameters(
		rv.SetHints(rv.HINT_HIDE,
			rv.NewIntRP("width", 0),
			rv.NewIntRP("height", 0),
		)...)
	m.AddParameters(
		rv.NewIntRP("page", 0),
		rv.NewIntRP("linewidth", 1),
		rv.NewIntRP("cellwidth", 5),
		rv.NewIntRP("mazewidth", 100),
		rv.NewIntRP("mazeheight", 100))
	m.InnerRender = func() {
		s := GetSignature(m)
		if sig != s {
			z := NewDepthFirstMaze(m.Params[5].GetValueInt(), m.Params[6].GetValueInt())
			m.Img = RenderMaze(m, z)
			sig = s
			m.RequestPaint()
		}
	}
	//driver.Main(rv.GetWidgetMainLoop(m))
	//rv.GtkWindowWithWidgetsInit(m)
	driver.Main(m)
}

func GetSignature(r *rv.BasicRenderModel) string {
	s := ""
	for i := 0; i < len(r.Params); i++ {
		p := r.Params[i]
		switch p.GetType() {
		case "int":
			s = fmt.Sprintf("%s%v", s, p.GetValueInt())
		case "uint32":
			s = fmt.Sprintf("%s%v", s, p.GetValueUInt32())
		case "float64":
			s = fmt.Sprintf("%s%v", s, p.GetValueFloat64())
		case "complex128":
			s = fmt.Sprintf("%s%v", s, p.GetValueComplex128())
		case "string":
			s = fmt.Sprintf("%s%v", s, p.GetValueString())
		}
	}
	return s
}

func RenderMaze(r rv.RenderModel, m *Maze) image.Image {
	//	w := m.width
	//	h := m.height
	mw := r.GetParameter("mazewidth").GetValueInt()
	mh := r.GetParameter("mazeheight").GetValueInt()
	lw := r.GetParameter("linewidth").GetValueInt()
	cw := r.GetParameter("cellwidth").GetValueInt()

	iw := (mw * cw) + ((mw + 2) * lw)
	ih := (mh * cw) + ((mh + 2) * lw)

	bounds := image.Rect(0, 0, iw, ih)
	b := image.NewRGBA(bounds)

	gc := draw2dimg.NewGraphicContext(b)
	gc.SetFillColor(color.White)
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(float64(lw))

	for x := 0; x < mw; x++ {
		for y := 0; y < mh; y++ {
			v := m.C(x, y)
			if v&W_NORTH == W_NORTH {
				gc.MoveTo(float64(x*(cw+(lw))), float64(y*(cw+(lw))))
				gc.LineTo(float64((x+1)*(cw+(lw))), float64(y*(cw+(lw))))
			}
			if v&W_EAST == W_EAST {
				gc.MoveTo(float64((x+1)*(cw+(lw))), float64(y*(cw+(lw))))
				gc.LineTo(float64((x+1)*(cw+(lw))), float64((y+1)*(cw+(lw))))
			}

			if v&W_SOUTH == W_SOUTH {
				gc.MoveTo(float64(x*(cw+(lw))), float64((y+1)*(cw+(lw))))
				gc.LineTo(float64((x+1)*(cw+(lw))), float64((y+1)*(cw+(lw))))
			}

			if v&W_WEST == W_WEST {
				gc.MoveTo(float64(x*(cw+(lw))), float64(y*(cw+(lw))))
				gc.LineTo(float64(x*(cw+(lw))), float64((y+1)*(cw+(lw))))
			}
		}
	}

	gc.Stroke()

	return b
}
