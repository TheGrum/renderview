// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package main

import (
	"github.com/TheGrum/renderview/driver"

	"github.com/TheGrum/renderview/examples/mandelbrot"
)

func main() {
	m := mandelbrot.NewMandelModel()
	driver.Main(m)
}
