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

func TestTransport_EmitsResponseHeaders(t *testing.T) {
	// Locks in FIX #4: the response dump must include headers, not just
	// the status line and body. This is what makes --debug useful for
	// diagnosing rate-limit, auth-redirect, and X-Request-Id issues.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req_xyz")
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	client := &http.Client{Transport: NewTransport(http.DefaultTransport, &buf)}
	resp, err := client.Get(srv.URL + "/anything")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	out := buf.String()
	if !strings.Contains(out, "X-Request-Id: req_xyz") {
		t.Errorf("response dump should include X-Request-Id header, got:\n%s", out)
	}
	if !strings.Contains(out, "X-Ratelimit-Remaining: 42") {
		t.Errorf("response dump should include X-RateLimit-Remaining header, got:\n%s", out)
	}
}

func TestTransport_FullBodyReachesCaller(t *testing.T) {
	// Locks in FIX #2: even when the response is larger than
	// maxBodySnippet, the caller (SDK) must see the full body — the
	// truncation is a *display* concern, not a functional one.
	big := bytes.Repeat([]byte("X"), maxBodySnippet*2+123)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(big)
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	client := &http.Client{Transport: NewTransport(http.DefaultTransport, &buf)}
	resp, err := client.Get(srv.URL + "/big")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != len(big) {
		t.Fatalf("caller received %d bytes, expected full body of %d", len(body), len(big))
	}
	if !bytes.Equal(body, big) {
		t.Fatal("caller received a different payload than the server sent")
	}

	// The dump should note the truncation so users are not misled into
	// thinking the SDK saw a short body.
	out := buf.String()
	if !strings.Contains(out, "truncated from dump") {
		t.Errorf("dump should mention truncation for large body, got last 200 chars:\n%s", tail(out, 200))
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
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
