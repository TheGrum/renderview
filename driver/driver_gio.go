// Copyright 2020 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build gio

package driver

import (
	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/gio"
)

func framebuffer(m rv.RenderModel) {
	gio.FrameBuffer(m)
}

func main(m rv.RenderModel) {
	gio.Main(m)
}
