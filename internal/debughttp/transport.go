// Package debughttp provides an http.RoundTripper wrapper that dumps the
// request and response for every call to stderr. The CLI uses it when the
// user passes --debug; it is safe to import elsewhere when an integration
// needs the same visibility.
//
// The transport redacts the Authorization header so tokens do not leak
// into logs. Response bodies are buffered in full so the SDK sees the same
// bytes it would without --debug, but only the first maxBodySnippet bytes
// are actually written to the dump — the terminal does not need a megabyte
// of JSON to be useful for debugging.
package debughttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

// maxBodySnippet bounds how much of a request/response body the debug
// transport writes to its output stream. The full body always reaches the
// SDK; this is a display cap only.
const maxBodySnippet = 64 * 1024

// Transport wraps another http.RoundTripper and prints each request and
// response to Out. If Out is nil, it writes to os.Stderr. Use NewTransport
// to construct one; the zero value is not usable because Base is required.
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
	// reads the header as-is, so we rewrite it on a clone. The clone shares
	// the Body reader with req — that's fine because DumpRequestOut copies
	// it into the dump without consuming the underlying source (it uses
	// GetBody() when available, which the stdlib sets for byte-slice
	// bodies, which is what the SDK produces).
	clone := req.Clone(req.Context())
	if v := clone.Header.Get("Authorization"); v != "" {
		clone.Header.Set("Authorization", redactToken(v))
	}

	if dump, err := httputil.DumpRequestOut(clone, true); err == nil {
		t.writeDump("--> ", dump)
	}

	resp, err := t.Base.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(t.out(), "<-- transport error after %s: %v\n\n", elapsed, err)
		return nil, err
	}

	// Buffer the entire body so the SDK sees the same bytes it would have
	// seen without --debug. Truncation happens only when we write to Out.
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		// Surface the read error upstream — the SDK will report a more
		// accurate diagnostic than we can synthesize here.
		fmt.Fprintf(t.out(), "<-- %s (body read error after %s: %v)\n\n", resp.Status, elapsed, readErr)
		return nil, readErr
	}
	if closeErr != nil {
		// Close errors are typically harmless (e.g., already closed), but
		// worth surfacing when someone has debugging on.
		fmt.Fprintf(t.out(), "(warning: closing response body returned: %v)\n", closeErr)
	}

	// Reinstate the body so the SDK's JSON parser sees it normally.
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// For the dump, replace the body with a reader over our buffered copy
	// and let DumpResponse do the formatting. Truncate only what we write
	// to the output stream.
	dumpResp := *resp
	dumpResp.Body = io.NopCloser(bytes.NewReader(body))
	if dump, err := httputil.DumpResponse(&dumpResp, true); err == nil {
		header := fmt.Sprintf("<-- %s (%s, %d bytes)\n", resp.Status, elapsed, len(body))
		t.writeHeader(header)
		t.writeDump("", dump)
	}

	return resp, nil
}

func (t *Transport) out() io.Writer {
	if t.Out != nil {
		return t.Out
	}
	return os.Stderr
}

func (t *Transport) writeHeader(header string) {
	fmt.Fprint(t.out(), header)
}

// writeDump writes up to maxBodySnippet bytes of payload to Out, preceded
// by prefix. It guarantees the output ends with a blank line so successive
// dumps are visually separated even when the dumped content does not end
// in a newline.
func (t *Transport) writeDump(prefix string, payload []byte) {
	w := t.out()
	fmt.Fprint(w, prefix)
	if len(payload) > maxBodySnippet {
		w.Write(payload[:maxBodySnippet])
		fmt.Fprintf(w, "\n... [%d bytes truncated from dump; full body delivered to SDK]\n", len(payload)-maxBodySnippet)
	} else {
		w.Write(payload)
		if len(payload) > 0 && payload[len(payload)-1] != '\n' {
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintln(w)
}

// redactToken replaces the value of an Authorization header with a short
// prefix + "<redacted>" so log readers can tell which scheme and which
// token (roughly) was used, without leaving the full secret in plain text.
func redactToken(headerValue string) string {
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
