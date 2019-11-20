// This file is part of go-wardley.
//
// Copyright (C) 2019  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package hcl

import (
	"reflect"
	"testing"
)

func TestParseHCL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		nNodes       int
		nConnectors  int
		idxNode      int
		Node         Node
		idxConnector int
		Connector    Connector
	}{
		{name: "empty", input: ``, nNodes: 0, nConnectors: 0, idxNode: -1, idxConnector: -1},
		{name: "node",
			input: `node id {
	  label = "label"
	  x     = 2
	  y     = 3
	}`, nNodes: 1, nConnectors: 0, idxNode: 0, idxConnector: -1, Node: Node{ID: "id", Label: "label", X: 2, Y: 3}},
		{name: "nodes",
			input: `node id {
	  label = "label"
	  x     = 2
	  y     = 3
	}
	node id2 {
	  label = "Label with spaces\nAnd multiline"
	  x     = 2
	  y     = 3
	}`, nNodes: 2, nConnectors: 0, idxNode: 1, idxConnector: -1, Node: Node{ID: "id2", Label: "Label with spaces\nAnd multiline", X: 2, Y: 3}},
		{name: "connector",
			input: `node id {
	  label = "label"
	  x     = 2
	  y     = 3
	}
connector {
	  label = "label"
	  from  = "id"
	  to    = "id2"
	}`, nNodes: 1, nConnectors: 1, idxNode: -1, idxConnector: 0, Connector: Connector{Label: "label", From: "id", To: "id2"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := ParseHCL([]byte(test.input))
			if err != nil {
				t.Fatalf("Unexpected error: %s\n", err)
			}
			if len(m.Nodes) != test.nNodes {
				t.Errorf("Missing elements in map: %+v\n", m)
			}
			if len(m.Connectors) != test.nConnectors {
				t.Errorf("Missing elements in map: %+v\n", m)
			}
			if test.idxNode != -1 {
				if !reflect.DeepEqual(m.Nodes[test.idxNode], &test.Node) {
					t.Errorf("Node differences: got %+v, want %+v", m.Nodes[test.idxNode], test.Node)
				}
			}
			if test.idxConnector != -1 {
				if !reflect.DeepEqual(m.Connectors[test.idxConnector], &test.Connector) {
					t.Errorf("Connector differences: got %+v, want %+v", m.Connectors[test.idxConnector], test.Connector)
				}
			}
		})
	}
}
