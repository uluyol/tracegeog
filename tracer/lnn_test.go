package tracer

import (
	"image"
	"math"
	"testing"
)

func TestLNNTracker(t *testing.T) {
	tracker := lnnTracker{
		AllowedGapPx:          10,
		AllowedAngleOffsetRad: 50 * math.Pi / 180,
		EWMAPointThresh:       2,
		DistPx: func(p1, p2 image.Point) float64 {
			return math.Hypot(float64(p1.X-p2.X), float64(p1.Y-p2.Y))
		},
	}

	p := func(x, y int) image.Point { return image.Pt(x, y) }

	// Near nodes
	tracker.AddPoint(p(0, 0), 0)
	tracker.AddPoint(p(10, 2), 0)
	tracker.AddPoint(p(100, 90), 0)

	tracker.AddPoint(p(2, 3), 1)
	tracker.AddPoint(p(8, 3), 1)
	tracker.AddPoint(p(95, 90), 1)

	tracker.AddPoint(p(4, 5), 2)
	tracker.AddPoint(p(6, 5), 2)
	tracker.AddPoint(p(90, 90), 2)

	tracker.AddPoint(p(6, 7), 3)
	tracker.AddPoint(p(4, 7), 3)
	tracker.AddPoint(p(85, 90), 3)

	tracker.AddPoint(p(8, 9), 4)
	tracker.AddPoint(p(2, 9), 4)
	tracker.AddPoint(p(80, 90), 4)

	tracker.AddPoint(p(10, 11), 5)
	tracker.AddPoint(p(0, 11), 5)
	tracker.AddPoint(p(75, 90), 5) // done

	tracker.AddPoint(p(12, 13), 6)
	tracker.AddPoint(p(0, 13), 6)

	tracker.AddPoint(p(14, 15), 7) // done
	tracker.AddPoint(p(0, 15), 7)  // done

	runs := tracker.Runs()

	if len(runs) != 3 {
		t.Fatalf("want 3 runs, have %d", len(runs))
	}

	mustEq(t, runs[0], p(0, 0), p(2, 3), p(4, 5), p(6, 7), p(8, 9), p(10, 11), p(12, 13), p(14, 15))
	mustEq(t, runs[1], p(10, 2), p(8, 3), p(6, 5), p(4, 7), p(2, 9), p(0, 11), p(0, 13), p(0, 15))
	mustEq(t, runs[2], p(100, 90), p(95, 90), p(90, 90), p(85, 90), p(80, 90), p(75, 90))
}

func mustEq(t *testing.T, run lineRun, seen ...image.Point) {
	t.Helper()

	match := true
	if len(run.SeenPoints) == len(seen) {
		for i, p := range run.SeenPoints {
			if p != seen[i] {
				match = false
				break
			}
		}
	} else {
		match = false
	}
	if !match {
		t.Errorf("do not match:\n\twant: %v\n\thave: %v", seen, run.SeenPoints)
	}
}
