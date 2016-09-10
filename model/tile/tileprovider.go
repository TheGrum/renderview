// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package model

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/llgcode/draw2d/draw2dimg"
)

type TileProvider interface {
	RenderTile(t Tile) image.Image
}

type TestTileProvider struct {
	Width  int
	Height int
}

func NewTestTileProvider(width, height int) *TestTileProvider {
	return &TestTileProvider{
		Width:  width,
		Height: height,
	}
}

func (p *TestTileProvider) RenderTile(t Tile) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	w, h := float64(p.Width), float64(p.Height)
	gc := draw2dimg.NewGraphicContext(img)
	gc.SetFillColor(color.White)
	gc.Clear()
	gc.SetStrokeColor(color.Black)
	gc.SetLineWidth(1)

	gc.MoveTo(0, 0)
	gc.LineTo(w, 0)
	gc.LineTo(w, h)
	gc.LineTo(0, h)
	gc.LineTo(0, 0)
	gc.Stroke()
	gc.SetFillColor(color.Black)
	gc.SetFontSize(12)
	gc.MoveTo(0, 0)
	gc.FillStringAt(fmt.Sprintf("Tile %dZ, %dY, %dX", t.Z, t.Y, t.X), 10, 70)
	fmt.Printf("Tile %dZ, %dY, %dX", t.Z, t.Y, t.X)
	return img
}

type OSMTileProvider struct {
	ServerURL string
}

func NewOSMTileProvider(ServerURL string) *OSMTileProvider {
	return &OSMTileProvider{
		ServerURL: ServerURL,
	}
}

func (o *OSMTileProvider) RenderTile(t Tile) image.Image {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				//log.Println("Dial!")
				return net.Dial(network, addr)
			},
			MaxIdleConnsPerHost: 50,
		},
	}
	request := o.ServerURL
	request = strings.Replace(request, "$Z", strconv.Itoa(int(t.Z)), -1)
	request = strings.Replace(request, "$Y", strconv.Itoa(int(t.Y)), -1)
	request = strings.Replace(request, "$X", strconv.Itoa(int(t.X)), -1)
	resp, err := client.Get(request)
	if err != nil {
		log.Println(err)
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	tile, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		log.Println(err)
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	return tile
}