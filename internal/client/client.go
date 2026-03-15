package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/itodca/marketo-cli/internal/config"
)

const (
	DefaultTimeout              = 30 * time.Second
	DefaultMaxRetries           = 2
	DefaultRateLimitSleep       = 20 * time.Second
	DefaultTokenExpirySkew      = 60 * time.Second
	DefaultAuthResponseLifetime = 3600
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (err *APIError) Error() string {
	message := strings.TrimSpace(err.Message)
	if message == "" {
		message = "request failed"
	}
	if err.Code == "" {
		return message
	}
	return fmt.Sprintf("[%s] %s", err.Code, message)
}

type Client struct {
	Config          config.Config
	TokenCache      *TokenCache
	HTTPClient      HTTPClient
	Timeout         time.Duration
	MaxRetries      int
	RateLimitSleep  time.Duration
	TokenExpirySkew time.Duration
	Sleep           func(time.Duration)
	Now             func() time.Time

	token       string
	tokenExpiry time.Time
}

func New(cfg config.Config) *Client {
	return &Client{
		Config:          cfg,
		TokenCache:      NewTokenCache(""),
		HTTPClient:      http.DefaultClient,
		Timeout:         DefaultTimeout,
		MaxRetries:      DefaultMaxRetries,
		RateLimitSleep:  DefaultRateLimitSleep,
		TokenExpirySkew: DefaultTokenExpirySkew,
		Sleep:           time.Sleep,
		Now:             time.Now,
	}
}

func (client *Client) Get(path string, params map[string]any) (map[string]any, error) {
	return client.request(http.MethodGet, path, params, nil)
}

func (client *Client) Post(path string, params map[string]any, body map[string]any) (map[string]any, error) {
	return client.request(http.MethodPost, path, params, body)
}

func (client *Client) Delete(path string, params map[string]any, body map[string]any) (map[string]any, error) {
	return client.request(http.MethodDelete, path, params, body)
}

func (client *Client) GetAllPages(path string, params map[string]any, limit int, batchSize int) (map[string]any, error) {
	pageParams := cloneParams(params)
	if pageParams == nil {
		pageParams = map[string]any{}
	}

	if batchSize <= 0 && limit > 0 {
		batchSize = limit
	}
	if batchSize > 0 {
		if batchSize > 300 {
			batchSize = 300
		}
		pageParams["batchSize"] = batchSize
	}

	results := []any{}
	warnings := []any{}
	requestID := ""

	for {
		data, err := client.Get(path, pageParams)
		if err != nil {
			return nil, err
		}

		if requestID == "" {
			if value, ok := data["requestId"].(string); ok {
				requestID = value
			}
		}
		if batchWarnings, ok := data["warnings"].([]any); ok {
			warnings = append(warnings, batchWarnings...)
		}
		if batchResults, ok := data["result"].([]any); ok {
			results = append(results, batchResults...)
		}

		if limit > 0 && len(results) >= limit {
			results = results[:limit]
			break
		}

		nextPageToken, _ := data["nextPageToken"].(string)
		if nextPageToken == "" {
			break
		}

		pageParams["nextPageToken"] = nextPageToken
	}

	response := map[string]any{
		"success": true,
		"result":  results,
	}
	if requestID != "" {
		response["requestId"] = requestID
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	return response, nil
}

func (client *Client) GetAllOffsetPages(path string, params map[string]any, limit int, pageSize int) (map[string]any, error) {
	if pageSize <= 0 {
		pageSize = 200
	}
	if pageSize > 200 {
		pageSize = 200
	}

	pageParams := cloneParams(params)
	if pageParams == nil {
		pageParams = map[string]any{}
	}
	offset := intValue(pageParams["offset"], 0)

	results := []any{}
	warnings := []any{}
	requestID := ""

	for {
		remaining := 0
		if limit > 0 {
			remaining = limit - len(results)
			if remaining <= 0 {
				break
			}
		}

		currentPageSize := pageSize
		if limit > 0 && remaining < currentPageSize {
			currentPageSize = remaining
		}

		pageParams["offset"] = offset
		pageParams["maxReturn"] = currentPageSize

		data, err := client.Get(path, pageParams)
		if err != nil {
			return nil, err
		}

		if requestID == "" {
			if value, ok := data["requestId"].(string); ok {
				requestID = value
			}
		}
		if batchWarnings, ok := data["warnings"].([]any); ok {
			warnings = append(warnings, batchWarnings...)
		}
		pageResults, _ := data["result"].([]any)
		results = append(results, pageResults...)

		if len(pageResults) < currentPageSize {
			break
		}

		offset += len(pageResults)
	}

	response := map[string]any{
		"success": true,
		"result":  results,
	}
	if requestID != "" {
		response["requestId"] = requestID
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	return response, nil
}

func (client *Client) request(method, path string, params map[string]any, body map[string]any) (map[string]any, error) {
	attempts := 0
	maxAttempts := 1 + client.maxRetries()
	forceRefresh := false

	for attempts < maxAttempts {
		attempts++

		request, cancel, err := client.newAPIRequest(method, path, params, body, forceRefresh)
		if err != nil {
			return nil, &APIError{Code: "request_failed", Message: err.Error()}
		}

		response, err := client.httpClient().Do(request)
		if err != nil {
			cancel()
			return nil, &APIError{Code: "request_failed", Message: err.Error()}
		}

		payload, readErr := readResponseBody(response)
		cancel()
		if readErr != nil {
			return nil, &APIError{Code: "request_failed", Message: readErr.Error(), HTTPStatus: response.StatusCode}
		}

		if response.StatusCode == http.StatusUnauthorized {
			client.invalidateToken()
			if attempts < maxAttempts {
				forceRefresh = true
				continue
			}
		}

		if response.StatusCode >= http.StatusBadRequest {
			return nil, &APIError{
				Code:       strconv.Itoa(response.StatusCode),
				Message:    responseMessage(response, payload),
				HTTPStatus: response.StatusCode,
			}
		}

		data, err := decodeJSONObject(payload, response.StatusCode)
		if err != nil {
			return nil, err
		}

		if success, ok := data["success"].(bool); !ok || success {
			return data, nil
		}

		code, message := extractMarketoError(data)
		if isMarketoAuthError(code) {
			client.invalidateToken()
			if attempts < maxAttempts {
				forceRefresh = true
				continue
			}
		}
		if code == "606" && attempts < maxAttempts {
			client.sleep(client.rateLimitSleep())
			forceRefresh = false
			continue
		}

		return nil, &APIError{Code: code, Message: message, HTTPStatus: response.StatusCode}
	}

	return nil, &APIError{Code: "request_failed", Message: "Request exhausted retries"}
}

func (client *Client) newAPIRequest(method, path string, params map[string]any, body map[string]any, forceRefresh bool) (*http.Request, context.CancelFunc, error) {
	accessToken, err := client.getToken(forceRefresh)
	if err != nil {
		return nil, nil, err
	}

	requestURL, err := client.buildURL(path, params)
	if err != nil {
		return nil, nil, err
	}

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		requestBody = bytes.NewReader(encoded)
	}

	ctx, cancel := client.timeoutContext()
	request, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	request.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return request, cancel, nil
}

func (client *Client) getToken(forceRefresh bool) (string, error) {
	if !forceRefresh && client.tokenIsValid() {
		return client.token, nil
	}

	if !forceRefresh {
		cached, ok, err := client.tokenCache().Load(client.Config.Profile)
		if err != nil {
			return "", err
		}
		if ok {
			client.token = cached.AccessToken
			client.tokenExpiry = time.Unix(int64(cached.Expiry), 0)
			return client.token, nil
		}
	}

	token, expiry, err := client.fetchToken()
	if err != nil {
		return "", err
	}
	client.token = token
	client.tokenExpiry = expiry
	return token, nil
}

func (client *Client) fetchToken() (string, time.Time, error) {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", client.Config.ClientID)
	values.Set("client_secret", client.Config.ClientSecret)

	requestURL := strings.TrimRight(client.Config.IdentityURL, "/") + "/oauth/token?" + values.Encode()
	ctx, cancel := client.timeoutContext()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		cancel()
		return "", time.Time{}, err
	}

	response, err := client.httpClient().Do(request)
	if err != nil {
		cancel()
		return "", time.Time{}, &APIError{Code: "auth_request_failed", Message: err.Error()}
	}

	payload, readErr := readResponseBody(response)
	cancel()
	if readErr != nil {
		return "", time.Time{}, &APIError{Code: "auth_request_failed", Message: readErr.Error(), HTTPStatus: response.StatusCode}
	}

	if response.StatusCode >= http.StatusBadRequest {
		return "", time.Time{}, &APIError{
			Code:       "auth_request_failed",
			Message:    responseMessage(response, payload),
			HTTPStatus: response.StatusCode,
		}
	}

	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", time.Time{}, &APIError{Code: "auth_response_invalid", Message: "Identity response was not valid JSON", HTTPStatus: response.StatusCode}
	}

	token, _ := data["access_token"].(string)
	if token == "" {
		return "", time.Time{}, &APIError{Code: "auth_response_invalid", Message: "Identity response did not include an access token", HTTPStatus: response.StatusCode}
	}

	expiresIn := intValue(data["expires_in"], DefaultAuthResponseLifetime)
	if err := client.tokenCache().Save(client.Config.Profile, token, expiresIn); err != nil {
		return "", time.Time{}, err
	}

	return token, client.expiryFromDuration(expiresIn), nil
}

func (client *Client) buildURL(path string, params map[string]any) (string, error) {
	normalized := strings.TrimSpace(path)
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	if strings.HasPrefix(normalized, "/rest/") {
		normalized = strings.TrimPrefix(normalized, "/rest")
	}

	requestURL, err := url.Parse(strings.TrimRight(client.Config.RestURL, "/") + normalized)
	if err != nil {
		return "", err
	}

	if len(params) == 0 {
		return requestURL.String(), nil
	}

	query := requestURL.Query()
	for key, value := range params {
		appendQueryValue(query, key, value)
	}
	requestURL.RawQuery = query.Encode()
	return requestURL.String(), nil
}

func appendQueryValue(values url.Values, key string, value any) {
	if value == nil {
		return
	}

	switch typed := value.(type) {
	case string:
		values.Add(key, typed)
		return
	case []string:
		for _, item := range typed {
			values.Add(key, item)
		}
		return
	case []any:
		for _, item := range typed {
			appendQueryValue(values, key, item)
		}
		return
	}

	reflectValue := reflect.ValueOf(value)
	if reflectValue.IsValid() && (reflectValue.Kind() == reflect.Slice || reflectValue.Kind() == reflect.Array) {
		for index := 0; index < reflectValue.Len(); index++ {
			appendQueryValue(values, key, reflectValue.Index(index).Interface())
		}
		return
	}

	values.Add(key, fmt.Sprint(value))
}

func cloneParams(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func extractMarketoError(data map[string]any) (string, string) {
	errorsValue, ok := data["errors"].([]any)
	if !ok || len(errorsValue) == 0 {
		return "unknown_error", "Unknown Marketo error"
	}

	first, ok := errorsValue[0].(map[string]any)
	if !ok {
		return "unknown_error", "Unknown Marketo error"
	}

	code, _ := first["code"].(string)
	if code == "" {
		code = "unknown_error"
	}

	message, _ := first["message"].(string)
	if message == "" {
		message = "Unknown Marketo error"
	}

	return code, message
}

func decodeJSONObject(payload []byte, statusCode int) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, &APIError{
			Code:       strconv.Itoa(statusCode),
			Message:    "Response was not valid JSON",
			HTTPStatus: statusCode,
		}
	}
	return data, nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func responseMessage(response *http.Response, payload []byte) string {
	message := strings.TrimSpace(string(payload))
	if message != "" {
		return message
	}
	return response.Status
}

func isMarketoAuthError(code string) bool {
	return code == "601" || code == "602"
}

func intValue(value any, fallback int) int {
	switch typed := value.(type) {
	case nil:
		return fallback
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	case string:
		parsed, err := strconv.Atoi(typed)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func (client *Client) tokenIsValid() bool {
	return client.token != "" && client.now().Before(client.tokenExpiry)
}

func (client *Client) tokenCache() *TokenCache {
	if client.TokenCache == nil {
		client.TokenCache = NewTokenCache("")
	}
	return client.TokenCache
}

func (client *Client) invalidateToken() {
	client.token = ""
	client.tokenExpiry = time.Time{}
	_ = client.tokenCache().Delete(client.Config.Profile)
}

func (client *Client) httpClient() HTTPClient {
	if client.HTTPClient == nil {
		client.HTTPClient = http.DefaultClient
	}
	return client.HTTPClient
}

func (client *Client) timeoutContext() (context.Context, context.CancelFunc) {
	timeout := client.timeout()
	if timeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), timeout)
}

func (client *Client) maxRetries() int {
	if client.MaxRetries < 0 {
		return 0
	}
	return client.MaxRetries
}

func (client *Client) timeout() time.Duration {
	if client.Timeout <= 0 {
		return DefaultTimeout
	}
	return client.Timeout
}

func (client *Client) rateLimitSleep() time.Duration {
	if client.RateLimitSleep <= 0 {
		return DefaultRateLimitSleep
	}
	return client.RateLimitSleep
}

func (client *Client) sleep(duration time.Duration) {
	if client.Sleep != nil {
		client.Sleep(duration)
		return
	}
	time.Sleep(duration)
}

func (client *Client) now() time.Time {
	if client.Now != nil {
		return client.Now()
	}
	return time.Now()
}

func (client *Client) expiryFromDuration(expiresIn int) time.Time {
	skew := client.TokenExpirySkew
	if skew < 0 {
		skew = 0
	}
	lifetime := time.Duration(expiresIn) * time.Second
	if lifetime > skew {
		lifetime -= skew
	} else {
		lifetime = 0
	}
	return client.now().Add(lifetime)
}
