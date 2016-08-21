// +build android

package driver

import (
    "renderview/driver/shiny"
    rv "renderview"
)

func framebuffer(m rv.RenderModel) {
    shiny.FrameBuffer(m)
}

func main(m rv.RenderModel) {
    shiny.Main(m)
}
