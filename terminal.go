package ntemoji

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

func bidiControl(r rune) bool {
	return r == '\u061c' || r == '\u200e' || r == '\u200f' ||
		r >= '\u202a' && r <= '\u202e' || r >= '\u2066' && r <= '\u2069'
}

// TerminalSafe escapes terminal controls and Unicode bidi controls as visible
// \uXXXX text, preserving newlines, tabs, and emoji joiners/variation selectors.
// Invalid UTF-8 is replaced with U+FFFD. This is for terminal display, not HTML,
// shell quoting, serialization, or round trips. Escape-sequence payloads remain
// visible text; terminal commands are never interpreted or executed here.
func TerminalSafe(text string) string {
	var out strings.Builder
	for _, r := range text {
		if (unicode.IsControl(r) && r != '\n' && r != '\t') || bidiControl(r) || r == '\u2028' || r == '\u2029' {
			fmt.Fprintf(&out, "\\u%04x", r)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

// NewTerminalSafeWriter returns a streaming TerminalSafe filter. Call Close to
// flush any trailing incomplete UTF-8. Close never closes dst. Buffered state is
// at most three UTF-8 bytes; output is streamed without retaining the input.
func NewTerminalSafeWriter(dst io.Writer) io.WriteCloser {
	return &terminalSafeWriter{dst: dst}
}

type terminalSafeWriter struct {
	dst     io.Writer
	pending []byte
	err     error
	closed  bool
}

func (w *terminalSafeWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	if w.err != nil {
		return 0, w.err
	}
	// Work in bounded chunks even if the caller supplies a very large write.
	total := len(p)
	for len(p) > 0 {
		n := min(len(p), 4096)
		data := append(w.pending, p[:n]...)
		p = p[n:]
		complete := 0
		for complete < len(data) && utf8.FullRune(data[complete:]) {
			_, size := utf8.DecodeRune(data[complete:])
			complete += size
		}
		w.err = writeText(w.dst, TerminalSafe(string(data[:complete])))
		w.pending = append(w.pending[:0], data[complete:]...)
		if w.err != nil {
			return 0, w.err
		}
	}
	return total, nil
}

func (w *terminalSafeWriter) Close() error {
	if w.closed {
		return w.err
	}
	w.closed = true
	if w.err == nil && len(w.pending) > 0 {
		w.err = writeText(w.dst, TerminalSafe(string(w.pending)))
	}
	w.pending = nil
	return w.err
}
