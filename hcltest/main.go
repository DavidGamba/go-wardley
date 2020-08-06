package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program())
}

func program() int {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false)
	remaining, err := opt.Parse(os.Args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)
	err = realMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

var input = `node id {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = 1 
			}
			node id2 {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = node.id.x + 1
			}
			node id3 {
				label = "label"
				visibility = 1
				evolution = "custom"
				x = node.id2.x + 1
			}`

type Map struct {
	Nodes []*Node `hcl:"node,block"`
}

type Node struct {
	ID          string `hcl:"label,label"`
	Label       string `hcl:"label"`
	Description string `hcl:"description,optional"`
	X           int
	Y           int
	Visibility  int    `hcl:"visibility"`
	Evolution   string `hcl:"evolution"`
	EvolutionX  int    `hcl:"x" cty:"x"`
	Fill        string `hcl:"fill,optional"`
	Color       string `hcl:"color,optional"`
}

func realMain() error {
	rootSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "node", LabelNames: []string{"name"}},
		},
	}

	buf := new(bytes.Buffer)
	parser, f, err := ParseHCL(buf, []byte(input), "test.hcl")
	if err != nil {
		return err
	}

	content, diags := f.Body.Content(rootSchema)
	err = handleDiags(parser, diags, os.Stderr)
	if err != nil {
		return err
	}
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}

	nodeType := cty.Object(map[string]cty.Type{
		"x": cty.Number,
	})
	fmt.Printf("ctx: %#v\n", ctx)

	for k, block := range content.Blocks {
		fmt.Println("-------------------------------------------------")
		fmt.Printf("key: %v block.type: %v\n", k, block.Type)
		fmt.Println("DefRange:", block.DefRange)
		fmt.Println("TypeRange:", block.TypeRange)
		fmt.Println("LabelRanges:", block.LabelRanges)
		fmt.Println("Labels:", block.Labels)
		switch block.Type {
		case "node":
			var node Node
			diags := gohcl.DecodeBody(block.Body, ctx, &node)
			err = handleDiags(parser, diags, os.Stderr)
			if err != nil {
				return err
			}
			fmt.Printf("node: %#v\n", node)

			v, err := gocty.ToCtyValue(node, nodeType)
			if err != nil {
				return err
			}
			fmt.Printf("v: %#v\n", v)

			var m map[string]cty.Value
			n, ok := ctx.Variables["node"]
			if !ok {
				m = map[string]cty.Value{
					block.Labels[0]: v,
				}
			} else {
				fmt.Printf("n: %#v\n", n)
				fmt.Printf("n: %#v\n", n.AsValueMap())
				m = n.AsValueMap()
				fmt.Printf("m: %#v\n", m)
				m[block.Labels[0]] = v
			}
			fmt.Printf("m: %#v\n", m)
			ctx.Variables["node"] = cty.MapVal(m)
			fmt.Printf("ctx: %#v\n", ctx)
		}
	}
	return nil
}

func ParseHCL(w io.Writer, data []byte, filename string) (*hclparse.Parser, *hcl.File, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		wr := hcl.NewDiagnosticTextWriter(
			w,              // writer to send messages to
			parser.Files(), // the parser's file cache, for source snippets
			100,            // wrapping width
			true,           // generate colored/highlighted output
		)
		wr.WriteDiagnostics(diags)
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
		fmt.Errorf("failure during decoding")
	}
	return nil
}
