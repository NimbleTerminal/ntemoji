package ntemoji

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

func assertTerminalSafe(t *testing.T, s string) {
	t.Helper()
	if !utf8.ValidString(s) {
		t.Fatal("invalid UTF-8")
	}
	for _, r := range s {
		if (unicode.IsControl(r) && r != '\n' && r != '\t') || bidiControl(r) || r == '\u2028' || r == '\u2029' {
			t.Fatalf("unsafe rune %U", r)
		}
	}
}

func TestTerminalSafe(t *testing.T) {
	input := "\x1b]52;c;payload\a\x1b[2J\r\b\x00\x7f\u009b\u202e\u2066\xff\n\t🧑🏽‍🚀"
	want := "\\u001b]52;c;payload\\u0007\\u001b[2J\\u000d\\u0008\\u0000\\u007f\\u009b\\u202e\\u2066�\n\t🧑🏽‍🚀"
	if got := TerminalSafe(input); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	for split := 0; split <= len(input); split++ {
		var out bytes.Buffer
		w := NewTerminalSafeWriter(&out)
		for _, part := range []string{input[:split], input[split:]} {
			if n, err := io.WriteString(w, part); n != len(part) || err != nil {
				t.Fatal(n, err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if out.String() != want {
			t.Fatalf("split %d: %q", split, out.String())
		}
		if _, err := w.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	}
}

type shortTerminalWriter struct{}

func (shortTerminalWriter) Write(p []byte) (int, error) { return 0, nil }

func TestTerminalSafeWriterFailure(t *testing.T) {
	w := NewTerminalSafeWriter(shortTerminalWriter{})
	if _, err := w.Write([]byte("x")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
	if err := w.Close(); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
	w = NewTerminalSafeWriter(shortTerminalWriter{})
	if _, err := w.Write([]byte{0xf0}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}

func TestWidgetUntrustedText(t *testing.T) {
	payload := "\x1b]52;c;payload\a\x1b[2J\u202e\r\n🚀"
	// Exercise direct exported-field mutation as well as options.
	for _, list := range []bool{false, true} {
		m := New(WithListMode(list), WithPresets([]string{payload}), WithPalettes([]Palette{{Name: payload, Icon: payload, Emojis: []string{payload}}}))
		m.Presets[0] = payload
		m.Palettes[0].Name = payload
		for _, focus := range []Focus{FocusPresets, FocusGrid, FocusCategories} {
			m.Focus = focus
			view := m.View().Content
			if strings.Contains(view, "]52;") || strings.Contains(view, "\x1b[2J") || strings.ContainsRune(view, '\u202e') {
				t.Fatalf("unsafe view: %q", view)
			}
		}
	}
	s := NewSwatchPicker(payload, payload)
	if got := s.SwatchView(); got != " 🚀 ▼" {
		t.Fatalf("swatch %q", got)
	}
	ti := NewTextInput()
	ti.Prompt = payload
	ti.Placeholder = payload
	ti.SetValue(payload)
	if strings.ContainsRune(ti.Value(), '\u202e') {
		t.Fatal("bidi in input")
	}
	ti.Focus()
	ti, _ = ti.Update(tea.PasteMsg{Content: payload})
	if strings.Contains(ti.View(), "]52;") {
		t.Fatal("escape payload in input")
	}
}

func FuzzTerminalSafeWriter(f *testing.F) {
	for _, seed := range []string{"", "🧑🏽‍🚀", "\x1b]52;c;payload\a\u202e", "\xff\xf0\x9f"} {
		f.Add(seed, uint16(1))
	}
	f.Fuzz(func(t *testing.T, input string, chunk uint16) {
		if len(input) > 16384 {
			t.Skip()
		}
		var out bytes.Buffer
		w := NewTerminalSafeWriter(&out)
		size := int(chunk)%64 + 1
		for start := 0; start < len(input); start += size {
			if _, err := io.WriteString(w, input[start:min(start+size, len(input))]); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if out.String() != TerminalSafe(input) {
			t.Fatalf("chunk-dependent output: %q", out.String())
		}
		assertTerminalSafe(t, out.String())
	})
}

func TestSearchInputLimit(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue(strings.Repeat("🚀", MaxSearchBytes))
	if got := ti.Value(); len(got) > MaxSearchBytes || !utf8.ValidString(got) {
		t.Fatal("unbounded or invalid input")
	}
	ti.Focus()
	ti, _ = ti.Update(tea.PasteMsg{Content: strings.Repeat("a", MaxSearchBytes)})
	if len(ti.Value()) > MaxSearchBytes {
		t.Fatal("paste exceeded limit")
	}
	ti.SetValue(strings.Repeat("a", MaxSearchBytes-1))
	ti, _ = ti.Update(tea.PasteMsg{Content: "🚀"})
	if len(ti.Value()) != MaxSearchBytes-1 {
		t.Fatal("split emoji at search limit")
	}
}

func TestSafeAccessibilitySnapshotsPreserveRawValues(t *testing.T) {
	payload := "\x1b]52;c;payload\a\u202e🚀"
	m := New(WithInitialEmoji(payload), WithPresets([]string{payload}))
	if m.Value() != payload {
		t.Fatal("raw selection changed")
	}
	state := m.AccessibilityState()
	for _, text := range []string{state.FocusedEmoji, state.FocusedName, state.SelectedEmoji, state.SelectedName, state.Category} {
		assertTerminalSafe(t, text)
		if strings.Contains(text, "]52;") {
			t.Fatal("OSC payload in snapshot")
		}
	}
	s := NewSwatchPicker(payload, payload)
	if s.Emoji() != payload {
		t.Fatal("raw swatch value changed")
	}
	snapshot := s.AccessibilityState()
	for _, text := range []string{snapshot.Label, snapshot.Emoji, snapshot.Name} {
		assertTerminalSafe(t, text)
	}
}

func TestTerminalSafeWriterLargeWrite(t *testing.T) {
	input := strings.Repeat("a", 4095) + "🧑🏽‍🚀" + strings.Repeat("\u202e\x1b[2J", 10000) + "\xf0\x9f"
	var out bytes.Buffer
	w := NewTerminalSafeWriter(&out)
	if n, err := io.WriteString(w, input); n != len(input) || err != nil {
		t.Fatal(n, err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if out.String() != TerminalSafe(input) {
		t.Fatal("large write changed UTF-8 handling")
	}
}

func FuzzSingleLine(f *testing.F) {
	for _, seed := range []string{"🚀", "\x1b]52;c;payload\a\u202e", "\xff\xf0\x9f", "\u009b2J\r\n\t"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 16384 {
			t.Skip()
		}
		result := singleLine(input)
		assertTerminalSafe(t, result)
		if strings.ContainsAny(result, "\r\n\t") {
			t.Fatal("multiline widget content")
		}
	})
}
