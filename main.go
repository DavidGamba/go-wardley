// This file is part of go-wardley.
//
// Copyright (C) 2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-wardley/hcl"
	svg "github.com/ajstarks/svgo"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

// Show guides in drawing
var showGuides bool

// Serve the drawing at localhost:8080
var serve bool

// Keeps count of connect path IDs
var connectID int

var outputFile string

var canvas *svg.SVG

var inputData *hcl.Map

var inputFile string

var width, height, margin int

func main() {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	opt.BoolVar(&showGuides, "guides", false, opt.Description("Show margins, limits and other guides in drawing"))
	opt.BoolVar(&serve, "serve", false, opt.Description("Serve the drawing at localhost:8080"))
	opt.StringVar(&outputFile, "output", "map.svg", opt.Description("Map svg output file"))
	opt.StringVar(&inputFile, "file", "map.hcl", opt.Description("Map input file"))
	opt.IntVar(&width, "width", 1280, opt.Description("Map width"))
	opt.IntVar(&height, "height", 768, opt.Description("Map height"))
	opt.IntVar(&margin, "margin", 40, opt.Description("Map margin"))
	remaining, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)
	err = realMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

func realMain() error {
	m, err := hcl.ParseHCLFile(inputFile)
	if err != nil {
		return err
	}
	inputData = m

	ofh, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer ofh.Close()
	drawing(ofh)
	ofh.Close()
	if serve {
		http.Handle("/", http.HandlerFunc(drawHandler))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func drawHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	drawing(w)
}

func drawing(w io.Writer) {
	canvas = svg.New(w)
	canvas.Start(width, height)
	canvas.Gstyle("font-family:sans-serif")
	grid(canvas, margin, width, height)
	canvas.Translate(margin*2, height-margin*2)
	canvas.Marker("connector-arrow", 17, 3, 12, 10, `orient="auto"`)
	canvas.Path("M0,0 L0,6 L12,3 z")
	canvas.MarkerEnd()
	canvas.Marker("connector-inertia", 0, 10, 20, 40, `orient="auto"`)
	canvas.Path("M-5,20 L-5,-20 L5,-20 L5,20")
	canvas.MarkerEnd()

	nodes := inputData.Nodes
	connectors := inputData.Connectors

	maxGenesis, maxCustom, maxProduct, maxCommodity := 0, 0, 0, 0
	maxY := 0
	for _, n := range nodes {
		if n.Evolution == "genesis" && n.EvolutionX > maxGenesis {
			maxGenesis = n.EvolutionX
		}
		if n.Evolution == "custom" && n.EvolutionX > maxCustom {
			maxCustom = n.EvolutionX
		}
		if n.Evolution == "product" && n.EvolutionX > maxProduct {
			maxProduct = n.EvolutionX
		}
		if n.Evolution == "commodity" && n.EvolutionX > maxCommodity {
			maxCommodity = n.EvolutionX
		}
		if n.Visibility > maxY {
			maxY = n.Visibility
		}
	}
	for _, n := range nodes {
		NodeXY(n, maxGenesis, maxCustom, maxProduct, maxCommodity, maxY)
	}
	for _, c := range connectors {
		var a, b *hcl.Node
		for _, n := range nodes {
			logger.Printf("node id: %s\n", n.ID)
			if n.ID == c.From {
				a = n
			}
			if n.ID == c.To {
				b = n
			}
		}
		if a == nil || b == nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't find node '%s'\n", c.From)
			continue
		}
		if b == nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't find node '%s'\n", c.To)
			continue
		}
		connect(c, a, b)
	}
	for _, n := range nodes {
		DrawNode(n)
	}
	canvas.Gend()
	canvas.Gend()
	canvas.End()
}

var mapGrid Grid

// Grid -
type Grid struct {
	XQuarterLenght int
	YLenght        int
	Genesis        int
	Custom         int
	Product        int
	Commodity      int
	Visible        int
}

func NodeXY(n *hcl.Node, maxGenesis, maxCustom, maxProduct, maxCommodity, maxY int) {
	switch n.Evolution {
	case "genesis":
		n.X = mapGrid.Genesis + mapGrid.XQuarterLenght/(maxCustom+1)*n.EvolutionX
	case "custom":
		n.X = mapGrid.Custom + mapGrid.XQuarterLenght/(maxCustom+1)*n.EvolutionX
	case "product":
		n.X = mapGrid.Product + mapGrid.XQuarterLenght/(maxCustom+1)*n.EvolutionX
	case "commodity":
		n.X = mapGrid.Commodity + mapGrid.XQuarterLenght/(maxCustom+1)*n.EvolutionX
	}

	n.Y = -mapGrid.YLenght / (maxY + 1) * (maxY + 1 - n.Visibility)
}

// DrawNode -
func DrawNode(n *hcl.Node) {
	nodeFontSize := 9
	canvas.Gstyle("text-shadow: 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white")
	if n.Description != "" {
		canvas.Title(n.Description)
	} else {
		canvas.Title(n.Label)
	}
	canvas.Circle(n.X, n.Y, 5, fmt.Sprintf("fill:%s;stroke:%s", n.Fill, n.Color))
	// canvas.Text(n.X+10, n.Y+3, n.Label, fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;text-shadow: -1px 0 white, 0 1px white, 1px 0 white, 0 -1px white", nodeFontSize))
	// canvas.Gstyle("text-shadow: -1px 0 white, 0 1px white, 1px 0 white, 0 -1px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white")
	canvas.Textlines(n.X+8, n.Y+10, strings.Split(n.Label, "\n"), nodeFontSize, nodeFontSize+3, "black", "left")
	canvas.Gend()

}

func connect(c *hcl.Connector, a, b *hcl.Node) {
	nodeFontSize := 9
	connectID++

	// Calculate midpoints
	x := a.X + (b.X-a.X)/2
	if a.X > b.X {
		x = b.X + (a.X-b.X)/2
	}
	y := a.Y + (b.Y-a.Y)/2
	if a.Y > b.Y {
		y = b.Y + (a.Y-b.Y)/2
	}

	// canvas.Def()
	// canvas.Path(fmt.Sprintf("M %d,%d %d,%d %d,%d", a.X, a.Y, x, y, b.X, b.Y), fmt.Sprintf(`id="%d"`, connectID))
	// canvas.DefEnd()
	switch c.Type {
	case "normal":
		canvas.Path(fmt.Sprintf("M %d,%d %d,%d", a.X, a.Y, b.X, b.Y),
			fmt.Sprintf(`id="%s-%s"`, a.ID, b.ID),
			fmt.Sprintf(`fill:none;stroke:%s;opacity:0.2`, c.Color))
	case "bold":
		canvas.Path(fmt.Sprintf("M %d,%d %d,%d", a.X, a.Y, b.X, b.Y), fmt.Sprintf(`fill:none;stroke:%s;opacity:0.8`, c.Color))
	case "change":
		canvas.Path(fmt.Sprintf("M %d,%d %d,%d %d,%d", a.X, a.Y, x, y, b.X, b.Y),
			fmt.Sprintf(`fill:white;stroke:%[1]s;opacity:0.6;stroke-dasharray:6,6;marker-end:url(#connector-arrow)`, c.Color))
	case "change-inertia":
		canvas.Path(fmt.Sprintf("M %d,%d %d,%d %d,%d", a.X, a.Y, x, y, b.X, b.Y),
			fmt.Sprintf(`fill:white;stroke:%[1]s;opacity:0.6;stroke-dasharray:6,6;marker-mid:url(#connector-inertia);marker-end:url(#connector-arrow)`, c.Color))
	}
	x += 8
	y += 10

	// if strings.Contains(c.Label, "\n") {
	canvas.Textlines(x, y, strings.Split(c.Label, "\n"), nodeFontSize, nodeFontSize+3, "black", "left")
	// } else {
	// 	s.Textpath(c.Label, fmt.Sprintf("#%d", connectID), `x="10" y="-5"`, fmt.Sprintf("text-align:left;font-size:%dpx;fill:black", nodeFontSize))
	// }
}

func grid(s *svg.SVG, margin, width, height int) {
	// TODO: Variable font size based on width and height
	fontSize := 12

	// Grid
	//   X
	xLenght := width - margin*4
	xZero := margin * 2
	xEnd := width - margin*2
	//   Y
	yLenght := height - margin*4
	yZero := height - margin*2
	yEnd := margin * 2

	mapGrid = Grid{
		XQuarterLenght: (width - margin*4) / 4,
		Genesis:        0,
		Custom:         (width - margin*4) / 4,
		Product:        (width - margin*4) * 2 / 4,
		Commodity:      (width - margin*4) * 3 / 4,
		YLenght:        height - margin*4,
		Visible:        0,
	}

	s.Rect(0, 0, width, height, "fill:white")

	if showGuides {
		// Limits Guide
		s.Line(0, height, width, height, "fill:none;stroke:red")
		s.Line(width, 0, width, height, "fill:none;stroke:red")
		// Margin guide
		s.Line(margin, height-margin, width-margin, height-margin, "fill:none;stroke:green")
		s.Line(width-margin, margin, width-margin, height-margin, "fill:none;stroke:green")
		s.Line(margin, margin, width-margin, margin, "fill:none;stroke:green")
		s.Line(margin, margin, margin, height-margin, "fill:none;stroke:green")

		s.Translate(xZero, yZero)
		s.Text(xLenght-40, -yLenght, fmt.Sprintf("%d,%d", xLenght, yLenght), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(0, 0, fmt.Sprintf("%d,%d", xZero, yZero), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLenght/4, 0, fmt.Sprintf("%d,%d", xLenght/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLenght*2/4, 0, fmt.Sprintf("%d,%d", 2*margin+xLenght*2/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLenght*3/4, 0, fmt.Sprintf("%d,%d", 2*margin+xLenght*3/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Gend()
	}

	// Grid
	s.Marker("arrow", 0, 3, 12, 10, `orient="auto"`)
	s.Path("M0,0 L0,6 L12,3 z", "fill:black")
	s.MarkerEnd()
	s.Line(xZero, yZero, xEnd, yZero, "fill:none;stroke:black;marker-end:url(#arrow)")
	s.Line(xZero, yZero, xZero, yEnd, "fill:blue;stroke:black;marker-end:url(#arrow)")

	s.Line(2*margin+xLenght/4, yZero, 2*margin+xLenght/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)
	s.Line(2*margin+xLenght*2/4, yZero, 2*margin+xLenght*2/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)
	s.Line(2*margin+xLenght*3/4, yZero, 2*margin+xLenght*3/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)

	// Text
	s.Text(xZero, height-margin, "Genesis", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLenght/4, height-margin, "Custom", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLenght*2/4, height-margin, "Product (+rental)", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLenght*3/4, height-margin, "Commodity (+utility)", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(xEnd-100, height-2*margin-5, "Evolution", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;font-weight:bold;font-family:serif", fontSize+2))

	s.TranslateRotate(xZero, yZero, 270)
	s.Text(0, -5, "Invisible", fmt.Sprintf("text-anchor:top;font-size:%dpx;fill:black", fontSize))
	s.Text(yLenght-50, -5, "Visible", fmt.Sprintf("text-anchor:top;font-size:%dpx;fill:black", fontSize))
	s.Text(yLenght-100, fontSize+2+5, "Value Chain", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;font-weight:bold;font-family:serif", fontSize+2))
	s.Gend()
}
