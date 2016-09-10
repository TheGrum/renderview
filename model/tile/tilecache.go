// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package model

import (
	"image"
	"sync"
)

type AdvancedTileProvider interface {
	TileProvider

	RenderTileRange(a Tile, b Tile) []TileImage
}

type StreamingTileProvider interface {
	TileProvider

	StreamTileRange(a Tile, b Tile) chan TileImage
}

type FallbackTileProvider interface {
	AdvancedTileProvider

	UsedFallback() bool
}

// TileImage is a simple struct returned by Advanced and Streaming Tile Providers
// that wraps a Tile and an Image together since these methods return more than one
// tile. The populated flag allows identification of zero value structs.
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

// TileCache provides a fixed size LRU cache for Tiles
// that calls a supplied TileProvider when necessary,
// and handles maintaining the cache, freshening accessed
// tiles and expiring the Least Recently Used tiles as
// necessary.
type TileCache struct {
	cache []TileImage

	provider TileProvider
	maxItems uint
}

// StreamingTileCache just wraps the TileCache and implements
// the StreamingTileProvider interface, which return a channel
// and sends Tiles down it as they are rendered instead of
// waiting for a complete tile set.
type StreamingTileCache struct {
	TileCache
}

// FallbackTileCache provides a fixed size LRU cache,
// similar to TileCache, but takes two TileProviders, a primary
// and a fallback. The fallback should be one that renders
// instantly. The primary can be a slower render that takes
// as long as needed. RenderTileRange renders exclusively from
// cached images from the primary tile provider and fills
// uncached spaces with fallback renderings that are never
// cached. A separate goroutine renders the primary tiles
// into the cache from a queue.
type FallbackTileCache struct {
	TileCache

	fallback     TileProvider
	requests     chan Tile
	usedFallback bool

	l sync.Mutex
}

func NewTileCache(provider TileProvider, maxItems uint) *TileCache {
	return &TileCache{
		cache:    make([]TileImage, 0, 100),
		provider: provider,
		maxItems: maxItems,
	}
}

func NewStreamingTileCache(provider TileProvider, maxItems uint) *StreamingTileCache {
	return &StreamingTileCache{
		TileCache: *NewTileCache(provider, maxItems),
	}
}

func NewFallbackTileCache(provider TileProvider, fallbackProvider TileProvider, maxItems uint) *FallbackTileCache {
	c := &FallbackTileCache{
		TileCache: *NewTileCache(provider, maxItems),

		fallback: fallbackProvider,
		requests: make(chan Tile, 1000),
	}
	go c.GoRender()
	return c
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
			found[(k.Tile.Y-a.Y)*w+(k.Tile.X-a.X)] = k
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
				found[i*int(w)+j] = k
				fcache = append(fcache, k)
			}
		}
	}

	// At the end of this process, fcache holds all the returned tiles
	// and nfcache holds what was not returned, so for LRU,
	// we want the last maxItems from nfcache+fcache
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

	return found
}

// Renders all tiles in the requested range and returns them
// This returns TileImages instead of merely images
// because the items are returned in an indeterminate order
// use the Tile to determine their position. a and b should
// be on the same zoom level
func (c *StreamingTileCache) StreamTileRange(a Tile, b Tile) chan TileImage {
	ch := make(chan TileImage, 20)
	go func() {
		if a.Z != b.Z {
			close(ch)
			return
		}

		// make sure that we can compare to a and b to determine if tile t is contained within
		a, b = SwapIfNeeded(a, b)
		w := b.X - a.X + 1
		h := b.Y - a.Y + 1

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
				found[(k.Tile.Y-a.Y)*w+(k.Tile.X-a.X)] = k
				ch <- k
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
					ch <- k
					found[i*int(w)+j] = k
					fcache = append(fcache, k)
				}
			}
		}

		// At the end of this process, fcache holds all the returned tiles
		// and nfcache holds what was not returned, so for LRU,
		// we want the last maxItems from nfcache+fcache
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
		close(ch)

	}()

	return ch
}

// Renders all tiles in the requested range and returns them
// This returns TileImages instead of merely images
// because the items are returned in an indeterminate order
// use the Tile to determine their position. a and b should
// be on the same zoom level
func (c *FallbackTileCache) RenderTileRange(a Tile, b Tile) []TileImage {
	if a.Z != b.Z {
		return nil
	}

	// make sure that we can compare to a and b to determine if tile t is contained within
	a, b = SwapIfNeeded(a, b)
	w := b.X - a.X + 1
	h := b.Y - a.Y + 1

	// We don't use ZoomFilter(c.cache, a.Z) here
	// because we want to efficiently move the found
	// items forward in the cache
	nfcache := make([]TileImage, 0, len(c.cache))
	fcache := make([]TileImage, 0, len(c.cache))

	// remember what we've found so far in here so
	// we know what to go and render at the end
	found := make([]TileImage, w*h, w*h)

	c.l.Lock()
	for _, k := range c.cache {
		if k.Tile.IsInside(a, b) {
			// this is one of the requested images
			found[(k.Tile.Y-a.Y)*w+(k.Tile.X-a.X)] = k
			fcache = append(fcache, k)
		} else {
			nfcache = append(nfcache, k)
		}
	}
	// At the end of this process, fcache holds all the returned tiles
	// and nfcache holds what was not returned, so for LRU,
	// we want the last maxItems from nfcache+fcache
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
	c.l.Unlock()

	c.usedFallback = false

	// Render all missing tiles using the fallback provider
	// Also add them to the request queue
	// Note that this results in items being added to the queue
	// more often than needed, so the queue servicing should
	// avoid rendering items that are still in the cache
	// don't add them to the fcache
	for i := 0; i < int(h); i++ {
		for j := 0; j < int(w); j++ {
			k := found[i*int(w)+j]
			if k.populated == false {
				tile := Tile{a.X + uint(j), a.Y + uint(i), a.Z}
				k = TileImage{true, tile, c.fallback.RenderTile(tile)}
				c.requests <- tile
				found[i*int(w)+j] = k
				c.usedFallback = true
				// Don't add to cache, because these are fallback images
				//fcache = append(fcache, k)
			}
		}
	}

	return found
}

func (c *FallbackTileCache) UsedFallback() bool {
	return c.usedFallback
}

// RenderTile returns an image from Cache, or if not found
// requests it from the TileProvider and returns it
func (c *FallbackTileCache) RenderTile(t Tile) image.Image {
	for i, k := range c.cache {
		if k.CompareTo(t) == 0 {
			// found the item, grab it
			r := k
			c.l.Lock()
			// pull it out of the list
			copy(c.cache[i:], c.cache[i+1:])
			// and put it back at the end
			c.cache = append(c.cache[:len(c.cache)-1], r)
			c.l.Unlock()
			return r.Img
		}
	}
	k := c.render(t)
	return k.Img
}

// Private function assumes that cache does NOT contain
// t; requests t from Provider and adds to cache,
// removing oldest item if necessary
func (c *FallbackTileCache) render(t Tile) TileImage {
	img := c.provider.RenderTile(t)
	c.l.Lock()
	if uint(len(c.cache)) >= c.maxItems {
		// drop oldest items
		c.cache = c.cache[(len(c.cache)-int(c.maxItems))+1:]
	}
	i := TileImage{true, t, img}
	c.cache = append(c.cache, i)
	c.l.Unlock()
	return i
}

// GoRender is called by NewFallbackTileCache and renders tiles
// from the request queue and stores them in the cache
func (c *FallbackTileCache) GoRender() {
	for k := range c.requests {
		// RenderTile has the behavior we want
		// of checking the cache first, and adding the image
		// from the primary TileProvider only
		// if it is not found in the cache
		_ = c.RenderTile(k)
	}
}
