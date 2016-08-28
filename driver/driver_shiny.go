// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build android shiny

package driver

import (
	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/shiny"
)

func framebuffer(m rv.RenderModel) {
	shiny.FrameBuffer(m)
}

func main(m rv.RenderModel) {
	shiny.Main(m)
}
