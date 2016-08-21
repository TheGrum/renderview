//package driver provides a generic interface to the renderview gui implementations
package driver

import (
    rv "renderview"
    )

// FrameBuffer sets up a full-window rendering method
func FrameBuffer(m rv.RenderModel) {
    framebuffer(m)
}

// Main sets up a window with automatic parameter editing widgets
func Main(m rv.RenderModel) {
    main(m)
}

