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
)

// Map -
type Map struct {
	Size       *Size        `hcl:"size,block"`
	Nodes      []*Node      `hcl:"node,block"`
	Connectors []*Connector `hcl:"connector,block"`
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
	FontSize: 9,
}

// Node -
type Node struct {
	ID          string `hcl:"label,label"`
	Label       string `hcl:"label"`
	Description string `hcl:"description,optional"`
	X           int
	Y           int
	Visibility  int    `hcl:"visibility"`
	Evolution   string `hcl:"evolution"`
	EvolutionX  int    `hcl:"x"`
	Fill        string `hcl:"fill,optional"`
	Color       string `hcl:"color,optional"`
}

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

var mapDefaults = Map{
	Size: &Size{
		Width:    1280,
		Height:   768,
		Margin:   40,
		FontSize: 12,
	},
}

func ParseHCL(w io.Writer, data []byte, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(data, filename)
	err := handleDiags(parser, diags, w)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

func ParseHCLFile(w io.Writer, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	err := handleDiags(parser, diags, w)
	if err != nil {
		return parser, f, fmt.Errorf("failure during input configuration parsing")
	}
	return parser, f, nil
}

func handleDiags(parser *hclparse.Parser, diags hcl.Diagnostics, w io.Writer) error {
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
	var mapDetails Map

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
	// TODO: Still using gohcl instead of hcldec
	// See: https://github.com/zclconf/go-cty/issues/38
	diags := gohcl.DecodeBody(f.Body, ctx, &mapDetails)
	err := handleDiags(parser, diags, w)
	if err != nil {
		return nil, fmt.Errorf("failure during decoding")
	}

	err = mergo.Merge(&mapDetails, mapDefaults)
	if err != nil {
		return nil, fmt.Errorf("failure calculating map defaults: %w", err)
	}

	// TODO: Figure why this fails
	// size := mapDetails.Size
	// err := mergo.Merge(size, sizeDefaults)
	// if err != nil {
	// 	return nil, fmt.Errorf("failure calculating size defaults: %w", err)
	// }

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
	return &mapDetails, nil
}
