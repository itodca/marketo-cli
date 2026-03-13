package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLeadGetUsesDefaultFieldsAndReturnsLead(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lead/123.json":
			rawQuery = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result": []map[string]any{
					{"id": 123, "email": "user@example.com"},
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
	}, nil, "lead", "get", "123")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if rawQuery != "fields=id%2Cemail%2CfirstName%2ClastName%2Ccompany%2Cunsubscribed%2CmarketingSuspended%2CemailInvalid%2CsfdcLeadId%2CsfdcContactId%2CcreatedAt%2CupdatedAt" {
		t.Fatalf("unexpected query string: %q", rawQuery)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["id"].(float64) != 123 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestLeadListEmailPaginatesResults(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			queries = append(queries, request.URL.RawQuery)
			if request.URL.Query().Get("nextPageToken") == "" {
				_ = json.NewEncoder(writer).Encode(map[string]any{
					"success":       true,
					"nextPageToken": "next-token",
					"result": []map[string]any{
						{"id": 1, "email": "user@example.com"},
					},
				})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result": []map[string]any{
					{"id": 2, "email": "second@example.com"},
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
	}, nil, "lead", "list", "--email", "user@example.com", "--limit", "5")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if len(queries) != 2 {
		t.Fatalf("expected two paginated requests, got %#v", queries)
	}
	if queries[0] != "batchSize=5&fields=id%2Cemail%2CfirstName%2ClastName%2Ccompany%2Cunsubscribed%2CmarketingSuspended%2CemailInvalid%2CsfdcLeadId%2CsfdcContactId%2CcreatedAt%2CupdatedAt&filterType=email&filterValues=user%40example.com" {
		t.Fatalf("unexpected first query: %q", queries[0])
	}
	if queries[1] != "batchSize=5&fields=id%2Cemail%2CfirstName%2ClastName%2Ccompany%2Cunsubscribed%2CmarketingSuspended%2CemailInvalid%2CsfdcLeadId%2CsfdcContactId%2CcreatedAt%2CupdatedAt&filterType=email&filterValues=user%40example.com&nextPageToken=next-token" {
		t.Fatalf("unexpected second query: %q", queries[1])
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected two records, got %#v", payload)
	}
}

func TestLeadDescribeDefaultsToDescribe2(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var path string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads/describe2.json":
			path = request.URL.Path
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"name": "email"}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "lead", "describe")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if path != "/rest/v1/leads/describe2.json" {
		t.Fatalf("unexpected path: %q", path)
	}
}

func TestLeadStaticListsPaginatesMemberships(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads/123/listMembership.json":
			queries = append(queries, request.URL.RawQuery)
			if request.URL.Query().Get("nextPageToken") == "" {
				_ = json.NewEncoder(writer).Encode(map[string]any{
					"success":       true,
					"nextPageToken": "next-token",
					"result":        []map[string]any{{"id": 10}},
				})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 11}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "lead", "static-lists", "123", "--limit", "5")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if len(queries) != 2 || queries[0] != "batchSize=5" || queries[1] != "batchSize=5&nextPageToken=next-token" {
		t.Fatalf("unexpected queries: %#v", queries)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected two memberships, got %#v", payload)
	}
}

func TestLeadProgramsUsesProgramFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var query string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads/123/programMembership.json":
			query = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"programId": 9}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, _, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "lead", "programs", "123", "--program-id", "9", "--program-id", "10")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if query != "filterType=programId&filterValues=9%2C10" {
		t.Fatalf("unexpected query: %q", query)
	}
}

func TestLeadSmartCampaignsFetchesMemberships(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads/123/smartCampaignMembership.json":
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"campaignId": 77}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "lead", "smart-campaigns", "123")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["campaignId"].(float64) != 77 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
