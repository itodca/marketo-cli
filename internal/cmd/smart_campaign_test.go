package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestSmartCampaignScheduleDefaultsToDryRun(t *testing.T) {
	exitCode, stdout, stderr := executeTestCommand(t, nil, nil, "smart-campaign", "schedule", "42", "--run-at", "2026-03-15T12:00:00Z")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if payload["dry_run"] != true || payload["action"] != "schedule" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload["campaign_id"].(float64) != 42 {
		t.Fatalf("unexpected campaign id: %#v", payload)
	}

	request := payload["request"].(map[string]any)
	input := request["input"].(map[string]any)
	if input["runAt"] != "2026-03-15T12:00:00Z" {
		t.Fatalf("unexpected request payload: %#v", request)
	}
}

func TestSmartCampaignTriggerRejectsTooManyLeads(t *testing.T) {
	args := []string{"smart-campaign", "trigger", "42"}
	for index := 1; index <= 101; index++ {
		args = append(args, "--lead", strconv.Itoa(index))
	}

	exitCode, _, stderr := executeTestCommand(t, nil, nil, args...)

	if exitCode != 1 {
		t.Fatalf("expected failure exit code, got %d", exitCode)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("stderr was not valid JSON: %v", err)
	}
	if payload["error"] != "[invalid_input] A maximum of 100 leads is allowed per trigger request" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}

func TestSmartCampaignTriggerExecutePostsLeadPayload(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var method string
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/campaigns/42/trigger.json":
			method = request.Method
			body, _ := io.ReadAll(request.Body)
			rawBody = string(body)
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"status": "triggered"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "smart-campaign", "trigger", "42", "--lead", "1", "--lead", "2", "--execute")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if method != http.MethodPost {
		t.Fatalf("expected POST request, got %s", method)
	}
	if rawBody != "{\"input\":{\"leads\":[{\"id\":1},{\"id\":2}]}}" {
		t.Fatalf("unexpected body: %q", rawBody)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["status"] != "triggered" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
