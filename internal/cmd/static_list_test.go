package cmd

import (
	"encoding/json"
	"io"
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

func TestStaticListAddDefaultsToDryRun(t *testing.T) {
	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "static-list", "add", "55", "--lead", "1", "--lead", "2")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if payload["dry_run"] != true || payload["action"] != "add" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload["list_id"].(float64) != 55 {
		t.Fatalf("unexpected list id: %#v", payload)
	}
}

func TestStaticListRemoveDryRunIncludesQueryParams(t *testing.T) {
	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "static-list", "remove", "55", "--lead", "1", "--lead", "2")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if payload["dry_run"] != true || payload["action"] != "remove" {
		t.Fatalf("unexpected payload: %#v", payload)
	}

	params := payload["params"].(map[string]any)
	ids := params["id"].([]any)
	if len(ids) != 2 || ids[0] != "1" || ids[1] != "2" {
		t.Fatalf("unexpected params payload: %#v", params)
	}
}

func TestStaticListRemoveExecuteSendsQueryAndBody(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var method string
	var query string
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lists/55/leads.json":
			method = request.Method
			query = request.URL.RawQuery
			body, _ := io.ReadAll(request.Body)
			rawBody = string(body)
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"status": "removed"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "static-list", "remove", "55", "--lead", "1", "--lead", "2", "--execute")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if method != http.MethodDelete {
		t.Fatalf("expected DELETE request, got %s", method)
	}
	if query != "id=1&id=2" {
		t.Fatalf("unexpected query: %q", query)
	}
	if rawBody != "{\"input\":[{\"id\":1},{\"id\":2}]}" {
		t.Fatalf("unexpected body: %q", rawBody)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["status"] != "removed" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
