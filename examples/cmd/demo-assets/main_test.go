package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"nimbleterminal.dev/ntfonts/packs/emoji"
	"nimbleterminal.dev/ntfonts/packs/emoji/sbix"
)

func TestStagedFontFormats(t *testing.T) {
	dir := t.TempDir()
	if err := stage(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "fonts", "specimens.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Faces) != 2 || m.Faces[0].Tech != "color-CBDT" || m.Faces[1].Tech != "color-sbix" {
		t.Fatalf("wrong format order: %+v", m.Faces)
	}
	for i, source := range []fs.FS{emoji.Files(), sbix.Files()} {
		face := m.Faces[i]
		original, err := fs.ReadFile(source, face.File)
		if err != nil {
			t.Fatal(err)
		}
		staged, err := os.ReadFile(filepath.Join(dir, "fonts", face.File))
		if err != nil || !bytes.Equal(original, staged) || face.Bytes != len(staged) {
			t.Fatalf("font %s changed: %v", face.File, err)
		}
	}
	for _, name := range []string{"OFL.txt", "OFL-sbix.txt", "NOTICE.txt"} {
		text, err := os.ReadFile(filepath.Join(dir, "fonts", name))
		if err != nil || len(text) == 0 {
			t.Fatalf("missing notice %s: %v", name, err)
		}
	}
	if len(m.Specimens) != 8 {
		t.Fatalf("missing specimens: %d", len(m.Specimens))
	}
	for _, sample := range m.Specimens {
		for _, locale := range m.Locales {
			if sample.Names[locale] == "" {
				t.Fatalf("missing %s name for %s", locale, sample.Text)
			}
		}
	}
}
