// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package model

import (
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
	modA, _ := m.mapper.BoundsFromTiles(c, d)
	ca, cb := m.mapper.BoundsFromTiles(c, c)

	var tileSizeX, tileSizeY float64
	var i, j int
	var offsetP image.Point
	m.Rendering = true
	w, h := d.X-c.X, d.Y-c.Y
	switch t := m.provider.(type) {
	case StreamingTileProvider:
		tilech := t.StreamTileRange(c, d)
		i = 0
		for k := range tilech {
			i2 := k.Img
			if (i == 0) && i2 != nil {
				tileSizeX = float64(i2.Bounds().Dx())
				tileLonWidth := tileSizeX / (cb.Lon - ca.Lon)
				offsetP.X = int(math.Floor((modA.Lon - a.Lon) * tileLonWidth))
				tileSizeY = float64(i2.Bounds().Dy())
				tileLatHeight := tileSizeY / (cb.Lat - ca.Lat)
				offsetP.Y = int(math.Floor((modA.Lat - a.Lat) * tileLatHeight))
				// Now correct right/bottom to actual screen range
				//				m.left.SetValueFloat64(a.Lon + tileLonWidth*tileSizeX*float64(w))
				//				m.bottom.SetValueFloat64(a.Lat + tileLatHeight*tileSizeX*float64(h))
				i = 1
			}
			i = int(k.Tile.Y - c.Y)
			j = int(k.Tile.X - c.X)
			i2 = k.Img
			draw.Draw(img, image.Rect(offsetP.X+j*int(tileSizeX), offsetP.Y+i*int(tileSizeY), offsetP.X+(j+1)*int(tileSizeX), offsetP.Y+(i+1)*int(tileSizeY)), i2, image.ZP, draw.Src)
			m.RequestPaint()
		}
	case FallbackTileProvider:
		tiles := t.RenderTileRange(c, d)
		usedFallback := t.UsedFallback()
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
			tileLonWidth := tileSizeX / (cb.Lon - ca.Lon)
			offsetP.X = int(math.Floor((modA.Lon - a.Lon) * tileLonWidth))
			tileSizeY = float64(i2.Bounds().Dy())
			tileLatHeight := tileSizeY / (cb.Lat - ca.Lat)
			offsetP.Y = int(math.Floor((modA.Lat - a.Lat) * tileLatHeight))
			// Now correct right/bottom to actual screen range
			//m.left.SetValueFloat64(a.Lon + tileLonWidth*float64(m.width.GetValueInt()))
			//m.bottom.SetValueFloat64(a.Lat + tileLatHeight*float64(m.height.GetValueInt()))
		}
		for _, k := range tiles {
			i = int(k.Tile.Y - c.Y)
			j = int(k.Tile.X - c.X)
			i2 := k.Img
			draw.Draw(img, image.Rect(offsetP.X+j*int(tileSizeX), offsetP.Y+i*int(tileSizeY), offsetP.X+(j+1)*int(tileSizeX), offsetP.Y+(i+1)*int(tileSizeY)), i2, image.ZP, draw.Src)
			m.RequestPaint()
		}
		if usedFallback {
			// There are fallback tiles present, so queue another render
			m.NeedsRender = true
			m.RequestPaint()
		}
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
			tileLonWidth := tileSizeX / (cb.Lon - ca.Lon)
			offsetP.X = int(math.Floor((modA.Lon - a.Lon) * tileLonWidth))
			tileSizeY = float64(i2.Bounds().Dy())
			tileLatHeight := tileSizeY / (cb.Lat - ca.Lat)
			offsetP.Y = int(math.Floor((modA.Lat - a.Lat) * tileLatHeight))
			// Now correct right/bottom to actual screen range
			//fmt.Printf("a.Lon %v modA.Lon %v ca.Lon %v cb.Lon %v tileLonWidth %v Expected right lon %v actual right lon %v offsetP.X %v\n", a.Lon, modA.Lon, ca.Lon, cb.Lon, tileLonWidth, a.Lon+float64(m.width.GetValueInt())/tileLonWidth, b.Lon, offsetP.X)
			//m.left.SetValueFloat64(a.Lon + float64(m.width.GetValueInt()-1)/tileLonWidth)
			//m.bottom.SetValueFloat64(a.Lat + float64(m.height.GetValueInt()-1)/tileLatHeight)
		}
		for _, k := range tiles {
			i = int(k.Tile.Y - c.Y)
			j = int(k.Tile.X - c.X)
			i2 := k.Img
			draw.Draw(img, image.Rect(offsetP.X+j*int(tileSizeX), offsetP.Y+i*int(tileSizeY), offsetP.X+(j+1)*int(tileSizeX), offsetP.Y+(i+1)*int(tileSizeY)), i2, image.ZP, draw.Src)
			m.RequestPaint()
		}
	case TileProvider:
		for i = 0; i < int(w); i++ {
			for j = 0; j < int(h); j++ {
				i2 := t.RenderTile(Tile{c.Z, c.Y + uint(i), c.X + uint(j)})
				if i == 0 && j == 0 {
					tileSizeX = float64(i2.Bounds().Dx())
					tileLonWidth := tileSizeX / (cb.Lon - ca.Lon)
					offsetP.X = int(math.Floor((modA.Lon - a.Lon) * tileLonWidth))
					tileSizeY = float64(i2.Bounds().Dy())
					tileLatHeight := tileSizeY / (cb.Lat - ca.Lat)
					offsetP.Y = int(math.Floor((modA.Lat - a.Lat) * tileLatHeight))
					// Now correct right/bottom to actual screen range
					//					m.left.SetValueFloat64(a.Lon + tileLonWidth*tileSizeX*float64(w))
					//					m.bottom.SetValueFloat64(a.Lat + tileLatHeight*tileSizeX*float64(h))
				}
				draw.Draw(img, image.Rect(0, 0, int(tileSizeX), int(tileSizeY)), i2, image.Point{offsetP.X + j*int(tileSizeX), offsetP.Y + i*int(tileSizeY)}, draw.Src)
				m.RequestPaint()
			}
		}
	}
	m.Rendering = false
}
