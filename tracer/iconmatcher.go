package tracer

import (
	"image"
	"image/color"
)

type IconMatcher struct {
	i   *image.RGBA
	off image.Point
}

var _ BlobMatcher = &IconMatcher{}

func NewIconMatcher(im image.Image) *IconMatcher {
	return &IconMatcher{
		copyToRGBA(im),
		image.Pt(im.Bounds().Dx()/2, im.Bounds().Dy()/2),
	}
}

func (m *IconMatcher) MatchStrength(x, y int, im *image.RGBA) float64 {
	sum := 0.0
	num := 0.0
	for j := 0; j < m.i.Rect.Dy(); j++ {
		for i := 0; i < m.i.Rect.Dx(); i++ {
			iconColor := m.i.RGBAAt(i+m.i.Rect.Min.X, j+m.i.Rect.Min.Y)
			if iconColor.A == 0 {
				continue
			}
			imPt := image.Pt(x+i, y+j).Sub(m.off)
			if !imPt.In(im.Rect) {
				return 0
			}
			imColor := im.RGBAAt(imPt.X, imPt.Y)
			if imColor.A != 0 {
				sum += 1 - colorDist(iconColor, imColor)
			}
			num++
		}
	}
	if num == 0 {
		return 0
	}
	return sum / num
}

func (m *IconMatcher) EraseMatch(x, y int, im *image.RGBA) {
	empty := color.RGBA{}
	for j := 0; j < m.i.Rect.Dy(); j++ {
		for i := 0; i < m.i.Rect.Dx(); i++ {
			iconColor := m.i.RGBAAt(i+m.i.Rect.Min.X, j+m.i.Rect.Min.Y)
			if iconColor.A == 0 {
				continue
			}
			imPt := image.Pt(x+i, y+j).Sub(m.off)
			if !imPt.In(im.Rect) {
				panic("EraseMatch should only be called if MatchStrength > 0")
			}
			im.SetRGBA(imPt.X, imPt.Y, empty)
		}
	}
}
