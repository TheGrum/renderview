#  TileRenderModel  #
=====================

TileRenderModel provides a model that consumes a set of interfaces to make displaying map tiles easy.

![map](http://i.imgur.com/MIwJRa5.png)

## LatLon 

A simple wrapper on a float64 latitude and longitude.

## Tile 

A Tile is a trio of integers defining an X, Y, and Z coordinate, where Z is a zoom, and X and Y vary between 0 and 2^Z-1. It provides helper functions for finding the tile at a lower zoom level that contains this tile, or the 4 child tiles contained within this tile at the next higher zoom level.

## TileMapper 

A TileMapper provides the functions necessary for converting between LatLons and Tiles.

Most usecases can just use tile.OSM for this.

## TileProvider 

A TileProvider renders a given Tile and returns an Image.

### AdvancedTileProvider 

A TileProvider that can also render a range of tiles.

### StreamingTileProvider 

A TileProvider that can render a range of tiles and return them down a channel.

### FallbackTileProvider 

An AdvancedTileProvider that can fall back to an alternate provider when a rendering fails.

### TestTileProvider 

Renders an image of a specified size with the tile coordinates drawn on it, for testing.

### OSMTileProvider 

The only concrete non-test provider currently. Accepts a Server URL for use to obtain tile images, replacing $X, $Y, and $Z in the URL with the tile coordinates.

### CompositingTileProvider 

Takes a list of TileProviders and renders them in order, compositing later images onto earlier ones in order.

## TileCache 

A LeastRecentlyUsed in-memory tile cache that fulfills the AdvancedTileProvider interface, takes a TileProvider to perform the actual rendering.

### StreamingTileCache 

Implements the StreamingTileProvider over a TileCache.

### FallbackTileCache 

Implements the FallbackTileProvider over a primary TileProvider and a fallback TileProvider, responding with fallback tiles and putting rendering requests in a queue to render in the background.

## TileRenderModel 

Takes a TileMapper and TileProvider and handles requesting and merging tiles into a single image to return to the RenderView.
