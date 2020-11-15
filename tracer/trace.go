package tracer

import (
	"container/heap"
	"image"
	"image/color"
	"math"
	"sort"
	"sync/atomic"
)

type BlobMatcher interface {
	MatchStrength(x, y int, im *image.RGBA) float64
	EraseMatch(x, y int, im *image.RGBA)
}

type NodeConfig struct {
	Matcher           BlobMatcher
	StrengthThreshold float64
	MaxCount          int
}

type LinkConfig struct {
	Color            color.Color
	MinColorAccuracy float64 // RGB value match, 0-1, 1 is an exact match
	MinWidthPx       int     // in pixels
	AllowedGapPx     int     // max gap allowed, in pixels
	NodeProximityPx  int     // min proximity to a node, in pixels

	// How many deg the line can move away from its current trajectory
	ExpectedDirectionDeg float64
}

type NodeTracer struct {
	c   NodeConfig
	im  *image.RGBA
	g   XYGraph
	log func(string, ...interface{})
}

type LinkTracer struct {
	c   LinkConfig
	im  *image.RGBA
	g   XYGraph
	log func(string, ...interface{})
}

func NewNode(c NodeConfig, tim image.Image, logfunc func(string, ...interface{})) *NodeTracer {
	t := &NodeTracer{c: c, im: copyToRGBA(tim), log: logfunc}
	t.g.Bounds = tim.Bounds()
	return t
}

func NewLink(c LinkConfig, tim image.Image, g *XYGraph, logfunc func(string, ...interface{})) *LinkTracer {
	t := &LinkTracer{c: c, im: copyToRGBA(tim), g: *g, log: logfunc}
	t.g.Bounds = tim.Bounds()
	return t
}

type Link struct {
	Src, Dst int // index of Src and Dst Nodes

	Points []image.Point // for debugging
}

// An XYGraph is a graph with points in the original image coordinates.
//
// X values range from 0 to RectMax.X and Y values range from 0 to RectMax.Y.
type XYGraph struct {
	Nodes []image.Point
	Links []Link

	Bounds image.Rectangle
}

type nodeCand struct {
	x, y  int
	score float64
}

func lessPt(p, q image.Point) bool {
	if p.X == q.X {
		return p.Y < q.Y
	}
	return p.X < q.X
}

// nodeCandHeap is max heap of nodeCand.
type nodeCandHeap []nodeCand

func (h nodeCandHeap) Len() int      { return len(h) }
func (h nodeCandHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h nodeCandHeap) Less(i, j int) bool {
	if h[i].score == h[j].score {
		return lessPt(
			image.Pt(h[i].x, h[i].y),
			image.Pt(h[j].x, h[j].y))
	}
	return h[i].score > h[j].score
}

func (h *nodeCandHeap) Push(x interface{}) { *h = append(*h, x.(nodeCand)) }

func (h *nodeCandHeap) Pop() interface{} {
	t := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return t
}

func (t *NodeTracer) Find() {
	b := t.g.Bounds
	nc := &t.c

	t.log("scoring candidate nodes")
	cc := make(chan nodeCand)
	numLeft := int32(b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		go func(y int) {
			for x := b.Min.X; x < b.Max.X; x++ {
				if score := nc.Matcher.MatchStrength(x, y, t.im); score > nc.StrengthThreshold {
					cc <- nodeCand{x, y, score}
				}
			}
			if atomic.AddInt32(&numLeft, -1) == 0 {
				close(cc)
			}
		}(y)
	}
	var cands []nodeCand
	for nc := range cc {
		cands = append(cands, nc)
	}
	h := nodeCandHeap(cands)
	heap.Init(&h)

	t.log("%d candidate nodes; selecting best", h.Len())
	for len(t.g.Nodes) <= nc.MaxCount && h.Len() > 0 && h[0].score > nc.StrengthThreshold {
		top := &h[0]
		score := nc.Matcher.MatchStrength(top.x, top.y, t.im)
		if top.score != score {
			// Score has changed (some pixels belonged to another node), record and fix heap.
			top.score = score
			heap.Fix(&h, 0)
			continue
		}

		// top is the best candidate
		t.g.Nodes = append(t.g.Nodes, image.Pt(top.x, top.y))
		nc.Matcher.EraseMatch(top.x, top.y, t.im)

		// Remove top
		heap.Remove(&h, 0)
	}

	sort.Slice(t.g.Nodes, func(i, j int) bool {
		return lessPt(t.g.Nodes[i], t.g.Nodes[j])
	})

	t.log("found %d nodes", len(t.g.Nodes))
}

func (t *LinkTracer) Find() {
	// Bad, should do a search from src to dst nodes
	lineRuns := t.findLineRuns(t.im)

	t.log("filtering %d candidate links", len(lineRuns))
	for i := 0; i < len(lineRuns); i++ {
		r := &lineRuns[i]
		src := closestNode(r.Src(), t.g.Nodes, float64(t.c.NodeProximityPx))
		dst := closestNode(r.Dst(), t.g.Nodes, float64(t.c.NodeProximityPx))
		if src < 0 || dst < 0 || src == dst {
			continue
		}
		t.g.Links = append(t.g.Links, Link{src, dst, r.SeenPoints})
	}
	t.log("found %d links", len(t.g.Links))
}

func (t *NodeTracer) Graph() *XYGraph {
	g := t.g
	return &g
}

func (t *LinkTracer) Graph() *XYGraph {
	g := t.g
	return &g
}

func closestNode(pt image.Point, n []image.Point, maxDistPx float64) int {
	minDistIdx := -1
	minDist := maxDistPx
	for i := range n {
		d := distPx(pt, n[i])
		if d < minDist {
			minDist = d
			minDistIdx = i
		}
	}
	return minDistIdx
}

type vec2 struct {
	X, Y float64
}

type lineRun struct {
	SeenPoints []image.Point

	time    int
	ewmaDir vec2
}

func (r *lineRun) Src() image.Point  { return r.SeenPoints[0] }
func (r *lineRun) Dst() image.Point  { return r.SeenPoints[len(r.SeenPoints)-1] }
func (r *lineRun) Add(p image.Point) { r.SeenPoints = append(r.SeenPoints, p) }

func newLineRun(x, y int) lineRun {
	p := image.Pt(x, y)
	return lineRun{
		SeenPoints: []image.Point{p},
	}
}

func (t *LinkTracer) findLineRuns(im *image.RGBA) []lineRun {
	b := im.Bounds()
	lc := &t.c

	wR, wG, wB, wA := lc.Color.RGBA()
	lineColor := color.RGBA{
		R: uint8(wR), G: uint8(wG), B: uint8(wB), A: uint8(wA),
	}

	matchesLine := func(x, y int) bool {
		sum := 0.0
		num := 0.0
		for j := y - lc.MinWidthPx; j < y+lc.MinWidthPx; j++ {
			if j < b.Min.Y || j > b.Max.Y {
				return false
			}
			for i := x - lc.MinWidthPx; i < x+lc.MinWidthPx; i++ {
				if i < b.Min.X || i > b.Max.X {
					return false
				}

				// Check if color is close enough
				imColor := im.RGBAAt(i, j)
				if imColor.A != 0 {
					sum += 1 - colorDist(lineColor, imColor)
				}
				num++
			}
		}
		if num == 0 {
			return false
		}
		// Mean pixel color in a min-max width range are close enough
		// (in both x and y directions).
		return sum/num >= lc.MinColorAccuracy
	}

	distPxWrapX := func(a, b image.Point) float64 {
		t1 := float64(a.X - b.X)
		t2 := float64(a.Y - b.Y)
		noWrap := math.Hypot(t1, t2)

		if b.X < a.X {
			a, b = b, a
		}

		t1 = float64(im.Bounds().Dx() + a.X - b.X)
		t2 = float64(a.Y - b.Y)
		wrapped := math.Hypot(t1, t2)

		return math.Min(noWrap, wrapped)
	}

	possibleLineLocs := make([]image.Point, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if matchesLine(x, y) {
				possibleLineLocs = append(possibleLineLocs, image.Pt(x, y))
			}
		}
	}
	t.log("%d points that possibly belong to lines", len(possibleLineLocs))

	type nodeRuns struct {
		nodeIdx int
		runs    []lineRun
	}
	cc := make(chan nodeRuns)
	numLeft := int32(len(t.g.Nodes))

	for i, n := range t.g.Nodes {
		go func(nodeIdx int, n image.Point) {
			t.log("searching for lines which begin at node (%d, %d)", n.X, n.Y)
			tracker := lnnTracker{
				AllowedGapPx:    float64(lc.AllowedGapPx),
				DistPx:          distPxWrapX,
				EWMAPointThresh: 8,
			}

			type pointWithTime struct {
				p    image.Point
				dist float64
				t    int
			}

			possibleLocs := make([]pointWithTime, len(possibleLineLocs))
			for i, pt := range possibleLineLocs {
				possibleLocs[i].p = pt
				possibleLocs[i].dist = distPxWrapX(n, pt)
				if possibleLocs[i].dist <= float64(lc.NodeProximityPx) {
					possibleLocs[i].t = 0
				} else {
					possibleLocs[i].t = int(possibleLocs[i].dist /
						float64(lc.AllowedGapPx))
				}
			}

			sort.Slice(possibleLocs, func(i, j int) bool {
				pi := &possibleLocs[i]
				pj := &possibleLocs[j]
				if pi.dist == pj.dist {
					if pi.p.Y == pj.p.Y {
						return pi.p.X < pj.p.X
					}
					return pi.p.Y < pj.p.Y
				}
				return pi.dist < pj.dist
			})
			for _, pt := range possibleLocs {
				tracker.AddPoint(pt.p, pt.t)
			}
			cc <- nodeRuns{nodeIdx, tracker.Runs()}
			if atomic.AddInt32(&numLeft, -1) == 0 {
				close(cc)
			}
		}(i, n)
	}

	runs := make([][]lineRun, len(t.g.Nodes))
	for nr := range cc {
		runs[nr.nodeIdx] = nr.runs
	}
	for i := 1; i < len(runs); i++ {
		runs[0] = append(runs[0], runs[i]...)
	}

	return runs[0]
}

func removeMarkedUnordered(inLine *bitmap2, points *[]image.Point) {
	for i := 0; i < len(*points); {
		pt := (*points)[i]
		if !inLine.Get(pt.X, pt.Y) {
			i++
		} else {
			(*points)[i] = (*points)[len(*points)-1]
			*points = (*points)[:len(*points)-1]
		}
	}
}

const alpha = 0.4

/*
func (c *Config) findLineRuns(im *image.RGBA) []lineRun {
	b := im.Bounds()

	wR, wG, wB, wA := c.Line.Color.RGBA()
	lineColor := color.RGBA{
		R: uint8(wR), G: uint8(wG), B: uint8(wB), A: uint8(wA),
	}

	matchesLine := func(x, y int) bool {
		for j := y - c.Line.MinWidthPx; j < y+c.Line.MinWidthPx; j++ {
			if j < b.Min.Y || j > b.Max.Y {
				return false
			}
			for i := x - c.Line.MinWidthPx; i < x+c.Line.MinWidthPx; i++ {
				if i < b.Min.X || i > b.Max.X {
					return false
				}

				// Check if color is close enough
				if colorDist(im.RGBAAt(i, j), lineColor) > c.Line.MinColorAccuracy {
					return false
				}
			}
		}
		// all pixels in a min-max width range are close enough (in both x and y directions)
		return true
	}

	expectedDirRad := c.Line.ExpectedDirectionDeg * math.Pi / 180
	ewmaPointThresh := 4 * c.Line.MinWidthPx

	var allRuns []lineRun
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if !matchesLine(x, y) {
				continue
			}
			matchedAny := false
			for i := range allRuns {
				r := &allRuns[i]
				if distPx(r.Dst(), image.Pt(x, y)) <= float64(c.Line.AllowedGapPx) {
					if len(r.SeenPoints) > ewmaPointThresh {
						dirAngle := offAngle(r.ewmaDir, vec2{float64(x - r.Dst().X), float64(y - r.Dst().Y)})
						if dirAngle > expectedDirRad {
							continue // skip this run, direction is off
						}
					}
					if len(r.SeenPoints) == ewmaPointThresh-1 {
						r.ewmaDir.X = float64(x - r.Src().X)
						r.ewmaDir.Y = float64(y - r.Src().Y)
					} else if len(r.SeenPoints) >= ewmaPointThresh {
						r.ewmaDir.X = alpha*float64(x-r.Dst().X) + (1-alpha)*r.ewmaDir.X
						r.ewmaDir.Y = alpha*float64(y-r.Dst().Y) + (1-alpha)*r.ewmaDir.Y
					}
					r.Add(image.Pt(x, y))

					matchedAny = true
				}
			}

			if !matchedAny {
				allRuns = append(allRuns, newLineRun(x, y))
			}
		}
	}
	return allRuns
}
*/

func offAngle(v1, v2 vec2) float64 {
	norm1 := math.Hypot(v1.X, v2.Y)
	norm2 := math.Hypot(v2.X, v2.Y)
	dot := v1.X*v2.X + v1.Y*v2.Y
	return math.Acos(dot / (norm1 * norm2))
}

func distPx(a, b image.Point) float64 {
	t1 := float64(a.X - b.X)
	t2 := float64(a.Y - b.Y)
	return math.Hypot(t1, t2)
}

func sqOfNormalDiff(a, b uint8) float64 {
	t := float64(a) - float64(b)
	t /= 255
	return t * t
}

func colorDist(a, b color.RGBA) float64 {
	sum := sqOfNormalDiff(a.R, b.R)
	sum += sqOfNormalDiff(a.G, b.G)
	sum += sqOfNormalDiff(a.B, b.B)
	sum += sqOfNormalDiff(a.A, a.A)
	return math.Sqrt(sum) / 2 // normalize range to [0, 1]
}

func copyToRGBA(im image.Image) *image.RGBA {
	b := im.Bounds()
	res := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			res.Set(x, y, im.At(x, y))
		}
	}
	return res
}
