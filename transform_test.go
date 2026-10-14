package ntemoji

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestDemojiText(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		opts     []TextOption
	}{
		{"Go 🚀", "Go rocket", nil},
		{"Go🚀now👍👍!", "Go rocket now thumbs up thumbs up!", nil},
		{"a # $ 1 + ©", "a # $ 1 + ©", nil},
		{"©️ ❤️", "copyright red heart", nil},
		{"🧑🏽‍🚀", "astronaut: medium skin tone", nil},
		{"Go 🚀", "Go :rocket:", []TextOption{WithDelimiters(":", ":")}},
		{"🚀", "[cohete]", []TextOption{WithDelimiters("[", "]"), WithTextLocale("es")}},
		{"🚀\u200d🚀", "🚀\u200d🚀", nil},
		{"❤\ufe0e", "❤\ufe0e", nil},
		{"\xff🚀\xfe", "\xffrocket\xfe", nil},
	} {
		got, err := DemojiString(tc.in, tc.opts...)
		if err != nil || got != tc.want {
			t.Errorf("%q: got %q (%v), want %q", tc.in, got, err, tc.want)
		}
		var out bytes.Buffer
		err = Demoji(&out, iotest.OneByteReader(strings.NewReader(tc.in)), tc.opts...)
		if err != nil || out.String() != tc.want {
			t.Errorf("chunked %q: got %q (%v)", tc.in, out.String(), err)
		}
	}
}

func TestRemojiText(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		opts     []TextOption
	}{
		{"Go :rocket:!", "Go 🚀!", nil},
		{"rocket :unknown: :rocket::thumbs up:", "rocket :unknown: 🚀👍", nil},
		{":astronaut: medium skin tone:", "🧑🏽‍🚀", nil},
		{"[cohete]", "🚀", []TextOption{WithDelimiters("[", "]"), WithTextLocale("es")}},
		{"<emoji>rocket</emoji>", "🚀", []TextOption{WithDelimiters("<emoji>", "</emoji>")}},
		{"::rocket:", ":🚀", nil},
		{"\xff:rocket:\xfe", "\xff🚀\xfe", nil},
	} {
		got, err := RemojiString(tc.in, tc.opts...)
		if err != nil || got != tc.want {
			t.Errorf("%q: got %q (%v), want %q", tc.in, got, err, tc.want)
		}
		var out bytes.Buffer
		err = Remoji(&out, iotest.OneByteReader(strings.NewReader(tc.in)), tc.opts...)
		if err != nil || out.String() != tc.want {
			t.Errorf("chunked %q: got %q (%v)", tc.in, out.String(), err)
		}
	}

}

func TestDemojiAllSequences(t *testing.T) {
	// One-byte reads exercise every UTF-8, modifier, flag, and ZWJ boundary.
	for key, emoji := range canonicalEmoji {
		if name := EmojiName(key); name != key {
			var out strings.Builder
			err := Demoji(&out, iotest.OneByteReader(strings.NewReader(emoji)), WithDelimiters("[", "]"))
			if err != nil || out.String() != "["+name+"]" {
				t.Fatalf("%q: %q (%v)", emoji, out.String(), err)
			}
		}
	}
}

func TestRemojiAmbiguousNames(t *testing.T) {
	// Exercise every non-ambiguous name in one stream, avoiding per-token indexes.
	for _, locale := range Locales() {
		reverse := map[string]string{}
		for key, emoji := range canonicalEmoji {
			name := EmojiNameFor(key, locale)
			if name == key {
				continue
			}
			if old, ok := reverse[name]; ok && old != emoji {
				reverse[name] = ""
			} else if !ok {
				reverse[name] = emoji
			}
		}
		for _, delimiters := range [][2]string{{"[", "]"}, {"", ""}} {
			var input, want strings.Builder
			for name, emoji := range reverse {
				token := delimiters[0] + name + delimiters[1]
				input.WriteString(token + "\n")
				if emoji == "" {
					want.WriteString(token + "\n")
				} else {
					want.WriteString(emoji + "\n")
				}
			}
			got, err := RemojiString(input.String(), WithDelimiters(delimiters[0], delimiters[1]), WithTextLocale(locale))
			if err != nil || got != want.String() {
				t.Fatalf("reverse names failed for %s %v: %v", locale, delimiters, err)
			}
		}

	}
}

type brokenReader struct {
	text string
	err  error
}

func (r *brokenReader) Read(p []byte) (int, error) {
	n := copy(p, r.text)
	r.text = r.text[n:]
	return n, r.err
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return 0, nil }

func TestTransformErrorsAndBytes(t *testing.T) {
	boom := errors.New("read failed")
	for _, transform := range []func(io.Writer, io.Reader, ...TextOption) error{Demoji, Remoji} {
		for _, text := range []string{"", "partial", ":rock", ":rocket:", "🚀"} {
			var out bytes.Buffer
			if err := transform(&out, &brokenReader{text, boom}); !errors.Is(err, boom) {
				t.Fatalf("lost reader error for %q: %v", text, err)
			}
		}
		if err := transform(shortWriter{}, strings.NewReader("hello")); !errors.Is(err, io.ErrShortWrite) {
			t.Fatalf("short writer: %v", err)
		}
	}
	source := []byte("🚀")
	encoded, err := DemojiBytes(source, WithDelimiters(":", ":"))
	if err != nil || string(encoded) != ":rocket:" {
		t.Fatal(string(encoded), err)
	}
	decoded, err := RemojiBytes(encoded)
	if err != nil || !bytes.Equal(source, decoded) {
		t.Fatal(string(decoded), err)
	}
	decoded[0] = 'x'
	if string(source) != "🚀" {
		t.Fatal("aliased input")
	}
}

type onceErrorReader struct{ err error }

func (r *onceErrorReader) Read([]byte) (int, error) {
	err := r.err
	r.err = io.EOF
	return 0, err
}

func TestTransformOneShotReadError(t *testing.T) {
	boom := errors.New("one-shot read error")
	for _, transform := range []func(io.Writer, io.Reader, ...TextOption) error{Demoji, Remoji} {
		if err := transform(io.Discard, &onceErrorReader{boom}); !errors.Is(err, boom) {
			t.Fatalf("lost one-shot error: %v", err)
		}
	}
}

func TestRemojiPlainGreedyBoundaries(t *testing.T) {
	for _, tc := range []struct{ input, want, locale string }{
		{"fire engine fire rocket", "🚒 🔥 🚀", "en"},
		{"man astronaut man", "👨‍🚀 👨", "en"},
		{"astronaut: medium skin tone", "🧑🏽‍🚀", "en"},
		{"rocketship xrocket rocket_2 rocket2 rocket\u0301", "rocketship xrocket rocket_2 rocket2 rocket\u0301", "en"},
		{"rocketЖ Жrocket rocket\u200d", "rocketЖ Жrocket rocket\u200d", "en"},
		{"(rocket), fire!", "(🚀), 🔥!", "en"},
		{"fire engines", "🔥 engines", "en"},
		{"cohete", "🚀", "es"},
		{"fusée fusées", "🚀 fusées", "fr"},
		{"火箭 发射火箭", "🚀 发射火箭", "zh"},
		{"\xffrocket\xfe", "\xff🚀\xfe", "en"},
		{"ROCKET rocket", "ROCKET 🚀", "en"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			opts := []TextOption{WithDelimiters("", ""), WithTextLocale(tc.locale)}
			got, err := RemojiString(tc.input, opts...)
			if err != nil || got != tc.want {
				t.Fatalf("got %q (%v), want %q", got, err, tc.want)
			}
			var out bytes.Buffer
			err = Remoji(&out, iotest.OneByteReader(strings.NewReader(tc.input)), opts...)
			if err != nil || out.String() != tc.want {
				t.Fatalf("chunked: %q (%v)", out.String(), err)
			}
		})
	}
}

func TestRemojiMissingOneDelimiter(t *testing.T) {
	for _, tc := range []struct{ input, open, close, want string }{
		{"rocket] xrocket]", "", "]", "🚀 xrocket]"},
		{"[rocket [rocketship", "[", "", "🚀 [rocketship"},
	} {
		got, err := RemojiString(tc.input, WithDelimiters(tc.open, tc.close))
		if err != nil || got != tc.want {
			t.Fatal(got, err)
		}
	}
}

func TestRemojiPlainReadErrorsAndBufferBoundary(t *testing.T) {
	boom := errors.New("read failed")
	var out bytes.Buffer
	err := Remoji(&out, &brokenReader{"rocket", boom}, WithDelimiters("", ""))
	if !errors.Is(err, boom) || out.String() != "🚀" {
		t.Fatal(out.String(), err)
	}
	prefix := strings.Repeat("x ", 4093)
	got, err := RemojiString(prefix+"fire engine rocketé rocket", WithDelimiters("", ""))
	if err != nil || got != prefix+"🚒 rocketé 🚀" {
		t.Fatal("buffer-boundary conversion failed", err)
	}
}

// Check output before a reader supplies another chunk, without timing-based
// assertions or closing the stream to force a flush.
type observedReader struct {
	first          string
	beforeNextRead func()
}

func (r *observedReader) Read(p []byte) (int, error) {
	if r.first != "" {
		n := copy(p, r.first)
		r.first = r.first[n:]
		return n, nil
	}
	r.beforeNextRead()
	return 0, io.EOF
}

func TestDemojiEmitsCompletedControlsBeforeNextRead(t *testing.T) {
	for _, ending := range []string{"\n", "\r\n", "\t", "\x00", "\u2028", "\u2029", "\u200e"} {
		t.Run(ending, func(t *testing.T) {
			var out bytes.Buffer
			reader := &observedReader{first: "🚀" + ending, beforeNextRead: func() {
				if got := out.String(); got != "rocket"+ending {
					t.Errorf("before next read: %q", got)
				}
			}}
			if err := Demoji(&out, reader); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSplitGraphemeWaitsForExtensions(t *testing.T) {
	for _, input := range []string{"\r", "🚀", "a", "\u200d", "\u0600", "\xe2\x80"} {
		advance, token, err := splitGrapheme([]byte(input), false)
		if advance != 0 || token != nil || err != nil {
			t.Fatalf("premature cluster %q: %d %q %v", input, advance, token, err)
		}
	}
	advance, token, err := splitGrapheme([]byte("\r\n"), false)
	if advance != 2 || string(token) != "\r\n" || err != nil {
		t.Fatal(advance, token, err)
	}
}
