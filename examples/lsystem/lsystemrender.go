package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	rv "renderview"

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

func RenderLSystem(m *rv.BasicRenderModel) image.Image {
	left := m.Params[0].GetValueFloat64()
	top := m.Params[1].GetValueFloat64()
	right := m.Params[2].GetValueFloat64()
	//bottom := m.Params[3].GetValueFloat64()
	width := m.Params[4].GetValueInt()
	height := m.Params[5].GetValueInt()

	//lsystem := m.GetParameter("lsystem").GetValueString()
	angle := m.GetParameter("angle").GetValueFloat64()
	//depth := m.GetParameter("depth").GetValueInt()
	magnitude := 100.0 / (right - left)

	bounds := image.Rect(0, 0, width, height)
	b := image.NewRGBA(bounds)

	gc := draw2dimg.NewGraphicContext(b)
	gc.SetFillColor(color.White)
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(1)

	location := FPoint{0 - left, 0 - top}
	nextLocation := FPoint{0, 0}
	direction := 90.0
	theta := direction * radcon

	finallsystem := "FX+FX+FX+FX"
	stack := make([]State, 0, 100)

	fmt.Printf("Rendering %s\n", finallsystem)
	for _, rn := range finallsystem {
		theta = direction * radcon
		switch rn {
		case '0', '1', '2', '3', '4', '5', 'A', 'B', 'C', 'D', 'E', 'F', 'G':
			nextLocation = location.Add(FPoint{magnitude * math.Cos(theta), magnitude * math.Sin(theta)})
			gc.MoveTo(location.X, location.Y)
			gc.LineTo(nextLocation.X, nextLocation.Y)
			fmt.Printf("Drawing from %v to %v\n", location, nextLocation)
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
			}
		}
	}
	gc.Stroke()
	return b
}
