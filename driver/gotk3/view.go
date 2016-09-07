// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build gotk3

package gotk3

import (
	"image"
	"image/draw"
	"log"

	rv "github.com/TheGrum/renderview"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// FrameBuffer sets up a GTK3 Window and runs a mainloop rendering the RenderModel
func FrameBuffer(m rv.RenderModel) {
	GtkWindowInit(m)
}

// Main sets up a GTK3 Window with widgets for editing parameters and runs a
// mainloop rendering the RenderModel
func Main(m rv.RenderModel) {
	GtkWindowWithWidgetsInit(m)
}

func GtkWindowInit(r rv.RenderModel) {
	gtk.Init(nil)
	window := GetGtkWindow(r, false)
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	window.SetDefaultSize(400, 400)
	window.ShowAll()
	gtk.Main()
}

func GtkWindowWithWidgetsInit(r rv.RenderModel) {
	gtk.Init(nil)
	window := GetGtkWindow(r, true)
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	window.SetSizeRequest(400, 400)
	window.ShowAll()
	gtk.Main()
}

func GetGtkWindow(r rv.RenderModel, addWidgets bool) *gtk.Window {
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
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
	parent, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1)
	sidebar, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	names := r.R.GetHintedParameterNamesWithFallback(rv.HINT_SIDEBAR | rv.HINT_FOOTER)
	if len(names) > 0 {
		label, _ := gtk.LabelNew("________Parameters________")
		sidebar.PackStart(label, false, false, 1)

		for i := 0; i < len(names); i++ {
			label, _ = gtk.LabelNew(names[i])
			sidebar.PackStart(label, false, false, 1)
			tv := NewGtkParamWidget(r.R.GetParameter(names[i]), r)
			r.ParamWidgets = append(r.ParamWidgets, tv)
			sidebar.PackStart(tv, false, false, 1)
		}

		parent.PackStart(sidebar, false, true, 0)
	}
	names = r.R.GetHintedParameterNames(rv.HINT_FULLTEXT)
	if len(names) > 0 {
		// we can only do this for one parameter, so ignore multiples
		tv := NewGtkParamWidget(r.R.GetParameter(names[0]), r)
		r.ParamWidgets = append(r.ParamWidgets, tv)
		parent.PackStart(tv, true, true, 1)
	}
	parent.PackEnd(r, true, true, 2)
	return parent
}

type GtkRenderWidget struct {
	*gtk.DrawingArea

	pixbuf *gdk.Pixbuf
	Image  image.Image

	index int
	R     rv.RenderModel

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
	page,
	zoomRate rv.RenderParameter

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

func NewGtkRenderWidget(r rv.RenderModel) *GtkRenderWidget {
	i, err := gtk.DrawingAreaNew()
	handleError(err)
	w := &GtkRenderWidget{
		DrawingArea: i,
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
	w.zoomRate = r.GetParameter("zoomRate")
	w.leftIsFloat64 = w.left.GetType() == "float64"
	w.zoomIsFloat64 = w.zoom.GetType() == "float64"
	w.Connect("draw", w.Draw)
	w.Connect("configure-event", w.Configure)
	w.Connect("motion-notify-event", w.OnMotion)
	w.Connect("button-press-event", w.OnButton)
	w.Connect("button-release-event", w.OnButton)
	w.Connect("scroll-event", w.OnScroll)
	w.Connect("key-press-event", w.OnKeyPress)
	w.R.SetRequestPaintFunc(func() {
		//w.UpdateParamWidgets()
		w.needsUpdate = true
		w.QueueDraw()
		//w.GetWindow().Invalidate(nil, false)
	})
	w.SetCanFocus(true)
	//	w.SetFocusOnClick(true) // missing?
	w.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.EXPOSURE_MASK | gdk.SCROLL_MASK | gdk.KEY_PRESS_MASK))
	return w
}

func (w *GtkRenderWidget) SetNeedsPaint() {
	w.needsPaint = true
	w.QueueDraw()
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
	allocation := w.GetAllocation()
	w.width.SetValueInt(allocation.GetWidth())
	w.height.SetValueInt(allocation.GetHeight())

	var err error
	w.pixbuf, err = gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, allocation.GetWidth(), allocation.GetHeight())
	if err != nil {
		log.Fatal(err)
	}
	w.needsPaint = true
	w.needsUpdate = true
}

func (w *GtkRenderWidget) Draw(da *gtk.DrawingArea, cr *cairo.Context) {
	if w.needsUpdate {
		w.UpdateParamWidgets()
	}
	if w.needsPaint || w.Image == nil {
		img := w.R.Render()
		if img == nil {
			return
		}
		switch a := img.(type) {
		case *image.RGBA:
			w.Image = a
			w.needsPaint = false
		case *image.NRGBA:
			w.Image = a
			w.needsPaint = false
		default:
			i2 := image.NewRGBA(img.Bounds())
			draw.Draw(i2, img.Bounds(), img, image.ZP, draw.Src)
			w.Image = i2
			w.needsPaint = false
		}
	}
	// copy from Go image to GDK image
	//w.pixbuf.Fill(0)
	switch a := w.Image.(type) {
	case *image.RGBA:
		GdkPixelCopy(a, w.pixbuf, image.ZR, image.ZP)
	case *image.NRGBA:
		GdkPixelCopyNRGBA(a, w.pixbuf, image.ZR, image.ZP)
	}

	gtk.GdkCairoSetSourcePixBuf(cr, w.pixbuf, 0, 0)
	cr.Paint()

	var err error
	allocation := w.GetAllocation()
	w.pixbuf, err = gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, allocation.GetWidth(), allocation.GetHeight())
	if err != nil {
		log.Fatal(err)
	}
}

func GdkPixelCopy(source *image.RGBA, target *gdk.Pixbuf, region image.Rectangle, targetOffset image.Point) {
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
	targetPix = target.GetPixels()
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

func GdkPixelCopyNRGBA(source *image.NRGBA, target *gdk.Pixbuf, region image.Rectangle, targetOffset image.Point) {
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
	targetPix = target.GetPixels()
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

func (w *GtkRenderWidget) OnScroll(da *gtk.DrawingArea, ge *gdk.Event) {
	e := &gdk.EventScroll{ge}
	if e.Direction() == gdk.SCROLL_DOWN {
		if w.zoomIsFloat64 {
			w.zoom.SetValueFloat64(w.zoom.GetValueFloat64() - 1)
		} else {
			w.zoom.SetValueInt(w.zoom.GetValueInt() - 1)
		}
		if w.options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || w.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
			mult := 1 + rv.ZOOM_RATE
			if w.zoomRate.GetValueFloat64() > 0 {
				mult = 1 + w.zoomRate.GetValueFloat64()
			}
			if w.leftIsFloat64 {

				zwidth := w.right.GetValueFloat64() - w.left.GetValueFloat64()
				zheight := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
				nzwidth := zwidth * mult
				nzheight := zheight * mult
				cx := float64(e.X()) / float64(w.width.GetValueInt())
				cy := float64(e.Y()) / float64(w.height.GetValueInt())
				if w.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
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
		if w.options.GetValueInt()&rv.OPT_AUTO_ZOOM == rv.OPT_AUTO_ZOOM || w.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
			mult := 1 - rv.ZOOM_RATE
			if w.zoomRate.GetValueFloat64() > 0 {
				mult = 1 - w.zoomRate.GetValueFloat64()
			}
			if w.leftIsFloat64 {
				zwidth := w.right.GetValueFloat64() - w.left.GetValueFloat64()
				zheight := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
				nzwidth := zwidth * mult
				nzheight := zheight * mult
				cx := float64(e.X()) / float64(w.width.GetValueInt())
				cy := float64(e.Y()) / float64(w.height.GetValueInt())
				if w.options.GetValueInt()&rv.OPT_CENTER_ZOOM == rv.OPT_CENTER_ZOOM {
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
	w.needsUpdate = true
	w.SetNeedsPaint()

}

func (w *GtkRenderWidget) OnMotion(da *gtk.DrawingArea, ge *gdk.Event) {
	e := &gdk.EventMotion{ge}
	//	fmt.Printf("Motion: X:%v Y:%v sx:%v sy:%v mouseIsDown:%v dragging:%v\n", e.X, e.Y, w.sx, w.sy, w.mouseIsDown, w.dragging)
	var X, Y float64
	X, Y = e.MotionVal()
	w.mouseX.SetValueFloat64(X)
	w.mouseY.SetValueFloat64(Y)

	if w.mouseIsDown && w.dragging == false {
		if ((X - w.sx) > 3) || ((w.sx - X) > 3) || ((Y - w.sy) > 3) || ((w.sy - Y) > 3) {
			w.dragging = true
		}
	}

	if w.dragging {
		if w.leftIsFloat64 {
			width := w.right.GetValueFloat64() - w.left.GetValueFloat64()
			height := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
			dx := width / float64(w.width.GetValueInt())
			dy := height / float64(w.height.GetValueInt())
			cx := float64(X-w.sx) * dx
			cy := float64(Y-w.sy) * dy
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
			cx := float64(X-w.sx) * dx
			cy := float64(Y-w.sy) * dy
			w.left.SetValueInt(int(float64(w.left.GetValueInt()) - cx))
			w.right.SetValueInt(int(float64(w.right.GetValueInt()) - cx))
			w.top.SetValueInt(int(float64(w.top.GetValueInt()) - cy))
			w.bottom.SetValueInt(int(float64(w.bottom.GetValueInt()) - cy))
		}
		//			Draw(r.Render(), buf.RGBA())
		w.needsUpdate = true
		w.SetNeedsPaint()

		w.sx = X
		w.sy = Y

	}

}

func (w *GtkRenderWidget) OnButton(da *gtk.DrawingArea, ge *gdk.Event) {
	e := &gdk.EventButton{ge}
	//	fmt.Printf("Button called with %v\n", e)i
	w.mouseX.SetValueFloat64(e.X())
	w.mouseY.SetValueFloat64(e.Y())
	w.GrabFocus()

	if gdk.EventType(e.Type()) == gdk.EVENT_BUTTON_PRESS {
		if w.dragging == false && w.mouseIsDown == false && e.Button() == 1 {
			//			fmt.Println("Mousedown")
			w.sx = e.X()
			w.sy = e.Y()
			w.mouseIsDown = true
		}
	}
	if gdk.EventType(e.Type()) == gdk.EVENT_BUTTON_RELEASE {
		//		fmt.Println("Mouseup")
		w.mouseIsDown = false
		w.dragging = false
	}
}

const (
	PAGE_UP   uint = 0xff55
	PAGE_DOWN uint = 0xff56
)

func (w *GtkRenderWidget) OnKeyPress(da *gtk.DrawingArea, ge *gdk.Event) {
	e := &gdk.EventKey{ge}
	if e.KeyVal() == PAGE_UP {
		w.page.SetValueInt(w.page.GetValueInt() - 1)
		w.needsUpdate = true
		w.SetNeedsPaint()
	}
	if e.KeyVal() == PAGE_DOWN {
		w.page.SetValueInt(w.page.GetValueInt() + 1)
		w.needsUpdate = true
		w.SetNeedsPaint()
	}
}

type GtkParamWidget struct {
	gtk.IWidget
	P rv.RenderParameter
}

func NewGtkParamWidget(p rv.RenderParameter, w *GtkRenderWidget) *GtkParamWidget {
	tv, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal(err)
	}
	r := &GtkParamWidget{
		IWidget: tv,
		P:       p,
	}
	tv.SetEditable(true)
	tv.SetCursorVisible(true)
	tb, err := tv.GetBuffer()
	if err != nil {
		log.Fatal(err)
	}
	tb.SetText(rv.GetParameterValueAsString(p))
	tb.Connect("changed", func() {
		var start, end *gtk.TextIter
		start, end = tb.GetBounds()
		s, err := tb.GetText(start, end, false)
		if err != nil {
			log.Fatal(err)
		}
		pValue := rv.GetParameterValueAsString(r.P)
		if s != pValue {
			rv.SetParameterValueFromString(r.P, s)
		}
		w.SetNeedsPaint()
	})
	return r
}

func (w *GtkParamWidget) Update() {
	switch a := w.IWidget.(type) {
	case *gtk.TextView:
		tb, err := a.GetBuffer()
		if err != nil {
			log.Fatal(err)
		}
		tb.SetText(rv.GetParameterValueAsString(w.P))
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
