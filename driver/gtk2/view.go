// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build !gotk3,!nogtk2 !shiny,!nogtk2

package gtk2

import (
	"image"
	"image/draw"
	"unsafe"

	rv "github.com/TheGrum/renderview"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

// FrameBuffer sets up a GTK2 Window and runs a mainloop rendering the RenderModel
func FrameBuffer(m rv.RenderModel) {
	GtkWindowInit(m)
}

// Main sets up a GTK2 Window with widgets for editing parameters and runs a
// mainloop rendering the RenderModel
func Main(m rv.RenderModel) {
	GtkWindowWithWidgetsInit(m)
}

func GtkWindowInit(r rv.RenderModel) {
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

func GtkWindowWithWidgetsInit(r rv.RenderModel) {
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

func GetGtkWindow(r rv.RenderModel, addWidgets bool) *gtk.Window {
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
	names := r.R.GetHintedParameterNamesWithFallback(rv.HINT_SIDEBAR | rv.HINT_FOOTER)
	if len(names) > 0 {
		sidebar.PackStart(gtk.NewLabel("________Parameters________"), false, false, 1)

		for i := 0; i < len(names); i++ {
			sidebar.PackStart(gtk.NewLabel(names[i]), false, false, 1)
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

	pixbuf *gdkpixbuf.Pixbuf
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
	w.zoomRate = r.GetParameter("zoomRate")
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
		w.needsPaint = true
		//w.QueueDraw()
		//w.GetWindow().Invalidate(nil, false)
	})
	w.SetCanFocus(true)
	//	w.SetFocusOnClick(true) // missing?
	w.SetEvents(int(gdk.POINTER_MOTION_MASK | gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK | gdk.EXPOSURE_MASK | gdk.SCROLL_MASK | gdk.KEY_PRESS_MASK))
	// This doesn't seem to actually work?
	glib.TimeoutAdd(3000, func() int {
		if w.needsPaint {
			w.QueueDraw()
		}
		return 1
	})

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
	//fmt.Printf("Configure called.\n")
	if w.pixbuf != nil {
		w.pixbuf.Unref()
	}
	allocation := w.GetAllocation()
	w.width.SetValueInt(allocation.Width)
	w.height.SetValueInt(allocation.Height)

	w.pixbuf = gdkpixbuf.NewPixbuf(gdkpixbuf.GDK_COLORSPACE_RGB, true, 8, allocation.Width, allocation.Height)
	w.needsPaint = true
	w.needsUpdate = true
}

func (w *GtkRenderWidget) Draw(ctx *glib.CallbackContext) {
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
	w.pixbuf.Fill(0)
	switch a := w.Image.(type) {
	case *image.RGBA:
		GdkPixelCopy(a, w.pixbuf, image.ZR, image.ZP)
	case *image.NRGBA:
		GdkPixelCopyNRGBA(a, w.pixbuf, image.ZR, image.ZP)
	}

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

func GdkPixelCopyNRGBA(source *image.NRGBA, target *gdkpixbuf.Pixbuf, region image.Rectangle, targetOffset image.Point) {
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
				cx := float64(e.X) / float64(w.width.GetValueInt())
				cy := float64(e.Y) / float64(w.height.GetValueInt())
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
				cx := float64(e.X) / float64(w.width.GetValueInt())
				cy := float64(e.Y) / float64(w.height.GetValueInt())
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

func (w *GtkRenderWidget) OnMotion(e *gdk.EventMotion) {
	//	fmt.Printf("Motion: X:%v Y:%v sx:%v sy:%v mouseIsDown:%v dragging:%v\n", e.X, e.Y, w.sx, w.sy, w.mouseIsDown, w.dragging)
	w.mouseX.SetValueFloat64(float64(e.X))
	w.mouseY.SetValueFloat64(float64(e.Y))

	if w.needsPaint {
		w.QueueDraw()
	}

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
		w.needsUpdate = true
		w.SetNeedsPaint()

		w.sx = e.X
		w.sy = e.Y

	}

}

func (w *GtkRenderWidget) OnButton(e *gdk.EventButton) {
	//	fmt.Printf("Button called with %v\n", e)
	w.mouseX.SetValueFloat64(float64(e.X))
	w.mouseY.SetValueFloat64(float64(e.Y))
	w.GrabFocus()

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
		w.needsUpdate = true
		w.SetNeedsPaint()
	}
	if e.Keyval == gdk.KEY_Page_Down {
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
	tv := gtk.NewTextView()
	r := &GtkParamWidget{
		IWidget: tv,
		P:       p,
	}
	tv.SetEditable(true)
	tv.SetCursorVisible(true)
	tb := tv.GetBuffer()
	tb.SetText(rv.GetParameterValueAsString(p))
	tb.Connect("changed", func() {
		var start, end gtk.TextIter
		tb.GetBounds(&start, &end)
		s := tb.GetText(&start, &end, false)
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
		tb := a.GetBuffer()
		tb.SetText(rv.GetParameterValueAsString(w.P))
	}
}
