package visualize

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/uluyol/tracegeog/tracer"
	"golang.org/x/image/font/gofont/gobold"
)

func DrawGraph(g *tracer.XYGraph) image.Image {
	ctx := gg.NewContext(g.Bounds.Dx(), g.Bounds.Dy())

	font, err := truetype.Parse(gobold.TTF)
	if err != nil {
		panic("failed to parse gobold")
	}
	ctx.SetFontFace(truetype.NewFace(font, &truetype.Options{
		Size: 12,
	}))

	nodeColor := color.RGBA{255, 165, 0, 255}       // orange
	transitColor := color.RGBA{100, 80, 230, 255}   // blue
	lineColor := color.RGBA{50, 200, 10, 255}       // green
	nodeLabelColor := color.RGBA{255, 0, 0, 255}    // red
	transitLabelColor := color.RGBA{0, 0, 255, 255} // blue

	ctx.SetColor(lineColor)
	ctx.SetLineWidth(3)
	for _, l := range g.Links {
		if len(l.Points) == 0 {
			ctx.DrawLine(
				float64(g.Nodes[l.Src].X),
				float64(g.Nodes[l.Src].Y),
				float64(g.Nodes[l.Dst].X),
				float64(g.Nodes[l.Dst].Y),
			)
			ctx.Stroke()
			continue
		}
		for pi := 1; pi < len(l.Points); pi++ {
			p := l.Points[pi-1]
			q := l.Points[pi]
			ctx.DrawLine(
				float64(p.X), float64(p.Y),
				float64(q.X), float64(q.Y),
			)
			ctx.Stroke()
		}
	}

	isTransit := make(map[int]bool)

	for _, ni := range g.TransitOnly {
		isTransit[ni] = true
		x := float64(g.Nodes[ni].X)
		y := float64(g.Nodes[ni].Y)
		ctx.SetColor(transitColor)
		ctx.DrawCircle(x, y, 10)
		ctx.Fill()
		ctx.SetColor(transitLabelColor)
		ctx.DrawStringAnchored(strconv.Itoa(ni), x, y-1, 0.5, 0.5)
	}

	for i, n := range g.Nodes {
		if isTransit[i] {
			continue
		}
		x := float64(n.X)
		y := float64(n.Y)
		ctx.SetColor(nodeColor)
		ctx.DrawCircle(x, y, 10)
		ctx.Fill()
		ctx.SetColor(nodeLabelColor)
		ctx.DrawStringAnchored(strconv.Itoa(i), x, y-1, 0.5, 0.5)
	}

	return ctx.Image()
}

func OverlayOn(im image.Image, base image.Image) image.Image {
	out := image.NewRGBA(im.Bounds())
	draw.Draw(out, im.Bounds(), base, image.ZP, draw.Src)
	draw.Draw(out, im.Bounds(), withAlpha{im, 50e3}, image.ZP, draw.Over)
	return out
}

type withAlpha struct {
	im image.Image
	a  uint32
}

func (im withAlpha) Bounds() image.Rectangle { return im.im.Bounds() }
func (im withAlpha) ColorModel() color.Model { return color.RGBAModel }

func (im withAlpha) At(i, j int) color.Color {
	c := im.im.At(i, j)
	r, g, b, _ := c.RGBA()
	return color.NRGBA64{
		uint16(r), uint16(g), uint16(b), uint16(im.a)}
}
