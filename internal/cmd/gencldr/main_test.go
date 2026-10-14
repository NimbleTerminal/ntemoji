package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
	if !bytes.Equal(got, want) {
		t.Fatal("metadata is stale; run go generate ./...")
	}
}

func TestParseAnnotations(t *testing.T) {
	entries := map[string]entry{}
	err := parse([]byte(`<ldml><identity><language type="en"/></identity><annotations>
<annotation cp="❤️" type="tts">red heart</annotation>
<annotation cp="❤"> heart | love | heart </annotation>
<annotation cp="🧑‍🚀" type="tts">astronaut</annotation>
<annotation cp="🧑‍🚀">space | rocket</annotation>
<annotation cp="&amp;" type="tts">ampersand</annotation>
</annotations></ldml>`), entries)
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
		if err := parse([]byte(content), entries); err == nil {
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
	if !bytes.Equal(got, want) {
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
	if !bytes.Equal(got, want) {
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
