package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.sporkops.com/v1"

const (
	maxRetries    = 3
	baseDelay     = 500 * time.Millisecond
	maxRetryAfter = 60
)

// maxResponseBodySize is the maximum response body size (10MB) to prevent
// unbounded memory allocation from malicious or misbehaving servers.
const maxResponseBodySize = 10 * 1024 * 1024

// APIError represents an error response from the Spork API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// Monitor represents an uptime monitor, aligned with the REST API.
type Monitor struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Type            string            `json:"type,omitempty"`
	Target          string            `json:"target,omitempty"`
	Method          string            `json:"method,omitempty"`
	ExpectedStatus  int               `json:"expected_status,omitempty"`
	Interval        int               `json:"interval,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	Regions         []string          `json:"regions,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	Keyword         string            `json:"keyword,omitempty"`
	KeywordType     string            `json:"keyword_type,omitempty"`
	SSLWarnDays     int               `json:"ssl_warn_days,omitempty"`
	AlertChannelIDs []string          `json:"alert_channel_ids,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	Paused          *bool             `json:"paused,omitempty"`
	Status          string            `json:"status,omitempty"`
	LastCheckedAt   string            `json:"last_checked_at,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

// Account represents the authenticated user's account info.
type Account struct {
	UID              string    `json:"uid"`
	Email            string    `json:"email"`
	Plan             string    `json:"plan"`
	MonitorLimit     int       `json:"monitor_limit"`
	CheckIntervalS   int       `json:"check_interval_s"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	HasPaymentMethod bool      `json:"has_payment_method"`
}

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"` // only present on creation
	Prefix     string     `json:"prefix"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// MonitorResult represents a single uptime check result from the API.
type MonitorResult struct {
	ID             string `json:"id"`
	MonitorID      string `json:"monitor_id"`
	Status         string `json:"status"`
	StatusCode     int    `json:"status_code"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	Region         string `json:"region"`
	ErrorMessage   string `json:"error_message,omitempty"`
	CheckedAt      string `json:"checked_at"`
}

// AlertChannel represents an alert channel for notifications.
type AlertChannel struct {
	ID                 string            `json:"id,omitempty"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	Config             map[string]string `json:"config"`
	Verified           bool              `json:"verified,omitempty"`
	Secret             string            `json:"secret,omitempty"`
	LastDeliveryStatus string            `json:"last_delivery_status,omitempty"`
	LastDeliveryAt     string            `json:"last_delivery_at,omitempty"`
	CreatedAt          string            `json:"created_at,omitempty"`
	UpdatedAt          string            `json:"updated_at,omitempty"`
}

// dataEnvelope wraps the standard API response: {"data": ...}
type dataEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// listEnvelope wraps the standard API list response: {"data": [...], "meta": {...}}
type listEnvelope struct {
	Data json.RawMessage `json:"data"`
	Meta struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"meta"`
}

// apiErrorEnvelope matches the REST API error format: {"error": {"code": ..., "message": ...}}
type apiErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Version is set by the CLI at startup for the User-Agent header.
var Version = "dev"

// Client is an HTTP client for the Spork API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client with the given auth token.
// The base URL can be overridden via SPORK_API_URL.
func NewClient(token string) *Client {
	baseURL := os.Getenv("SPORK_API_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if !strings.HasPrefix(baseURL, "https://") {
		fmt.Fprintf(os.Stderr, "Warning: SPORK_API_URL must use https://, ignoring %q\n", baseURL)
		baseURL = defaultBaseURL
	}

	parsedBase, _ := url.Parse(baseURL)

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				if parsedBase != nil && req.URL.Host != parsedBase.Host {
					req.Header.Del("Authorization")
				}
				return nil
			},
		},
	}
}

func (c *Client) CreateMonitor(m *Monitor) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("POST", "/monitors", m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListMonitors() ([]Monitor, error) {
	var result []Monitor
	if err := c.doList("GET", "/monitors", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetMonitor(id string) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("GET", "/monitors/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateMonitor(id string, m *Monitor) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("PATCH", "/monitors/"+url.PathEscape(id), m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteMonitor(id string) error {
	return c.doRaw("DELETE", "/monitors/"+url.PathEscape(id), nil)
}

func (c *Client) GetMonitorResults(id string, limit int) ([]MonitorResult, error) {
	path := fmt.Sprintf("/monitors/%s/results?per_page=%d", url.PathEscape(id), limit)
	var result []MonitorResult
	if err := c.doList("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetAccount() (*Account, error) {
	var result Account
	if err := c.doSingle("GET", "/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateAPIKey(name string, expiresDays int) (*APIKey, error) {
	req := struct {
		Name          string `json:"name"`
		ExpiresInDays *int   `json:"expires_in_days,omitempty"`
	}{Name: name}
	if expiresDays > 0 {
		req.ExpiresInDays = &expiresDays
	}
	var result APIKey
	if err := c.doSingle("POST", "/api-keys", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListAPIKeys() ([]APIKey, error) {
	var result []APIKey
	if err := c.doList("GET", "/api-keys", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) DeleteAPIKey(id string) error {
	return c.doRaw("DELETE", "/api-keys/"+url.PathEscape(id), nil)
}

func (c *Client) ListAlertChannels() ([]AlertChannel, error) {
	var result []AlertChannel
	if err := c.doList("GET", "/alert-channels", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetAlertChannel(id string) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("GET", "/alert-channels/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateAlertChannel(ch *AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("POST", "/alert-channels", ch, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateAlertChannel(id string, ch *AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("PUT", "/alert-channels/"+url.PathEscape(id), ch, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteAlertChannel(id string) error {
	return c.doRaw("DELETE", "/alert-channels/"+url.PathEscape(id), nil)
}

func (c *Client) TestAlertChannel(id string) error {
	return c.doRaw("POST", "/alert-channels/"+url.PathEscape(id)+"/test", nil)
}

func (c *Client) doSingle(method, path string, body, result any) error {
	respBody, err := c.rawRequest(method, path, body)
	if err != nil {
		return err
	}
	if result != nil && len(respBody) > 0 {
		var envelope dataEnvelope
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return fmt.Errorf("parsing response envelope: %w", err)
		}
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("parsing response data: %w", err)
		}
	}
	return nil
}

func (c *Client) doList(method, path string, body any, result any) error {
	respBody, err := c.rawRequest(method, path, body)
	if err != nil {
		return err
	}
	if result != nil && len(respBody) > 0 {
		var envelope listEnvelope
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return fmt.Errorf("parsing response envelope: %w", err)
		}
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("parsing response data: %w", err)
		}
	}
	return nil
}

func (c *Client) doRaw(method, path string, body any) error {
	_, err := c.rawRequest(method, path, body)
	return err
}

// rawRequest performs the HTTP request with retry logic for transient errors.
// Response bodies are capped at maxResponseBodySize to prevent unbounded memory use.
func (c *Client) rawRequest(method, path string, body any) ([]byte, error) {
	var jsonBytes []byte
	if body != nil {
		var err error
		jsonBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
	}

	reqURL := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			time.Sleep(delay)
		}

		var reqBody io.Reader
		if jsonBytes != nil {
			reqBody = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequest(method, reqURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("User-Agent", "spork-cli/"+Version)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		limitedBody := io.LimitReader(resp.Body, maxResponseBodySize)
		respBody, err := io.ReadAll(limitedBody)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					if seconds > maxRetryAfter {
						seconds = maxRetryAfter
					}
					if seconds > 0 {
						time.Sleep(time.Duration(seconds) * time.Second)
					}
				}
			}
			lastErr = fmt.Errorf("API error (HTTP %d): transient error", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			msg := string(respBody)
			var errResp apiErrorEnvelope
			if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
				msg = errResp.Error.Message
			}
			return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
		}

		if resp.StatusCode == http.StatusNoContent {
			return nil, nil
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}
