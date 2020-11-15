// +build ignore

package tracer

import (
	"fmt"
	"image"
	"math"
)

const debugMHT = true

type pointHyp struct {
	p     image.Point
	time  int
	lines []*lineHyp

	minScore     float64
	minScoreLine *lineHyp
}

type status int32

const (
	stWIP status = iota
	stDead
	stDone
)

type lineHyp struct {
	p       *pointHyp
	ewmaDir vec2
	status  status
	doneIdx int32 // index in multipleHypTracker.done
	len     int32
	head    *lineHyp
	prev    *lineHyp
}

func (h *lineHyp) Src() image.Point { return h.head.p.p }
func (h *lineHyp) Dst() image.Point { return h.p.p }
func (h *lineHyp) Len() int         { return int(h.len) }

func (h *lineHyp) IsDead() bool {
	for h != nil {
		if h.status == stDead {
			return true
		} else if h.status == stDone {
			return false
		}
		h = h.prev
	}
	return false
}

type multipleHypTracker struct {
	AllowedGapPx float64
	DistPx       func(p1, p2 image.Point) float64

	done []*lineHyp

	wipLines  []*lineHyp
	wipPoints []*pointHyp
}

const ewmaPointThresh = 8

func (t *multipleHypTracker) AddPoint(p image.Point, time int) {

	ph := &pointHyp{p: p, time: time, minScore: math.Inf(1)}
	t.wipPoints = append(t.wipPoints, ph)

	// Initially, just record points as new lines
	if time == 0 {
		lh := new(lineHyp)
		lh.p = ph
		ph.lines = append(ph.lines, lh)
		lh.status = stWIP
		lh.doneIdx = -1
		lh.len = 1
		lh.head = lh
		lh.prev = nil
		t.wipLines = append(t.wipLines, lh)
		return
	}

	// Afterwards, points need to be connected to existing lines
	for _, lh := range t.wipLines {
		if t.DistPx(lh.Dst(), p) <= t.AllowedGapPx {
			lh2 := new(lineHyp)
			lh2.p = ph
			lh2.len = lh.len + 1
			lh2.head = lh.head
			lh2.prev = lh

			if lh2.len == ewmaPointThresh {
				lh2.ewmaDir.X = float64(p.X - lh.Src().X)
				lh2.ewmaDir.Y = float64(p.Y - lh.Src().Y)
			} else if lh2.len > ewmaPointThresh {
				lh2.ewmaDir.X = alpha*float64(p.X-lh.Dst().X) + (1-alpha)*lh.ewmaDir.X
				lh2.ewmaDir.Y = alpha*float64(p.Y-lh.Dst().Y) + (1-alpha)*lh.ewmaDir.Y
			}

			t.wipLines = append(t.wipLines, lh2)
		}
	}
}

const mhtLookahead = 3

func (t *multipleHypTracker) PruneAt(time int) {
	if time < mhtLookahead {
		return
	}

	// Score different hypotheses
	for _, lh := range t.wipLines {
		end := lh.p.p
		for lh != nil && lh.p.time > time-mhtLookahead {
			lh = lh.prev
		}
		if lh == nil {
			continue
		}
		//println(lh.len)
		score := math.Pi
		if lh.len > ewmaPointThresh {
			avgDir := vec2{
				X: float64(end.X - lh.p.p.X),
				Y: float64(end.Y - lh.p.p.Y),
			}
			score = offAngle(avgDir, lh.prev.ewmaDir)
		}
		ph := lh.p
		if score < ph.minScore {
			ph.minScore = score
			ph.minScoreLine = lh
		}
	}

	stopIndex := len(t.wipPoints)
	// Pick best line for each point
	for i, ph := range t.wipPoints {
		if ph.time <= time-mhtLookahead {
			for _, lh := range ph.lines {
				if lh != ph.minScoreLine {
					lh.status = stDead
				}
			}
			ph.lines = nil
			if ph.minScoreLine != nil {
				if debugMHT && ph.minScoreLine.prev != nil {
					fmt.Printf("attach %v to %v\n", ph.p, ph.minScoreLine.prev.p.p)
				}
				ph.minScoreLine.status = stDone

				// Add to done slice, taking over a prefix's index
				idx := int32(-1)
				if ph.minScoreLine.prev != nil {
					idx = ph.minScoreLine.prev.doneIdx
				}
				if idx == -1 {
					idx = int32(len(t.done))
					t.done = append(t.done, nil)
				}
				t.done[idx] = ph.minScoreLine
				ph.minScoreLine.doneIdx = idx
			} else {
				if debugMHT {
					fmt.Printf("throw away %v\n", ph.p)
				}
			}
		}
		stopIndex = i
	}

	// Clear decided points
	t.wipPoints = t.wipPoints[stopIndex:]

	newWIP := make([]*lineHyp, 0, len(t.wipLines))
	// Remove dead and done lines
	for _, lh := range t.wipLines {
		if lh.status == stWIP && !lh.IsDead() {
			newWIP = append(newWIP, lh)
		}
	}
	t.wipLines = newWIP
}

func (t *multipleHypTracker) Finalize() {
	if len(t.wipPoints) == 0 {
		return
	}
	time := t.wipPoints[0].time
	for len(t.wipPoints) > 0 {
		t.PruneAt(time)
		time++
	}
}

func (t *multipleHypTracker) Runs() []lineRun {
	runs := make([]lineRun, len(t.done))
	for i, lh := range t.done {
		r := lineRun{SeenPoints: make([]image.Point, lh.Len())}
		for t := lh; t != nil; t = t.prev {
			r.SeenPoints[t.len-1] = t.p.p
		}
		if r.Src() != lh.Src() || r.Dst() != lh.Dst() {
			panic("unexpected error")
		}
		runs[i] = r
	}
	return runs
}
