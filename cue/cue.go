package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueErrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/DavidGamba/go-getoptions"
)

//go:embed wardley-schema.cue
var f embed.FS

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("quiet", false, opt.GetEnv("QUIET"))
	opt.SetCommandFn(Run)
	opt.String("config", "", opt.Required())
	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Parsing cue config")

	configFilename := opt.Value("config").(string)

	c := cuecontext.New()
	schema, err := f.ReadFile("wardley-schema.cue")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}
	// compile our schema first
	s := c.CompileBytes(schema)

	configData, err := os.ReadFile(configFilename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	// compile our value with scope
	v := c.CompileBytes(configData, cue.Scope(s))
	if v.Err() != nil {
		return fmt.Errorf("failed to compile config: %s", cueErrors.Details(v.Err(), nil))
	}
	err = v.Validate(
		cue.Final(),
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Hidden(true),
		cue.Optional(true),
	)
	if err != nil {
		// TODO: We could print the line with the error by parsing the possitions response.
		// Logger.Printf("possitions: %v\n", cueErrors.Positions(err))
		return fmt.Errorf("failed config validation of file '%s': %s", configFilename, cueErrors.Details(err, nil))
	}

	w := Wardley{}

	g := gocodec.New((*cue.Runtime)(c), nil)
	err = g.Encode(v, &w)
	if err != nil {
		return err
	}
	fmt.Printf("map: %v\n", w)
	pretty, err := json.MarshalIndent(w, "", "\t")
	if err != nil {
		return err
	}
	fmt.Printf("map: %v\n", string(pretty))

	i, err := v.Fields()
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}
	for i.Next() {
		fmt.Printf("value %s: %v\n", i.Value().Path(), i.Value())
	}

	return nil
}

type Wardley struct {
	Map struct {
		Size      Size
		Node      map[string]Node
		Connector map[string]Connector
	}
}

type Size struct {
	Width    int
	Height   int
	Margin   int
	FontSize int `json:"font_size"`
}

type Node struct {
	ID          string
	Label       string
	Visibility  int
	Evolution   string
	X           int
	Description string
	Fill        string
	Color       string
}

type Connector struct {
	ID    string
	From  string
	To    string
	Label string
	Color string
	Type  string
}
