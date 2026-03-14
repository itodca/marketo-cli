package output

import (
	"encoding/json"
	"errors"
	"io"
)

type Format string

const (
	FormatJSON    Format = "json"
	FormatCompact Format = "compact"
	FormatRaw     Format = "raw"
)

func ResolveFormat(jsonOutput, compactOutput, rawOutput bool) (Format, error) {
	selected := 0
	if jsonOutput {
		selected++
	}
	if compactOutput {
		selected++
	}
	if rawOutput {
		selected++
	}
	if selected > 1 {
		return "", errors.New("choose only one of --json, --compact, or --raw")
	}
	if compactOutput {
		return FormatCompact, nil
	}
	if rawOutput {
		return FormatRaw, nil
	}
	return FormatJSON, nil
}

func Payload(data any, format Format) any {
	if format == FormatRaw {
		return data
	}

	if envelope, ok := data.(map[string]any); ok {
		success, hasSuccess := envelope["success"].(bool)
		if hasSuccess && success {
			if result, hasResult := envelope["result"]; hasResult {
				return result
			}
		}
	}

	return data
}

func PrintResult(writer io.Writer, data any, format Format, fields []string) error {
	filtered := filterFields(data, fields)

	switch format {
	case FormatJSON:
		encoded, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return err
		}
		_, err = writer.Write(append(encoded, '\n'))
		return err
	case FormatCompact:
		switch items := filtered.(type) {
		case []any:
			for _, item := range items {
				if err := writeJSONLine(writer, item); err != nil {
					return err
				}
			}
			return nil
		case []map[string]any:
			for _, item := range items {
				if err := writeJSONLine(writer, item); err != nil {
					return err
				}
			}
			return nil
		default:
			return writeJSONLine(writer, filtered)
		}
	case FormatRaw:
		return writeJSONLine(writer, filtered)
	default:
		return errors.New("unsupported output format")
	}
}

func PrintError(writer io.Writer, message string) error {
	return writeJSONLine(writer, map[string]string{"error": message})
}

func writeJSONLine(writer io.Writer, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = writer.Write(append(encoded, '\n'))
	return err
}

func filterFields(data any, fields []string) any {
	if len(fields) == 0 {
		return data
	}

	fieldSet := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		fieldSet[field] = struct{}{}
	}

	switch value := data.(type) {
	case map[string]any:
		return filterMap(value, fieldSet)
	case []map[string]any:
		filtered := make([]map[string]any, 0, len(value))
		for _, item := range value {
			filtered = append(filtered, filterMap(item, fieldSet))
		}
		return filtered
	case []any:
		filtered := make([]any, 0, len(value))
		for _, item := range value {
			if record, ok := item.(map[string]any); ok {
				filtered = append(filtered, filterMap(record, fieldSet))
				continue
			}
			filtered = append(filtered, item)
		}
		return filtered
	default:
		return data
	}
}

func filterMap(data map[string]any, fields map[string]struct{}) map[string]any {
	filtered := make(map[string]any)
	for key, value := range data {
		if _, ok := fields[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}
