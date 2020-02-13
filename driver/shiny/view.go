// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build android shiny

package shiny

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	rv "github.com/TheGrum/renderview"

	"golang.org/x/exp/shiny/widget/theme"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/unit"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

// FrameBuffer sets up a Shiny screen and runs a mainloop rendering the rv.RenderModel
func FrameBuffer(m rv.RenderModel) {
	driver.Main(GetMainLoop(m))
}

// Main sets up a Shiny screen and runs a mainloop rendering the rv.RenderModel; with widgets
// when widgets are functional in Shiny
func Main(m rv.RenderModel) {
	//driver.Main(GetMainLoopWithWidgets(m))
	driver.Main(GetMainLoop(m))
}

func GetMainLoop(r rv.RenderModel) func(s screen.Screen) {
	return func(s screen.Screen) {
		MainLoop(s, r)
	}
}

func MainLoop(s screen.Screen, r rv.RenderModel) {
	var needsPaint = false
	var mouseIsDown = false
	var dragging bool = false
	var dx, dy float64
	var sx, sy float32

	var left, top, right, bottom, zoom rv.RenderParameter
	left = r.GetParameter("left")
	top = r.GetParameter("top")
	right = r.GetParameter("right")
	bottom = r.GetParameter("bottom")
	zoom = r.GetParameter("zoom")
	mouseX := r.GetParameter("mouseX")
	mouseY := r.GetParameter("mouseY")
	options := r.GetParameter("options")
	page := r.GetParameter("page")
	//	offsetX := r.GetParameter("offsetX")
	//	offsetY := r.GetParameter("offsetY")

	leftIsFloat64 := left.GetType() == "float64"
	zoomIsFloat64 := zoom.GetType() == "float64"

	w, err := s.NewWindow(nil)
	if err != nil {
		handleError(err)
		return
	}

	buf := screen.Buffer(nil)
	defer func() {
		if buf != nil {
			buf.Release()
		}
		w.Release()
	}()

	r.SetRequestPaintFunc(func() {
		if buf == nil || r == nil || w == nil {
			return
		}
		needsPaint = true
		w.Send(paint.Event{})
	})

	for {
		switch e := w.NextEvent().(type) {
		case lifecycle.Event:
			if e.To == lifecycle.StageDead {
				return
			}

		case paint.Event:
			w.Upload(image.Point{}, buf, buf.Bounds())
			w.Publish()

		case key.Event:
			if e.Code == key.CodeEscape {
				return
			}
			if e.Code == key.CodePageUp && e.Direction == key.DirPress {
				page.SetValueInt(page.GetValueInt() - 1)
				needsPaint = true
			}
			if e.Code == key.CodePageDown && e.Direction == key.DirPress {
				page.SetValueInt(page.GetValueInt() + 1)
				needsPaint = true
			}

		case mouse.Event:
			//fmt.Printf("mouse pos(%v)\n", e)
			mouseX.SetValueFloat64(float64(e.X))
			mouseY.SetValueFloat64(float64(e.Y))

			if dragging == false && e.Direction == mouse.DirPress && e.Button == mouse.ButtonLeft {
				//				fmt.Printf("mouse down left(%v)\n", e)
				sx = e.X
				sy = e.Y
				mouseIsDown = true
			}
			if e.Button == mouse.ButtonWheelDown {
				if zoomIsFloat64 {
					zoom.SetValueFloat64(zoom.GetValueFloat64() - 1)
				} else {
					zoom.SetValueInt(zoom.GetValueInt() - 1)
				}
				if options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
					mult := 1 + rv.ZOOM_RATE
					if leftIsFloat64 {

						zwidth := right.GetValueFloat64() - left.GetValueFloat64()
						zheight := bottom.GetValueFloat64() - top.GetValueFloat64()
						nzwidth := zwidth * mult
						nzheight := zheight * mult
						cx := float64(e.X) / float64(buf.Size().X)
						cy := float64(e.Y) / float64(buf.Size().Y)
						if options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
							cx = 0.5
							cy = 0.5
						}
						//fmt.Printf("zoomOut: mult: %v zwidth: %v nzwidth: %v cx: %v left: %v nleft: %v\n", mult, zwidth, nzwidth, cx, left.GetValueFloat64(), left.GetValueFloat64()-((nzwidth-zwidth)*cx))
						left.SetValueFloat64(left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
						top.SetValueFloat64(top.GetValueFloat64() - ((nzheight - zheight) * cy))
						right.SetValueFloat64(left.GetValueFloat64() + nzwidth)
						bottom.SetValueFloat64(top.GetValueFloat64() + nzheight)
						needsPaint = true

					}
				}

			}
			if e.Button == mouse.ButtonWheelUp {
				if zoomIsFloat64 {
					zoom.SetValueFloat64(zoom.GetValueFloat64() + 1)
				} else {
					zoom.SetValueInt(zoom.GetValueInt() + 1)
				}
				if options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
					mult := 1 - rv.ZOOM_RATE
					if leftIsFloat64 {
						zwidth := right.GetValueFloat64() - left.GetValueFloat64()
						zheight := bottom.GetValueFloat64() - top.GetValueFloat64()
						nzwidth := zwidth * mult
						nzheight := zheight * mult
						cx := float64(e.X) / float64(buf.Size().X)
						cy := float64(e.Y) / float64(buf.Size().Y)
						if options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
							cx = 0.5
							cy = 0.5
						}
						left.SetValueFloat64(left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
						top.SetValueFloat64(top.GetValueFloat64() - ((nzheight - zheight) * cy))
						right.SetValueFloat64(left.GetValueFloat64() + nzwidth)
						bottom.SetValueFloat64(top.GetValueFloat64() + nzheight)
						needsPaint = true
					}
				}
			}
			if e.Direction == mouse.DirNone && mouseIsDown {
				//				fmt.Printf("mouse drag(%v) dragging (%v)\n", e, dragging)
				if dragging == false {
					//					fmt.Printf("Checking %v, %v, %v, %v\n", e.X, e.Y, sx, sy)
					if ((e.X - sx) > 3) || ((sx - e.X) > 3) || ((e.Y - sy) > 3) || ((sy - e.Y) > 3) {
						dragging = true
						//						fmt.Printf("Dragging.\n")
					}
				} else {
					if leftIsFloat64 {
						width := right.GetValueFloat64() - left.GetValueFloat64()
						height := bottom.GetValueFloat64() - top.GetValueFloat64()
						dx = width / float64(buf.Size().X)
						dy = height / float64(buf.Size().Y)
						cx := float64(e.X-sx) * dx
						cy := float64(e.Y-sy) * dy
						//						fmt.Printf("dx %v dy %v cx %v cy %v\n", dx, dy, cx, cy)
						left.SetValueFloat64(left.GetValueFloat64() - cx)
						right.SetValueFloat64(right.GetValueFloat64() - cx)
						top.SetValueFloat64(top.GetValueFloat64() - cy)
						bottom.SetValueFloat64(bottom.GetValueFloat64() - cy)
					} else {
						width := right.GetValueInt() - left.GetValueInt()
						height := bottom.GetValueInt() - top.GetValueInt()
						dx = float64(width) / float64(buf.Size().X)
						dy = float64(height) / float64(buf.Size().Y)
						cx := float64(e.X-sx) * dx
						cy := float64(e.Y-sy) * dy
						left.SetValueInt(int(float64(left.GetValueInt()) - cx))
						right.SetValueInt(int(float64(right.GetValueInt()) - cx))
						top.SetValueInt(int(float64(top.GetValueInt()) - cy))
						bottom.SetValueInt(int(float64(bottom.GetValueInt()) - cy))
					}
					Draw(r.Render(), buf.RGBA())

					sx = e.X
					sy = e.Y

				}
			}
			if e.Direction == mouse.DirRelease {
				dragging = false
				mouseIsDown = false
			}

		case size.Event:
			if buf != nil {
				buf.Release()
			}
			r.GetParameter("width").SetValueInt(e.Size().X)
			r.GetParameter("height").SetValueInt(e.Size().Y)
			buf, err = s.NewBuffer(e.Size())
			if err != nil {
				log.Fatal(err)
			}
			Draw(r.Render(), buf.RGBA())
		default:

		}
		if needsPaint {
			needsPaint = false
			Draw(r.Render(), buf.RGBA())
			w.Send(paint.Event{})
		}
	}
}

func Draw(mimg image.Image, bimg *image.RGBA) {
	if !(mimg == nil) && !(bimg == nil) {
		r := mimg.Bounds()
		r2 := bimg.Bounds()
		mx := r.Dx()
		my := r.Dy()
		if r2.Dx() < mx {
			mx = r2.Dx()
		}
		if r2.Dy() < my {
			my = r2.Dy()
		}
		r3 := image.Rect(0, 0, mx, my)
		draw.Draw(bimg, r3, mimg, image.ZP, draw.Src)
	}
}

func handleError(e error) {
	log.Fatal(e)
}

func GetWidgetMainLoop(r rv.RenderModel) func(s screen.Screen) {
	return func(s screen.Screen) {
		WidgetMainLoop(s, r)
	}
}

func expand(n node.Node, expandAlongWeight int) node.Node {
	return widget.WithLayoutData(n, widget.FlowLayoutData{
		ExpandAcross:      true,
		AlongWeight: expandAlongWeight,
	})
}

type RenderWidget struct {
	node.LeafEmbed
	index int
	r     rv.RenderModel

	left,
	top,
	right,
	bottom,
	width,
	height,
	zoom,
	mouseX,
	mouseY,
	options,
	page rv.RenderParameter

	sx,
	sy float32

	leftIsFloat64,
	zoomIsFloat64,
	mouseIsDown,
	dragging bool
}

func NewRenderWidget(r rv.RenderModel) *RenderWidget {
	w := &RenderWidget{
		r: r,
	}
	w.Wrapper = w
	w.left = r.GetParameter("left")
	w.top = r.GetParameter("top")
	w.right = r.GetParameter("right")
	w.bottom = r.GetParameter("bottom")
	w.width = r.GetParameter("width")
	w.height = r.GetParameter("height")
	w.zoom = r.GetParameter("zoom")
	w.mouseX = r.GetParameter("mouseX")
	w.mouseY = r.GetParameter("mouseY")
	w.options = r.GetParameter("options")
	w.page = r.GetParameter("page")
	w.leftIsFloat64 = w.left.GetType() == "float64"
	w.zoomIsFloat64 = w.zoom.GetType() == "float64"
	r.SetRequestPaintFunc(func() {
		w.Mark(node.MarkNeedsPaintBase)
	})
	return w
}

func (m *RenderWidget) OnInputEvent(e interface{}, origin image.Point) node.EventHandled {
	switch e := e.(type) {
	case key.Event:
		if e.Code == key.CodeEscape {
			return node.NotHandled
		}
		if e.Code == key.CodePageUp && e.Direction == key.DirPress {
			m.page.SetValueInt(m.page.GetValueInt() - 1)
			m.Mark(node.MarkNeedsPaintBase)
		}
		if e.Code == key.CodePageDown && e.Direction == key.DirPress {
			m.page.SetValueInt(m.page.GetValueInt() + 1)
			m.Mark(node.MarkNeedsPaintBase)
		}

	case mouse.Event:
		//fmt.Printf("mouse pos(%v)\n", e)
		m.mouseX.SetValueFloat64(float64(e.X))
		m.mouseY.SetValueFloat64(float64(e.Y))

		if m.dragging == false && e.Direction == mouse.DirPress && e.Button == mouse.ButtonLeft {
			//				fmt.Printf("mouse down left(%v)\n", e)
			m.sx = e.X
			m.sy = e.Y
			m.mouseIsDown = true
		}
		if e.Button == mouse.ButtonWheelDown {
			if m.zoomIsFloat64 {
				m.zoom.SetValueFloat64(m.zoom.GetValueFloat64() - 1)
			} else {
				m.zoom.SetValueInt(m.zoom.GetValueInt() - 1)
			}
			if m.options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || m.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
				mult := 1 + rv.ZOOM_RATE
				if m.leftIsFloat64 {

					zwidth := m.right.GetValueFloat64() - m.left.GetValueFloat64()
					zheight := m.bottom.GetValueFloat64() - m.top.GetValueFloat64()
					nzwidth := zwidth * mult
					nzheight := zheight * mult
					cx := float64(e.X) / float64(m.width.GetValueInt())
					cy := float64(e.Y) / float64(m.height.GetValueInt())
					if m.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
						cx = 0.5
						cy = 0.5
					}
					//fmt.Printf("zoomOut: mult: %v zwidth: %v nzwidth: %v cx: %v left: %v nleft: %v\n", mult, zwidth, nzwidth, cx, left.GetValueFloat64(), left.GetValueFloat64()-((nzwidth-zwidth)*cx))
					m.left.SetValueFloat64(m.left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
					m.top.SetValueFloat64(m.top.GetValueFloat64() - ((nzheight - zheight) * cy))
					m.right.SetValueFloat64(m.left.GetValueFloat64() + nzwidth)
					m.bottom.SetValueFloat64(m.top.GetValueFloat64() + nzheight)

					m.Mark(node.MarkNeedsPaintBase)

				}
			}

		}
		if e.Button == mouse.ButtonWheelUp {
			if m.zoomIsFloat64 {
				m.zoom.SetValueFloat64(m.zoom.GetValueFloat64() + 1)
			} else {
				m.zoom.SetValueInt(m.zoom.GetValueInt() + 1)
			}
			if m.options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || m.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
				mult := 1 - rv.ZOOM_RATE
				if m.leftIsFloat64 {
					zwidth := m.right.GetValueFloat64() - m.left.GetValueFloat64()
					zheight := m.bottom.GetValueFloat64() - m.top.GetValueFloat64()
					nzwidth := zwidth * mult
					nzheight := zheight * mult
					cx := float64(e.X) / float64(m.width.GetValueInt())

					cy := float64(e.Y) / float64(m.height.GetValueInt())
					if m.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
						cx = 0.5
						cy = 0.5
					}
					m.left.SetValueFloat64(m.left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
					m.top.SetValueFloat64(m.top.GetValueFloat64() - ((nzheight - zheight) * cy))
					m.right.SetValueFloat64(m.left.GetValueFloat64() + nzwidth)
					m.bottom.SetValueFloat64(m.top.GetValueFloat64() + nzheight)
					m.Mark(node.MarkNeedsPaintBase)
				}
			}
		}
		if e.Direction == mouse.DirNone && m.mouseIsDown {
			//				fmt.Printf("mouse drag(%v) dragging (%v)\n", e, dragging)
			if m.dragging == false {
				//					fmt.Printf("Checking %v, %v, %v, %v\n", e.X, e.Y, sx, sy)
				if ((e.X - m.sx) > 3) || ((m.sx - e.X) > 3) || ((e.Y - m.sy) > 3) || ((m.sy - e.Y) > 3) {
					m.dragging = true
					//						fmt.Printf("Dragging.\n")
				}
			} else {
				if m.leftIsFloat64 {
					width := m.right.GetValueFloat64() - m.left.GetValueFloat64()
					height := m.bottom.GetValueFloat64() - m.top.GetValueFloat64()
					dx := width / float64(m.width.GetValueInt())
					dy := height / float64(m.height.GetValueInt())
					cx := float64(e.X-m.sx) * dx
					cy := float64(e.Y-m.sy) * dy
					//						fmt.Printf("dx %v dy %v cx %v cy %v\n", dx, dy, cx, cy)
					m.left.SetValueFloat64(m.left.GetValueFloat64() - cx)
					m.right.SetValueFloat64(m.right.GetValueFloat64() - cx)
					m.top.SetValueFloat64(m.top.GetValueFloat64() - cy)
					m.bottom.SetValueFloat64(m.bottom.GetValueFloat64() - cy)
				} else {
					width := m.right.GetValueInt() - m.left.GetValueInt()
					height := m.bottom.GetValueInt() - m.top.GetValueInt()
					dx := float64(width) / float64(m.Rect.Dx())
					dy := float64(height) / float64(m.Rect.Dy())
					cx := float64(e.X-m.sx) * dx
					cy := float64(e.Y-m.sy) * dy
					m.left.SetValueInt(int(float64(m.left.GetValueInt()) - cx))
					m.right.SetValueInt(int(float64(m.right.GetValueInt()) - cx))
					m.top.SetValueInt(int(float64(m.top.GetValueInt()) - cy))
					m.bottom.SetValueInt(int(float64(m.bottom.GetValueInt()) - cy))
				}
				m.Mark(node.MarkNeedsPaintBase)
				//			Draw(r.Render(), buf.RGBA())

				m.sx = e.X
				m.sy = e.Y

			}
		}
		if e.Direction == mouse.DirRelease {
			m.dragging = false
			m.mouseIsDown = false
		}

	}
	return node.NotHandled
}

func (m *RenderWidget) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	m.width.SetValueInt(m.Rect.Dx())
	m.height.SetValueInt(m.Rect.Dy())
	m.Marks.UnmarkNeedsPaintBase()
	Draw(m.r.Render(), ctx.Dst)
	return nil
}

func WidgetMainLoop(s screen.Screen, r rv.RenderModel) {
	w := GetRenderWidgetWithSidebar(r)
	if err := widget.RunWindow(s, w, nil); err != nil {
		log.Fatal(err)
	}
}

func GetRenderWidgetWithSidebar(r rv.RenderModel) node.Node {
	sideflow := widget.NewFlow(widget.AxisVertical)
	names := r.GetParameterNames()

	for i := 0; i < len(names); i++ {
		sideflow.Insert(widget.NewLabel(names[i]), nil)
		//e := widget.NewLabel("test")
		//e := editor.NewEditor(inconsolata.Regular8x16, nil)
		//e := NewTextEdit(r.GetParameter(names[i]).GetValueString(), inconsolata.Regular8x16, nil)
		//e := NewTextEdit("text", inconsolata.Regular8x16, nil)
		//		e.Text = r.GetParameter(names[i]).GetValueString()
		//e.Text = "Test"
		//e.Rect = image.Rectangle{image.ZP, image.Point{50, 16}}
		//		w := expand(e, 0
		//w := widget.NewSizer(unit.Pixels(50), unit.Pixels(30), e)
		//			widget.NewLabel(r.GetParameter(names[i]).GetValueString()))
		//sideflow.Insert(w, nil)
	}
	sidebar := widget.NewUniform(theme.StaticColor(color.RGBA{0xff, 0xff, 0xff, 0xff}), sideflow)

	//	sidebar := widget.NewUniform(theme.Neutral,
	//		widget.NewPadder(widget.AxisBoth, unit.Ems(0.5),
	//			sideflow,
	//		),
	//	)
	divider := widget.NewSizer(unit.Value{}, unit.DIPs(2),
		widget.NewUniform(theme.StaticColor(color.RGBA{0xbf, 0xbf, 0xb0, 0xff}), nil))

	body := widget.NewUniform(theme.StaticColor(color.RGBA{0xbf, 0xbf, 0xb0, 0xff}), NewRenderWidget(r))

	w := widget.NewFlow(widget.AxisHorizontal,
		expand(widget.NewSheet(sidebar), 0),
		expand(widget.NewSheet(divider), 0),
		expand(widget.NewSheet(body), 1),
	)

	return w
}

type ParamEdits struct {
	P []ParamEdit
}

type ParamEdit struct {
	R rv.RenderModel
	P rv.RenderParameter
	N node.Node
}
