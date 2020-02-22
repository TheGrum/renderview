// Copyright 2020 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build fyne

package driver

import (
	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/fyne"
)

func framebuffer(m rv.RenderModel) {
	fyne.FrameBuffer(m)
}

func main(m rv.RenderModel) {
	fyne.Main(m)
}
