// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

//package driver provides a generic interface to the renderview gui implementations
package driver

import (
	rv "github.com/TheGrum/renderview"
)

// FrameBuffer sets up a full-window rendering method
func FrameBuffer(m rv.RenderModel) {
	framebuffer(m)
}

// Main sets up a window with automatic parameter editing widgets
func Main(m rv.RenderModel) {
	main(m)
}
