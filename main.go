// This file is part of go-wardley.
//
// Copyright (C) 2019-2020  David Gamba Rios
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
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/go-wardley/hcl"
	svg "github.com/ajstarks/svgo"
	"github.com/fsnotify/fsnotify"
)

// BuildMetadata - Provides the metadata part of the version information.
var BuildMetadata string = "dev"

var version string = "0.3.0"

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

// Show guides in drawing
var showGuides bool

// Add jitter to nodes' X and Y positions to reduce overlapping
var jitter bool

// Keeps count of connect path IDs
var connectID int

var canvas *svg.SVG

func main() {
	var inputFile, outputFile string
	var port int

	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false, opt.Description("Show debug logs"))
	opt.BoolVar(&jitter, "jitter", false, opt.Description("Add jitter to node positions"))
	opt.IntVarOptional(&port, "serve", 8080, opt.Description("Serve the drawing at localhost:<port>"), opt.ArgName("port"))
	opt.Bool("version", false, opt.Alias("V"), opt.Description("Print version information"))
	opt.Bool("watch", false, opt.Description("Watch file for changes"))
	opt.BoolVar(&showGuides, "guides", false, opt.Description("Show margins, limits and other guides in drawing"))
	opt.StringVar(&inputFile, "file", "", opt.Description("Map input file"), opt.Required(""), opt.ArgName("filename"))
	opt.StringVar(&outputFile, "output", "", opt.Description("Map svg output file, by default replaces input file extension to .svg"), opt.ArgName("filename"))
	_, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		os.Exit(1)
	}
	if opt.Called("version") {
		fmt.Printf("Version: %s+%s\n", version, BuildMetadata)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
		hcl.Logger.SetOutput(os.Stderr)
	}

	if opt.Called("serve") {
		fmt.Printf("Serving content on: http://localhost:%d\n", port)
		serveFile(inputFile, port)
		os.Exit(0)
	}

	if opt.Called("watch") {
		absFile, err := filepath.Abs(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to get Absolute path from input file '%s': %s\n", inputFile, err)
			os.Exit(1)
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to stablish watcher: %s\n", err)
			os.Exit(1)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					logger.Printf("watcher event: %s\n", event.String())
					if event.Name == absFile && event.Op&fsnotify.Write == fsnotify.Write {
						err := render(absFile, outputFile)
						if err != nil {
							fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					logger.Printf("Watcher error: %s\n", err)
				}
			}
		}()

		// Vim deletes the file when saving it making fsnotify loose the pointer to it.
		// Have to watch the dir.
		fmt.Printf("Starting watcher on: %s\n", filepath.Dir(absFile))
		err = watcher.Add(filepath.Dir(absFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: watcher error: %s\n", err)
			os.Exit(1)
		}
		err = render(absFile, outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		}
		<-done
	} else {
		err := render(inputFile, outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
	}
}

func render(inputFile, outputFile string) error {
	m, err := parseInputFile(inputFile)
	if err != nil {
		return err
	}
	if outputFile == "" {
		outputFile = strings.Replace(inputFile, filepath.Ext(inputFile), ".svg", 1)
		logger.Printf("output file: %s\n", outputFile)
	}
	err = renderFile(m, outputFile)
	if err != nil {
		return err
	}
	return nil
}

func parseInputFile(name string) (*hcl.Map, error) {
	parser, f, err := hcl.ParseHCLFile(os.Stderr, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read '%s': %w", name, err)
	}
	m, err := hcl.DecodeMap(os.Stderr, parser, f)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func renderFile(m *hcl.Map, outputFile string) error {
	ofh, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to write to '%s': %w", outputFile, err)
	}
	defer ofh.Close()
	drawing(ofh, m)
	fmt.Printf("Updated file: %s\n", outputFile)
	return nil
}

func serveFile(inputFile string, port int) error {
	http.Handle("/", http.HandlerFunc(drawHandler(inputFile)))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		return err
	}
	return nil
}

func drawHandler(inputFile string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		m, err := parseInputFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		drawing(w, m)
	}
}

func drawing(w io.Writer, m *hcl.Map) {
	canvas = svg.New(w)
	canvas.Start(m.Size.Width, m.Size.Height)
	canvas.Gstyle("font-family:sans-serif")
	grid(canvas, m.Size.Margin, m.Size.Width, m.Size.Height, m.Size.FontSize+2)
	canvas.Translate(m.Size.Margin*2, m.Size.Height-m.Size.Margin*2)
	canvas.Marker("connector-arrow", 17, 3, 12, 10, `orient="auto"`)
	canvas.Path("M0,0 L0,6 L12,3 z")
	canvas.MarkerEnd()
	canvas.Marker("connector-inertia", 0, 10, 20, 40, `orient="auto"`)
	canvas.Path("M-5,20 L-5,-20 L5,-20 L5,20")
	canvas.MarkerEnd()

	nodes := m.Nodes
	connectors := m.Connectors

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
		connect(c, a, b, m.Size.FontSize)
	}
	for _, n := range nodes {
		DrawNode(n, m.Size.FontSize)
	}
	canvas.Gend()
	canvas.Gend()
	canvas.End()
}

var mapGrid Grid

// Grid -
type Grid struct {
	XQuarterLength int
	YLength        int
	Genesis        int
	Custom         int
	Product        int
	Commodity      int
	Visible        int
}

func NodeXY(n *hcl.Node, maxGenesis, maxCustom, maxProduct, maxCommodity, maxY int) {
	switch n.Evolution {
	case "genesis":
		n.X = mapGrid.Genesis + mapGrid.XQuarterLength/(maxGenesis+1)*n.EvolutionX
	case "custom":
		n.X = mapGrid.Custom + mapGrid.XQuarterLength/(maxCustom+1)*n.EvolutionX
	case "product":
		n.X = mapGrid.Product + mapGrid.XQuarterLength/(maxProduct+1)*n.EvolutionX
	case "commodity":
		n.X = mapGrid.Commodity + mapGrid.XQuarterLength/(maxCommodity+1)*n.EvolutionX
	}

	n.Y = -mapGrid.YLength / (maxY + 1) * (maxY + 1 - n.Visibility)
	if jitter {
		jitterSizeX := mapGrid.XQuarterLength / 20
		jitterX := -jitterSizeX + rand.Intn(jitterSizeX*2+1)
		jitterSizeY := mapGrid.YLength / 100
		jitterY := -jitterSizeY + rand.Intn(jitterSizeY*2+1)
		n.X += jitterX
		n.Y += jitterY
	}
}

// DrawNode -
func DrawNode(n *hcl.Node, fontSize int) {
	canvas.Gstyle("text-shadow: 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white")
	if n.Description != "" {
		canvas.Title(n.Description)
	} else {
		canvas.Title(n.Label)
	}
	canvas.Circle(n.X, n.Y, 5, fmt.Sprintf("fill:%s;stroke:%s", n.Fill, n.Color))
	// canvas.Text(n.X+10, n.Y+3, n.Label, fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;text-shadow: -1px 0 white, 0 1px white, 1px 0 white, 0 -1px white", nodeFontSize))
	// canvas.Gstyle("text-shadow: -1px 0 white, 0 1px white, 1px 0 white, 0 -1px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white, 0 0 3px white")
	canvas.Textlines(n.X+8, n.Y+10, strings.Split(n.Label, "\n"), fontSize, fontSize+3, "black", "left")
	canvas.Gend()

}

func connect(c *hcl.Connector, a, b *hcl.Node, fontSize int) {
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
	canvas.Textlines(x, y, strings.Split(c.Label, "\n"), fontSize, fontSize+3, "black", "left")
	// } else {
	// 	s.Textpath(c.Label, fmt.Sprintf("#%d", connectID), `x="10" y="-5"`, fmt.Sprintf("text-align:left;font-size:%dpx;fill:black", nodeFontSize))
	// }
}

func grid(s *svg.SVG, margin, width, height, fontSize int) {
	// Grid
	//   X
	xLength := width - margin*4
	xZero := margin * 2
	xEnd := width - margin*2
	//   Y
	yLength := height - margin*4
	yZero := height - margin*2
	yEnd := margin * 2

	mapGrid = Grid{
		XQuarterLength: (width - margin*4) / 4,
		Genesis:        0,
		Custom:         (width - margin*4) / 4,
		Product:        (width - margin*4) * 2 / 4,
		Commodity:      (width - margin*4) * 3 / 4,
		YLength:        height - margin*4,
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
		s.Text(xLength-40, -yLength, fmt.Sprintf("%d,%d", xLength, yLength), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(0, 0, fmt.Sprintf("%d,%d", xZero, yZero), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLength/4, 0, fmt.Sprintf("%d,%d", xLength/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLength*2/4, 0, fmt.Sprintf("%d,%d", 2*margin+xLength*2/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Text(xLength*3/4, 0, fmt.Sprintf("%d,%d", 2*margin+xLength*3/4, 0), fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:green", fontSize))
		s.Gend()
	}

	// Grid
	s.Marker("arrow", 0, 3, 12, 10, `orient="auto"`)
	s.Path("M0,0 L0,6 L12,3 z", "fill:black")
	s.MarkerEnd()
	s.Line(xZero, yZero, xEnd, yZero, "fill:none;stroke:black;marker-end:url(#arrow)")
	s.Line(xZero, yZero, xZero, yEnd, "fill:blue;stroke:black;marker-end:url(#arrow)")

	s.Line(2*margin+xLength/4, yZero, 2*margin+xLength/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)
	s.Line(2*margin+xLength*2/4, yZero, 2*margin+xLength*2/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)
	s.Line(2*margin+xLength*3/4, yZero, 2*margin+xLength*3/4, yEnd, `fill:none;stroke:gray;stroke-dasharray:1,10`)

	// Text
	s.Text(xZero, height-margin, "Genesis", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLength/4, height-margin, "Custom", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLength*2/4, height-margin, "Product (+rental)", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(2*margin+xLength*3/4, height-margin, "Commodity (+utility)", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black", fontSize))
	s.Text(xEnd-100, height-2*margin-5, "Evolution", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;font-weight:bold;font-family:serif", fontSize+2))

	s.TranslateRotate(xZero, yZero, 270)
	s.Text(0, -5, "Invisible", fmt.Sprintf("text-anchor:top;font-size:%dpx;fill:black", fontSize))
	s.Text(yLength-50, -5, "Visible", fmt.Sprintf("text-anchor:top;font-size:%dpx;fill:black", fontSize))
	s.Text(yLength-100, fontSize+2+5, "Value Chain", fmt.Sprintf("text-anchor:left;font-size:%dpx;fill:black;font-weight:bold;font-family:serif", fontSize+2))
	s.Gend()
}
