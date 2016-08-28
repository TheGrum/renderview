// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

<<<<<<< HEAD
// +build !gotk3,linux,!android
=======
// +build linux,!android
>>>>>>> e0fc460b97ddda8c1fa010b4ad8573c945ecde99

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
