package model

import "math"

type Tile struct {
	X uint
	Y uint
	Z uint
}

func (t Tile) CompareTo(a Tile) int {
	if t.Z < a.Z {
		return -1
	}
	if t.Z > a.Z {
		return 1
	}
	// zoomLevels are equal
	if t.Y < a.Y {
		return -1
	}
	if t.Y > a.Y {
		return 1
	}
	// latitudes are equal
	if t.X < a.X {
		return -1
	}
	if t.X > a.X {
		return 1
	}
	// all are equal
	return 0
}

func (t Tile) IsInside(a Tile, b Tile) bool {
	if t.Z != a.Z || t.Z != b.Z {
		return false
	}
	if (t.X >= a.X) && (t.X <= b.X) && (t.Y >= a.Y) && (t.Y <= b.Y) {
		return true
	}
	return false
}

func (t Tile) GetParentTile() Tile {
	if t.Z == 0 {
		return t
	} else {
		return Tile{
			Z: t.Z - 1,
			Y: t.Y / 2,
			X: t.X / 2,
		}
	}
}

func (t Tile) GetChildTile(left bool, top bool) Tile {
	switch true {
	case left && top:
		return Tile{
			Z: t.Z + 1,
			Y: t.Y * 2,
			X: t.X * 2,
		}
	case !left && top:
		return Tile{
			Z: t.Z + 1,
			Y: t.Y * 2,
			X: t.X*2 + 1,
		}
	case left && !top:
		return Tile{
			Z: t.Z + 1,
			Y: t.Y*2 + 1,
			X: t.X * 2,
		}
	default:
		return Tile{
			Z: t.Z + 1,
			Y: t.Y*2 + 1,
			X: t.X*2 + 1,
		}
	}
}

type LatLon struct {
	Lat float64
	Lon float64
}

type TileMapper interface {
	LatLon(t Tile) LatLon
	Tile(l LatLon, Zoom uint) Tile
	TilesFromBounds(a LatLon, b LatLon, maxWidth uint, maxHeight uint) (Tile, Tile)
	BoundsFromTiles(a Tile, b Tile) (LatLon, LatLon)
}

func ShiftToBottomRight(t Tile) Tile {
	max := uint(2<<uint(t.Z) - 1)
	if t.X < max {
		t.X += 1
	}
	if t.Y < max {
		t.Y += 1
	}
	return t
}

// A generic, mechanical method for determining a proper zoom level and set of
// tiles for a given image dimension and coordinate range by iteration, such that
// the image resulting from rendering the given tiles and cutting to the
// specified lat and lon will be smaller in both dimensions than the maxWidth
// and maxHeight given
func GenericTilesFromBounds(t TileMapper, a LatLon, b LatLon, maxWidth uint, maxHeight uint, tileSize uint, maxZoomLevel uint) (Tile, Tile) {
	//maxXTiles := maxWidth / tileSize
	//maxYTiles := maxHeight / tileSize

	w := math.Abs(b.Lon - a.Lon)
	h := math.Abs(b.Lat - a.Lat)
	var c, d Tile
	zoomLevel := maxZoomLevel
	for {
		// Get the tiles containing a and b at current zoomLevel
		c = t.Tile(a, zoomLevel)
		d = t.Tile(b, zoomLevel)
		//fmt.Printf("c,d: %v, %v\n", c, d)
		// Obtain the resulting lat/lon outer bounds
		e, f := t.BoundsFromTiles(c, d)
		// Calculate the resulting width and height
		w2 := math.Abs(f.Lon - e.Lon)
		h2 := math.Abs(f.Lat - e.Lat)
		// Get the pixel dimensions of the whole image
		g := math.Abs(float64((d.X - c.X) * tileSize))
		i := math.Abs(float64((d.Y - c.Y) * tileSize))
		// And adjust it to match the results of clipping to a and b
		l := uint(math.Ceil(g * w / w2))
		m := uint(math.Ceil(i * h / h2))

		// if we have exceeded either bound, loop again, decreasing the zoomLevel,
		// unless we are already at the minimum zoom
		if (zoomLevel > 0) && ((l > maxWidth) || (m > maxHeight)) {
			zoomLevel -= 1
		} else {
			break
		}
	}
	// push d to bottom right since left/top can actually be
	// within tile c, shifting image up and to the left
	return c, ShiftToBottomRight(d)
}

// GenericBoundsFromTiles returns the lat/lon pairs that encompass
// a set of tiles
func GenericBoundsFromTiles(t TileMapper, a Tile, b Tile, rightbottom LatLon) (LatLon, LatLon) {
	// left is easy
	var c, d LatLon
	c = t.LatLon(a)
	max := uint(2<<uint(b.Z) - 1)
	if b.X == max {
		if b.Y == max {
			d = rightbottom
		} else {
			d = t.LatLon(Tile{b.X, b.Y + 1, b.Z})
			d.Lon = rightbottom.Lon
		}
	} else {
		if b.Y == max {
			d = t.LatLon(Tile{b.X + 1, b.Y, b.Z})
			d.Lat = rightbottom.Lat
		} else {
			d = t.LatLon(Tile{b.X + 1, b.Y + 1, b.Z})
		}
	}
	return c, d
}

type OSMTileMapper struct {
	TileSize uint
}

var OSM = OSMTileMapper{256}

// Calculates the appropriate OSM Tile position that, at the given zoomLevel, contains the given lat/lon.
func (o OSMTileMapper) Tile(l LatLon, zoomLevel uint) Tile {
	var fx, fy float64
	var x, y uint
	var shift float64 = math.Exp2(float64(zoomLevel)) //float64(uint(1) << zoomLevel)

	fx = (l.Lon + 180) / 360 * shift
	fy = (1.0 - math.Log(math.Tan(l.Lat*math.Pi/180.0)+1.0/math.Cos(l.Lat*math.Pi/180.0))/math.Pi) / 2.0 * shift

	x = uint(math.Floor(fx))
	y = uint(math.Floor(fy))
	if x < 0 || x > (2<<zoomLevel)-1 {
		// We hit a NaN or Inf in calculation
		x = (2 << zoomLevel) - 1
	}
	if y < 0 || y > (2<<zoomLevel)-1 {
		// We hit a NaN or Inf in calculation
		y = (2 << zoomLevel) - 1
	}
	//fmt.Printf("Calculated Tile(%v, %v, %v) for LatLon(%v)\n", fx, fy, zoomLevel, l)
	return Tile{x, y, zoomLevel}
}

// Calculates the lat/lon of the upper-left point of the given OSM Tile
func (o OSMTileMapper) LatLon(t Tile) LatLon {
	var lat, lon float64
	var n float64 = math.Pi - 2*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))

	lat = 180 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	lon = float64(t.X)/math.Exp2(float64(t.Z))*360 - 180
	//fmt.Printf("Calculated LatLon(%v, %v) for Tile(%v)\n", lat, lon, t)
	return LatLon{lat, lon}
}

func (o OSMTileMapper) TilesFromBounds(a LatLon, b LatLon, maxWidth uint, maxHeight uint) (Tile, Tile) {
	return GenericTilesFromBounds(o, a, b, maxWidth, maxHeight, o.TileSize, 18)
}

func (o OSMTileMapper) BoundsFromTiles(a Tile, b Tile) (LatLon, LatLon) {
	return GenericBoundsFromTiles(o, a, b, LatLon{-85.0511, 180})
}
