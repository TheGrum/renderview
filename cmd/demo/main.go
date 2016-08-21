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
