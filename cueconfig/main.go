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

	configs := []CueConfigFile{}

	schemaFilename := "wardley-schema.cue"
	schemaFH, err := f.Open(schemaFilename)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", schemaFilename, err)
	}
	defer schemaFH.Close()
	configs = append(configs, CueConfigFile{schemaFH, schemaFilename})

	for _, configFilename := range configFilenames {
		configFH, err := os.Open(configFilename)
		if err != nil {
			return fmt.Errorf("failed to open '%s': %w", configFilename, err)
		}
		defer configFH.Close()
		configs = append(configs, CueConfigFile{configFH, configFilename})
	}

	w := Wardley{}
	err = Unmarshal(configs, &w)
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

type CueConfigFile struct {
	Data io.Reader
	Name string
}

func Unmarshal(configs []CueConfigFile, v any) error {
	c := cuecontext.New()
	value := cue.Value{}
	for i, reader := range configs {
		d, err := io.ReadAll(reader.Data)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
		Logger.Printf("compiling %s\n", reader.Name)
		var t cue.Value
		if i == 0 {
			t = c.CompileBytes(d, cue.Filename(reader.Name))
		} else {
			t = c.CompileBytes(d, cue.Filename(reader.Name), cue.Scope(value))
		}
		value = value.Unify(t)
	}
	if value.Err() != nil {
		return fmt.Errorf("failed to compile: %s", cueErrors.Details(value.Err(), nil))
	}
	err := value.Validate(
		cue.Final(),
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Hidden(true),
		cue.Optional(true),
	)
	if err != nil {
		return fmt.Errorf("failed config validation: %s", cueErrors.Details(err, nil))
	}

	g := gocodec.New((*cue.Runtime)(c), nil)
	err = g.Encode(value, &v)
	if err != nil {
		return fmt.Errorf("failed to encode cue values: %w", err)
	}
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
