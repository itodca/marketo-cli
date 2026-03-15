package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/itodca/marketo-cli/internal/config"
)

func testConfig(baseURL, profile string) config.Config {
	return config.Config{
		MunchkinID:   "123-ABC-456",
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RestURL:      baseURL + "/rest",
		IdentityURL:  baseURL + "/identity",
		Profile:      profile,
	}
}

func TestClientAuthenticatesAndCachesTokenInMemory(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	tokenCalls := 0
	apiCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch request.URL.Path {
		case "/identity/oauth/token":
			tokenCalls++
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			apiCalls++
			if request.Header.Get("Authorization") != "Bearer tok123" {
				http.Error(writer, "missing token", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": apiCalls}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	first, err := client.Get("/v1/leads.json", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	second, err := client.Get("/v1/leads.json", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if first["result"].([]any)[0].(map[string]any)["id"].(float64) != 1 {
		t.Fatalf("unexpected first payload: %#v", first)
	}
	if second["result"].([]any)[0].(map[string]any)["id"].(float64) != 2 {
		t.Fatalf("unexpected second payload: %#v", second)
	}
	if tokenCalls != 1 {
		t.Fatalf("expected one token request, got %d", tokenCalls)
	}
	if apiCalls != 2 {
		t.Fatalf("expected two API calls, got %d", apiCalls)
	}
}

func TestClientUsesProfileScopedTokenCache(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	first := New(testConfig(server.URL, "profile-a"))
	first.TokenCache = NewTokenCache(cacheDir)
	if _, err := first.Get("/v1/leads.json", nil); err != nil {
		t.Fatalf("first Get returned error: %v", err)
	}

	second := New(testConfig(server.URL, "profile-b"))
	second.TokenCache = NewTokenCache(cacheDir)
	if _, err := second.Get("/v1/leads.json", nil); err != nil {
		t.Fatalf("second Get returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(cacheDir, "token-profile-a.json")); err != nil {
		t.Fatalf("expected profile-a cache file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "token-profile-b.json")); err != nil {
		t.Fatalf("expected profile-b cache file: %v", err)
	}
}

func TestClientRefreshesTokenAfter401(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	tokenCalls := 0
	apiCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch request.URL.Path {
		case "/identity/oauth/token":
			tokenCalls++
			token := "stale-token"
			if tokenCalls > 1 {
				token = "fresh-token"
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": token, "expires_in": 3600})
		case "/rest/v1/leads.json":
			apiCalls++
			if request.Header.Get("Authorization") == "Bearer stale-token" {
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 1}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	result, err := client.Get("/v1/leads.json", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result["result"].([]any)[0].(map[string]any)["id"].(float64) != 1 {
		t.Fatalf("unexpected payload: %#v", result)
	}
	if tokenCalls != 2 {
		t.Fatalf("expected two token requests, got %d", tokenCalls)
	}
	if apiCalls != 2 {
		t.Fatalf("expected two API calls, got %d", apiCalls)
	}
}

func TestClientRefreshesStaleTokenWithoutContextCanceled(t *testing.T) {
	t.Parallel()

	tokenCalls := 0
	apiCalls := 0

	client := New(testConfig("https://example.com", "default"))
	client.TokenCache = NewTokenCache(t.TempDir())
	if err := client.TokenCache.Save("default", "stale-token", 3600); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	client.HTTPClient = scriptedHTTPClient(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			tokenCalls++
			return contextAwareResponse(request.Context(), http.StatusOK, `{"access_token":"fresh-token","expires_in":3600}`), nil
		case "/rest/v1/leads.json":
			apiCalls++
			if request.Header.Get("Authorization") == "Bearer stale-token" {
				return staticResponse(http.StatusUnauthorized, "Unauthorized"), nil
			}
			if request.Header.Get("Authorization") != "Bearer fresh-token" {
				return staticResponse(http.StatusUnauthorized, "Unauthorized"), nil
			}
			return contextAwareResponse(request.Context(), http.StatusOK, `{"success":true,"result":[{"id":1}]}`), nil
		default:
			return staticResponse(http.StatusNotFound, "not found"), nil
		}
	})

	result, err := client.Get("/v1/leads.json", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result["result"].([]any)[0].(map[string]any)["id"].(float64) != 1 {
		t.Fatalf("unexpected payload: %#v", result)
	}
	if tokenCalls != 1 {
		t.Fatalf("expected one token refresh, got %d", tokenCalls)
	}
	if apiCalls != 2 {
		t.Fatalf("expected two API calls, got %d", apiCalls)
	}
}

func TestClientRefreshesTokenAfterMarketoAuthError(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"601", "602"} {
		code := code
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			tokenCalls := 0
			apiCalls := 0

			client := New(testConfig("https://example.com", "default"))
			client.TokenCache = NewTokenCache(t.TempDir())
			if err := client.TokenCache.Save("default", "stale-token", 3600); err != nil {
				t.Fatalf("Save returned error: %v", err)
			}
			client.HTTPClient = scriptedHTTPClient(func(request *http.Request) (*http.Response, error) {
				switch request.URL.Path {
				case "/identity/oauth/token":
					tokenCalls++
					return contextAwareResponse(request.Context(), http.StatusOK, `{"access_token":"fresh-token","expires_in":3600}`), nil
				case "/rest/v1/leads.json":
					apiCalls++
					if request.Header.Get("Authorization") == "Bearer stale-token" {
						return contextAwareResponse(request.Context(), http.StatusOK, `{"success":false,"errors":[{"code":"`+code+`","message":"Access token invalid"}]}`), nil
					}
					if request.Header.Get("Authorization") != "Bearer fresh-token" {
						return staticResponse(http.StatusUnauthorized, "Unauthorized"), nil
					}
					return contextAwareResponse(request.Context(), http.StatusOK, `{"success":true,"result":[{"id":1}]}`), nil
				default:
					return staticResponse(http.StatusNotFound, "not found"), nil
				}
			})

			result, err := client.Get("/v1/leads.json", nil)
			if err != nil {
				t.Fatalf("Get returned error: %v", err)
			}

			if result["result"].([]any)[0].(map[string]any)["id"].(float64) != 1 {
				t.Fatalf("unexpected payload: %#v", result)
			}
			if tokenCalls != 1 {
				t.Fatalf("expected one token refresh, got %d", tokenCalls)
			}
			if apiCalls != 2 {
				t.Fatalf("expected two API calls, got %d", apiCalls)
			}

			cached, ok, err := client.TokenCache.Load("default")
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if !ok {
				t.Fatal("expected refreshed token to be cached")
			}
			if cached.AccessToken != "fresh-token" {
				t.Fatalf("expected fresh-token, got %s", cached.AccessToken)
			}
		})
	}
}

func TestClientRetriesRateLimitErrors(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	apiCalls := 0
	sleeps := []time.Duration{}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			apiCalls++
			if apiCalls == 1 {
				_ = json.NewEncoder(writer).Encode(map[string]any{
					"success": false,
					"errors":  []map[string]any{{"code": "606", "message": "Rate limit"}},
				})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{{"id": 1}}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())
	client.Sleep = func(duration time.Duration) {
		sleeps = append(sleeps, duration)
	}

	result, err := client.Get("/v1/leads.json", nil)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result["result"].([]any)[0].(map[string]any)["id"].(float64) != 1 {
		t.Fatalf("unexpected payload: %#v", result)
	}
	if len(sleeps) != 1 || sleeps[0] != DefaultRateLimitSleep {
		t.Fatalf("expected one rate-limit sleep, got %#v", sleeps)
	}
}

func TestClientDeleteSupportsQueryAndJSONBody(t *testing.T) {
	t.Parallel()

	var method string
	var rawQuery string
	var rawBody string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/lists/10/leads.json":
			method = request.Method
			rawQuery = request.URL.RawQuery
			body, _ := io.ReadAll(request.Body)
			rawBody = string(body)
			_ = json.NewEncoder(writer).Encode(map[string]any{"success": true, "result": []map[string]any{}})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	if _, err := client.Delete("/v1/lists/10/leads.json", map[string]any{"id": []int{1, 2}}, map[string]any{"input": []map[string]any{{"id": 1}, {"id": 2}}}); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if method != http.MethodDelete {
		t.Fatalf("expected DELETE request, got %s", method)
	}
	if rawQuery != "id=1&id=2" {
		t.Fatalf("unexpected query string: %q", rawQuery)
	}
	if !strings.Contains(rawBody, "\"input\":[{\"id\":1},{\"id\":2}]") {
		t.Fatalf("unexpected body: %q", rawBody)
	}
}

func TestClientGetAllPagesFollowsNextPageToken(t *testing.T) {
	t.Parallel()

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
					"requestId":     "abc",
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

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	result, err := client.GetAllPages("/v1/leads.json", map[string]any{"filterType": "email", "filterValues": "user@example.com"}, 0, 0)
	if err != nil {
		t.Fatalf("GetAllPages returned error: %v", err)
	}

	records, ok := result["result"].([]any)
	if !ok || len(records) != 2 {
		t.Fatalf("unexpected result payload: %#v", result)
	}
	if queries[0] != "batchSize=300&filterType=email&filterValues=user%40example.com" && queries[0] != "filterType=email&filterValues=user%40example.com" {
		t.Fatalf("unexpected first query: %q", queries[0])
	}
	if queries[1] != "filterType=email&filterValues=user%40example.com&nextPageToken=next-token" && queries[1] != "batchSize=300&filterType=email&filterValues=user%40example.com&nextPageToken=next-token" {
		t.Fatalf("unexpected second query: %q", queries[1])
	}
}

func TestClientGetAllOffsetPagesFollowsOffsetPagination(t *testing.T) {
	t.Parallel()

	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/asset/v1/programs.json":
			queries = append(queries, request.URL.RawQuery)
			if request.URL.Query().Get("offset") == "0" {
				_ = json.NewEncoder(writer).Encode(map[string]any{
					"success": true,
					"result":  []map[string]any{{"id": 1}, {"id": 2}},
				})
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"success": true,
				"result":  []map[string]any{{"id": 3}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	result, err := client.GetAllOffsetPages("/asset/v1/programs.json", nil, 3, 2)
	if err != nil {
		t.Fatalf("GetAllOffsetPages returned error: %v", err)
	}

	records, ok := result["result"].([]any)
	if !ok || len(records) != 3 {
		t.Fatalf("unexpected result payload: %#v", result)
	}
	if len(queries) != 2 {
		t.Fatalf("expected two offset requests, got %#v", queries)
	}
	if queries[0] != "maxReturn=2&offset=0" {
		t.Fatalf("unexpected first query: %q", queries[0])
	}
	if queries[1] != "maxReturn=1&offset=2" {
		t.Fatalf("unexpected second query: %q", queries[1])
	}
}

func TestClientReturnsHelpfulErrorForInvalidJSONResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/identity/oauth/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "tok123", "expires_in": 3600})
		case "/rest/v1/leads.json":
			writer.Header().Set("Content-Type", "text/plain")
			_, _ = writer.Write([]byte("not json"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := New(testConfig(server.URL, "default"))
	client.TokenCache = NewTokenCache(t.TempDir())

	_, err := client.Get("/v1/leads.json", nil)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}

	if err.Error() != "[200] Response was not valid JSON" {
		t.Fatalf("unexpected error: %v", err)
	}
}

type scriptedHTTPClient func(*http.Request) (*http.Response, error)

func (client scriptedHTTPClient) Do(request *http.Request) (*http.Response, error) {
	return client(request)
}

type contextAwareBody struct {
	ctx    context.Context
	data   []byte
	offset int
}

func (body *contextAwareBody) Read(buffer []byte) (int, error) {
	if err := body.ctx.Err(); err != nil {
		return 0, err
	}
	if body.offset >= len(body.data) {
		return 0, io.EOF
	}

	read := copy(buffer, body.data[body.offset:])
	body.offset += read
	return read, nil
}

func (body *contextAwareBody) Close() error {
	return nil
}

func contextAwareResponse(ctx context.Context, status int, payload string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body: &contextAwareBody{
			ctx:  ctx,
			data: []byte(payload),
		},
	}
}

func staticResponse(status int, payload string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(payload)),
	}
}
