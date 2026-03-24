package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type credentials struct {
	Token   string `json:"token"`
	SavedAt string `json:"saved_at"`
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "spork")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "spork")
}

// CredentialsPath returns the path to the credentials file.
func CredentialsPath() (string, error) {
	return filepath.Join(configDir(), "credentials.json"), nil
}

// SaveToken writes the auth token to the credentials file.
func SaveToken(token string) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	creds := credentials{
		Token:   token,
		SavedAt: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	return nil
}

// LoadToken reads the auth token from the credentials file.
// It also accepts a token via the SPORK_API_KEY environment variable.
func LoadToken() (string, error) {
	if token := os.Getenv("SPORK_API_KEY"); token != "" {
		return token, nil
	}

	path, err := CredentialsPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading credentials: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parsing credentials: %w", err)
	}

	return creds.Token, nil
}

// ClearToken deletes the credentials file.
func ClearToken() error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	return nil
}
