package renderview

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"strconv"
	"strings"
	"unsafe"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"

	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font/inconsolata"

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

func GetWidgetMainLoop(r RenderModel) func(s screen.Screen) {
	return func(s screen.Screen) {
		WidgetMainLoop(s, r)
	}
}

func expand(n node.Node, expandAlongWeight int) node.Node {
	return widget.WithLayoutData(n, widget.FlowLayoutData{
		ExpandAcross:      true,
		ExpandAlongWeight: expandAlongWeight,
	})
}

type RenderWidget struct {
	node.LeafEmbed
	index int
	r     RenderModel

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
	page RenderParameter

	sx,
	sy float32

	leftIsFloat64,
	zoomIsFloat64,
	mouseIsDown,
	dragging bool
}

func NewRenderWidget(r RenderModel) *RenderWidget {
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
			if m.options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || m.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
				mult := 1 + ZOOM_RATE
				if m.leftIsFloat64 {

					zwidth := m.right.GetValueFloat64() - m.left.GetValueFloat64()
					zheight := m.bottom.GetValueFloat64() - m.top.GetValueFloat64()
					nzwidth := zwidth * mult
					nzheight := zheight * mult
					cx := float64(e.X) / float64(m.width.GetValueInt())
					cy := float64(e.Y) / float64(m.height.GetValueInt())
					if m.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
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
			if m.options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || m.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
				mult := 1 - ZOOM_RATE
				if m.leftIsFloat64 {
					zwidth := m.right.GetValueFloat64() - m.left.GetValueFloat64()
					zheight := m.bottom.GetValueFloat64() - m.top.GetValueFloat64()
					nzwidth := zwidth * mult
					nzheight := zheight * mult
					cx := float64(e.X) / float64(m.width.GetValueInt())

					cy := float64(e.Y) / float64(m.height.GetValueInt())
					if m.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
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

func WidgetMainLoop(s screen.Screen, r RenderModel) {
	w := GetRenderWidgetWithSidebar(r)
	if err := widget.RunWindow(s, w, nil); err != nil {
		log.Fatal(err)
	}
}

func GetParameterValueAsString(p RenderParameter) string {
	switch p.GetType() {
	case "int":
		return fmt.Sprintf("%v", p.GetValueInt())
	case "uint32":
		return fmt.Sprintf("%v", p.GetValueUInt32())
	case "float64":
		return fmt.Sprintf("%v", p.GetValueFloat64())
	case "complex128":
		return fmt.Sprintf("%v", p.GetValueComplex128())
	case "string":
		return p.GetValueString()
	default:
		return p.GetValueString()
	}
}

func ParseComplex(v string) (complex128, error) {
	v = strings.Replace(v, ",", "+", -1)
	l := strings.Split(v, "+")
	r, err := strconv.ParseFloat(l[0], 64)
	if err != nil {
		return 0, err
	}
	if len(l) > 1 {
		l[1] = strings.Replace(l[1], "i", "", -1)
		i, err := strconv.ParseFloat(l[1], 64)
		if err != nil {
			return 0, err
		}
		return complex(r, i), nil
	} else {
		return complex(r, 0), nil
	}

}
func SetParameterValueFromString(p RenderParameter, v string) {
	switch p.GetType() {
	case "int":
		i, err := strconv.Atoi(v)
		if err == nil {
			p.SetValueInt(i)
		}
	case "uint32":
		i, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			p.SetValueUInt32(uint32(i))
		}
	case "float64":
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			p.SetValueFloat64(f)
		}
	case "complex128":
		c, err := ParseComplex(v)
		if err == nil {
			p.SetValueComplex128(c)
		}
	case "string":
		p.SetValueString(v)
	default:
		p.SetValueString(v)

	}
}

func GetRenderWidgetWithSidebar(r RenderModel) node.Node {
	sideflow := widget.NewFlow(widget.AxisVertical)
	names := r.GetParameterNames()

	for i := 0; i < len(names); i++ {
		sideflow.Insert(widget.NewLabel(names[i]), nil)
		//e := widget.NewLabel("test")
		//e := editor.NewEditor(inconsolata.Regular8x16, nil)
		//e := NewTextEdit(r.GetParameter(names[i]).GetValueString(), inconsolata.Regular8x16, nil)
		e := NewTextEdit("text", inconsolata.Regular8x16, nil)
		//		e.Text = r.GetParameter(names[i]).GetValueString()
		e.Text = "Test"
		e.Rect = image.Rectangle{image.ZP, image.Point{50, 16}}
		//		w := expand(e, 0
		w := widget.NewSizer(unit.Pixels(50), unit.Pixels(30), e)
		//			widget.NewLabel(r.GetParameter(names[i]).GetValueString()))
		sideflow.Insert(w, nil)
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
	R RenderModel
	P RenderParameter
	N node.Node
}

func GtkWindowInit(r RenderModel) {
	gtk.Init(nil)
	window := GetGtkWindow(r, false)
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		//		println("got destroy!", ctx.Data().(string))
		gtk.MainQuit()
	}, "foo")

	window.SetSizeRequest(400, 400)
	window.ShowAll()
	gtk.Main()
}

func GtkWindowWithWidgetsInit(r RenderModel) {
	gtk.Init(nil)
	window := GetGtkWindow(r, true)
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		//		println("got destroy!", ctx.Data().(string))
		gtk.MainQuit()
	}, "foo")

	window.SetSizeRequest(400, 400)
	window.ShowAll()
	gtk.Main()
}

func GetGtkWindow(r RenderModel, addWidgets bool) *gtk.Window {
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	var child gtk.IWidget
	render := NewGtkRenderWidget(r)
	child = render
	if addWidgets {
		child = WrapRenderWidget(render)
	}
	window.Add(child)
	return window
}

func WrapRenderWidget(r *GtkRenderWidget) gtk.IWidget {
	parent := gtk.NewHBox(false, 1)
	sidebar := gtk.NewVBox(false, 1)
	names := r.R.GetParameterNames()
	sidebar.PackStart(gtk.NewLabel("________Parameters________"), false, false, 1)

	for i := 0; i < len(names); i++ {
		sidebar.PackStart(gtk.NewLabel(names[i]), false, false, 1)
		tv := NewGtkParamWidget(r.R.GetParameter(names[i]))
		r.ParamWidgets = append(r.ParamWidgets, tv)
		sidebar.PackStart(tv, false, false, 1)
	}

	parent.PackStart(sidebar, false, true, 1)
	parent.PackEnd(r, true, true, 1)
	return parent
}

type GtkRenderWidget struct {
	*gtk.DrawingArea

	pixbuf *gdkpixbuf.Pixbuf
	Image  *image.RGBA

	index int
	R     RenderModel

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
	page RenderParameter

	sx,
	sy float64

	leftIsFloat64,
	zoomIsFloat64,
	mouseIsDown,
	dragging,
	needsPaint bool
	needsUpdate bool

	ParamWidgets []*GtkParamWidget
}

func NewGtkRenderWidget(r RenderModel) *GtkRenderWidget {
	w := &GtkRenderWidget{
		DrawingArea: gtk.NewDrawingArea(),
		R:           r,
	}
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
	w.Connect("expose-event", w.Draw)
	w.Connect("configure-event", w.Configure)
	w.Connect("motion-notify-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		mev := *(**gdk.EventMotion)(unsafe.Pointer(&arg))
		w.OnMotion(mev)
	})
	w.Connect("button-press-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		mev := *(**gdk.EventButton)(unsafe.Pointer(&arg))
		w.OnButton(mev)
	})
	w.Connect("button-release-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		mev := *(**gdk.EventButton)(unsafe.Pointer(&arg))
		w.OnButton(mev)
	})
	w.Connect("scroll-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		sev := *(**gdk.EventScroll)(unsafe.Pointer(&arg))
		w.OnScroll(sev)
	})
	w.Connect("key-press-event", func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		kev := *(**gdk.EventKey)(unsafe.Pointer(&arg))
		w.OnKeyPress(kev)
	})
	w.R.SetRequestPaintFunc(func() {
		//w.UpdateParamWidgets()
		w.needsUpdate = true
		w.GetWindow().Invalidate(nil, false)
	})
	w.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.EXPOSURE_MASK | gdk.SCROLL_MASK | gdk.KEY_PRESS_MASK))
	return w
}

func (w *GtkRenderWidget) SetNeedsPaint() {
	w.needsPaint = true
	allocation := w.GetAllocation()
	w.GetWindow().Invalidate((*gdk.Rectangle)(allocation), false)
}

func (w *GtkRenderWidget) UpdateParamWidgets() {
	w.R.Lock()
	defer w.R.Unlock()
	for i := 0; i < len(w.ParamWidgets); i++ {
		w.ParamWidgets[i].Update()
	}
	w.needsUpdate = false
}

func (w *GtkRenderWidget) Configure() {
	//fmt.Printf("Configure called.\n")
	if w.pixbuf != nil {
		w.pixbuf.Unref()
	}
	allocation := w.GetAllocation()
	w.width.SetValueInt(allocation.Width)
	w.height.SetValueInt(allocation.Height)

	w.pixbuf = gdkpixbuf.NewPixbuf(gdkpixbuf.GDK_COLORSPACE_RGB, true, 8, allocation.Width, allocation.Height)
	w.needsPaint = true
}

func (w *GtkRenderWidget) Draw(ctx *glib.CallbackContext) {
	if w.needsPaint || w.Image == nil {
		if w.needsUpdate {
			w.UpdateParamWidgets()
		}
		img := w.R.Render()
		if img == nil {
			return
		}
		switch a := img.(type) {
		case *image.RGBA:
			w.Image = a
			w.needsPaint = false
		default:
			//todo: copy to an RGBA here
			fmt.Printf("Missing type handler for (%t)\n", a)
			return
		}
	}
	// copy from Go image to GDK image
	w.pixbuf.Fill(0)
	GdkPixelCopy(w.Image, w.pixbuf, image.ZR, image.ZP)

	// Draw GDK image on window
	win := w.GetWindow()
	draw := win.GetDrawable()
	gc := gdk.NewGC(draw)
	draw.DrawPixbuf(gc, w.pixbuf, 0, 0, 0, 0, w.pixbuf.GetWidth(), w.pixbuf.GetHeight(), gdk.RGB_DITHER_NONE, 0, 0)
}

func GdkPixelCopy(source *image.RGBA, target *gdkpixbuf.Pixbuf, region image.Rectangle, targetOffset image.Point) {
	if source == nil {
		return
	}
	if target == nil {
		return
	}
	sourceBounds := source.Rect
	targetBounds := image.Rect(0, 0, target.GetWidth(), target.GetHeight())
	sourceRowstride := source.Stride
	targetRowstride := target.GetRowstride()
	//Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4]
	var targetPix []byte
	var sourcePix []uint8
	targetPix = target.GetPixelsWithLength()
	sourcePix = source.Pix

	if region.Dx() == 0 {
		region = sourceBounds.Intersect(targetBounds.Sub(targetOffset))
	}
	region = region.Intersect(sourceBounds).Intersect(targetBounds.Sub(targetOffset))
	var x, y int

	for y = region.Min.Y; y < region.Max.Y; y++ {
		for x = region.Min.X; x < region.Max.X; x++ {
			targetPix[(y+targetOffset.Y)*targetRowstride+(x+targetOffset.X)*4] = sourcePix[y*sourceRowstride+x*4]
			targetPix[(y+targetOffset.Y)*targetRowstride+(x+targetOffset.X)*4+1] = sourcePix[y*sourceRowstride+x*4+1]
			targetPix[(y+targetOffset.Y)*targetRowstride+(x+targetOffset.X)*4+2] = sourcePix[y*sourceRowstride+x*4+2]
			targetPix[(y+targetOffset.Y)*targetRowstride+(x+targetOffset.X)*4+3] = sourcePix[y*sourceRowstride+x*4+3]
		}
	}
}

func (w *GtkRenderWidget) OnScroll(e *gdk.EventScroll) {
	// the case of SCROLL_Down is incorrect in gdk.go
	// todo: fix this when it is fixed upstream
	// e.Direction always has same value, and does not match documentation
	if gdk.ModifierType(e.State) > (1 << 30) {
		if w.zoomIsFloat64 {
			w.zoom.SetValueFloat64(w.zoom.GetValueFloat64() - 1)
		} else {
			w.zoom.SetValueInt(w.zoom.GetValueInt() - 1)
		}
		if w.options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || w.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
			mult := 1 + ZOOM_RATE
			if w.leftIsFloat64 {

				zwidth := w.right.GetValueFloat64() - w.left.GetValueFloat64()
				zheight := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
				nzwidth := zwidth * mult
				nzheight := zheight * mult
				cx := float64(e.X) / float64(w.width.GetValueInt())
				cy := float64(e.Y) / float64(w.height.GetValueInt())
				if w.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
					cx = 0.5
					cy = 0.5
				}
				//fmt.Printf("zoomOut: mult: %v zwidth: %v nzwidth: %v cx: %v left: %v nleft: %v\n", mult, zwidth, nzwidth, cx, left.GetValueFloat64(), left.GetValueFloat64()-((nzwidth-zwidth)*cx))
				w.left.SetValueFloat64(w.left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
				w.top.SetValueFloat64(w.top.GetValueFloat64() - ((nzheight - zheight) * cy))
				w.right.SetValueFloat64(w.left.GetValueFloat64() + nzwidth)
				w.bottom.SetValueFloat64(w.top.GetValueFloat64() + nzheight)
			}
		}

		//if gdk.ScrollDirection(e.Direction) == gdk.SCROLL_UP {
	} else {
		if w.zoomIsFloat64 {
			w.zoom.SetValueFloat64(w.zoom.GetValueFloat64() + 1)
		} else {
			w.zoom.SetValueInt(w.zoom.GetValueInt() + 1)
		}
		if w.options.GetValueInt()&OPT_AUTO_ZOOM == OPT_AUTO_ZOOM || w.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
			mult := 1 - ZOOM_RATE
			if w.leftIsFloat64 {
				zwidth := w.right.GetValueFloat64() - w.left.GetValueFloat64()
				zheight := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
				nzwidth := zwidth * mult
				nzheight := zheight * mult
				cx := float64(e.X) / float64(w.width.GetValueInt())
				cy := float64(e.Y) / float64(w.height.GetValueInt())
				if w.options.GetValueInt()&OPT_CENTER_ZOOM == OPT_CENTER_ZOOM {
					cx = 0.5
					cy = 0.5
				}
				w.left.SetValueFloat64(w.left.GetValueFloat64() - ((nzwidth - zwidth) * cx))
				w.top.SetValueFloat64(w.top.GetValueFloat64() - ((nzheight - zheight) * cy))
				w.right.SetValueFloat64(w.left.GetValueFloat64() + nzwidth)
				w.bottom.SetValueFloat64(w.top.GetValueFloat64() + nzheight)
			}
		}
	}
	w.SetNeedsPaint()

}

func (w *GtkRenderWidget) OnMotion(e *gdk.EventMotion) {
	//	fmt.Printf("Motion: X:%v Y:%v sx:%v sy:%v mouseIsDown:%v dragging:%v\n", e.X, e.Y, w.sx, w.sy, w.mouseIsDown, w.dragging)
	w.mouseX.SetValueFloat64(float64(e.X))
	w.mouseY.SetValueFloat64(float64(e.Y))

	if w.mouseIsDown && w.dragging == false {
		if ((e.X - w.sx) > 3) || ((w.sx - e.X) > 3) || ((e.Y - w.sy) > 3) || ((w.sy - e.Y) > 3) {
			w.dragging = true
		}
	}

	if w.dragging {
		if w.leftIsFloat64 {
			width := w.right.GetValueFloat64() - w.left.GetValueFloat64()
			height := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
			dx := width / float64(w.width.GetValueInt())
			dy := height / float64(w.height.GetValueInt())
			cx := float64(e.X-w.sx) * dx
			cy := float64(e.Y-w.sy) * dy
			//						fmt.Printf("dx %v dy %v cx %v cy %v\n", dx, dy, cx, cy)
			w.left.SetValueFloat64(w.left.GetValueFloat64() - cx)
			w.right.SetValueFloat64(w.right.GetValueFloat64() - cx)
			w.top.SetValueFloat64(w.top.GetValueFloat64() - cy)
			w.bottom.SetValueFloat64(w.bottom.GetValueFloat64() - cy)
		} else {
			width := w.right.GetValueInt() - w.left.GetValueInt()
			height := w.bottom.GetValueInt() - w.top.GetValueInt()
			dx := float64(width) / float64(w.width.GetValueInt())
			dy := float64(height) / float64(w.height.GetValueInt())
			cx := float64(e.X-w.sx) * dx
			cy := float64(e.Y-w.sy) * dy
			w.left.SetValueInt(int(float64(w.left.GetValueInt()) - cx))
			w.right.SetValueInt(int(float64(w.right.GetValueInt()) - cx))
			w.top.SetValueInt(int(float64(w.top.GetValueInt()) - cy))
			w.bottom.SetValueInt(int(float64(w.bottom.GetValueInt()) - cy))
		}
		//			Draw(r.Render(), buf.RGBA())
		w.SetNeedsPaint()

		w.sx = e.X
		w.sy = e.Y

	}

}

func (w *GtkRenderWidget) OnButton(e *gdk.EventButton) {
	//	fmt.Printf("Button called with %v\n", e)
	w.mouseX.SetValueFloat64(float64(e.X))
	w.mouseY.SetValueFloat64(float64(e.Y))

	if gdk.EventType(e.Type) == gdk.BUTTON_PRESS {
		if w.dragging == false && w.mouseIsDown == false && e.Button == 1 {
			//			fmt.Println("Mousedown")
			w.sx = e.X
			w.sy = e.Y
			w.mouseIsDown = true
		}
	}
	if gdk.EventType(e.Type) == gdk.BUTTON_RELEASE {
		//		fmt.Println("Mouseup")
		w.mouseIsDown = false
		w.dragging = false
	}
}

func (w *GtkRenderWidget) OnKeyPress(e *gdk.EventKey) {
	if e.Keyval == gdk.KEY_Page_Up {
		w.page.SetValueInt(w.page.GetValueInt() - 1)
		w.needsPaint = true
	}
	if e.Keyval == gdk.KEY_Page_Down {
		w.page.SetValueInt(w.page.GetValueInt() + 1)
		w.needsPaint = true
	}
}

type GtkParamWidget struct {
	gtk.IWidget
	P RenderParameter
}

func NewGtkParamWidget(p RenderParameter) *GtkParamWidget {
	tv := gtk.NewTextView()
	r := &GtkParamWidget{
		IWidget: tv,
		P:       p,
	}
	tv.SetEditable(true)
	tv.SetCursorVisible(true)
	tb := tv.GetBuffer()
	tb.SetText(GetParameterValueAsString(p))
	tb.Connect("changed", func() {
		var start, end gtk.TextIter
		tb.GetBounds(&start, &end)
		s := tb.GetText(&start, &end, false)
		pValue := GetParameterValueAsString(r.P)
		if s != pValue {
			SetParameterValueFromString(r.P, s)
		}
	})
	return r
}

func (w *GtkParamWidget) Update() {
	switch a := w.IWidget.(type) {
	case *gtk.TextView:
		tb := a.GetBuffer()
		tb.SetText(GetParameterValueAsString(w.P))

	}
}
