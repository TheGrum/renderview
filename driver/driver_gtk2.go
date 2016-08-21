// +build linux,!android

package driver

import (
    "renderview/driver/gtk2"
    rv "renderview"
)

func framebuffer(m rv.RenderModel) {
    gtk2.FrameBuffer(m)
}

func main(m rv.RenderModel) {
    gtk2.Main(m)
}

