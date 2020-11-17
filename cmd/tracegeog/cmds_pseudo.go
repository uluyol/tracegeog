package main

import (
	"encoding/json"
	"flag"
	"image"
	"log"
	"os"

	"github.com/uluyol/tracegeog/tracer"
	"github.com/uluyol/tracegeog/unproject"
)

type ImageReadingCmd struct {
	InputPath string
	im        image.Image
}

type GraphReadingCmd struct {
	InputGraph string
	graph      tracer.XYGraph
}

type GeoGraphReadingCmd struct {
	InputGraph string
	graph      unproject.GeoGraph
}

type GraphWritingCmd struct{ OutputPath string }

func (c *ImageReadingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.InputPath, "i", "", "path to input image (png or jpeg)")
}

func (c *GraphReadingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.InputGraph, "g", "", "path to input graph")
}

func (c *GeoGraphReadingCmd) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.InputGraph, "g", "", "path to input geo graph")
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

func (c *GeoGraphReadingCmd) Prepare() {
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
