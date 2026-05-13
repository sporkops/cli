// Package config persists user-facing CLI preferences that don't belong
// in the OS keyring (which is reserved for secrets — see internal/auth).
//
// Right now the only stored preference is the active organization ID, set
// by `spork org use <id>` and consumed by every org-scoped command when
// neither --org nor SPORK_ORG_ID is supplied. The active org is not a
// secret — it's the same value that already flows through --org and the
// env var — so plaintext on disk is appropriate; this is the same posture
// the GitHub CLI takes for ~/.config/gh/hosts.yml (host preferences in
// plaintext, OAuth tokens in the OS keyring).
//
// Reading is fault-tolerant: a missing file returns the zero State, which
// callers treat as "no preferences yet" and fall through to env / auto-
// resolve. A malformed file returns a real error — silently ignoring
// corruption would hide problems users want to notice and fix.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// State is the on-disk shape. JSON for the same reasons gh / aws / stripe
// use it: human-readable, easy to grep, friction-free to extend with new
// preferences (last-used profile, default region, etc.) later.
type State struct {
	// ActiveOrganizationID is the org that org-scoped commands default to
	// when neither --org nor SPORK_ORG_ID is supplied. Set by
	// `spork org use <id>`; unset means "no preference — fall back to
	// SDK auto-resolve, which picks the only org for API-key callers
	// and errors with a candidate list for multi-org Firebase users."
	ActiveOrganizationID string `json:"active_organization_id,omitempty"`
}

// Path returns the absolute path to the config file. The lookup order:
//
//  1. SPORK_CONFIG_DIR (test isolation override; also useful for users
//     who keep their dotfiles under a non-standard location).
//  2. os.UserConfigDir() / spork / config.json — XDG_CONFIG_HOME on
//     Linux, ~/Library/Application Support on macOS, %AppData% on
//     Windows. Matches what `gh`, `stripe`, and friends do.
//
// Returns the path even when the file doesn't exist; callers handle
// fs.ErrNotExist.
func Path() (string, error) {
	if dir := os.Getenv("SPORK_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "config.json"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(base, "spork", "config.json"), nil
}

// Load reads the persisted state from disk. A missing file returns the
// zero State and a nil error — first-run is not an error. A malformed
// file is returned as a real error so users notice corruption rather
// than silently lose preferences.
func Load() (State, error) {
	path, err := Path()
	if err != nil {
		return State{}, err
	}
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var s State
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return State{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return s, nil
}

// Save writes the state to disk atomically. The parent directory is
// created with 0o700 so future additions (per-profile tokens, OAuth
// state) can land here without revisiting permissions — even though the
// file itself doesn't contain secrets today, the directory is shared.
//
// Atomicity matters because the CLI may be invoked concurrently (CI
// matrix jobs running `spork org use ...` in parallel against the same
// $HOME). os.CreateTemp + Rename gives us a single-step swap on POSIX;
// Windows promotes Rename to a replace as of Go 1.5.
func Save(s State) error {
	path, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "config.*.tmp")
	if err != nil {
		return fmt.Errorf("create tempfile: %w", err)
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("encode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("close tempfile: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// SetActiveOrganization is the most common write — it loads, sets the
// active org, and saves in one shot. Wrapping the common case keeps
// callers from having to round-trip the State struct when they only
// care about one field.
func SetActiveOrganization(orgID string) error {
	s, err := Load()
	if err != nil {
		return err
	}
	s.ActiveOrganizationID = orgID
	return Save(s)
}

// ClearActiveOrganization removes the persisted active org, falling back
// to SDK auto-resolve on the next call. Equivalent to `spork org use
// --clear` (or the documented "unset" subcommand if we ever add one).
func ClearActiveOrganization() error {
	return SetActiveOrganization("")
}
