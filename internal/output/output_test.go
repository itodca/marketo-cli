package output

import (
	"bytes"
	"testing"
)

func TestResolveFormatRejectsMultipleFlags(t *testing.T) {
	t.Parallel()

	if _, err := ResolveFormat(true, true, false); err == nil {
		t.Fatal("expected format validation error")
	}
}

func TestPrintResultCompactPrintsOneRecordPerLine(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	err := PrintResult(&buffer, []any{
		map[string]any{"id": 1},
		map[string]any{"id": 2},
	}, FormatCompact, nil)
	if err != nil {
		t.Fatalf("PrintResult returned error: %v", err)
	}

	expected := "{\"id\":1}\n{\"id\":2}\n"
	if buffer.String() != expected {
		t.Fatalf("expected %q, got %q", expected, buffer.String())
	}
}

func TestPayloadUnwrapsSuccessEnvelopeUnlessRaw(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"success": true,
		"result":  []any{map[string]any{"profile": "default"}},
	}

	unwrapped := Payload(data, FormatJSON)
	if _, ok := unwrapped.([]any); !ok {
		t.Fatalf("expected result slice, got %#v", unwrapped)
	}

	raw := Payload(data, FormatRaw)
	if _, ok := raw.(map[string]any); !ok {
		t.Fatalf("expected full envelope in raw mode, got %#v", raw)
	}
}

func TestPrintResultFieldFilteringLeavesNonObjectItemsUntouched(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	err := PrintResult(&buffer, []any{
		map[string]any{"id": 1, "name": "alpha"},
		"not-an-object",
		map[string]any{"id": 2, "name": "beta"},
	}, FormatCompact, []string{"id"})
	if err != nil {
		t.Fatalf("PrintResult returned error: %v", err)
	}

	expected := "{\"id\":1}\n\"not-an-object\"\n{\"id\":2}\n"
	if buffer.String() != expected {
		t.Fatalf("expected %q, got %q", expected, buffer.String())
	}
}
