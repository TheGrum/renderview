package renderview

import (
	"fmt"
	"image"
	"image/draw"
	"log"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

const (
	OPT_NONE        = iota      // 0
	OPT_CENTER_ZOOM = 1 << iota // 1
	OPT_AUTO_ZOOM   = 1 << iota // 2
)

const ZOOM_RATE = 0.1

func GetMainLoop(r RenderModel) func(s screen.Screen) {
	return func(s screen.Screen) {
		MainLoop(s, r)
	}
}

func MainLoop(s screen.Screen, r RenderModel) {
	var needsPaint = false
	var mouseIsDown = false
	var dragging bool = false
	var dx, dy float64
	var sx, sy float32

	var left, top, right, bottom, zoom RenderParameter
	left = r.GetParameter("left")
	top = r.GetParameter("top")
	right = r.GetParameter("right")
	bottom = r.GetParameter("bottom")
	zoom = r.GetParameter("zoom")
	mouseX := r.GetParameter("mouseX")
	mouseY := r.GetParameter("mouseY")
	options := r.GetParameter("options")

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
		//		w.Send(paint.Event{})
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
				if options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
					mult := 1 + ZOOM_RATE
					if leftIsFloat64 {

						zwidth := right.GetValueFloat64() - left.GetValueFloat64()
						zheight := bottom.GetValueFloat64() - top.GetValueFloat64()
						nzwidth := zwidth * mult
						nzheight := zheight * mult
						cx := float64(e.X) / float64(buf.Size().X)
						cy := float64(e.Y) / float64(buf.Size().Y)
						if options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
							cx = 0.5
							cy = 0.5
						}
						fmt.Printf("zoomOut: mult: %v zwidth: %v nzwidth: %v cx: %v left: %v nleft: %v\n", mult, zwidth, nzwidth, cx, left.GetValueFloat64(), left.GetValueFloat64()-((nzwidth-zwidth)*cx))
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
				if options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
					mult := 1 - ZOOM_RATE
					if leftIsFloat64 {
						zwidth := right.GetValueFloat64() - left.GetValueFloat64()
						zheight := bottom.GetValueFloat64() - top.GetValueFloat64()
						nzwidth := zwidth * mult
						nzheight := zheight * mult
						cx := float64(e.X) / float64(buf.Size().X)
						cy := float64(e.Y) / float64(buf.Size().Y)
						if options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
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
