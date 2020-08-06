// This file is part of go-wardley.
//
// Copyright (C) 2019-2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// package hcl - implements the HCL related functionality.
// Input parsing, input error handling.
package hcl

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/imdario/mergo"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Map -
type Map struct {
	Size       *Size        `hcl:"size,block"`
	Nodes      []*Node      `hcl:"node,block"`
	Connectors []*Connector `hcl:"connector,block"`
}

var mapSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "size"},
		{Type: "node", LabelNames: []string{"id"}},
		{Type: "connector"},
	},
}

var mapDefaults = Map{
	Size: &Size{
		Width:    1280,
		Height:   768,
		Margin:   40,
		FontSize: 12,
	},
}

type Size struct {
	Width    int `hcl:"width,optional"`
	Height   int `hcl:"height,optional"`
	Margin   int `hcl:"margin,optional"`
	FontSize int `hcl:"font_size,optional"`
}

var sizeDefaults = Size{
	Width:    1280,
	Height:   768,
	Margin:   40,
	FontSize: 12,
}

// Node -
type Node struct {
	ID          string `hcl:"id,label"`
	Label       string `hcl:"label"`
	Description string `hcl:"description,optional"`
	X           int
	Y           int
	Visibility  int    `hcl:"visibility" cty:"visibility"`
	Evolution   string `hcl:"evolution"`
	EvolutionX  int    `hcl:"x" cty:"x"`
	Fill        string `hcl:"fill,optional"`
	Color       string `hcl:"color,optional"`
}

var nodeType = cty.Object(map[string]cty.Type{
	"x":          cty.Number,
	"visibility": cty.Number,
})

var nodeDefaults = Node{
	Fill:  "white",
	Color: "black",
}

// Connector -
type Connector struct {
	Label string `hcl:"label,optional"`
	From  string `hcl:"from"`
	To    string `hcl:"to"`
	Color string `hcl:"color,optional"`
	Type  string `hcl:"type,optional"`
	// From hcl.Expression `hcl:"from,attr"`
	// To   hcl.Expression `hcl:"to,attr"`
}

var connectorDefaults = Connector{
	Color: "black",
	Type:  "normal",
}

func ParseHCL(w io.Writer, data []byte, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(data, filename)
	err := handleDiags(w, parser, diags)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

func ParseHCLFile(w io.Writer, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	err := handleDiags(w, parser, diags)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

func handleDiags(w io.Writer, parser *hclparse.Parser, diags hcl.Diagnostics) error {
	if diags.HasErrors() {
		wr := hcl.NewDiagnosticTextWriter(
			w,              // writer to send messages to
			parser.Files(), // the parser's file cache, for source snippets
			100,            // wrapping width
			true,           // generate colored/highlighted output
		)
		wr.WriteDiagnostics(diags)
		return fmt.Errorf("errors found")
	}
	return nil
}

func DecodeMap(w io.Writer, parser *hclparse.Parser, f *hcl.File) (*Map, error) {
	mapDetails := &Map{}

	content, diags := f.Body.Content(mapSchema)
	err := handleDiags(w, parser, diags)
	if err != nil {
		return nil, err
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "size":
			var size Size
			diags := gohcl.DecodeBody(block.Body, ctx, &size)
			err = handleDiags(w, parser, diags)
			if err != nil {
				return mapDetails, err
			}
			mapDetails.Size = &size
		case "node":
			var node Node
			diags := gohcl.DecodeBody(block.Body, ctx, &node)
			err = handleDiags(w, parser, diags)
			if err != nil {
				return mapDetails, err
			}
			node.ID = block.Labels[0]
			mapDetails.Nodes = append(mapDetails.Nodes, &node)

			v, err := gocty.ToCtyValue(node, nodeType)
			if err != nil {
				return mapDetails, err
			}
			// fmt.Printf("v: %#v\n", v)

			var m map[string]cty.Value
			n, ok := ctx.Variables["node"]
			if !ok {
				m = map[string]cty.Value{
					block.Labels[0]: v,
				}
			} else {
				// fmt.Printf("n: %#v\n", n)
				// fmt.Printf("n: %#v\n", n.AsValueMap())
				m = n.AsValueMap()
				// fmt.Printf("m: %#v\n", m)
				m[block.Labels[0]] = v
			}
			// fmt.Printf("m: %#v\n", m)
			ctx.Variables["node"] = cty.MapVal(m)
			// fmt.Printf("ctx: %#v\n", ctx)
		case "connector":
			var connector Connector
			diags := gohcl.DecodeBody(block.Body, ctx, &connector)
			err = handleDiags(w, parser, diags)
			if err != nil {
				return mapDetails, err
			}
			mapDetails.Connectors = append(mapDefaults.Connectors, &connector)
		}
	}

	// TODO: New way to calculate defaults
	// err = mergo.Merge(&mapDetails, mapDefaults)
	// if err != nil {
	// 	return nil, fmt.Errorf("failure calculating map defaults: %w", err)
	// }

	// TODO: Figure why this fails
	if mapDetails.Size == nil {
		mapDetails.Size = &sizeDefaults
	}
	size := mapDetails.Size
	err = mergo.Merge(size, sizeDefaults)
	if err != nil {
		return nil, fmt.Errorf("failure calculating size defaults: %w", err)
	}

	for _, node := range mapDetails.Nodes {
		err := mergo.Merge(node, nodeDefaults)
		if err != nil {
			return nil, fmt.Errorf("failure calculating node defaults: %w", err)
		}
	}
	for _, connector := range mapDetails.Connectors {
		err := mergo.Merge(connector, connectorDefaults)
		if err != nil {
			return nil, fmt.Errorf("failure calculating connector defaults: %w", err)
		}
	}
	return mapDetails, nil
}
