package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSmartListListUsesFolderFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var folder string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/smartLists.json":
			folder = request.URL.Query().Get("folder")
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 7}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "smart-list", "list", "--folder-id", "7", "--folder-type", "Program", "--limit", "1")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if folder != "{\"id\":7,\"type\":\"Program\"}" {
		t.Fatalf("unexpected folder value: %q", folder)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["id"].(float64) != 7 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestSmartListGetCanIncludeRules(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var includeRules string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/smartList/12.json":
			includeRules = request.URL.Query().Get("includeRules")
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 12}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "smart-list", "get", "12", "--include-rules")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if includeRules != "true" {
		t.Fatalf("unexpected includeRules value: %q", includeRules)
	}
}
