package textcli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamingCommands(t *testing.T) {
	for _, tc := range []struct {
		name        string
		reverse     bool
		args        []string
		input, want string
	}{
		{"plain reverse", true, []string{"--open", "", "--close", ""}, "fire engine rocket", "🚒 🚀"},
		{"plain", false, nil, "Go 🚀\n", "Go rocket\n"},
		{"localized", false, []string{"--locale", "es", "--open", "[", "--close", "]"}, "🚀\n👍", "[cohete]\n[pulgar hacia arriba]"},
		{"reverse", true, nil, "Go :rocket:\n:unknown:", "Go 🚀\n:unknown:"},
		{"localized reverse", true, []string{"--locale", "fr", "--open", "[", "--close", "]", "-"}, "[fusée]", "🚀"},
		{"fallback", false, []string{"--locale", "unknown"}, "🚀", "rocket"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			code := Run(tc.reverse, tc.args, strings.NewReader(tc.input), &out, &diagnostics)
			if code != 0 || out.String() != tc.want || diagnostics.Len() != 0 {
				t.Fatalf("code=%d output=%q diagnostics=%q", code, out.String(), diagnostics.String())
			}
		})
	}
}
func TestFileHelpAndErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("🚀"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := Run(false, []string{path}, strings.NewReader("ignored"), &out, &diagnostics); code != 0 || out.String() != "rocket" {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"--help"}, 0}, {[]string{"--list-locales"}, 0}, {[]string{"--bad-flag"}, 2},
		{[]string{"a", "b"}, 2}, {[]string{"--open", ""}, 0}, {[]string{path + "missing"}, 1},
	} {
		out.Reset()
		diagnostics.Reset()
		if code := Run(true, tc.args, strings.NewReader(""), &out, &diagnostics); code != tc.code {
			t.Fatalf("%v: %d", tc.args, code)
		}
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestOutputError(t *testing.T) {
	var diagnostics bytes.Buffer
	if code := Run(false, nil, strings.NewReader("🚀"), brokenWriter{}, &diagnostics); code != 1 || !strings.Contains(diagnostics.String(), "write failed") {
		t.Fatal(code, diagnostics.String())
	}
}

func TestPflagSyntax(t *testing.T) {
	for _, args := range [][]string{
		{"--locale=es", "--open=[", "--close=]"},
		{"-", "--locale", "es", "--open", "[", "--close", "]"},
	} {
		var out, diagnostics bytes.Buffer
		if code := Run(false, args, strings.NewReader("🚀"), &out, &diagnostics); code != 0 || out.String() != "[cohete]" {
			t.Fatal(code, out.String(), diagnostics.String())
		}
	}
	for _, args := range [][]string{{"--help"}, {"-h"}} {
		var out, diagnostics bytes.Buffer
		if code := Run(false, args, strings.NewReader(""), &out, &diagnostics); code != 0 || !strings.Contains(diagnostics.String(), "--locale") || out.Len() != 0 {
			t.Fatal(code, out.String(), diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := Run(false, []string{"-locale", "es"}, strings.NewReader("🚀"), &out, &diagnostics); code != 2 || diagnostics.Len() == 0 || out.Len() != 0 {
		t.Fatal(code, out.String(), diagnostics.String())
	}
}

func TestVersionFlag(t *testing.T) {
	t.Cleanup(func(prev string) func() { return func() { version = prev } }(version))
	version = "v1.2.3"
	for _, tc := range []struct {
		reverse bool
		args    []string
		want    string
	}{
		{false, []string{"--version"}, "demoji v1.2.3\n"},
		{true, []string{"-V"}, "remoji v1.2.3\n"},
	} {
		var out, diagnostics bytes.Buffer
		if code := Run(tc.reverse, tc.args, strings.NewReader("🚀"), &out, &diagnostics); code != 0 || out.String() != tc.want || diagnostics.Len() != 0 {
			t.Fatalf("%v: code=%d output=%q diagnostics=%q", tc.args, code, out.String(), diagnostics.String())
		}
	}
}
