package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/itodca/marketo-cli/internal/client"
	"github.com/itodca/marketo-cli/internal/config"
	"github.com/itodca/marketo-cli/internal/output"
)

type Runtime struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Cwd    string
	Getenv func(string) string
}

type RootOptions struct {
	Profile string
	JSON    bool
	Compact bool
	Raw     bool
}

func NewRuntime() *Runtime {
	return &Runtime{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getenv: os.Getenv,
	}
}

func (runtime *Runtime) CurrentDir() (string, error) {
	if runtime.Cwd != "" {
		return runtime.Cwd, nil
	}
	return os.Getwd()
}

func (runtime *Runtime) Env(key string) string {
	if runtime.Getenv != nil {
		return runtime.Getenv(key)
	}
	return os.Getenv(key)
}

func writeResult(runtime *Runtime, options *RootOptions, data any) error {
	return writeResultFields(runtime, options, data, nil)
}

func writeResultFields(runtime *Runtime, options *RootOptions, data any, fields []string) error {
	format, err := output.ResolveFormat(options.JSON, options.Compact, options.Raw)
	if err != nil {
		return err
	}

	return output.PrintResult(runtime.Stdout, output.Payload(data, format), format, fields)
}

func writeError(runtime *Runtime, err error) {
	if err == nil {
		return
	}
	message := strings.TrimSpace(err.Error())
	_ = output.PrintError(runtime.Stderr, message)
}

func loadClient(runtime *Runtime, profileName string) (*client.Client, error) {
	currentDir, err := runtime.CurrentDir()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(profileName, currentDir, runtime.Env)
	if err != nil {
		return nil, err
	}

	return client.New(cfg), nil
}

func parseFields(raw string) []string {
	if raw == "" {
		return nil
	}

	values := []string{}
	for _, field := range strings.Split(raw, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			values = append(values, field)
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func parseKVPairs(values []string) (map[string]any, error) {
	if len(values) == 0 {
		return nil, nil
	}

	result := map[string]any{}
	for _, value := range values {
		key, raw, ok := strings.Cut(value, "=")
		if !ok {
			return nil, fmt.Errorf("Invalid key=value pair: %s", value)
		}

		key = strings.TrimSpace(key)
		raw = strings.TrimSpace(raw)
		if existing, exists := result[key]; exists {
			switch typed := existing.(type) {
			case []string:
				result[key] = append(typed, raw)
			case string:
				result[key] = []string{typed, raw}
			default:
				result[key] = []string{fmt.Sprint(existing), raw}
			}
			continue
		}

		result[key] = raw
	}

	return result, nil
}

func parseIntArg(label, raw string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("Invalid %s: %s", label, raw)
	}
	return value, nil
}

func loadJSONInput(runtime *Runtime, inputPath string) (map[string]any, error) {
	if inputPath == "" {
		return nil, nil
	}

	var payload []byte
	var err error
	if inputPath == "-" {
		payload, err = io.ReadAll(runtime.Stdin)
	} else {
		payload, err = os.ReadFile(inputPath)
	}
	if err != nil {
		return nil, err
	}

	var decoded any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	result, ok := decoded.(map[string]any)
	if !ok || result == nil {
		return nil, fmt.Errorf("JSON input must be an object")
	}
	return result, nil
}
