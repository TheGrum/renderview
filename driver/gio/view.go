// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build gio

package gio

import (
	"image"
	//"image/color"
	"image/draw"
	"log"
	//"fmt"

	rv "github.com/TheGrum/renderview"
	//"golang.org/x/exp/shiny/screen"
	//"golang.org/x/exp/shiny/widget/theme"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/paint"
	//"gioui.org/unit"
	"gioui.org/widget"
	//"gioui.org/widget/material"
	//"gioui.org/font/gofont"
)

// FrameBuffer sets up a Gio Window and runs a mainloop rendering the rv.RenderModel
func FrameBuffer(m rv.RenderModel) {
	MainLoop(m)
}

// Main sets up a Gio Window and runs a mainloop rendering the rv.RenderModel with widgets
func Main(m rv.RenderModel) {
	MainLoop(m)
	//MainLoopWithWidgets(m)
}

func MainLoop(r rv.RenderModel) {
	var needsPaint = false
	var mouseIsDown = false
	var dragging bool = false
	var dx, dy float64
	var sx, sy float32
	var wx, wy int

	var left, top, right, bottom, width, height, zoom rv.RenderParameter
	left = r.GetParameter("left")
	top = r.GetParameter("top")
	right = r.GetParameter("right")
	bottom = r.GetParameter("bottom")
	width = r.GetParameter("width")
	height = r.GetParameter("height")
	zoom = r.GetParameter("zoom")
	mouseX := r.GetParameter("mouseX")
	mouseY := r.GetParameter("mouseY")
	options := r.GetParameter("options")
	page := r.GetParameter("page")
	//	offsetX := r.GetParameter("offsetX")
	//	offsetY := r.GetParameter("offsetY")
	//fmt.Printf("left.GetType() %v", left.GetType())
	leftIsFloat64 := left.GetType() == "float64"
	zoomIsFloat64 := zoom.GetType() == "float64"

	w := app.NewWindow()
	go func() {
		//th := material.NewTheme()
		gtx := layout.NewContext(w.Queue())
		for e := range w.Events() {
			switch e := e.(type) {
			case system.DestroyEvent:
				return

			case system.FrameEvent:
				wx = e.Size.X
				wy = e.Size.Y
				width.SetValueInt(e.Size.X)
				height.SetValueInt(e.Size.Y)
				gtx.Reset(e.Config, e.Size)
				img := r.Render()
				if img != nil {
				ni := paint.NewImageOp(img)
                                ni.Add(gtx.Ops)
				po := paint.PaintOp{f32.Rectangle{f32.Point{0, 0}, f32.Point{float32(img.Bounds().Size().X), float32(img.Bounds().Size().Y)}}}
                                po.Add(gtx.Ops)
	//		form(gtx, th)
}
				e.Frame(gtx.Ops)

				//		case paint.Event:
				//			w.Upload(image.Point{}, buf, buf.Bounds())
				//			w.Publish()

			case key.Event:
				if e.Name == "⎋ " {
					//				return nil
				}
				if e.Name == "⇞ " {
					page.SetValueInt(page.GetValueInt() - 1)
					needsPaint = true
				}
				if e.Name == "⇟ " {
					page.SetValueInt(page.GetValueInt() + 1)
					needsPaint = true
				}

			case pointer.Event:
				//fmt.Printf("mouse pos(%v)\n", e)
				mouseX.SetValueFloat64(float64(e.Position.X))
				mouseY.SetValueFloat64(float64(e.Position.Y))

				if mouseIsDown == false && dragging == false && (e.Buttons&pointer.ButtonLeft) == pointer.ButtonLeft {
					//fmt.Printf("mouse down left(%v)\n", e)
					sx = e.Position.X
					sy = e.Position.Y
					mouseIsDown = true
				}
				if e.Scroll.Y > 0 {
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
							cx := float64(e.Position.X) / float64(wx)
							cy := float64(e.Position.Y) / float64(wy)
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
				if e.Scroll.Y < 0 {
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
							cx := float64(e.Position.X) / float64(wx)
							cy := float64(e.Position.Y) / float64(wy)
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
				if mouseIsDown {
					//				fmt.Printf("mouse drag(%v) dragging (%v)\n", e, dragging)
					if dragging == false {
								//			fmt.Printf("Checking %v, %v, %v, %v\n", e.Position.X, e.Position.Y, sx, sy)
						if ((e.Position.X - sx) > 3) || ((sx - e.Position.X) > 3) || ((e.Position.Y - sy) > 3) || ((sy - e.Position.Y) > 3) {
							dragging = true
								//					fmt.Printf("Dragging.\n")
						}
					} else {
						if leftIsFloat64 {
							width := right.GetValueFloat64() - left.GetValueFloat64()
							height := bottom.GetValueFloat64() - top.GetValueFloat64()
							dx = width / float64(wx)
							dy = height / float64(wy)
							cx := float64(e.Position.X-sx) * dx
							cy := float64(e.Position.Y-sy) * dy
							//fmt.Printf("dx %v dy %v cx %v cy %v\n", dx, dy, cx, cy)
							left.SetValueFloat64(left.GetValueFloat64() - cx)
							right.SetValueFloat64(right.GetValueFloat64() - cx)
							top.SetValueFloat64(top.GetValueFloat64() - cy)
							bottom.SetValueFloat64(bottom.GetValueFloat64() - cy)
//							fmt.Printf("left %v right %v top %v bottom %v", left.GetValueFloat64(), right.GetValueFloat64(), top.GetValueFloat64(), bottom.GetValueFloat64())
						} else {
							width := right.GetValueInt() - left.GetValueInt()
							height := bottom.GetValueInt() - top.GetValueInt()
							dx = float64(width) / float64(wx)
							dy = float64(height) / float64(wy)
							cx := float64(e.Position.X-sx) * dx
							cy := float64(e.Position.Y-sy) * dy
							//fmt.Printf("dx %v dy %v cx %v cy %v\n", dx, dy, cx, cy)
							left.SetValueInt(int(float64(left.GetValueInt()) - cx))
							right.SetValueInt(int(float64(right.GetValueInt()) - cx))
							top.SetValueInt(int(float64(top.GetValueInt()) - cy))
							bottom.SetValueInt(int(float64(bottom.GetValueInt()) - cy))
//							fmt.Printf("left %v right %v top %v bottom %v", left.GetValueInt(), right.GetValueInt(), top.GetValueInt(), bottom.GetValueInt())
						}
						ni := paint.NewImageOp(r.Render())
						ni.Add(gtx.Ops)
						po := paint.PaintOp{f32.Rectangle{f32.Point{0, 0}, f32.Point{float32(wx), float32(wy)}}}
						po.Add(gtx.Ops)
						w.Invalidate()
						//					Draw(r.Render(), buf.RGBA())

						sx = e.Position.X
						sy = e.Position.Y

					}
				}
				if e.Buttons == 0 {
					dragging = false
					mouseIsDown = false
				}

				//		case size.Event:
				//			if buf != nil {
				//				buf.Release()
				//			}
				//			r.GetParameter("width").SetValueInt(e.Size().X)
				//			r.GetParameter("height").SetValueInt(e.Size().Y)
				//			buf, err = s.NewBuffer(e.Size())
				//			if err != nil {
				//				log.Fatal(err)
				//			}
				//			Draw(r.Render(), buf.RGBA())
			default:

			}
			if needsPaint {
				needsPaint = false
				ni := paint.NewImageOp(r.Render())
				ni.Add(gtx.Ops)
				po := paint.PaintOp{f32.Rectangle{f32.Point{0, 0}, f32.Point{float32(wx), float32(wy)}}}
				po.Add(gtx.Ops)
				w.Invalidate()
			}
		}
	}()

	r.SetRequestPaintFunc(func() {
		if r == nil || w == nil {
			return
		}
		needsPaint = true
		w.Invalidate()
		//		w.Send(paint.Event{})
	})
	app.Main()
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

type RenderWidget struct {
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

	ParamEdits []*ParamEdit
}

func NewRenderWidget(r rv.RenderModel) *RenderWidget {
	w := &RenderWidget{
		r: r,
	}
	//w.Wrapper = w
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

	return w
}

func WidgetMainLoop(r rv.RenderModel) {
//	w := GetRenderWidgetWithSidebar(r)
//	if err := widget.RunWindow(s, w, nil); err != nil {
//		log.Fatal(err)
//	}
}

/*
func GetRenderWidgetWithSidebar(r rv.RenderModel) []func() {
	widgets := []func(){}
	//widgets = append(widgets, func() {
	//	layout.Flex{Axis: layout.Vertical}
	//}
	flow := layout.Flex{Axis: layout.Vertical}
	names := r.GetParameterNames()

	for i := 0; i < len(names); i++ {
		flow.Layout
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
*/

type ParamEdits struct {
	P []ParamEdit
}

type ParamEdit struct {
	R rv.RenderModel
	P rv.RenderParameter
	N widget.Editor
}
