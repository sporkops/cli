package monitor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sporkops/spork-go"
)

func TestWriteMonitorHCL_SkipsDefaults(t *testing.T) {
	// A minimal monitor (all defaults except the required fields) should
	// emit the shortest possible resource block.
	m := &spork.Monitor{
		ID:     "mon_abc123",
		Name:   "My API",
		Target: "https://api.example.com",
		Type:   "http",   // default — should be omitted
		Method: "GET",    // default — should be omitted
	}
	var buf bytes.Buffer
	if err := writeMonitorHCL(&buf, m, ""); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, `resource "sporkops_monitor" "my_api"`) {
		t.Errorf("expected resource header with derived name, got %q", out)
	}
	if !strings.Contains(out, `name              = "My API"`) {
		t.Errorf("expected name attribute, got %q", out)
	}
	if !strings.Contains(out, `target            = "https://api.example.com"`) {
		t.Errorf("expected target attribute, got %q", out)
	}
	if strings.Contains(out, "type") {
		t.Errorf("default http type should not be emitted, got %q", out)
	}
	if strings.Contains(out, "method") {
		t.Errorf("default GET method should not be emitted, got %q", out)
	}
	if strings.Contains(out, "interval") {
		t.Errorf("default 60s interval should not be emitted, got %q", out)
	}
}

func TestWriteMonitorHCL_EmitsNonDefaultAttributes(t *testing.T) {
	paused := true
	m := &spork.Monitor{
		ID:              "mon_xyz",
		Name:            "keyword check",
		Target:          "https://shop.example.com/cart",
		Type:            "keyword",
		Method:          "POST",
		ExpectedStatus:  201,
		Interval:        300,
		Timeout:         60,
		Regions:         []string{"us-central1", "europe-west1"},
		AlertChannelIDs: []string{"ach_1", "ach_2", "ach_3"},
		Tags:            []string{"prod", "checkout"},
		Paused:          &paused,
		Keyword:         "Add to cart",
		KeywordType:     "exists",
		Headers: map[string]string{
			"X-Test": "1",
			"X-Env":  "prod",
		},
	}
	var buf bytes.Buffer
	if err := writeMonitorHCL(&buf, m, "prod_cart"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	want := []string{
		`resource "sporkops_monitor" "prod_cart"`,
		`type              = "keyword"`,
		`method            = "POST"`,
		`expected_status   = 201`,
		`interval          = 300`,
		`timeout           = 60`,
		`regions           = ["us-central1", "europe-west1"]`,
		`alert_channel_ids = [`,
		`  "ach_1",`,
		`  "ach_2",`,
		`  "ach_3",`,
		`]`,
		`tags              = ["prod", "checkout"]`,
		`paused            = true`,
		`keyword           = "Add to cart"`,
		`keyword_type      = "exists"`,
		`headers           = {`,
		`  "X-Env" = "prod"`, // sorted alphabetically
		`  "X-Test" = "1"`,
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("expected substring %q in output:\n%s", w, out)
		}
	}
}

func TestHclIdentifier(t *testing.T) {
	tests := map[string]string{
		"My API":                "my_api",
		"  Spaces   ":           "spaces",
		"monitor.example.com":   "monitor_example_com",
		"100-alpha":             "m_100_alpha",
		"":                      "",
		"_leading":              "leading",
		"!!! weird !!!":         "weird",
	}
	for input, want := range tests {
		if got := hclIdentifier(input); got != want {
			t.Errorf("hclIdentifier(%q) = %q, want %q", input, got, want)
		}
	}
}
