package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.sporkops.com/v1"

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
	// Enforce HTTPS to prevent sending auth tokens over plaintext
	if !strings.HasPrefix(baseURL, "https://") {
		fmt.Fprintf(os.Stderr, "Warning: SPORK_API_URL must use https://, ignoring %q\n", baseURL)
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateMonitor creates a new uptime monitor.
func (c *Client) CreateMonitor(m *Monitor) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("POST", "/monitors", m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListMonitors returns all monitors for the authenticated user.
func (c *Client) ListMonitors() ([]Monitor, error) {
	var result []Monitor
	if err := c.doList("GET", "/monitors", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMonitor returns a single monitor by ID.
func (c *Client) GetMonitor(id string) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("GET", "/monitors/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMonitor partially updates a monitor by ID.
func (c *Client) UpdateMonitor(id string, m *Monitor) (*Monitor, error) {
	var result Monitor
	if err := c.doSingle("PATCH", "/monitors/"+url.PathEscape(id), m, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMonitor deletes a monitor by ID.
func (c *Client) DeleteMonitor(id string) error {
	return c.doRaw("DELETE", "/monitors/"+url.PathEscape(id), nil)
}

// GetMonitorResults returns recent check results for a monitor.
func (c *Client) GetMonitorResults(id string, limit int) ([]MonitorResult, error) {
	path := fmt.Sprintf("/monitors/%s/results?per_page=%d", url.PathEscape(id), limit)
	var result []MonitorResult
	if err := c.doList("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAccount returns the authenticated user's account info.
func (c *Client) GetAccount() (*Account, error) {
	var result Account
	if err := c.doSingle("GET", "/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateAPIKey creates a new API key. expiresDays=0 means no expiry.
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

// ListAPIKeys returns all API keys for the authenticated user.
func (c *Client) ListAPIKeys() ([]APIKey, error) {
	var result []APIKey
	if err := c.doList("GET", "/api-keys", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteAPIKey deletes an API key by ID.
func (c *Client) DeleteAPIKey(id string) error {
	return c.doRaw("DELETE", "/api-keys/"+url.PathEscape(id), nil)
}

// ListAlertChannels returns all alert channels for the authenticated user.
func (c *Client) ListAlertChannels() ([]AlertChannel, error) {
	var result []AlertChannel
	if err := c.doList("GET", "/alert-channels", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAlertChannel returns a single alert channel by ID.
func (c *Client) GetAlertChannel(id string) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("GET", "/alert-channels/"+url.PathEscape(id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateAlertChannel creates a new alert channel.
func (c *Client) CreateAlertChannel(ch *AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("POST", "/alert-channels", ch, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateAlertChannel updates an alert channel by ID.
func (c *Client) UpdateAlertChannel(id string, ch *AlertChannel) (*AlertChannel, error) {
	var result AlertChannel
	if err := c.doSingle("PUT", "/alert-channels/"+url.PathEscape(id), ch, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAlertChannel deletes an alert channel by ID.
func (c *Client) DeleteAlertChannel(id string) error {
	return c.doRaw("DELETE", "/alert-channels/"+url.PathEscape(id), nil)
}

// TestAlertChannel sends a test notification to an alert channel.
func (c *Client) TestAlertChannel(id string) error {
	return c.doRaw("POST", "/alert-channels/"+url.PathEscape(id)+"/test", nil)
}

// doSingle performs a request and unwraps a single-item {data: ...} envelope.
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

// doList performs a request and unwraps a list {data: [...], meta: {...}} envelope.
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

// doRaw performs a request expecting no response body (e.g., DELETE → 204).
func (c *Client) doRaw(method, path string, body any) error {
	_, err := c.rawRequest(method, path, body)
	return err
}

// rawRequest performs the HTTP request and returns the raw response body.
// It handles error status codes and returns a parsed APIError when appropriate.
func (c *Client) rawRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
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
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Parse structured error: {"error": {"code": "...", "message": "..."}}
		msg := string(respBody)
		var errResp apiErrorEnvelope
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			msg = errResp.Error.Message
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	// 204 No Content
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return respBody, nil
}
