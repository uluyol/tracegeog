/*

Package repetita converts graphs to the Repetita [1] format.

Taken from (https://github.com/ngvozdiev/ncode-net/blob/master/src/net_gen.h#L57):

The first line of the file is "NODES XX" followed by a comment line and
XX lines for each of the nodes in the graph.
Each node line is of the format <node_name> <x> <y> where x and y are the x,y
coordinates of the nodes.
The nodes section is followed by and empty line and "EDGES XX" on a new line.
A comment line is next, followed by for each edge
"<label> <src> <dst> <weight> <bw> <delay>".
The bandwidth is in kbps and the delay in microseconds.

[1] https://github.com/svissicchio/Repetita

*/
package repetita

import (
	"fmt"
	"io"
	"sort"

	"github.com/uluyol/tracegeog/unproject"
)

type Exporter struct {
	// Speed of Light / Speed of Light in Fiber
	RefractiveIndex float64

	MakeSymmetric bool
}

func (e *Exporter) WriteGeo(g *unproject.GeoGraph, w io.Writer) error {
	var err error
	writef := func(format string, args ...interface{}) {
		if err != nil {
			return
		}
		_, e := fmt.Fprintf(w, format, args...)
		if e != nil {
			err = e
		}
	}

	isTransit := make(map[int]bool)
	for _, n := range g.TransitOnly {
		isTransit[n] = true
	}

	writef("NODES %d\n", len(g.Nodes))
	writef("label x y\n")
	for i := range g.Nodes {
		name := "node"
		if isTransit[i] {
			name = "transit"
		}
		writef("%s_%d 0 0\n", name, i)
	}

	links := g.Links
	if e.MakeSymmetric {
		links = makeSym(links)
	}
	links = append([]unproject.Link(nil), links...)

	sort.Slice(links, func(i, j int) bool {
		if links[i].Src == links[j].Src {
			return links[i].Dst < links[j].Dst
		}
		return links[i].Src < links[j].Src
	})

	writef("\nEDGES %d\n", len(links))
	writef("label src dest weight bw delay\n")
	for i, l := range links {
		writef("edge_%d %d %d 0 1000000 %d\n", i, l.Src, l.Dst, e.delayMicros(g, l))
	}

	return err
}

func (e *Exporter) delayMicros(g *unproject.GeoGraph, l unproject.Link) int64 {
	const SpeedOfLight = 299_792_458 // meters / sec
	metersPerSec := SpeedOfLight / e.RefractiveIndex

	n1 := g.Nodes[l.Src]
	n2 := g.Nodes[l.Dst]
	distKM := greatCircleDistance(n1.Lat, n1.Lon, n2.Lat, n2.Lon)

	delaySec := distKM * 1e3 / metersPerSec
	return int64(delaySec * 1e6)
}

func makeSym(in []unproject.Link) []unproject.Link {
	out := make([]unproject.Link, len(in), 2*len(in))
	has := make(map[[2]int]bool)

	for i, l := range in {
		out[i] = l
		has[[2]int{l.Src, l.Dst}] = true
	}

	for _, l := range in {
		if !has[[2]int{l.Dst, l.Src}] {
			has[[2]int{l.Dst, l.Src}] = true
			out = append(out, unproject.Link{
				Src: l.Dst,
				Dst: l.Src,
			})
		}
	}
	return out
}

var DefaultExporter = Exporter{
	RefractiveIndex: 1.467,
	MakeSymmetric:   true,
}
