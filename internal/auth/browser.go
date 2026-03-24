package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	loginTimeout    = 5 * time.Minute
	pollInterval    = 5 * time.Second
	apiBase         = "https://api.sporkops.com/v1"
	browserAuthURL  = "https://sporkops.com/cli-auth"
)

type createSessionResponse struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	ExpiresIn  int    `json:"expires_in"`
	Interval   int    `json:"interval"`
}

type pollResponse struct {
	Status string `json:"status"`
	APIKey string `json:"api_key"`
}

// Login performs the RFC 8628-style device-code login flow.
// It creates a session, prints the user code, opens the browser,
// then polls until the user confirms or the timeout expires.
func Login() (string, error) {
	// Step 1: create a device auth session
	resp, err := http.Post(apiBase+"/cli/auth/sessions", "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("creating auth session: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status from auth server: %s", resp.Status)
	}

	var session createSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", fmt.Errorf("reading auth session response: %w", err)
	}

	// Step 2: print URL and user code; do not open the browser
	fmt.Printf("Visit this URL to authenticate:\n%s\n\nEnter this code: %s\n\nWaiting for confirmation...\n", browserAuthURL, session.UserCode)

	// Step 4: poll
	deadline := time.Now().Add(loginTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)

		apiKey, done, err := pollSession(session.DeviceCode)
		if err != nil {
			return "", err
		}
		if done {
			return apiKey, nil
		}
	}

	return "", fmt.Errorf("login timed out after %s — please run 'spork login' again", loginTimeout)
}

func pollSession(deviceCode string) (apiKey string, done bool, err error) {
	body, _ := json.Marshal(map[string]string{"device_code": deviceCode})
	resp, err := http.Post(apiBase+"/cli/auth/sessions/poll", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", false, fmt.Errorf("polling auth session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		return "", false, fmt.Errorf("session expired — please run 'spork login' again")
	}
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("unexpected status while polling: %s", resp.Status)
	}

	var pr pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return "", false, fmt.Errorf("reading poll response: %w", err)
	}

	if pr.Status == "complete" {
		return pr.APIKey, true, nil
	}
	return "", false, nil
}
