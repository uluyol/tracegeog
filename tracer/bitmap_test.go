package tracer

import "testing"

func TestBitmapSimple(t *testing.T) {
	type idx struct {
		x, y int
	}

	toset := map[idx]bool{
		{0, 2}:  true,
		{50, 2}: true,
		{19, 2}: true,
		{1, 1}:  true,
	}

	b := newBitmap2(50, 3)
	for idx := range toset {
		b.Set(idx.x, idx.y)
	}

	t.Log(b)

	for idx := range toset {
		if !b.Get(idx.x, idx.y) {
			t.Errorf("(%d,%d) is not set", idx.x, idx.y)
		}
	}

	for x := 0; x < 10; x++ {
		for y := 0; y < 3; y++ {
			if b.Get(x, y) != toset[idx{x, y}] {
				t.Errorf("(%d,%d) != %t", x, y, toset[idx{x, y}])
			}
		}
	}
}
