package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedMetadataIsCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	got, err := generate(filepath.Join(root, "internal", "cldr"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "metadata_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(normalizeGeneratedSource(t, got), normalizeGeneratedSource(t, want)) {
		t.Fatal("metadata is stale; run go generate ./...")
	}
}

func TestParseAnnotations(t *testing.T) {
	entries := map[string]entry{}
	err := parseLocale([]byte(`<ldml><identity><language type="en"/></identity><annotations>
<annotation cp="❤️" type="tts">red heart</annotation>
<annotation cp="❤"> heart | love | heart </annotation>
<annotation cp="🧑‍🚀" type="tts">astronaut</annotation>
<annotation cp="🧑‍🚀">space | rocket</annotation>
<annotation cp="&amp;" type="tts">ampersand</annotation>
</annotations></ldml>`), entries, "en")
	if err != nil {
		t.Fatal(err)
	}
	if entries["❤"].name != "red heart" || len(entries["❤"].tags) != 2 || entries["❤"].tags[1] != "love" {
		t.Fatalf("normalization or keyword parsing failed: %+v", entries)
	}
	if entries["🧑‍🚀"].name != "astronaut" || entries["&"].name != "ampersand" {
		t.Fatal("lost sequence or XML entity")
	}
	for _, content := range []string{
		`<ldml><identity><language type="fr"/></identity><annotations><annotation cp="x" type="tts">x</annotation></annotations></ldml>`,
		`<ldml><identity><language type="en"/></identity><annotations><annotation cp="❤" type="tts">↑↑↑</annotation></annotations></ldml>`,
		`<ldml><identity><language type="en"/></identity><annotations><annotation cp="❤" type="tts">conflicting name</annotation></annotations></ldml>`,
	} {
		if err := parseLocale([]byte(content), entries, "en"); err == nil {
			t.Errorf("accepted invalid data: %s", content)
		}
	}
}

func TestRejectChangedSnapshot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "en.xml"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sources.json"), []byte(`[{"file":"en.xml","locale":"en","sha256":"wrong"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(dir); err == nil {
		t.Fatal("accepted an unverified source snapshot")
	}
}

func TestGeneratedTranslationsAreCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	got, err := generateTranslations(filepath.Join(root, "internal", "cldr"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "locales_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(normalizeGeneratedSource(t, got), normalizeGeneratedSource(t, want)) {
		t.Fatal("translations stale; run task generate-cldr")
	}
}

func TestRejectChangedTranslationSnapshot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sources.json"), []byte(`[{"file":"fr.xml","locale":"fr","sha256":"wrong"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fr.xml"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := generateTranslations(dir); err == nil {
		t.Fatal("accepted changed translation snapshot")
	}
}

func TestGeneratedEmojiSequencesAreCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	got, err := generateEmoji(filepath.Join(root, "internal", "cldr"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "emoji_sequences_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(normalizeGeneratedSource(t, got), normalizeGeneratedSource(t, want)) {
		t.Fatal("emoji sequences stale; run task generate-cldr")
	}
}

func TestLocaleManifestRejectsUnsupportedTags(t *testing.T) {
	for _, locale := range []string{"pt-BR", "zh_Hant", "en;panic()", "", "EN"} {
		dir := t.TempDir()
		manifest := fmt.Sprintf(`[{"file":"test.xml","locale":%q,"sha256":"unused"}]`, locale)
		if err := os.WriteFile(filepath.Join(dir, "sources.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := loadLocales(dir)
		if err == nil || !strings.Contains(err.Error(), "invalid or duplicate locale source") {
			t.Fatalf("locale %q: %v", locale, err)
		}
	}
}

// Go releases can disagree about which Unicode characters are printable in
// quoted strings. Compare source with ASCII string literals and formatting
// produced by the same toolchain, preserving declarations, data, and comments.
func normalizeGeneratedSource(t *testing.T, source []byte) []byte {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatal(err)
		}
		literal.Value = strconv.QuoteToASCII(value)
		return true
	})
	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestNormalizeGeneratedSource(t *testing.T) {
	source := []byte("// Generated data.\npackage example\nvar names = map[string][]string{\n\"🫯\": {\"fight cloud\", \"é\"},\n}\n")
	escaped := []byte(`// Generated data.
package example
var names = map[string][]string{
"\U0001faef": {"fight cloud", "\u00e9"},
}
`)
	want := normalizeGeneratedSource(t, source)
	if !bytes.Equal(want, normalizeGeneratedSource(t, escaped)) {
		t.Fatal("equivalent Unicode string spellings differ")
	}
	for _, change := range []struct{ old, new string }{
		{"fight cloud", "wrong name"},
		{"🫯", "☁"},
		{"names", "renamed"},
		{"Generated data.", "Changed provenance."},
		{`, "é"`, ""},
	} {
		changed := strings.Replace(string(source), change.old, change.new, 1)
		if bytes.Equal(want, normalizeGeneratedSource(t, []byte(changed))) {
			t.Errorf("missed change from %q to %q", change.old, change.new)
		}
	}
}
