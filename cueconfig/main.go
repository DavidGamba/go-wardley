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

	"github.com/DavidGamba/dgtools/cueutils"
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
	opt.StringSlice("config", 1, 99, opt.Required())
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

	configFilenames := opt.Value("config").([]string)

	configs := []cueutils.CueConfigFile{}

	schemaFilename := "wardley-schema.cue"
	schemaFH, err := f.Open(schemaFilename)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", schemaFilename, err)
	}
	defer schemaFH.Close()
	configs = append(configs, cueutils.CueConfigFile{Data: schemaFH, Name: schemaFilename})

	for _, configFilename := range configFilenames {
		configFH, err := os.Open(configFilename)
		if err != nil {
			return fmt.Errorf("failed to open '%s': %w", configFilename, err)
		}
		defer configFH.Close()
		configs = append(configs, cueutils.CueConfigFile{Data: configFH, Name: configFilename})
	}

	w := Wardley{}
	err = cueutils.Unmarshal(configs, &w)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	pretty, err := json.MarshalIndent(w, "", "\t")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", string(pretty))

	return nil
}

type Wardley struct {
	Size      Size
	Node      map[string]Node
	Connector map[string]Connector
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
