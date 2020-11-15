package tracer

import (
	"image"
	"math"
)

type lnnTracker struct {
	AllowedGapPx          float64
	AllowedAngleOffsetRad float64
	EWMAPointThresh       int
	DistPx                func(p1, p2 image.Point) float64

	runs []lineRun
}

func (t *lnnTracker) AddPoint(pt image.Point, time int) bool {
	if time == 0 {
		t.runs = append(t.runs, lineRun{SeenPoints: []image.Point{pt}, time: time})
		return true
	}

	bestDist := math.Inf(1)
	bestIdx := -1
	for i := range t.runs {
		r := &t.runs[i]

		if r.time >= time {
			continue
		}

		if len(r.SeenPoints) > t.EWMAPointThresh {
			dirAngle := offAngle(
				r.ewmaDir,
				vec2{float64(pt.X - r.Dst().X), float64(pt.Y - r.Dst().Y)})
			if dirAngle > t.AllowedAngleOffsetRad {
				continue // skip this run, direction is off
			}
		}

		dist := t.DistPx(t.runs[i].Dst(), pt)
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}

	if bestIdx == -1 || bestDist > t.AllowedGapPx {
		return false // No good line found
	}

	// Otherwise, add to best candidate.
	r := &t.runs[bestIdx]
	if len(r.SeenPoints) == t.EWMAPointThresh-1 {
		r.ewmaDir.X = float64(pt.X - r.Src().X)
		r.ewmaDir.Y = float64(pt.Y - r.Src().Y)
	} else if len(r.SeenPoints) >= t.EWMAPointThresh {
		r.ewmaDir.X = alpha*float64(pt.X-r.Dst().X) + (1-alpha)*r.ewmaDir.X
		r.ewmaDir.Y = alpha*float64(pt.Y-r.Dst().Y) + (1-alpha)*r.ewmaDir.Y
	}

	r.SeenPoints = append(r.SeenPoints, pt)
	r.time = time

	return true
}

func (t *lnnTracker) Runs() []lineRun { return t.runs }
