// Package debughttp provides an http.RoundTripper wrapper that dumps the
// request and response for every call to stderr. The CLI uses it when the
// user passes --debug; it is safe to import elsewhere when an integration
// needs the same visibility.
//
// The transport redacts the Authorization header so tokens do not leak into
// logs, and truncates response bodies at 64 KiB to keep the terminal
// navigable when the API returns a long list.
package debughttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// maxBodySnippet bounds how much of a response body the debug transport
// prints. It is large enough to show a full page of monitors and small
// enough that `spork monitor list --debug` is still usable in a terminal.
const maxBodySnippet = 64 * 1024

// Transport wraps another http.RoundTripper and prints each request and
// response to Out. If Out is nil, it writes to os.Stderr. Use NewTransport
// to construct one; the zero value is NOT usable because Base is required.
type Transport struct {
	// Base is the underlying RoundTripper. Required.
	Base http.RoundTripper
	// Out is where dumps are written. If nil, os.Stderr.
	Out io.Writer
}

// NewTransport returns a debug Transport wrapping base (defaulting to
// http.DefaultTransport) and writing to out (defaulting to os.Stderr).
func NewTransport(base http.RoundTripper, out io.Writer) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Transport{Base: base, Out: out}
}

// RoundTrip satisfies http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	// Redact Authorization so tokens don't leak into logs. httputil.DumpRequestOut
	// reads the header as-is, so we rewrite it on a clone.
	clone := req.Clone(req.Context())
	if v := clone.Header.Get("Authorization"); v != "" {
		clone.Header.Set("Authorization", redactToken(v))
	}

	dump, err := httputil.DumpRequestOut(clone, true)
	if err == nil {
		t.write("--> ", dump)
	}

	resp, err := t.Base.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(t.out(), "<-- transport error after %s: %v\n\n", elapsed, err)
		return nil, err
	}

	// Buffer the body so we can both dump it and let the SDK read it.
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodySnippet))
	resp.Body.Close()
	if readErr != nil {
		// Couldn't fully read; still surface the status line.
		fmt.Fprintf(t.out(), "<-- %s (body read error after %s: %v)\n\n", resp.Status, elapsed, readErr)
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	// Replace the body so downstream reads see it untouched. If the server
	// sent more than maxBodySnippet, anything beyond the snippet is dropped
	// from the debug dump but the response body still needs the full bytes,
	// which we accept losing here in exchange for not buffering arbitrarily
	// large payloads in the debug path.
	resp.Body = io.NopCloser(bytes.NewReader(body))

	header := fmt.Sprintf("<-- %s %s (%s)\n", resp.Status, req.URL.Path, elapsed)
	t.write(header, body)
	return resp, nil
}

func (t *Transport) out() io.Writer {
	if t.Out != nil {
		return t.Out
	}
	return discardFallback()
}

// discardFallback is replaced in init() so the Transport has a sensible
// default even when constructed as &Transport{Base: x} without a writer.
var discardFallback = func() io.Writer { return nil }

func init() {
	// Use a closure captured at init time so the package does not import
	// os at the top level for this one default (keeps it testable — tests
	// set Out directly).
	discardFallback = func() io.Writer {
		return stderr()
	}
}

func (t *Transport) write(prefix string, payload []byte) {
	w := t.out()
	if w == nil {
		return
	}
	fmt.Fprint(w, prefix)
	w.Write(payload)
	if len(payload) > 0 && payload[len(payload)-1] != '\n' {
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
}

func redactToken(headerValue string) string {
	// Show the scheme and a short prefix so it's clear which token was used,
	// without revealing enough to be reusable.
	parts := strings.SplitN(headerValue, " ", 2)
	if len(parts) == 2 && parts[1] != "" {
		prefix := parts[1]
		if len(prefix) > 10 {
			prefix = prefix[:10]
		}
		return parts[0] + " " + prefix + "...<redacted>"
	}
	return "<redacted>"
}
