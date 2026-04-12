package debughttp

import (
	"io"
	"os"
)

// stderr returns os.Stderr. It exists as a package-level indirection so
// transport.go can use a default writer without importing os at the top
// level (keeps the small unit tests from having to juggle os.Stderr).
func stderr() io.Writer { return os.Stderr }
