// This file is part of go-wardley.
//
// Copyright (C) 2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package hcl

import (
	"fmt"

	// "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	// "github.com/zclconf/go-cty/cty"
)

// Map -
type Map struct {
	Nodes      []*Node      `hcl:"node,block"`
	Connectors []*Connector `hcl:"connector,block"`
}

// Node -
type Node struct {
	ID          string `hcl:"label,label"`
	Label       string `hcl:"label"`
	Description string `hcl:"description"`
	X           int
	Y           int
	Visibility  int    `hcl:"visibility"`
	Evolution   string `hcl:"evolution"`
	EvolutionX  int    `hcl:"x"`
	Fill        string `hcl:"fill"`
	Color       string `hcl:"color"`
}

// Connector -
type Connector struct {
	Label string `hcl:"label"`
	From  string `hcl:"from"`
	To    string `hcl:"to"`
	Color string `hcl:"color"`
	Type  string `hcl:"type"`
	// From hcl.Expression `hcl:"from,attr"`
	// To   hcl.Expression `hcl:"to,attr"`
}

// ParseHCL -
func ParseHCL(data []byte) (*Map, error) {
	var m Map
	// ctx := &hcl.EvalContext{
	// 	Variables: map[string]cty.Value{
	// 		"hola": cty.StringVal("hola"),
	// 	},
	// }
	err := hclsimple.Decode("map.hcl", data, nil, &m)
	if err != nil {
		return &m, fmt.Errorf("Failed to load configuration: %w", err)
	}

	// Code if the definition of From and To looked like:
	//     From    hcl.Expression `hcl:"from,attr"`
	//     To      hcl.Expression `hcl:"to,attr"`
	// Would allow for: from = node.node_name.id
	//     for _, c := range m.Connectors {
	//     	traversal, diags := hcl.AbsTraversalForExpr(c.To)
	//     	if len(diags) != 0 {
	//     		for _, diag := range diags {
	//     			fmt.Printf("diag: - %s", diag)
	//     		}
	//     		return &m, fmt.Errorf("unexpected diagnostics extracting To")
	//     	}
	//     	fmt.Printf("%+v, %+v\n", traversal.RootName(), diags)
	//     }

	return &m, nil
}

// ParseHCLFile -
func ParseHCLFile(file string) (*Map, error) {
	var m Map
	// ctx := &hcl.EvalContext{
	// 	Variables: map[string]cty.Value{
	// 		"name": cty.StringVal("hola"),
	// 	},
	// }
	err := hclsimple.DecodeFile(file, nil, &m)
	if err != nil {
		return &m, fmt.Errorf("Failed to load configuration: %w", err)
	}
	return &m, nil
}
