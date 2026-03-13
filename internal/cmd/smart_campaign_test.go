package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSmartCampaignListAllOmitsActiveFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var isActive string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/smartCampaigns.json":
			isActive = request.URL.Query().Get("isActive")
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 99}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "smart-campaign", "list", "--all", "--limit", "1")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if isActive != "" {
		t.Fatalf("unexpected isActive value: %q", isActive)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["id"].(float64) != 99 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestSmartCampaignListRejectsActiveAndAll(t *testing.T) {
	exitCode, _, stderr := executeTestCommand(t, nil, nil, "smart-campaign", "list", "--active", "--all")

	if exitCode != 1 {
		t.Fatalf("expected failure exit code, got %d", exitCode)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("stderr was not valid JSON: %v", err)
	}
	if payload["error"] != "Choose only one of --active or --all" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}
