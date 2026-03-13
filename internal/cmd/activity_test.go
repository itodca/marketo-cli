package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestActivityTypesListsAvailableTypes(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/activities/types.json":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 1, "name": "Visit Web Page"}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "activity", "types")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["name"] != "Visit Web Page" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestActivityListUsesPagingTokenAndTypeFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var pagingQuery string
	var activityQueries []string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/activities/pagingtoken.json":
			pagingQuery = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{"nextPageToken": "start-token"})
		case "/rest/v1/activities.json":
			activityQueries = append(activityQueries, request.URL.RawQuery)
			if request.URL.Query().Get("nextPageToken") == "start-token" {
				_ = json.NewEncoder(writer).Encode(map[string]any{
					"success":       true,
					"nextPageToken": "next-token",
					"result":        []map[string]any{{"id": 1}},
				})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 2}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "activity", "list", "123", "--type-id", "1", "--type-id", "2", "--limit", "5")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if !strings.HasPrefix(pagingQuery, "sinceDatetime=") {
		t.Fatalf("expected sinceDatetime paging query, got %q", pagingQuery)
	}
	if len(activityQueries) != 2 {
		t.Fatalf("expected two activity requests, got %#v", activityQueries)
	}
	if activityQueries[0] != "activityTypeIds=1%2C2&batchSize=5&leadIds=123&nextPageToken=start-token" {
		t.Fatalf("unexpected first activity query: %q", activityQueries[0])
	}
	if activityQueries[1] != "activityTypeIds=1%2C2&batchSize=5&leadIds=123&nextPageToken=next-token" {
		t.Fatalf("unexpected second activity query: %q", activityQueries[1])
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected two activities, got %#v", payload)
	}
}

func TestActivityChangesUsesWatchLeadAndListFilters(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	writeTestConfig(t, homeDir)

	var pagingQuery string
	var changesQuery string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/activities/pagingtoken.json":
			pagingQuery = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{"nextPageToken": "start-token"})
		case "/rest/v1/activities/leadchanges.json":
			changesQuery = request.URL.RawQuery
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"activityTypeId": 13}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	exitCode, stdout, stderr := executeTestCommand(t, map[string]string{
		"MARKETO_REST_URL":     server.URL + "/rest",
		"MARKETO_IDENTITY_URL": server.URL + "/identity",
	}, nil, "activity", "changes", "--watch", "email", "--watch", "company", "--lead-id", "1", "--lead-id", "2", "--list-id", "99")

	if exitCode != 0 {
		t.Fatalf("expected success, got stderr %q", stderr)
	}
	if !strings.HasPrefix(pagingQuery, "sinceDatetime=") {
		t.Fatalf("expected sinceDatetime paging query, got %q", pagingQuery)
	}
	expectedQuery := "fields=email%2Ccompany&leadIds=1%2C2&listId=99&nextPageToken=start-token"
	if changesQuery != expectedQuery {
		t.Fatalf("unexpected lead changes query: %q", changesQuery)
	}

	var payload []map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout was not valid JSON: %v", err)
	}
	if len(payload) != 1 || payload[0]["activityTypeId"].(float64) != 13 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
