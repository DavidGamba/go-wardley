package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/DavidGamba/go-getoptions"
)

var Logger = log.New(io.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false, opt.GetEnv("DEBUG"))
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		Logger.SetOutput(os.Stderr)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func run(ctx context.Context) error {
	c := cuecontext.New()
	schema, err := os.ReadFile("./wardley-definitions.cue")
	if err != nil {
		return fmt.Errorf("failed to read definitions: %w", err)
	}
	// compile our schema first
	s := c.CompileBytes(schema)

	values, err := os.ReadFile("./wardley.cue")
	if err != nil {
		return fmt.Errorf("failed to read values: %w", err)
	}
	// compile our value with scope
	v := c.CompileBytes(values, cue.Scope(s))
	// fmt.Printf("---\n%v\n---\n", v)

	i, err := v.Fields()
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}
	for i.Next() {
		fmt.Printf("value %s: %v\n", i.Value().Path(), i.Value())
	}

	return nil
}
