// Copyright 2020 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build fyne

package fyne

import (
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	rv "github.com/TheGrum/renderview"
	"github.com/mattn/go-gtk/gdkpixbuf"
)

// FrameBuffer sets up a Fyne Window and runs a mainloop rendering the RenderModel
func FrameBuffer(m rv.RenderModel) {
	FyneWindowInit(m)
}

// Main sets up a Fyne2 Window with widgets for editing parameters and runs a
// mainloop rendering the RenderModel
func Main(m rv.RenderModel) {
	FyneWindowWithWidgetsInit(m)
}

func FyneWindowInit(r rv.RenderModel) {
	app := app.New()
	window := GetFyneWindow(r, app, false)
	window.Resize(fyne.NewSize(600, 400))

	window.ShowAndRun()
}

func FyneWindowWithWidgetsInit(r rv.RenderModel) {
	app := app.New()
	window := GetFyneWindow(r, app, true)
	window.Resize(fyne.NewSize(600, 400))

	window.ShowAndRun()
}

func GetFyneWindow(r rv.RenderModel, app fyne.App, addWidgets bool) fyne.Window {
	window := app.NewWindow("Renderview")
	var child fyne.CanvasObject
	render := NewFyneRenderWidget(r, window)
	child = render
	if addWidgets {
		child = WrapRenderWidget(render)
	}
	window.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), child))
	return window
}

func WrapRenderWidget(r *FyneRenderWidget) fyne.CanvasObject {
	parent := widget.NewHBox()
	sidebar := widget.NewVBox()
	names := r.R.GetHintedParameterNamesWithFallback(rv.HINT_SIDEBAR | rv.HINT_FOOTER)
	if len(names) > 0 {
		sidebar.Append(widget.NewLabel("________Parameters________"))

		for i := 0; i < len(names); i++ {
			sidebar.Append(widget.NewLabel(names[i]))
			tv := NewFyneParamWidget(r.R.GetParameter(names[i]), r, false)
			r.ParamWidgets = append(r.ParamWidgets, tv)
			sidebar.Append(tv)
		}

		parent.Append(sidebar)
	}
	names = r.R.GetHintedParameterNames(rv.HINT_FULLTEXT)
	if len(names) > 0 {
		// we can only do this for one parameter, so ignore multiples
		tv := NewFyneParamWidget(r.R.GetParameter(names[0]), r, true)
		r.ParamWidgets = append(r.ParamWidgets, tv)
		parent.Append(tv)
	}
	//r.SetMinSize(fyne.Size{Width: 100, Height: 100})
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, parent, nil), parent, r)
}

type FyneRenderWidgetRenderer struct {
	objects []fyne.CanvasObject
	*canvas.Raster

	*FyneRenderWidget
}

func (w *FyneRenderWidgetRenderer) MinSize() fyne.Size {
	return fyne.Size{Width: 100, Height: 100}
}

func (w *FyneRenderWidgetRenderer) Layout(size fyne.Size) {
	w.width.SetValueInt(size.Width)
	w.height.SetValueInt(size.Height)
	w.Raster.Resize(size)
	w.needsPaint = true
}

func (w *FyneRenderWidgetRenderer) ApplyTheme() {

}

func (w *FyneRenderWidgetRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (w *FyneRenderWidgetRenderer) Refresh() {
	canvas.Refresh(w.Raster)
}

func (w *FyneRenderWidgetRenderer) Objects() []fyne.CanvasObject {
	return w.objects
}

func (w *FyneRenderWidgetRenderer) Destroy() {

}

type FyneRenderWidget struct {
	widget.BaseWidget

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
	isFocused,
	needsPaint,
	needsUpdate bool

	ParamWidgets []*FyneParamWidget
}

func NewFyneRenderWidget(r rv.RenderModel, window fyne.Window) *FyneRenderWidget {
	w := &FyneRenderWidget{
		R: r,
	}
	w.ExtendBaseWidget(w)
	//w.Raster = canvas.NewRaster(w.Render)
	//w.objects = []fyne.CanvasObject{w.Raster}
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
	w.R.SetRequestPaintFunc(func() {
		w.needsUpdate = true
		w.needsPaint = true
		window.Canvas().Refresh(w)
	})

	var _ fyne.Focusable = w
	var _ fyne.Scrollable = w
	var _ fyne.Draggable = w
	var _ desktop.Hoverable = w

	return w
}

func (w *FyneRenderWidget) CreateRenderer() fyne.WidgetRenderer {
	renderer := &FyneRenderWidgetRenderer{FyneRenderWidget: w}
	render := canvas.NewRaster(w.Render)
	renderer.Raster = render
	renderer.objects = []fyne.CanvasObject{render}
	renderer.ApplyTheme()

	return renderer
}

func (w *FyneRenderWidget) Render(width int, height int) image.Image {
	if w.needsUpdate {
		w.UpdateParamWidgets()
		w.needsUpdate = false
	}
	if w.Image == nil || w.needsPaint {
		//	fmt.Printf("Rendering\n")
		img := w.R.Render()
		if img != nil {
			w.Image = img
		}
		w.needsPaint = false
	}
	//fmt.Printf("Drawing\n")
	i2 := image.NewRGBA(image.Rect(0, 0, width, height))
	if w.Image != nil {
		draw.Draw(i2, w.Image.Bounds(), w.Image, image.ZP, draw.Src)
	}
	return i2
}

func (w *FyneRenderWidget) SetNeedsPaint() {
	w.needsPaint = true
	canvas.Refresh(w)
}

func (w *FyneRenderWidget) UpdateParamWidgets() {
	w.R.Lock()
	defer w.R.Unlock()
	for i := 0; i < len(w.ParamWidgets); i++ {
		w.ParamWidgets[i].Update()
	}
	w.needsUpdate = false
}

func (w *FyneRenderWidget) Scrolled(e *fyne.ScrollEvent) {
	if e.DeltaY < 0 {
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
				cx := float64(float64(e.Position.X)) / float64(w.width.GetValueInt())
				cy := float64(float64(e.Position.Y)) / float64(w.height.GetValueInt())
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
				cx := float64(float64(e.Position.X)) / float64(w.width.GetValueInt())
				cy := float64(float64(e.Position.Y)) / float64(w.height.GetValueInt())
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

func (w *FyneRenderWidget) Dragged(e *fyne.DragEvent) {
	if w.leftIsFloat64 {
		width := w.right.GetValueFloat64() - w.left.GetValueFloat64()
		height := w.bottom.GetValueFloat64() - w.top.GetValueFloat64()
		dx := width / float64(w.width.GetValueInt())
		dy := height / float64(w.height.GetValueInt())
		cx := float64(float64(e.DraggedX) * dx)
		cy := float64(float64(e.DraggedY) * dy)
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
		cx := float64(float64(e.DraggedX) * dx)
		cy := float64(float64(e.DraggedY) * dy)
		w.left.SetValueInt(int(float64(w.left.GetValueInt()) - cx))
		w.right.SetValueInt(int(float64(w.right.GetValueInt()) - cx))
		w.top.SetValueInt(int(float64(w.top.GetValueInt()) - cy))
		w.bottom.SetValueInt(int(float64(w.bottom.GetValueInt()) - cy))
	}
	//			Draw(r.Render(), buf.RGBA())
	w.needsUpdate = true
	w.SetNeedsPaint()

	w.sx = float64(e.Position.X)
	w.sy = float64(e.Position.Y)

}

func (w *FyneRenderWidget) DragEnd() {

}

func (w *FyneRenderWidget) MouseIn(e *desktop.MouseEvent) {}

func (w *FyneRenderWidget) MouseOut() {}

func (w *FyneRenderWidget) MouseMoved(e *desktop.MouseEvent) {
	w.mouseX.SetValueFloat64(float64(e.Position.X))
	w.mouseY.SetValueFloat64(float64(e.Position.Y))

	// Fyne has Draggable, so we don't need to detect and handle
	// dragging from mouse moves
}

func (w *FyneRenderWidget) MouseDown(e *desktop.MouseEvent) {
	w.mouseX.SetValueFloat64(float64(float64(e.Position.X)))
	w.mouseY.SetValueFloat64(float64(float64(e.Position.Y)))

	if w.dragging == false && w.mouseIsDown == false && e.Button == desktop.LeftMouseButton {
		//			fmt.Println("Mousedown")
		w.sx = float64(e.Position.X)
		w.sy = float64(e.Position.Y)
		w.mouseIsDown = true
	}

}

func (w *FyneRenderWidget) MouseUp(e *desktop.MouseEvent) {
	w.mouseX.SetValueFloat64(float64(float64(e.Position.X)))
	w.mouseY.SetValueFloat64(float64(float64(e.Position.Y)))

	w.mouseIsDown = false
	w.dragging = false
}

func (w *FyneRenderWidget) FocusGained() {
	w.isFocused = true
}
func (w *FyneRenderWidget) FocusLost() {
	w.isFocused = false
}

func (w *FyneRenderWidget) Focused() bool {
	return w.isFocused == true
}

func (w *FyneRenderWidget) TypedRune(rune) {

}

func (w *FyneRenderWidget) TypedKey(*fyne.KeyEvent) {

}

func (w *FyneRenderWidget) KeyDown(e *fyne.KeyEvent) {
	if e.Name == fyne.KeyPageUp {
		w.page.SetValueInt(w.page.GetValueInt() - 1)
		w.needsUpdate = true
		w.SetNeedsPaint()
	}
	if e.Name == fyne.KeyPageDown {
		w.page.SetValueInt(w.page.GetValueInt() + 1)
		w.needsUpdate = true
		w.SetNeedsPaint()
	}
}

func (w *FyneRenderWidget) KeyUp(e *fyne.KeyEvent) {
}

type FyneParamWidget struct {
	*widget.Entry
	P rv.RenderParameter
}

func NewFyneParamWidget(p rv.RenderParameter, w *FyneRenderWidget, multiLine bool) *FyneParamWidget {
	var tv *widget.Entry
	if multiLine {
		tv = widget.NewMultiLineEntry()
	} else {
		tv = widget.NewEntry()
	}
	r := &FyneParamWidget{
		Entry: tv,
		P:     p,
	}
	tv.OnChanged = func(s string) {
		pValue := rv.GetParameterValueAsString(r.P)
		if s != pValue {
			rv.SetParameterValueFromString(r.P, s)
		}
		w.SetNeedsPaint()
	}
	return r
}

func (w *FyneParamWidget) Update() {
	w.Entry.SetText(rv.GetParameterValueAsString(w.P))
}
