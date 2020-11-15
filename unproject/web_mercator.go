package unproject

import (
	"image"
	"math"
	"sync"

	"github.com/wroge/wgs84"
)

var webMercatorConst struct {
	once  sync.Once
	width float64
}

func webMercatorWidth() float64 {
	webMercatorConst.once.Do(func() {
		f := wgs84.LonLat().To(wgs84.WebMercator())
		west, _, _ := f(-180, 0, 0)
		east, _, _ := f(180, 0, 0)
		webMercatorConst.width = math.Abs(west) + math.Abs(east)
	})
	return webMercatorConst.width
}

type WebMercator struct {
	Bounds      image.Rectangle
	ExtraMargin struct {
		Left, Right int
	}
	PrimeMeridianX int // after adding extra margin
	EquatorY       int
}

func (w *WebMercator) scalingFactor() float64 {
	width := w.Bounds.Dx() + w.ExtraMargin.Left + w.ExtraMargin.Right
	return webMercatorWidth() / float64(width)
}

func (w *WebMercator) ToLatLon(p image.Point) LatLon {
	wbToLL := wgs84.WebMercator().To(wgs84.LonLat())

	// x goes left to right, same as lat
	x := float64(p.X - w.Bounds.Min.X - w.PrimeMeridianX)

	// y goes down, so invert to get lon
	y := float64(w.EquatorY - (p.Y - w.Bounds.Min.Y))

	c := w.scalingFactor()

	lon, lat, _ := wbToLL(x*c, y*c, 0)
	return LatLon{
		Lat: lat,
		Lon: lon,
	}
}
