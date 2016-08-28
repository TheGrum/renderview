// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package main

import "math/rand"

// Walls
const (
	W_NONE    = iota
	W_NORTH   = 1 << iota
	W_EAST    = 1 << iota
	W_SOUTH   = 1 << iota
	W_WEST    = 1 << iota
	W_VISITED = 1 << iota
	W_START   = 1 << iota
	W_END     = 1 << iota
	W_ALL     = W_NORTH | W_EAST | W_SOUTH | W_WEST
)

// Directions
const (
	D_NORTH = iota
	D_EAST  = iota
	D_SOUTH = iota
	D_WEST  = iota
)

type Maze struct {
	width  int
	height int

	cells []byte
}

// C returns the current value of cell x,y
func (m *Maze) C(x int, y int) byte {
	return m.cells[y*m.width+x]
}

// A adds flag v to the cell x,y
func (m *Maze) A(x int, y int, v byte) {
	m.cells[y*m.width+x] = m.cells[y*m.width+x] | v
}

// U removes flag v from the cell x,y
func (m *Maze) U(x int, y int, v byte) {
	m.cells[y*m.width+x] = m.cells[y*m.width+x] & (^v)
}

// S returns whether the cell x,y has flag v set
func (m *Maze) S(x int, y int, v byte) bool {
	return ((m.cells[y*m.width+x] & v) == v)
}

func NewDepthFirstMaze(width int, height int) *Maze {
	m := Maze{
		width:  width,
		height: height,

		cells: make([]byte, width*height),
	}

	for i := 0; i < width*height; i++ {
		m.cells[i] = W_EAST | W_SOUTH
	}

	x, y := 0, 0
	x = rand.Intn(width)
	y = rand.Intn(height)
	m.cells[y*width+x] = m.cells[y*width+x] | W_START
	DepthFirstStep(&m, x, y)
	for i := 0; i < width; i++ {
		m.A(i, 0, W_NORTH)
	}
	for i := 0; i < height; i++ {
		m.A(0, i, W_WEST)
	}
	return &m
}

func DepthFirstStep(m *Maze, x int, y int) {
	//	fmt.Printf("x: %v, y: %v, s: %v %v %v\n", x, y, m.S(x, y, W_VISITED), W_VISITED, m.C(x, y)&W_VISITED)
	m.A(x, y, W_VISITED)
	//	fmt.Printf("v: %v, s: %v\n", m.C(x, y), m.S(x, y, W_VISITED))
	// Loop over each direction, recursing into each
	// cell
	for _, d := range rand.Perm(4) {
		switch d {
		case D_NORTH:
			// if we aren't at the top already
			if y > 0 {
				if !m.S(x, y-1, W_VISITED) {
					m.U(x, y-1, W_SOUTH)
					DepthFirstStep(m, x, y-1)
				}

			}
		case D_EAST:
			if x < m.width-1 {
				if !m.S(x+1, y, W_VISITED) {
					m.U(x, y, W_EAST)
					DepthFirstStep(m, x+1, y)
				}
			}
		case D_SOUTH:
			if y < m.height-1 {
				if !m.S(x, y+1, W_VISITED) {
					m.U(x, y, W_SOUTH)
					DepthFirstStep(m, x, y+1)
				}
			}
		case D_WEST:
			if x > 0 {
				if !m.S(x-1, y, W_VISITED) {
					m.U(x-1, y, W_EAST)
					DepthFirstStep(m, x-1, y)
				}
			}
		}

	}
}
