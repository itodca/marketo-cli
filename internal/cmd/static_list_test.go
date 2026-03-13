package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticListMembersSupportsFieldFiltering(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var query string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lists/55/leads.json":
			query = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 1, "email": "user@example.com"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "static-list", "members", "55", "--fields", "id")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if query != "fields=id" {
		t.Fatalf("unexpected query: %q", query)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || len(payload[0]) != 1 || payload[0]["id"].(float64) != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestStaticListCheckEncodesLeadIDs(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var query string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lists/55/leads/ismember.json":
			query = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 1, "isMember": true}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "static-list", "check", "55", "--lead", "1", "--lead", "2")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if query != "id=1&id=2" {
		t.Fatalf("unexpected query: %q", query)
	}
}
