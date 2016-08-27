// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package main

import (
	"renderview"
	"renderview/examples/mandelbrot"
)

func main() {
	m := mandelbrot.NewMandelModel()
	renderview.GtkWindowWithWidgetsInit(m)
	//	driver.Main(renderview.GetMainLoop(m))

}
