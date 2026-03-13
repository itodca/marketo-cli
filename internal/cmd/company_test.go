package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompanyListUsesNameFilterAndFieldFiltering(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var query string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/companies.json":
			query = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 10, "company": "Acme"}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "company", "list", "--name", "Acme", "--fields", "id")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if query != "fields=id&filterType=company&filterValues=Acme" {
		t.Fatalf("unexpected query: %q", query)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || len(payload[0]) != 1 || payload[0]["id"].(float64) != 10 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestCompanyDescribeFetchesSchema(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/companies/describe.json":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"name": "company"}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "company", "describe")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["name"] != "company" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
