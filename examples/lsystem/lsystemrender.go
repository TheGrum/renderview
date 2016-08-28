// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

// +build example

package main

import (
	"image"
	"image/color"
	"math"

	rv "github.com/TheGrum/renderview"

	"github.com/llgcode/draw2d/draw2dimg"
)

const radcon = math.Pi / 180

type State struct {
	Location  FPoint
	Direction float64
}

type FPoint struct {
	X, Y float64
}

func (f FPoint) Add(a FPoint) FPoint {
	return FPoint{f.X + a.X, f.Y + a.Y}
}

func (f FPoint) Sub(a FPoint) FPoint {
	return FPoint{f.X - a.X, f.Y - a.Y}
}

func FMin(a float64, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func FMax(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func RenderLSystemModel(m *rv.BasicRenderModel, c *rv.ChangeMonitor) image.Image {
	left := m.Params[0].GetValueFloat64()
	top := m.Params[1].GetValueFloat64()
	right := m.Params[2].GetValueFloat64()
	bottom := m.Params[3].GetValueFloat64()
	width := m.Params[4].GetValueInt()
	height := m.Params[5].GetValueInt()

	lsystem := m.GetParameter("lsystem").GetValueString()
	angle := m.GetParameter("angle").GetValueFloat64()
	depth := m.GetParameter("depth").GetValueInt()
	if depth > 20 {
		depth = 20
		m.GetParameter("depth").SetValueInt(20)
	}
	bounds := image.Rect(0, 0, width, height)
	result := m.GetParameter("LSystemResult").GetValueString()
	magnitude := 1.0 //5 * float64(width) / (right - left)

	if c.HasChanged() {
		// lsystem or depth has changed, recalculate
		result = Calculate(lsystem, depth)
		m.GetParameter("LSystemResult").SetValueString(result)
		_, minX, minY, dx, dy := RenderLSystem(left, top, right, bottom, bounds, angle, 1, result)
		//fmt.Printf("Applying %v,%v %vx%v mag:%v calmag:%v\n", minX, minY, dx, dy, magnitude, 5*(dx/float64(width)))
		//mult := (float64(width) / dx) / 5
		left = minX - 1 //* mult
		top = minY - 1  //* mult
		//		mult := magnitude
		right = left + dx + 2 //float64(width)         //* magnitude)
		bottom = top + dy + 2 //*(dx/float64(width)) //* magnitude)
		//fmt.Printf("Final %v,%v,%v,%v\n", left, top, right, bottom)
		magnitude = 1.0 //5 * float64(width) / (right - left)

		m.Params[0].SetValueFloat64(left)
		m.Params[1].SetValueFloat64(top)
		m.Params[2].SetValueFloat64(right)
		m.Params[3].SetValueFloat64(bottom)
		m.RequestPaint()

	}
	img, _, _, _, _ := RenderLSystem(left, top, right, bottom, bounds, angle, magnitude, result)
	return img
}

func RenderLSystem(left float64, top float64, right float64, bottom float64, bounds image.Rectangle, angle float64, magnitude float64, lsystem string) (*image.RGBA, float64, float64, float64, float64) {

	b := image.NewRGBA(bounds)

	gc := draw2dimg.NewGraphicContext(b)
	gc.SetFillColor(color.White)
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(1)

	dx := right - left
	//dy := bottom - top
	mx := float64(bounds.Dx()) / dx
	my := float64(bounds.Dx()) / dx

	location := FPoint{0, 0}
	nextLocation := FPoint{0, 0}
	direction := 90.0
	theta := direction * radcon

	maxX := 0.0
	minX := 0.0
	maxY := 0.0
	minY := 0.0

	stack := make([]State, 0, 100)

	for _, rn := range lsystem {
		theta = direction * radcon
		switch rn {
		case '0', '1', '2', '3', '4', '5', 'A', 'B', 'C', 'D', 'E', 'F', 'G':
			nextLocation = location.Add(FPoint{magnitude * math.Cos(theta), magnitude * math.Sin(theta)})
			gc.MoveTo((location.X-left)*mx, (location.Y-top)*my)
			gc.LineTo((nextLocation.X-left)*mx, (nextLocation.Y-top)*my)
			//fmt.Printf("Drawing from %v to %v\n", location, nextLocation)
			location = nextLocation
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g':
			nextLocation = location.Add(FPoint{magnitude * math.Cos(theta), magnitude * math.Sin(theta)})
			location = nextLocation
		case '+':
			direction -= angle
		case '-':
			direction += angle
		case '[':
			stack = append(stack, State{location, direction})
		case ']':
			if len(stack) > 0 {
				state := stack[len(stack)-1]
				location = state.Location
				direction = state.Direction
				stack = stack[:len(stack)-1]
			}
		}
		maxX = FMax(maxX, location.X)
		minX = FMin(minX, location.X)
		maxY = FMax(maxY, location.Y)
		minY = FMin(minY, location.Y)
		//fmt.Printf("(%v),", location)
	}
	gc.Stroke()
	//fmt.Printf("%v, %v, %v, %v, %v, %v\n", minX, minY, maxX, maxY, location.X, location.Y)

	return b, minX, minY, maxX - minX, maxY - minY

}
