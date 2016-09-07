// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package model

import (
	"fmt"
	"image"
)

type TileImage struct {
	populated bool

	Tile Tile
	Img  image.Image
}

func (c TileImage) CompareTo(t Tile) int {
	return c.Tile.CompareTo(t)
}

// 3200x2400 window in 256x256 px frames is approx 117 frames
// 1600x1200 window is approx 30 frames
// storing less than one windows worth of frames
// means a single window would flush the cache, making it
// pointless, unless the user of the cache can make
// its request in a way that takes full advantage of
// what is left in the cache (which means explicitly NOT
// simply iterating in a loop pair.)

type TileCache struct {
	cache []TileImage

	provider TileProvider
	maxItems uint
}

type AdvancedTileProvider interface {
	TileProvider

	RenderTileRange(a Tile, b Tile) []TileImage
}

func NewTileCache(provider TileProvider, maxItems uint) *TileCache {
	return &TileCache{
		cache:    make([]TileImage, 0, 100),
		provider: provider,
		maxItems: maxItems,
	}
}

// RenderTile returns an image from Cache, or if not found
// requests it from the TileProvider and returns it
func (c *TileCache) RenderTile(t Tile) image.Image {
	for i, k := range c.cache {
		if k.CompareTo(t) == 0 {
			// found the item, grab it
			r := k
			// pull it out of the list
			copy(c.cache[i:], c.cache[i+1:])
			// and put it back at the end
			c.cache = append(c.cache[:len(c.cache)-1], r)
			return r.Img
		}
	}
	k := c.render(t)
	return k.Img
}

// Private function assumes that cache does NOT contain
// t; requests t from Provider and adds to cache,
// removing oldest item if necessary
func (c *TileCache) render(t Tile) TileImage {
	img := c.provider.RenderTile(t)
	if uint(len(c.cache)) >= c.maxItems {
		// drop oldest items
		c.cache = c.cache[(len(c.cache)-int(c.maxItems))+1:]
	}
	i := TileImage{true, t, img}
	c.cache = append(c.cache, i)
	return i
}

/*
func ZoomFilter(c []TileImage, zoomLevel uint) []TileImage {
	r := make([]TileImage, 0, len(c))
	for _, k := range c {
		if k.Tile.Z == zoomLevel {
			r = append(r, k)
		}
	}
	return r
}
*/

// SwapIfNeeded is a utility function that ensures we have the upper left
// and bottom right tiles
func SwapIfNeeded(a Tile, b Tile) (Tile, Tile) {
	if b.Y < a.Y {
		a, b = b, a
	}
	// Now a.Y must be less then b.Y,
	// we want a.X to be less then b.X as well
	if b.X < a.X {
		a.X, b.X = b.X, a.X
	}
	return a, b
}

// Renders all tiles in the requested range and returns them
// This returns TileImages instead of merely images
// because the items are returned in an indeterminate order
// use the Tile to determine their position. a and b should
// be on the same zoom level
func (c *TileCache) RenderTileRange(a Tile, b Tile) []TileImage {
	if a.Z != b.Z {
		return nil
	}

	// make sure that we can compare to a and b to determine if tile t is contained within
	a, b = SwapIfNeeded(a, b)
	w := b.X - a.X + 1
	h := b.Y - a.Y + 1
	fmt.Printf("a(%v) b(%v) w(%v) h(%v)\n", a, b, w, h)

	// We don't use ZoomFilter(c.cache, a.Z) here
	// because we want to efficiently move the found
	// items forward in the cache
	nfcache := make([]TileImage, 0, len(c.cache))
	fcache := make([]TileImage, 0, len(c.cache))

	// remember what we've found so far in here so
	// we know what to go and render at the end
	found := make([]TileImage, w*h, w*h)

	for _, k := range c.cache {
		if k.Tile.IsInside(a, b) {
			// this is one of the requested images
			//fmt.Printf("checking a(%v) k(%v) pos %d\n", a, k, (k.Tile.Y-a.Y)*w+(k.Tile.X-a.X))
			found[(k.Tile.Y-a.Y)*w+(k.Tile.X-a.X)] = k
			//fmt.Printf("Found tile (%v)\n", k)
			fcache = append(fcache, k)
		} else {
			nfcache = append(nfcache, k)
		}
	}

	// Render all missing tiles and add them to the fcache
	for i := 0; i < int(h); i++ {
		for j := 0; j < int(w); j++ {
			//fmt.Printf("Checking i %d, j %d, kpos %d\n", i, j, i*int(w)+j)
			k := found[i*int(w)+j]
			//fmt.Printf("Checking k(%v)\n", k)
			if k.populated == false {
				tile := Tile{a.X + uint(j), a.Y + uint(i), a.Z}
				k = TileImage{true, tile, c.provider.RenderTile(tile)}
				fmt.Printf("%v\n", k)
				found[i*int(w)+j] = k
				fcache = append(fcache, k)
			}
		}
	}

	// At the end of this process, fcache holds all the returned tiles
	// and nfcache holds what was not returned, so for LRU,
	// we want the last maxItems from nfcache+fcache
	//fmt.Printf("maxItems(%d), fcache(%v), nfcache(%v)\n", c.maxItems, fcache, nfcache)
	fnum := len(fcache)
	nnum := len(nfcache)

	if fnum > int(c.maxItems) {
		fnum = int(c.maxItems)
		nnum = 0
	} else if nnum+fnum > int(c.maxItems) {
		nnum = int(c.maxItems) - fnum
	}

	if nnum > 0 {
		c.cache = append(nfcache[len(nfcache)-nnum:], fcache...)
	} else {
		c.cache = fcache[len(fcache)-fnum:]
	}

	//	fmt.Printf("Returning %v\n", found)
	return found
}
