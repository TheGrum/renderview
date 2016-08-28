// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build example

package main

import (
	"flag"
	"image/gif"
	"log"
	"math"
	"os"
	"time"

	rv "github.com/TheGrum/renderview"
	"github.com/TheGrum/renderview/driver/shiny"
)

func main() {
	flag.Parse()

	var (
		f   *os.File
		err error
	)
	if flag.NArg() > 0 {
		f, err = os.Open(flag.Arg(0))
		handleError(err)
	} else {
		f, err = os.Open("test.gif")
		handleError(err)
	}

	images, err := gif.DecodeAll(f)
	handleError(err)
	f.Close()

	numImages := len(images.Image)

	start := time.Now()
	m := rv.NewBasicRenderModel()
	m.AddParameters(rv.NewIntRP("page", 0))
	page := m.GetParameter("page")
	m.InnerRender = func() {
		p := page.GetValueInt()
		if p > numImages {
			p = numImages
			page.SetValueInt(p)
		} else if p < 0 {
			p = 0
			page.SetValueInt(p)
		}
		if p > 0 {
			m.Img = images.Image[p-1]
		} else {
			d := int(math.Floor(time.Since(start).Seconds()*250)) % numImages
			m.Img = images.Image[d]
		}
	}
	go func(m *rv.BasicRenderModel) {
		ticker := time.NewTicker(time.Millisecond * 250)
		for _ = range ticker.C {
			m.RequestPaint()
		}
	}(m)
	shiny.FrameBuffer(m)
}

func handleError(err error) {
	if !(err == nil) {
		log.Fatal(err)
	}
}
