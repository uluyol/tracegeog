package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"github.com/google/subcommands"
	"github.com/uluyol/tracegeog/tracer"
	"github.com/uluyol/tracegeog/unproject"
	"github.com/uluyol/tracegeog/visualize"
)

type ImageReadingCmd struct {
	InputPath string
	im        image.Image
}

type GraphReadingCmd struct {
	InputGraph string
	graph      tracer.XYGraph
}

type GraphWritingCmd struct{ OutputPath string }

func (c *ImageReadingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.InputPath, "i", "", "path to input image (png or jpeg)")
}

func (c *GraphReadingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.InputGraph, "g", "", "path to input graph")
}

func (c *GraphWritingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.OutputPath, "o", "", "path to output json")
}

func (c *ImageReadingCmd) Prepare() {
	im, err := readImage(c.InputPath)
	if err != nil {
		log.Fatalf("unable to read input: %v", err)
	}
	c.im = im
}

func (c *GraphReadingCmd) Prepare() {
	f, err := os.Open(c.InputGraph)
	if err != nil {
		log.Fatalf("unable to open input graph %s: %v",
			c.InputGraph, err)
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&c.graph); err != nil {
		log.Fatalf("failed to read input graph: %v", err)
	}
	f.Close() // non-fatal if errors
}

type TraceNodes struct {
	ImageReadingCmd
	GraphWritingCmd

	NodeIconPath      string
	NodeColorAccuracy float64
	MaxNodeCount      int
}

func (c *TraceNodes) Name() string     { return "trace-nodes" }
func (c *TraceNodes) Synopsis() string { return "trace nodes from an image" }
func (c *TraceNodes) Usage() string    { return c.Synopsis() + "\n" }

func (c *TraceNodes) SetFlags(fs *flag.FlagSet) {
	c.ImageReadingCmd.SetFlags(fs)
	c.GraphWritingCmd.SetFlags(fs)

	fs.StringVar(&c.NodeIconPath, "icon", "", "path to not icon image (png or jpeg)")
	fs.Float64Var(&c.NodeColorAccuracy, "node-color-accuracy", 0.8, "minimum node color accuracy")
	fs.IntVar(&c.MaxNodeCount, "max-node-count", 0, "max node count (prunes if more than this are availabe)")

}

type TraceLinks struct {
	ImageReadingCmd
	GraphReadingCmd
	GraphWritingCmd

	LineColorString      string
	LineColorAccuracy    float64
	LineWidthPx          int
	LineAllowedGapPx     int
	NodeProximityPx      int
	ExpectedDirectionDeg float64
}

func (c *TraceLinks) Name() string     { return "trace-links" }
func (c *TraceLinks) Synopsis() string { return "trace links from an image" }
func (c *TraceLinks) Usage() string    { return c.Synopsis() + "\n" }

func (c *TraceLinks) SetFlags(fs *flag.FlagSet) {
	c.ImageReadingCmd.SetFlags(fs)
	c.GraphReadingCmd.SetFlags(fs)
	c.GraphWritingCmd.SetFlags(fs)

	fs.StringVar(&c.LineColorString, "line-color", "#000000", "line color")
	fs.Float64Var(&c.LineColorAccuracy, "line-color-accuracy", 0.85, "minimum color accuracy to match line")
	fs.IntVar(&c.LineWidthPx, "line-width", 3, "minimum line width (pixels)")
	fs.IntVar(&c.LineAllowedGapPx, "line-gap", 1, "maximum line gap (pixels)")
	fs.IntVar(&c.NodeProximityPx, "line-node-dist", 1, "maximum distance between line and node (pixels)")
	fs.Float64Var(&c.ExpectedDirectionDeg, "line-dir-deg", 10, "maximum permitted change in line direction")
}

type Vis struct {
	ImageReadingCmd
	GraphReadingCmd

	OutputImagePath        string
	OutputOverlayImagePath string
}

func (c *Vis) Name() string     { return "vis" }
func (c *Vis) Synopsis() string { return "visualize a graph" }
func (c *Vis) Usage() string    { return c.Synopsis() + "\n" }

func (c *Vis) SetFlags(fs *flag.FlagSet) {
	c.ImageReadingCmd.SetFlags(fs)
	c.GraphReadingCmd.SetFlags(fs)

	fs.StringVar(&c.OutputImagePath, "png", "", "path to output png")
	fs.StringVar(&c.OutputOverlayImagePath, "overlaypng", "", "path to output overlay png")
}

type Unproj struct {
	GraphReadingCmd
	GraphWritingCmd

	Projection  string
	ExtraMargin struct {
		Left, Right int
	}
	ScaleY         float64
	PrimeMeridianX int // after adding extra margin
	EquatorY       int
}

func (c *Unproj) Name() string     { return "unproj" }
func (c *Unproj) Synopsis() string { return "unproject a graph to lat, lon pairs" }
func (c *Unproj) Usage() string    { return c.Synopsis() + "\n" }

func (c *Unproj) SetFlags(fs *flag.FlagSet) {
	c.GraphReadingCmd.SetFlags(fs)
	c.GraphWritingCmd.SetFlags(fs)

	fs.StringVar(&c.Projection, "proj", "web-mercator",
		"map projection to invert (web-mercator)")
	fs.IntVar(&c.ExtraMargin.Left, "extra-margin-left", 0, "margin to add to the left")
	fs.IntVar(&c.ExtraMargin.Right, "extra-margin-right", 0, "margin to add to the left")
	fs.Float64Var(&c.ScaleY, "scale-y", 1, "multiply y-values by this before converting")
	fs.IntVar(&c.PrimeMeridianX, "prime-meridian-x", -1,
		"prime meridian x value after adding margins (leave -1 to use image center)")
	fs.IntVar(&c.EquatorY, "equator-y", -1,
		"equator y value (leave -1 to use image center)")
}

func readImage(p string) (image.Image, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return im, f.Close()
}

func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	return
}

func (c *TraceNodes) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c.ImageReadingCmd.Prepare()

	icon, err := readImage(c.NodeIconPath)
	if err != nil {
		log.Fatalf("unable to read icon: %v", err)
	}

	tracer := tracer.NewNode(tracer.NodeConfig{
		Matcher:           tracer.NewIconMatcher(icon),
		StrengthThreshold: c.NodeColorAccuracy,
		MaxCount:          c.MaxNodeCount,
	}, c.im, log.Printf)

	tracer.Find()
	graph := tracer.Graph()

	if err := writeGraphTo(graph, c.OutputPath); err != nil {
		log.Fatalf("unable to write output json file %s: %v",
			c.OutputPath, err)
	}
	return subcommands.ExitSuccess
}

func (c *TraceLinks) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c.ImageReadingCmd.Prepare()
	c.GraphReadingCmd.Prepare()

	lineColor, err := parseHexColor(c.LineColorString)
	if err != nil {
		log.Fatalf("bad line color: %v", err)
	}

	tracer := tracer.NewLink(tracer.LinkConfig{
		Color:                lineColor,
		MinColorAccuracy:     c.LineColorAccuracy,
		MinWidthPx:           c.LineWidthPx,
		AllowedGapPx:         c.LineAllowedGapPx,
		NodeProximityPx:      c.NodeProximityPx,
		ExpectedDirectionDeg: c.ExpectedDirectionDeg,
	}, c.im, &c.graph, log.Printf)

	tracer.Find()
	graph := tracer.Graph()

	if err := writeGraphTo(graph, c.OutputPath); err != nil {
		log.Fatalf("unable to write output json file %s: %v",
			c.OutputPath, err)
	}

	return subcommands.ExitSuccess
}

func (c *Vis) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c.ImageReadingCmd.Prepare()
	c.GraphReadingCmd.Prepare()

	outIm := visualize.DrawGraph(&c.graph)
	if err := writePngTo(outIm, c.OutputImagePath); err != nil {
		log.Fatalf("unable to write png to %s: %v",
			c.OutputImagePath, err)
	}
	outIm = visualize.OverlayOn(outIm, c.im)
	if err := writePngTo(outIm, c.OutputOverlayImagePath); err != nil {
		log.Fatalf("unable to write overlayed png to %s: %v",
			c.OutputOverlayImagePath, err)
	}
	return subcommands.ExitSuccess
}

func (c *Unproj) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	c.GraphReadingCmd.Prepare()

	g := c.GraphReadingCmd.graph

	if c.Projection != "web-mercator" {
		log.Fatalf("only web-mercator projection is supported")
	}

	wm := unproject.WebMercator{
		Bounds:         g.Bounds,
		ScaleY:         c.ScaleY,
		PrimeMeridianX: c.PrimeMeridianX,
		EquatorY:       c.EquatorY,
	}
	wm.ExtraMargin.Left = c.ExtraMargin.Left
	wm.ExtraMargin.Right = c.ExtraMargin.Right
	if c.PrimeMeridianX == -1 {
		wm.PrimeMeridianX =
			(g.Bounds.Dx() + c.ExtraMargin.Left + c.ExtraMargin.Right) / 2
	}
	if c.EquatorY == -1 {
		wm.EquatorY = g.Bounds.Dy() / 2
	}

	geog := unproject.ToGeoGraph(&g, wm.ToLatLon)
	if err := writeGraphTo(geog, c.OutputPath); err != nil {
		log.Fatalf("failed to write output to %s: %v", c.OutputPath, err)
	}

	return subcommands.ExitSuccess
}

func writeGraphTo(graph interface{}, p string) error {
	log.Printf("writing graph to %s", p)

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(graph); err != nil {
		return err
	}
	return f.Close()
}

func writePngTo(im image.Image, p string) error {
	log.Printf("writing image to %s", p)

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if err := png.Encode(f, im); err != nil {
		return err
	}
	return f.Close()
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&TraceNodes{}, "")
	subcommands.Register(&TraceLinks{}, "")
	subcommands.Register(&Vis{}, "")
	subcommands.Register(&Unproj{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("tracegeog: ")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
