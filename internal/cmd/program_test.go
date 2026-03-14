package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProgramListByNameUsesExactLookup(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var query string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/program/byName.json":
			query = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 42, "name": "Welcome"}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "program", "list", "--name", "Welcome")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if query != "name=Welcome" {
		t.Fatalf("unexpected query: %q", query)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["name"] != "Welcome" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestProgramListUsesOffsetPagination(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/programs.json":
			queries = append(queries, request.URL.RawQuery)
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 1}, {"id": 2}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "program", "list", "--limit", "2")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if len(queries) != 1 || queries[0] != "maxReturn=2&offset=0" {
		t.Fatalf("unexpected queries: %#v", queries)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected two programs, got %#v", payload)
	}
}

func TestProgramGetFetchesProgramByID(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/program/123.json":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 123, "name": "Newsletter"}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "program", "get", "123")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["id"].(float64) != 123 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
