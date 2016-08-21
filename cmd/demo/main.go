package main

import (
	"renderview"
	"renderview/examples/mandelbrot"
)

func main() {
	m := mandelbrot.NewMandelView()
	renderview.GtkWindowWithWidgetsInit(m)
	//	driver.Main(renderview.GetMainLoop(m))

}
