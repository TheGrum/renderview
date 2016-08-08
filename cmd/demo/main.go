package main

import (
	"renderview"
	"renderview/examples/mandelbrot"

	"golang.org/x/exp/shiny/driver"
)

func main() {
	m := mandelbrot.NewMandelView()
	driver.Main(renderview.GetMainLoop(m))
}
