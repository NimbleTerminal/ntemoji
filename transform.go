package ntemoji

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// TextOption configures Demoji and Remoji and their String/Bytes variants.
type TextOption func(*textConfig)

type textConfig struct{ before, after, locale string }

// WithDelimiters wraps descriptive names in Demoji and identifies names in
// Remoji. Demoji defaults to no delimiters; Remoji defaults to colons.
// Empty delimiters enable greedy plain-name matching in Remoji. An omitted
// delimiter is replaced by a Unicode word-boundary check. No escaping is performed.
func WithDelimiters(before, after string) TextOption {
	return func(c *textConfig) { c.before, c.after = before, after }
}

// WithTextLocale selects CLDR names using the same fallback as EmojiNameFor.
// Remoji recognizes names in that locale, including English fallback entries.
func WithTextLocale(locale string) TextOption {
	return func(c *textConfig) { c.locale = locale }
}

// ErrEmptyDelimiter is retained for source compatibility.
// Deprecated: Remoji now accepts empty delimiters and never returns this error.
var ErrEmptyDelimiter = errors.New("ntemoji: Remoji requires nonempty opening and closing delimiters")

// Demoji writes src to dst, replacing recognized whole emoji graphemes with
// descriptive names. Ordinary symbols and unknown sequences pass through.
// Text-default single symbols (such as © or ❤) require emoji presentation VS16.
// Plain names are separated from adjacent words/names by spaces. Delimited
// output preserves surrounding spacing. Text-presentation (VS15) is preserved.
// Memory is bounded to a grapheme (maximum 1 MiB); larger graphemes return a
// scanner error. Output is flushed at completed control graphemes, including
// line feeds. Neither stream is closed. Errors may leave partial output.
func Demoji(dst io.Writer, src io.Reader, opts ...TextOption) (err error) {
	output := bufio.NewWriter(dst)
	defer func() { err = errors.Join(err, output.Flush()) }()
	dst = output
	cfg := textConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	scanner.Split(splitGrapheme)
	var previous rune
	previousChanged := false
	plain := cfg.before == "" && cfg.after == ""
	for scanner.Scan() {
		original := scanner.Text()
		text := original
		changed := false
		key := metadataKey(original)
		if canonical, ok := canonicalEmoji[key]; ok && (original == canonical || original != key || utf8.RuneCountInString(key) > 1) {
			if name := EmojiNameFor(original, cfg.locale); name != original {
				text = cfg.before + name + cfg.after
				changed = true
			}
		}
		first, _ := utf8.DecodeRuneInString(text)
		if plain && (changed || previousChanged) && wordRune(previous) && wordRune(first) {
			if err := writeText(dst, " "); err != nil {
				return err
			}
		}
		if err := writeText(dst, text); err != nil {
			return err
		}
		if endsClosedGrapheme(original) {
			if err := output.Flush(); err != nil {
				return err
			}
		}
		previous, _ = utf8.DecodeLastRuneInString(text)
		previousChanged = changed
	}
	return scanner.Err()
}

func wordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) }

// endsClosedGrapheme recognizes GCB Control/LF using the segmenter's own
// property tables: only these force a break before an Extend character. CR
// must wait for a possible LF (GB3), even though it breaks before Extend.
func endsClosedGrapheme(text string) bool {
	r, _ := utf8.DecodeLastRuneInString(text)
	if r == '\r' {
		return false
	}
	_, rest, _, _ := uniseg.FirstGraphemeClusterInString(string(r)+"\u0300", -1)
	return rest != ""
}

// Keep the final extensible cluster until more bytes arrive: a modifier or ZWJ
// may still extend it. Incomplete UTF-8 must not create a false boundary.
func splitGrapheme(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return 0, nil, nil
	}
	available := data
	if !atEOF {
		for i := max(0, len(data)-utf8.UTFMax); i < len(data); i++ {
			if utf8.RuneStart(data[i]) && !utf8.FullRune(data[i:]) {
				available = data[:i]
				break
			}
		}
	}
	if len(available) == 0 {
		return 0, nil, nil
	}
	cluster, rest, _, _ := uniseg.FirstGraphemeCluster(available, -1)
	if len(rest) == 0 && !atEOF && !endsClosedGrapheme(string(cluster)) {
		return 0, nil, nil
	}
	return len(cluster), cluster, nil
}

// Remoji replaces exact, case-sensitive CLDR names with canonical
// emoji presentation sequences. Defaults to :name:. Unknown and ambiguous names
// remain unchanged; keywords and emoticons are not decoded. Longest matching
// tokens win, including names containing delimiter characters. This is not a
// lossless encoding: literal tokens may convert and presentation is normalized.
// WithDelimiters("", "") enables plain names: matching is left-to-right and
// greedy, using the longest complete name at each position. Missing delimiters
// require boundaries against Unicode letters, numbers, marks, connector
// punctuation and joiners. No case folding or linguistic word segmentation is
// performed. Plain mode can replace ordinary prose; it is not a reliable inverse
// of plain Demoji output.
// Memory is bounded by the metadata index and longest token, not input size.
// Neither stream is closed. Errors may leave partial output.
func Remoji(dst io.Writer, src io.Reader, opts ...TextOption) (err error) {
	cfg := textConfig{before: ":", after: ":"}
	for _, opt := range opts {
		opt(&cfg)
	}
	output := bufio.NewWriter(dst)
	defer func() { err = errors.Join(err, output.Flush()) }()
	dst = output
	index := remojiIndexFor(cfg)
	root, longest := index.root, index.longest
	reader := &textReader{Reader: bufio.NewReaderSize(src, longest+utf8.UTFMax)}
	var previous rune
	for {
		if reader.err != nil && reader.Buffered() == 0 {
			if reader.err == io.EOF {
				return nil
			}
			return reader.err
		}
		node, matched, replacement := root, 0, ""
		// An omitted opening delimiter requires the start of a whole name.
		if cfg.before != "" || !nameWordRune(previous) {
			for depth := 1; depth <= longest; depth++ {
				data := reader.peek(depth)
				if len(data) < depth {
					break
				}
				node = node.next[data[depth-1]]
				if node == nil {
					break
				}
				if node.terminal {
					literal := ""
					if node.value == "" {
						literal = string(data[:depth])
					}
					var next rune
					if cfg.after == "" {
						next, _ = reader.runeAt(depth)
					}
					if cfg.after != "" || !nameWordRune(next) {
						matched, replacement = depth, node.value
						// An ambiguous longest name shields shorter matches too.
						if replacement == "" {
							replacement = literal
						}
					}
				}
				if len(node.next) == 0 {
					break
				}
			}
		}
		if matched > 0 {
			data := reader.peek(matched)
			previous, _ = utf8.DecodeLastRune(data[:matched])
			if err := writeText(dst, replacement); err != nil {
				return err
			}
			_, _ = reader.Discard(matched)
			continue
		}
		// Consume one original rune, preserving invalid UTF-8 bytes verbatim.
		r, size := reader.runeAt(0)
		if size == 0 {
			continue
		}
		if err := writeText(dst, string(reader.peek(size))); err != nil {
			return err
		}
		_, _ = reader.Discard(size)
		previous = r
	}
}

// Bound the cache because callers may supply arbitrary delimiters. Indexes are
// immutable after construction and remain safe to use after eviction.
const remojiCacheLimit = 8

type remojiIndex struct {
	root    *textNode
	longest int
}
type cachedRemojiIndex struct {
	config textConfig
	index  *remojiIndex
}

var remojiCache struct {
	sync.Mutex
	entries []cachedRemojiIndex
}

func remojiIndexFor(cfg textConfig) *remojiIndex {
	cfg.locale = metadataLocale(cfg.locale)
	// Avoid retaining unusually large, caller-controlled delimiter strings.
	if len(cfg.before)+len(cfg.after) > 128 {
		return buildRemojiIndex(cfg)
	}
	remojiCache.Lock()
	defer remojiCache.Unlock()
	for i, entry := range remojiCache.entries {
		if entry.config == cfg {
			copy(remojiCache.entries[1:i+1], remojiCache.entries[:i])
			remojiCache.entries[0] = entry
			return entry.index
		}
	}
	index := buildRemojiIndex(cfg)
	remojiCache.entries = append([]cachedRemojiIndex{{cfg, index}}, remojiCache.entries...)
	if len(remojiCache.entries) > remojiCacheLimit {
		remojiCache.entries[remojiCacheLimit] = cachedRemojiIndex{}
		remojiCache.entries = remojiCache.entries[:remojiCacheLimit]
	}
	return index
}

func buildRemojiIndex(cfg textConfig) *remojiIndex {
	tokens := map[string]string{}
	for key, emoji := range canonicalEmoji {
		name := EmojiNameFor(key, cfg.locale)
		if name == key {
			continue
		}
		token := cfg.before + name + cfg.after
		if old, exists := tokens[token]; exists && old != emoji {
			tokens[token] = ""
		} else if !exists {
			tokens[token] = emoji
		}
	}
	root := &textNode{}
	longest := 1
	for token, emoji := range tokens {
		longest = max(longest, len(token))
		node := root
		for i := range len(token) {
			if node.next == nil {
				node.next = map[byte]*textNode{}
			}
			if node.next[token[i]] == nil {
				node.next[token[i]] = &textNode{}
			}
			node = node.next[token[i]]
		}
		node.value = emoji
		node.terminal = true
	}
	return &remojiIndex{root: root, longest: longest}
}

func nameWordRune(r rune) bool {
	return wordRune(r) || unicode.Is(unicode.Pc, r) || r == '\u200c' || r == '\u200d'
}

// Keep read errors until buffered bytes have been processed. bufio.Peek can
// otherwise consume a one-shot error while looking beyond a potential token.
type textReader struct {
	*bufio.Reader
	err error
}

func (r *textReader) peek(n int) []byte {
	if r.err != nil && n > r.Buffered() {
		n = r.Buffered()
	}
	data, err := r.Peek(n)
	if err != nil {
		r.err = err
	}
	return data
}

func (r *textReader) runeAt(offset int) (rune, int) {
	for size := 1; size <= utf8.UTFMax; size++ {
		data := r.peek(offset + size)
		if len(data) <= offset {
			return 0, 0
		}
		if utf8.FullRune(data[offset:]) || r.err != nil {
			return utf8.DecodeRune(data[offset:])
		}
	}
	return utf8.RuneError, 1
}

type textNode struct {
	terminal bool
	next     map[byte]*textNode
	value    string
}

func writeText(dst io.Writer, text string) error {
	n, err := io.WriteString(dst, text)
	if err == nil && n != len(text) {
		return io.ErrShortWrite
	}
	return err
}

// DemojiString is Demoji for a string. On error it returns any partial output.
func DemojiString(src string, opts ...TextOption) (string, error) {
	var out strings.Builder
	err := Demoji(&out, strings.NewReader(src), opts...)
	return out.String(), err
}

// RemojiString is Remoji for a string. On error it returns any partial output.
func RemojiString(src string, opts ...TextOption) (string, error) {
	var out strings.Builder
	err := Remoji(&out, strings.NewReader(src), opts...)
	return out.String(), err
}

// DemojiBytes is Demoji for bytes; it does not modify or alias src.
func DemojiBytes(src []byte, opts ...TextOption) ([]byte, error) {
	var out bytes.Buffer
	err := Demoji(&out, bytes.NewReader(src), opts...)
	return out.Bytes(), err
}

// RemojiBytes is Remoji for bytes; it does not modify or alias src.
func RemojiBytes(src []byte, opts ...TextOption) ([]byte, error) {
	var out bytes.Buffer
	err := Remoji(&out, bytes.NewReader(src), opts...)
	return out.Bytes(), err
}
