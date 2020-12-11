package unproject

import (
	"image"

	"github.com/uluyol/tracegeog/tracer"
)

type LatLon struct {
	Lat, Lon float64
}

type Link struct {
	Src, Dst int
}

type GeoGraph struct {
	Nodes       []LatLon
	TransitOnly []int // indices of nodes that are transit-only
	Links       []Link
}

type InversionFunc = func(image.Point) LatLon

func ToGeoGraph(g *tracer.XYGraph, invertFn InversionFunc) *GeoGraph {
	geo := new(GeoGraph)
	geo.Nodes = make([]LatLon, len(g.Nodes))
	for i, n := range g.Nodes {
		geo.Nodes[i] = invertFn(n)
	}
	geo.TransitOnly = append([]int(nil), g.TransitOnly...)
	geo.Links = make([]Link, len(g.Links))
	for i, l := range g.Links {
		geo.Links[i] = Link{Src: l.Src, Dst: l.Dst}
	}
	return geo
}
