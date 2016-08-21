package renderview

import (
	"image"
	"log"

	"golang.org/x/exp/shiny/widget/node"
	"golang.org/x/exp/shiny/widget/theme"
	"golang.org/x/image/font"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"

	"sigint.ca/graphics/editor"
	"sigint.ca/graphics/editor/address"
)

// TextEdit is a Shiny leaf widget that holds a sigint.ca/graphics/editor (text editor)
type TextEdit struct {
	node.LeafEmbed
	editor *editor.Editor

	NeedsRender bool
	Text        string
	ThemeColor  theme.Color
}

// NewTextEdit returns a new TextEdit widget.
func NewTextEdit(text string, face font.Face, opts *editor.OptionSet) *TextEdit {
	t := &TextEdit{
		editor:      editor.NewEditor(face, opts),
		Text:        text,
		NeedsRender: true,
	}
	t.editor.Buffer.InsertString(address.Simple{0, 0}, text)
	t.Wrapper = t
	return t
}

func (t *TextEdit) PaintBase(ctx *node.PaintBaseContext, origin image.Point) error {
	t.Marks.UnmarkNeedsPaintBase()
	dst := ctx.Dst.SubImage(t.Rect.Add(origin)).(*image.RGBA)
	if dst.Bounds().Empty() {
		return nil
	}
	log.Printf("dst.Bounds(): %v", dst.Bounds())
	t.editor.Draw(dst, dst.Bounds())
	t.NeedsRender = false
	return nil
}

func (t *TextEdit) OnInputEvent(ie interface{}, origin image.Point) node.EventHandled {
	log.Printf("Sending (%v) to (%v)\n", ie, t.Rect)
	switch e := ie.(type) {
	case key.Event:
		if e.Direction == key.DirPress || e.Direction == key.DirNone {
			t.editor.SendKeyEvent(e)
			t.Text = string(t.editor.Buffer.Contents())
			t.Mark(node.MarkNeedsPaintBase)
			t.NeedsRender = true
			return true
		}
	case mouse.Event:
		//t.editor.SendMouseEvent(e)
	}
	return false
}
