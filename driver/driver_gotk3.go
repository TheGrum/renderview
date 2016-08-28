// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build gotk3,linux,!android

package driver

import (
	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/gotk3"
)

func framebuffer(m rv.RenderModel) {
	gotk3.FrameBuffer(m)
}

func main(m rv.RenderModel) {
	gotk3.Main(m)
}
