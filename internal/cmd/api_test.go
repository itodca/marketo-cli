package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthCheckValidatesConfiguredProfile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/activities/types.json":
			if request.Header.Get("Authorization") != "Bearer tok123" {
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result": []map[string]any{
					{"id": 1},
					{"id": 2},
				},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "auth", "check")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected one result row, got %#v", payload)
	}
	if payload[0]["status"] != "ok" {
		t.Fatalf("unexpected status payload: %#v", payload[0])
	}
	if payload[0]["activity_types_available"].(float64) != 2 {
		t.Fatalf("unexpected activity type count: %#v", payload[0])
	}
}

func TestAPIGetParsesQueryAndFiltersFields(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			rawQuery = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result": []map[string]any{
					{"id": 1, "email": "user@example.com"},
				},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "api", "get", "/v1/leads.json", "--query", "filterType=email", "--query", "filterValues=user@example.com", "--fields", "id")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if rawQuery != "filterType=email&filterValues=user%40example.com" {
		t.Fatalf("unexpected query string: %q", rawQuery)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || len(payload[0]) != 1 || payload[0]["id"].(float64) != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestAPIPostSendsBodyPairs(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var method string
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads/push.json":
			method = request.Method
			body, _ := io.ReadAll(request.Body)
			rawBody = string(body)
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"status": "queued"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "api", "post", "/v1/leads/push.json", "--body", "name=from-flag")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if method != http.MethodPost {
		t.Fatalf("expected POST request, got %s", method)
	}
	if rawBody != "{\"name\":\"from-flag\"}" {
		t.Fatalf("unexpected body: %q", rawBody)
	}
}

func TestAPIDeleteReadsJSONInputFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	payloadPath := filepath.Join(t.TempDir(), "payload.json")
	if err := os.WriteFile(payloadPath, []byte(`{"input":[{"id":1}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var method string
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lists/123/leads.json":
			method = request.Method
			body, _ := io.ReadAll(request.Body)
			rawBody = string(body)
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "api", "delete", "/v1/lists/123/leads.json", "--input", payloadPath)

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if method != http.MethodDelete {
		t.Fatalf("expected DELETE request, got %s", method)
	}
	if rawBody != "{\"input\":[{\"id\":1}]}" {
		t.Fatalf("unexpected body: %q", rawBody)
	}
}

func TestAPIPostRejectsBodyAndInputTogether(t *testing.T) {
	payloadPath := filepath.Join(t.TempDir(), "payload.json")
	if err := os.WriteFile(payloadPath, []byte(`{"name":"from-file"}`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	exitCode, _, stderr := executeTestCommand(t, nil, nil, "api", "post", "/v1/leads/push.json", "--body", "name=from-flag", "--input", payloadPath)

	if exitCode != 1 {
		t.Fatalf("expected failure exit code, got %d", exitCode)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("stderr was not valid JSON: %v", err)
	}
	if payload["error"] != "Use either --body or --input, not both" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}

func executeTestCommand(t *testing.T, env map[string]string, stdin *bytes.Buffer, args ...string) (int, string, string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	input := bytes.NewBuffer(nil)
	if stdin != nil {
		input = stdin
	}

	runtime := &Runtime{
		Stdin:  input,
		Stdout: &stdout,
		Stderr: &stderr,
		Cwd:    t.TempDir(),
		Getenv: func(key string) string {
			return env[key]
		},
	}

	command := NewRootCmd(runtime)
	command.SetArgs(args)

	if err := command.Execute(); err != nil {
		var exitErr *exitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), stdout.String(), stderr.String()
		}
		writeError(runtime, err)
		return 1, stdout.String(), stderr.String()
	}

	return 0, stdout.String(), stderr.String()
}

func writeTestConfig(t *testing.T, homeDir string) {
	t.Helper()

	configDir := filepath.Join(homeDir, ".config", "mrkto")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(strings.Join([]string{
		"MARKETO_MUNCHKIN_ID=123-ABC-456",
		"MARKETO_CLIENT_ID=test-id",
		"MARKETO_CLIENT_SECRET=test-secret",
		"",
	}, "\n")), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}
