package tracer

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

type bitmapImage [][]int

func (im bitmapImage) At(i, j int) color.Color {
	color := color.RGBA{255, 255, 255, 255}
	if im[j][i] > 0 {
		color.R = 0
		color.G = 0
		color.B = 0
	}
	return color
}

func (im bitmapImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, len(im[0]), len(im))
}

func (im bitmapImage) ColorModel() color.Model { return color.RGBAModel }

var _ image.Image = bitmapImage{}

func TestIconMatcherSimple(t *testing.T) {
	icon := bitmapImage{
		{0, 1, 0},
		{1, 1, 0},
		{0, 0, 1},
	}

	bim := bitmapImage{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
	}

	im := copyToRGBA(bim)

	m := NewIconMatcher(icon)

	locs := [][2]int{
		{2, 2},
		{9, 1},
		{20, 2},
	}

	for _, loc := range locs {
		if st := m.MatchStrength(loc[0], loc[1], im); st != 1 {
			t.Errorf("failed to find icon at (%d, %d): strength is %f", loc[0], loc[1], st)
		}
	}

	for y := bim.Bounds().Min.Y; y < bim.Bounds().Max.Y; y++ {
		for x := bim.Bounds().Min.X; x < bim.Bounds().Max.X; x++ {
			isLoc := false
			for _, loc := range locs {
				if loc[0] == x && loc[1] == y {
					isLoc = true
					break
				}
			}

			if isLoc {
				continue // skip points that are perfect matches
			}

			if st := m.MatchStrength(x, y, im); st >= 1 || st < 0 {
				t.Errorf("bad strength at (%d, %d): %f", x, y, st)
			}
		}
	}

	for _, loc := range locs {
		m.EraseMatch(loc[0], loc[1], im)
		if st := m.MatchStrength(loc[0], loc[1], im); st != 0 {
			t.Errorf("failed to erase icon at (%d, %d): strength is %f; image:\n%s", loc[0], loc[1], st, imToString(im))
		}
	}
}

func imToString(im *image.RGBA) string {
	buf := strings.Builder{}
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			c := im.RGBAAt(x, y)
			avg := 10 * (int(c.R) + int(c.G) + int(c.B) + int(c.A)) / 4
			avg /= 255
			switch avg {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9:
				buf.WriteByte('0' + byte(avg))
			case 10:
				buf.WriteByte('X')
			default:
				buf.WriteByte('U')
			}
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}
