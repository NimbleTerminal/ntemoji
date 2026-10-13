package ntemoji

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestEmojiMetadataLookup(t *testing.T) {
	for emoji, want := range map[string]string{
		"👍":      "thumbs up",
		"👎":      "thumbs down",
		"🐱":      "cat face",
		"🐈":      "cat",
		"🚀":      "rocket",
		"":       "",
		"custom": "custom",
		"🧑‍🚀":    "astronaut",
		"👍🏽":     "thumbs up: medium skin tone",
		"🇺🇸":     "flag: United States",
		"❤️":     "red heart",
		"❤":      "red heart",
		"⚙️":     "gear",
		"🖥️":     "desktop computer",
	} {
		if got := EmojiName(emoji); got != want {
			t.Errorf("EmojiName(%q) = %q, want %q", emoji, got, want)
		}
	}
	for _, unknown := range []string{"", "custom", "not an emoji ❤️"} {
		if got := Tags(unknown); got != nil {
			t.Errorf("Tags(%q) = %v, want nil", unknown, got)
		}
	}
}

func TestTagsAreIndependent(t *testing.T) {
	// Exercise both CLDR keywords and a symbol with a name-only fallback.
	for _, emoji := range []string{EmojiThumbsUp, "₲"} {
		original := Tags(emoji)
		got := Tags(emoji)
		if len(got) == 0 {
			t.Fatalf("missing tags for %q", emoji)
		}
		got[0] = "caller mutation"
		got = append(got, "caller addition")
		if !slices.Equal(Tags(emoji), original) || MatchEmoji(emoji, "caller") {
			t.Fatalf("mutating Tags(%q) changed shared metadata", emoji)
		}
		if !MatchEmoji(emoji, original[0]) {
			t.Fatalf("original keyword stopped matching %q", emoji)
		}
	}
}

func TestBuiltInMetadataCoverage(t *testing.T) {
	check := func(emoji string) {
		t.Helper()
		name := EmojiName(emoji)
		if name == "" || name == emoji || len(Tags(emoji)) == 0 {
			t.Errorf("missing metadata for %q", emoji)
		}
		if !MatchEmoji(emoji, strings.ToUpper(name)) {
			t.Errorf("name %q does not match %q case-insensitively", name, emoji)
		}
	}
	// Check exported constants as well as the palettes: some constants are
	// intentionally absent from the default picker but still need labels.
	file, err := parser.ParseFile(token.NewFileSet(), "constants.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, value := range spec.Values {
			literal, ok := value.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Fatal("expected emoji constants to be string literals")
			}
			emoji, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			check(emoji)
		}
		return false
	})
	for _, palette := range DefaultPalettes {
		check(palette.Icon)
		for _, emoji := range palette.Emojis {
			check(emoji)
		}
	}
	for emoji, tags := range emojiTags {
		check(emoji)
		if !slices.Equal(Tags(emoji), tags) {
			t.Errorf("Tags(%q) does not match generated keywords", emoji)
		}
	}
}
