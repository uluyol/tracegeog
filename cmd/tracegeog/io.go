package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

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
