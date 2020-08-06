// This file is part of go-wardley.
//
// Copyright (C) 2019-2020  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package hcl

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDecodeMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Map
	}{
		{"empty", "", &mapDefaults},
		{"empty", "size {}", &mapDefaults},
		{"optional", "size { width = 7 }", &Map{
			Size: &Size{Width: 7, Height: 768, Margin: 40, FontSize: 12},
		}},
		{"node", `node id {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = 1 
			}`, &Map{
			Size:  &Size{Width: 1280, Height: 768, Margin: 40, FontSize: 12},
			Nodes: []*Node{{ID: "id", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 1, Fill: "white", Color: "black"}},
		}},
		{"connector", `connector {
				label = "label"
				to = "to"
				from = "from"
			}`, &Map{
			Size:       &Size{Width: 1280, Height: 768, Margin: 40, FontSize: 12},
			Connectors: []*Connector{{Label: "label", To: "to", From: "from", Color: "black", Type: "normal"}},
		}},
		{"all", `node id {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = 1 
			}
			node id2 {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = 1 
			}
			connector {
				label = "label"
				to = "to"
				from = "from"
			}`, &Map{
			Size: &Size{Width: 1280, Height: 768, Margin: 40, FontSize: 12},
			Nodes: []*Node{
				{ID: "id", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 1, Fill: "white", Color: "black"},
				{ID: "id2", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 1, Fill: "white", Color: "black"},
			},
			Connectors: []*Connector{{Label: "label", To: "to", From: "from", Color: "black", Type: "normal"}},
		}},
		{"references", `node id {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = 1 
			}
			node id2 {
				label = "label"
				visibility = node.id.visibility
				evolution = "custom"
				x = node.id.x + 1
			}
			node id3 {
				label = "label"
				visibility = node.id.visibility
				evolution = "custom"
				x = node.id2.x + 1
			}
			connector {
				label = "label"
				to = "to"
				from = "from"
			}`, &Map{
			Size: &Size{Width: 1280, Height: 768, Margin: 40, FontSize: 12},
			Nodes: []*Node{
				{ID: "id", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 1, Fill: "white", Color: "black"},
				{ID: "id2", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 2, Fill: "white", Color: "black"},
				{ID: "id3", Label: "label", Visibility: 1, Evolution: "custom", EvolutionX: 3, Fill: "white", Color: "black"},
			},
			Connectors: []*Connector{{Label: "label", To: "to", From: "from", Color: "black", Type: "normal"}},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			parser, f, err := ParseHCL(buf, []byte(test.input), "test.hcl")
			if err != nil {
				t.Fatalf("%s\n%s\n", err, buf.String())
			}
			mapDetails, err := DecodeMap(buf, parser, f)
			if buf.String() != "" {
				t.Errorf("output didn't match expected value: %s", buf.String())
			}
			if err != nil {
				out := new(bytes.Buffer)
				spew.Fdump(out, mapDetails)
				t.Fatalf("Error: %s, %s", err, out.String())
			}
			if !reflect.DeepEqual(mapDetails, test.expected) {
				out := new(bytes.Buffer)
				exp := new(bytes.Buffer)
				spew.Fdump(out, mapDetails)
				spew.Fdump(exp, test.expected)
				t.Fatalf("unexpected value:\n%s!=\n%s", out.String(), exp.String())
			}
		})
	}
}
