// Copyright 2016 Sonia Keys. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE_skeys file.

package mandelbrot

// https://soniacodes.wordpress.com/
// https://rosettacode.org/wiki/Mandelbrot_set#Go
// minor changes to make it a callable function

import (
	"image"
	"image/color"
	"image/draw"
	"math/cmplx"
)

func mandelbrot(a complex128, maxEsc float64) float64 {
	i := 0.0
	for z := a; cmplx.Abs(z) < 2 && i < maxEsc; i++ {
		z = z*z + a
	}
	return float64(maxEsc-i) / maxEsc
}

func generateMandelbrot(rMin, iMin, rMax, iMax float64, width, red, green, blue int, maxEsc int) image.Image {
	scale := float64(width) / (rMax - rMin)
	height := int(scale * (iMax - iMin))
	bounds := image.Rect(0, 0, width, height)
	b := image.NewRGBA(bounds)
	draw.Draw(b, bounds, image.NewUniform(color.Black), image.ZP, draw.Src)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			fEsc := mandelbrot(complex(
				float64(x)/scale+rMin,
				float64(y)/scale+iMin), float64(maxEsc))
			b.Set(x, y, color.RGBA{uint8(float64(red) * fEsc),
				uint8(float64(green) * fEsc), uint8(float64(blue) * fEsc), 255})

		}
	}
	return b
}
