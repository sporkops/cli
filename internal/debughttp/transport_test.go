package debughttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransport_DumpsRequestAndResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req_abc")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":{"id":"mon_1"}}`))
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	tr := NewTransport(http.DefaultTransport, &buf)
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", srv.URL+"/monitors/mon_1", nil)
	req.Header.Set("Authorization", "Bearer sk_live_supersecret_dont_leak_me")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"data":{"id":"mon_1"}}` {
		t.Errorf("transport swallowed body: got %q", body)
	}

	out := buf.String()
	if !strings.Contains(out, "--> ") {
		t.Errorf("expected request dump marker, got %q", out)
	}
	if !strings.Contains(out, "<-- 200 OK") {
		t.Errorf("expected response dump marker, got %q", out)
	}
	if !strings.Contains(out, `"id":"mon_1"`) {
		t.Errorf("expected response body in dump, got %q", out)
	}
	if strings.Contains(out, "sk_live_supersecret_dont_leak_me") {
		t.Errorf("Authorization leaked into dump: %q", out)
	}
	if !strings.Contains(out, "Bearer sk_live_su") || !strings.Contains(out, "<redacted>") {
		t.Errorf("expected redacted token prefix, got %q", out)
	}
}

func TestRedactToken(t *testing.T) {
	tests := map[string]string{
		"Bearer sk_live_abc123def456":  "Bearer sk_live_ab...<redacted>",
		"Bearer abc":                   "Bearer abc...<redacted>",
		"":                             "<redacted>",
		"MalformedHeaderWithNoSpace":   "<redacted>",
	}
	for input, want := range tests {
		if got := redactToken(input); got != want {
			t.Errorf("redactToken(%q) = %q, want %q", input, got, want)
		}
	}
}
