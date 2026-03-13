package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatsUsageWeeklyUsesLast7DaysPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var path string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/stats/usage/last7days.json":
			path = request.URL.Path
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"calls": 42}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "stats", "usage", "--weekly")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if path != "/rest/v1/stats/usage/last7days.json" {
		t.Fatalf("unexpected path: %q", path)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["calls"].(float64) != 42 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
