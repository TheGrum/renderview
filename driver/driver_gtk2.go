// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build !nogtk2,!gotk3 !nogtk2,!android !nogtk2,!shiny

package driver

import (
	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/gtk2"
)

func framebuffer(m rv.RenderModel) {
	gtk2.FrameBuffer(m)
}

func main(m rv.RenderModel) {
	gtk2.Main(m)
}
