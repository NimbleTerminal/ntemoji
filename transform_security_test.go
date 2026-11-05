package ntemoji

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

type forbiddenRead struct{ t *testing.T }

func (r forbiddenRead) Read([]byte) (int, error) {
	r.t.Fatal("read before delimiter validation")
	return 0, io.EOF
}

func TestDelimiterLimits(t *testing.T) {
	for _, convert := range []func(io.Writer, io.Reader, ...TextOption) error{Demoji, Remoji} {
		for _, delimiters := range [][2]string{{strings.Repeat("[", 129), ""}, {"", strings.Repeat("]", 129)}, {strings.Repeat("é", 64), "]"}} {
			var out bytes.Buffer
			err := convert(&out, forbiddenRead{t}, WithDelimiters(delimiters[0], delimiters[1]))
			if !errors.Is(err, ErrDelimiterTooLong) || out.Len() != 0 {
				t.Fatal(err, out.String())
			}
		}
		if err := convert(io.Discard, strings.NewReader(""), WithDelimiters(strings.Repeat("[", 64), strings.Repeat("]", 64))); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRawConversionPreservesControls(t *testing.T) {
	input := "\x1b]52;c;payload\a\u202e\xff"
	for _, convert := range []func(string, ...TextOption) (string, error){DemojiString, RemojiString} {
		got, err := convert(input)
		if err != nil || got != input {
			t.Fatalf("raw conversion: %q %v", got, err)
		}
	}
}

func TestDemojiGraphemeLimit(t *testing.T) {
	err := Demoji(io.Discard, strings.NewReader("a"+strings.Repeat("\u0301", 1<<19)))
	if err == nil {
		t.Fatal("accepted oversized grapheme")
	}
}

func FuzzConversionsStreaming(f *testing.F) {
	for _, seed := range []string{"", "🚀 :rocket:", "\x1b]52;c;payload\a\u202e", "\xff\xf0\x9f", "a\u0301🧑🏽‍🚀"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			t.Skip()
		}
		for _, convert := range []func(io.Writer, io.Reader, ...TextOption) error{Demoji, Remoji} {
			var whole, chunked bytes.Buffer
			err1 := convert(&whole, strings.NewReader(input))
			err2 := convert(&chunked, iotest.OneByteReader(strings.NewReader(input)))
			if (err1 == nil) != (err2 == nil) || !bytes.Equal(whole.Bytes(), chunked.Bytes()) {
				t.Fatalf("chunk-dependent conversion: %q / %q, %v / %v", whole.String(), chunked.String(), err1, err2)
			}
		}
	})
}
