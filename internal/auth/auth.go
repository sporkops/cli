// Package auth handles CLI credential storage.
//
// Tokens live in the operating system's secure credential store — macOS
// Keychain, the Linux Secret Service (libsecret / GNOME Keyring / KWallet),
// or the Windows Credential Manager — via github.com/99designs/keyring.
//
// For headless / CI environments where no keyring daemon is reachable,
// the CLI falls back to the SPORK_API_KEY environment variable, which
// every command already honours. That is the same pattern `gh`, `stripe`,
// and `aws` use: interactive logins write to the OS keyring, CI pipelines
// pass credentials through environment.
//
// Plain-text credential files on disk are NOT supported. The historical
// ~/.config/spork/credentials.json is ignored; migration is intentionally
// a non-goal per the v1 scope decision.
package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

// ErrNoKeyring is returned when the OS keyring is unavailable (typical in
// headless CI environments). Callers should surface a helpful message
// pointing users at SPORK_API_KEY.
var ErrNoKeyring = errors.New("no OS keyring available")

// keyringService is the identifier written to every keyring backend. It
// doubles as the label users see in Keychain Access / seahorse, so keep
// it human-readable.
const keyringService = "sporkops-cli"

// tokenKey is the single item key under which we store the current API
// token. Future expansion (multiple profiles, org-specific tokens) can
// extend this; v1 stores exactly one token per user.
const tokenKey = "api-token"

// openKeyring opens the OS keyring, preferring the most-native backend
// available on the current platform. Order mirrors gh and stripe CLI
// defaults:
//
//   - macOS: Keychain
//   - Linux: Secret Service (libsecret), then KWallet
//   - Windows: Credential Manager
//
// We intentionally do NOT include the FileBackend — plain-text or
// passphrase-prompted file storage is a DX hazard (silent prompts in
// scripts, weaker than a real keyring, ambiguous location on disk).
// Callers without a keyring use SPORK_API_KEY.
func openKeyring() (keyring.Keyring, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,       // macOS
			keyring.SecretServiceBackend,  // Linux (GNOME Keyring, KWallet via libsecret bridge)
			keyring.KWalletBackend,        // Linux (KDE)
			keyring.WinCredBackend,        // Windows
		},
		// macOS Keychain specifics: use the default user keychain and a
		// stable service label so items group cleanly in Keychain Access.
		KeychainName: "login",
		// Linux Secret Service specifics: the collection is the default
		// "login" collection, unlocked on user login.
		LibSecretCollectionName: "login",
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoKeyring, err)
	}
	return ring, nil
}

// SaveToken stores the API token in the OS keyring. It replaces any
// existing token under the same key.
func SaveToken(token string) error {
	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("opening keyring: %w", err)
	}
	return ring.Set(keyring.Item{
		Key:         tokenKey,
		Data:        []byte(token),
		Label:       "Sporkops CLI API token",
		Description: "Used by the `spork` CLI to authenticate API requests.",
	})
}

// LoadToken returns the stored API token.
//
// Resolution order:
//
//  1. SPORK_API_KEY environment variable (always wins, so CI pipelines
//     and ephemeral containers can run without a keyring).
//  2. OS keyring entry written by `spork login`.
//
// Returns "" with no error when neither is populated, so RequireAuth can
// surface the standard "log in first" message instead of a raw keyring
// error.
func LoadToken() (string, error) {
	if token := os.Getenv("SPORK_API_KEY"); token != "" {
		return token, nil
	}
	ring, err := openKeyring()
	if err != nil {
		if errors.Is(err, ErrNoKeyring) {
			// No keyring, no env var → not logged in, same as "empty
			// keyring" for the caller's purposes.
			return "", nil
		}
		return "", fmt.Errorf("opening keyring: %w", err)
	}
	item, err := ring.Get(tokenKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("reading token from keyring: %w", err)
	}
	return string(item.Data), nil
}

// ClearToken removes the stored API token. It is a no-op when no token
// is stored (so `spork logout` on a fresh machine does not error).
func ClearToken() error {
	ring, err := openKeyring()
	if err != nil {
		if errors.Is(err, ErrNoKeyring) {
			return nil
		}
		return fmt.Errorf("opening keyring: %w", err)
	}
	if err := ring.Remove(tokenKey); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return nil
		}
		return fmt.Errorf("removing token from keyring: %w", err)
	}
	return nil
}
