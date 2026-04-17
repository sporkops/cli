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
	"encoding/json"
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

	// Build a redacted clone for dumping. Authorization is redacted on the
	// header, and if the request body is JSON we walk it and redact values
	// at known-sensitive keys. The original req is untouched so the SDK
	// sends the real bytes on the wire.
	clone := req.Clone(req.Context())
	if v := clone.Header.Get("Authorization"); v != "" {
		clone.Header.Set("Authorization", redactToken(v))
	}
	if req.Body != nil && req.GetBody != nil {
		if body, err := req.GetBody(); err == nil {
			if raw, err := io.ReadAll(body); err == nil {
				_ = body.Close()
				redacted := redactJSONBody(raw)
				clone.Body = io.NopCloser(bytes.NewReader(redacted))
				clone.ContentLength = int64(len(redacted))
			}
		}
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

	// For the dump, replace the body with the redacted copy and let
	// DumpResponse do the formatting. Truncate only what we write to the
	// output stream.
	dumpBody := redactJSONBody(body)
	dumpResp := *resp
	dumpResp.Body = io.NopCloser(bytes.NewReader(dumpBody))
	dumpResp.ContentLength = int64(len(dumpBody))
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

// sensitiveJSONKeys is the set of JSON field names whose values are replaced
// with "<redacted>" before we write a request or response body to the debug
// stream. Match is case-insensitive and substring-based so nested variants
// ("api_key", "bot_token", "integration_key") all land in the net.
var sensitiveJSONKeys = []string{
	"api_key", "apikey",
	"token", "bot_token", "access_token", "refresh_token", "id_token",
	"secret", "webhook_secret", "signing_secret", "client_secret",
	"password", "passphrase",
	"integration_key", "routing_key",
	"authorization",
}

// redactJSONBody tries to parse body as JSON and replace values at known
// sensitive keys with "<redacted>". If parsing fails, the body is returned
// unchanged — the caller should expect that non-JSON bodies (binary blobs,
// form encoding) are logged as-is. The SDK only ever sends JSON, so this is
// the common path.
func redactJSONBody(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	// Cheap check: skip any non-JSON payload so we don't garble form posts.
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return body
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return body
	}
	redacted := redactValue(parsed)
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	// Don't escape < > & — the dump is consumed by humans on a terminal,
	// not a browser, so we want "<redacted>" to appear literally.
	enc.SetEscapeHTML(false)
	if err := enc.Encode(redacted); err != nil {
		return body
	}
	// Encoder appends a trailing newline; strip it so the dump matches the
	// shape of the original body.
	return bytes.TrimRight(out.Bytes(), "\n")
}

func redactValue(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(vv))
		for k, val := range vv {
			if isSensitiveKey(k) {
				out[k] = "<redacted>"
				continue
			}
			out[k] = redactValue(val)
		}
		return out
	case []any:
		out := make([]any, len(vv))
		for i, el := range vv {
			out[i] = redactValue(el)
		}
		return out
	default:
		return v
	}
}

func isSensitiveKey(k string) bool {
	lk := strings.ToLower(k)
	for _, needle := range sensitiveJSONKeys {
		if strings.Contains(lk, needle) {
			return true
		}
	}
	return false
}
