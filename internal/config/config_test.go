package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The config package owns the on-disk shape that ` + "`spork org use`" + ` writes
// and the root command reads back at startup, so these tests pin the
// fault-tolerance properties callers depend on:
//
//   - Missing file → zero State, nil error (first-run is normal).
//   - Round-trip preserves ActiveOrganizationID.
//   - Save creates parent directories.
//   - Save is atomic (the temp+rename leaves no .tmp residue on success).
//   - Malformed JSON is reported as an error, not silently dropped.
//
// Each test uses SPORK_CONFIG_DIR + t.TempDir so it doesn't touch the
// user's real config and Go reclaims the directory on cleanup.

func withTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SPORK_CONFIG_DIR", dir)
	return dir
}

func TestPath_HonoursOverride(t *testing.T) {
	dir := withTempConfigDir(t)
	got, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	want := filepath.Join(dir, "config.json")
	if got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestLoad_MissingFileReturnsZeroState(t *testing.T) {
	withTempConfigDir(t)
	s, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if s.ActiveOrganizationID != "" {
		t.Errorf("ActiveOrganizationID = %q, want empty", s.ActiveOrganizationID)
	}
}

func TestSave_RoundTripsActiveOrg(t *testing.T) {
	withTempConfigDir(t)
	if err := Save(State{ActiveOrganizationID: "org_acme"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.ActiveOrganizationID != "org_acme" {
		t.Errorf("ActiveOrganizationID = %q, want %q", got.ActiveOrganizationID, "org_acme")
	}
}

func TestSave_CreatesParentDirectory(t *testing.T) {
	// SPORK_CONFIG_DIR may point at a path that doesn't exist yet
	// (first-run on a fresh machine). MkdirAll on Save must create it.
	tmp := t.TempDir()
	t.Setenv("SPORK_CONFIG_DIR", filepath.Join(tmp, "nested", "dir"))
	if err := Save(State{ActiveOrganizationID: "org_x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "nested", "dir", "config.json")); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestSave_LeavesNoTempFileOnSuccess(t *testing.T) {
	dir := withTempConfigDir(t)
	if err := Save(State{ActiveOrganizationID: "org_x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file %s should have been renamed away", e.Name())
		}
	}
}

func TestSetActiveOrganization_PreservesOtherFields(t *testing.T) {
	// SetActiveOrganization is the common-case helper; it loads → mutates
	// → saves. When State grows new fields, this round-trip must not
	// silently zero them out. Today there's only one field, so this test
	// pins the helper signature works even when the file already exists.
	withTempConfigDir(t)
	if err := Save(State{ActiveOrganizationID: "org_old"}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	if err := SetActiveOrganization("org_new"); err != nil {
		t.Fatalf("SetActiveOrganization: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.ActiveOrganizationID != "org_new" {
		t.Errorf("ActiveOrganizationID = %q, want %q", got.ActiveOrganizationID, "org_new")
	}
}

func TestClearActiveOrganization_RemovesValue(t *testing.T) {
	withTempConfigDir(t)
	if err := SetActiveOrganization("org_pinned"); err != nil {
		t.Fatalf("seed SetActiveOrganization: %v", err)
	}
	if err := ClearActiveOrganization(); err != nil {
		t.Fatalf("ClearActiveOrganization: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.ActiveOrganizationID != "" {
		t.Errorf("ActiveOrganizationID = %q, want empty after clear", got.ActiveOrganizationID)
	}
}

func TestLoad_MalformedFileReturnsError(t *testing.T) {
	// Silently ignoring corruption would hide a problem users want to
	// notice. Surface a real error so `spork org current` says "decode
	// failed" instead of pretending the preference vanished.
	dir := withTempConfigDir(t)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("not json"), 0o600); err != nil {
		t.Fatalf("write malformed file: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
