// Copyright 2020 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build gio

package gio

import (
	"image"

	//"image/color"
	"image/draw"
	"log"

	rv "github.com/TheGrum/renderview"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget/material"

	"gioui.org/unit"
	"gioui.org/widget"
	//"gioui.org/widget/material"
)

// FrameBuffer sets up a Gio Window and runs a mainloop rendering the rv.RenderModel
func FrameBuffer(m rv.RenderModel) {
	MainLoop(m)
}

// Main sets up a Gio Window and runs a mainloop rendering the rv.RenderModel with widgets
func Main(m rv.RenderModel) {
	//MainLoop(m)
	MainLoopWithWidgets(m)
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

const SIDEBAR_WIDTH = 160

func MainLoopWithWidgets(r rv.RenderModel) {
	var needsPaint = true
	var editorsChanged = true
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
	sidebarWidth := r.GetParameter("sidebarWidth")
	//	offsetX := r.GetParameter("offsetX")
	//	offsetY := r.GetParameter("offsetY")
	//fmt.Printf("left.GetType() %v", left.GetType())
	leftIsFloat64 := left.GetType() == "float64"
	zoomIsFloat64 := zoom.GetType() == "float64"
	paramEditors := []ParamEdit{}
	var fullTextEditor ParamEdit
	paramList := &layout.List{
		Axis: layout.Vertical,
	}
	lx := 0
	sbw := 0
	gofont.Register()

	pnames := r.GetHintedParameterNames(rv.HINT_FULLTEXT)
	if len(pnames) > 0 {
		param := r.GetParameter(pnames[0])
		fullTextEditor = ParamEdit{
			P: param,
			N: new(widget.Editor),
		}
		fullTextEditor.N.SetText(rv.GetParameterValueAsString(fullTextEditor.P))
	}
	for _, pname := range r.GetHintedParameterNamesWithFallback(rv.HINT_SIDEBAR | rv.HINT_FOOTER) {
		param := r.GetParameter(pname)
		paramEdit := ParamEdit{
			P: param,
			N: &widget.Editor{
				SingleLine: true,
			}}
		paramEditors = append(paramEditors, paramEdit)
		paramEdit.N.SetText(rv.GetParameterValueAsString(paramEdit.P))
	}

	w := app.NewWindow()
	go func() {
		th := material.NewTheme()
		gtx := layout.NewContext(w.Queue())
		for e := range w.Events() {
			if editorsChanged == true {
				// may also need parameters updated
				if fullTextEditor.N != nil {
					fullTextEditor.N.SetText(rv.GetParameterValueAsString(fullTextEditor.P))
				}
				for _, pe := range paramEditors {
					pe.N.SetText(rv.GetParameterValueAsString(pe.P))
				}
				editorsChanged = false
			}
			switch e := e.(type) {
			case system.DestroyEvent:
				return

			case system.FrameEvent:
				wx = e.Size.X
				wy = e.Size.Y
				if len(paramEditors) > 0 {
					sbw = sidebarWidth.GetValueInt()
					if sbw == 0 {
						sbw = SIDEBAR_WIDTH
						if fullTextEditor.N == nil {
							// just the sidebar - if we have one?
							//width.SetValueInt(e.Size.X - sbw)
							lx = sbw
						} else {
							//width.SetValueInt(e.Size.X - sbw - (SIDEBAR_WIDTH * 2))
							lx = sbw + (SIDEBAR_WIDTH * 2)
						}
					}
				}
				//	fmt.Printf("lx: %v", lx)
				width.SetValueInt(e.Size.X)
				height.SetValueInt(e.Size.Y)
				gtx.Reset(e.Config, e.Size)
				gtx.Constraints.Width.Max = sbw
				widgetList := []func(){}
				widgetList = append(widgetList, func() {
					th.Label(unit.Dp(15), "____Parameters____").Layout(gtx)
				})
				for _, pe := range paramEditors {
					//fmt.Printf("pe: %v %v %v\n", pe, pe.P.GetName(), pe.N.Text())
					widgetList = append(widgetList, func(pe ParamEdit) func() {
						return func() {
							th.Label(unit.Dp(15), pe.P.GetName()).Layout(gtx)
						}
					}(pe))
					widgetList = append(widgetList, func(pe ParamEdit) func() {
						return func() {
							th.Editor(pe.P.GetName()).Layout(gtx, pe.N)
							for range pe.N.Events(gtx) {
								rv.SetParameterValueFromString(pe.P, pe.N.Text())
							}
						}
					}(pe))
				}
				paramList.Layout(gtx, len(widgetList), func(i int) {
					layout.UniformInset(unit.Dp(1)).Layout(gtx, widgetList[i])
				})
				//gtx.Reset(e.Config, e.Size)
				if fullTextEditor.N != nil {
					gtx.Constraints.Width.Max = sbw + (SIDEBAR_WIDTH * 2) - 5
					layout.Inset{Top: unit.Dp(2), Left: unit.Dp(float32(sbw + 5))}.Layout(gtx, func() {
						layout.Stack{Alignment: layout.N}.Layout(gtx,
							//							layout.Stacked(func() {
							//								th.Label(unit.Dp(15), fullTextEditor.P.GetName()).Layout(gtx)
							//							}),
							layout.Expanded(func() {
								th.Editor(fullTextEditor.P.GetName()).Layout(gtx, fullTextEditor.N)
								//fullTextEditor.N.SetText(fullTextEditor.P.GetValueString())
								for range fullTextEditor.N.Events(gtx) {
									rv.SetParameterValueFromString(fullTextEditor.P, fullTextEditor.N.Text())
								}
							}))
					})
				}
				//gtx.Reset(e.Config, e.Size)
				gtx.Constraints.Width.Max = e.Size.X
				img := r.Render()
				if img != nil {
					ni := paint.NewImageOp(img)
					ni.Add(gtx.Ops)
					po := paint.PaintOp{f32.Rectangle{f32.Point{float32(lx), 0}, f32.Point{float32(img.Bounds().Size().X), float32(img.Bounds().Size().Y)}}}
					po.Add(gtx.Ops)
				}

				needsPaint = false

				e.Frame(gtx.Ops)

			case key.Event:
				if e.Name == "⎋ " {
					//				return nil
				}
				if e.Name == "⇞ " {
					page.SetValueInt(page.GetValueInt() - 1)
					editorsChanged = true
					needsPaint = true
				}
				if e.Name == "⇟ " {
					page.SetValueInt(page.GetValueInt() + 1)
					editorsChanged = true
					needsPaint = true
				}

			case pointer.Event:
				//fmt.Printf("mouse pos(%v)\n", e)
				mouseX.SetValueFloat64(float64(e.Position.X))
				mouseY.SetValueFloat64(float64(e.Position.Y))

				if e.Position.X <= float32(lx) {
					break
				}
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
					editorsChanged = true
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
					editorsChanged = true
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
						editorsChanged = true
						needsPaint = true
						// ni := paint.NewImageOp(r.Render())
						// ni.Add(gtx.Ops)
						// po := paint.PaintOp{f32.Rectangle{f32.Point{0, 0}, f32.Point{float32(wx), float32(wy)}}}
						// po.Add(gtx.Ops)
						// w.Invalidate()
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

type ParamEdit struct {
	P rv.RenderParameter
	N *widget.Editor
}
