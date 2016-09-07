// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package model

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	rv "github.com/TheGrum/renderview"
)

type TileRenderModel struct {
	rv.EmptyRenderModel

	RequestRender chan interface{}
	Rendering     bool
	NeedsRender   bool
	Img           image.Image
	mapper        TileMapper
	provider      TileProvider

	lastTopLeft     Tile
	lastBottomRight Tile
	//	lastWidth       int
	//	lastHeight      int

	left    rv.RenderParameter
	right   rv.RenderParameter
	top     rv.RenderParameter
	bottom  rv.RenderParameter
	width   rv.RenderParameter
	height  rv.RenderParameter
	started bool
}

func NewTileRenderModel(mapper TileMapper, provider TileProvider, leftTop LatLon, bottomRight LatLon) *TileRenderModel {
	m := TileRenderModel{
		EmptyRenderModel: rv.EmptyRenderModel{
			Params: make([]rv.RenderParameter, 0, 10),
		},
		RequestRender: make(chan interface{}, 10),
		mapper:        mapper,
		provider:      provider,
		started:       false,
	}
	m.AddParameters(rv.DefaultParameters(false, rv.HINT_HIDE, rv.OPT_AUTO_ZOOM, leftTop.Lon, leftTop.Lat, bottomRight.Lon, bottomRight.Lat)...)
	m.AddParameters(rv.SetHints(rv.HINT_HIDE, rv.NewFloat64RP("zoomRate", 0.50))...)
	m.left = m.GetParameter("left")
	m.top = m.GetParameter("top")
	m.right = m.GetParameter("right")
	m.bottom = m.GetParameter("bottom")
	m.width = m.GetParameter("width")
	m.height = m.GetParameter("height")
	return &m
}

func (m *TileRenderModel) Render() image.Image {
	if !m.Rendering {
		m.RequestRender <- true
		m.NeedsRender = false
	} else {
		m.NeedsRender = true
	}
	return m.Img
}

func (m *TileRenderModel) Start() {
	if !m.started {
		m.started = true
		go m.GoRender()
	}
}

func (m *TileRenderModel) GoRender() {
	for {
		select {
		case <-m.RequestRender:
			m.InnerRender()
		}
	}
}

func (m *TileRenderModel) InnerRender() {

	img, ok := m.Img.(*image.RGBA)
	if img == nil || !ok {
		i2 := image.NewRGBA(image.Rect(0, 0, m.width.GetValueInt(), m.height.GetValueInt()))
		img = i2
		m.Img = img
	} else {
		b := img.Bounds()
		if b.Dx() != m.width.GetValueInt() || b.Dy() != m.height.GetValueInt() {
			i2 := image.NewRGBA(image.Rect(0, 0, m.width.GetValueInt(), m.height.GetValueInt()))
			// maybe do something here to copy/move the previous image?
			img = i2
			m.Img = img
		}
	}
	a := LatLon{m.top.GetValueFloat64(), m.left.GetValueFloat64()}
	b := LatLon{m.bottom.GetValueFloat64(), m.right.GetValueFloat64()}
	c, d := m.mapper.TilesFromBounds(a, b, uint(img.Bounds().Dx()), uint(img.Bounds().Dy()))
	modA, modB := m.mapper.BoundsFromTiles(c, d)
	ca, cb := m.mapper.BoundsFromTiles(c, c)

	var tileSizeX, tileSizeY float64
	var i, j int
	var offsetP image.Point
	m.Rendering = true
	w, h := d.X-c.X, d.Y-c.Y
	switch t := m.provider.(type) {
	case AdvancedTileProvider:
		tiles := t.RenderTileRange(c, d)
		if len(tiles) > 0 {
			i2 := tiles[0].Img
			if i2 == nil {
				for _, k := range tiles {
					if k.Img != nil {
						i2 = k.Img
						break
					}
				}
				if i2 == nil {
					panic("Unable to find image")
				}
			}
			tileSizeX = float64(i2.Bounds().Dx())
			actualWidth := tileSizeX * float64(w)
			actualLonRange := modB.Lon - modA.Lon
			tileLonWidth := tileSizeX / (cb.Lon - ca.Lon) // actualWidth / actualLonRange
			offsetP.X = int(math.Floor((modA.Lon - a.Lon) * tileLonWidth))
			fmt.Printf("targetLon %v caLon %v cbLon %v tileSizeX %v, actualWidth %v, actualLonRange %v, tileLonWidth %v, offsetP.X %v\n", a.Lon, ca.Lon, cb.Lon, tileSizeX, actualWidth, actualLonRange, tileLonWidth, offsetP.X)
			tileSizeY = float64(i2.Bounds().Dy())
			//			actualHeight := tileSizeY * float64(h)
			//			actualLatRange := modB.Lat - modA.Lat
			tileLatHeight := tileSizeY / (cb.Lat - ca.Lat) //actualHeight / actualLatRange
			offsetP.Y = int(math.Floor((modA.Lat - a.Lat) * tileLatHeight))
			fmt.Printf("Offset calculated %v\n", offsetP)
			//offsetP.X = 0
			//offsetP.Y = 0
		}
		for _, k := range tiles {
			i = int(k.Tile.Y - c.Y)
			j = int(k.Tile.X - c.X)
			i2 := k.Img
			draw.Draw(img, image.Rect(offsetP.X+j*int(tileSizeX), offsetP.Y+i*int(tileSizeY), offsetP.X+(j+1)*int(tileSizeX), offsetP.Y+(i+1)*int(tileSizeY)), i2, image.ZP, draw.Src)
		}
	case TileProvider:
		for i = 0; i < int(w); i++ {
			for j = 0; j < int(h); j++ {
				i2 := t.RenderTile(Tile{c.Z, c.Y + uint(i), c.X + uint(j)})
				if i == 0 && j == 0 {
					tileSizeX = float64(i2.Bounds().Dx())
					actualWidth := tileSizeX * float64(w)
					actualLonRange := modB.Lon - modA.Lon
					tileLonWidth := actualLonRange / actualWidth
					offsetP.X = int(math.Floor((a.Lon - modA.Lon) * tileSizeX / tileLonWidth))
					tileSizeY = float64(i2.Bounds().Dy())
					actualHeight := tileSizeY * float64(h)
					actualLatRange := modB.Lat - modA.Lat
					tileLatHeight := actualLatRange / actualHeight
					offsetP.Y = int(math.Floor((a.Lat - modA.Lat) * tileSizeY / tileLatHeight))
				}
				draw.Draw(img, image.Rect(0, 0, int(tileSizeX), int(tileSizeY)), i2, image.Point{offsetP.X + j*int(tileSizeX), offsetP.Y + i*int(tileSizeY)}, draw.Src)
			}
		}
	}
	m.Rendering = false
}
