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

func configDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "spork"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".config", "spork"), nil
}

// CredentialsPath returns the path to the credentials file.
func CredentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// SaveToken writes the auth token to the credentials file.
// It refuses to write if the target path is a symlink and uses atomic
// rename to prevent TOCTOU race conditions.
func SaveToken(token string) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Refuse to write through a symlink to prevent symlink attacks.
	if fi, err := os.Lstat(path); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write credentials: %s is a symlink", path)
		}
	}

	creds := credentials{
		Token:   token,
		SavedAt: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	// Write to a temp file in the same directory and atomic-rename into place
	// to avoid partial writes and TOCTOU races.
	tmp, err := os.CreateTemp(dir, ".credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing credentials to temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file to credentials: %w", err)
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

	// Check file permissions and warn if more permissive than 0600.
	if fi, err := os.Stat(path); err == nil {
		perm := fi.Mode().Perm()
		if perm&0o177 != 0 {
			fmt.Fprintf(os.Stderr, "Warning: credentials file %s has permissions %04o, expected 0600. "+
				"Consider running: chmod 600 %s\n", path, perm, path)
		}
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
