// Demo-assets stages optional browser fonts and localized sequence specimens.
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"nimbleterminal.dev/ntemoji"
	"nimbleterminal.dev/ntfonts/packs/emoji"
	"nimbleterminal.dev/ntfonts/packs/emoji/sbix"
)

type specimen struct {
	Text  string            `json:"text"`
	Names map[string]string `json:"names"`
}

type fontFace struct {
	File   string `json:"file"`
	Weight int    `json:"weight"`
	Style  string `json:"style"`
	Tech   string `json:"tech"`
	Bytes  int    `json:"bytes"`
}

type manifest struct {
	Family    string     `json:"family"`
	Faces     []fontFace `json:"faces"`
	Locales   []string   `json:"locales"`
	Specimens []specimen `json:"specimens"`
}

func stage(dir string) error {
	fonts := filepath.Join(dir, "fonts")
	if err := os.MkdirAll(fonts, 0755); err != nil {
		return err
	}
	cbdtFace, sbixFace := emoji.WebFaces()[0], sbix.WebFaces()[0]
	m := manifest{Family: cbdtFace.Family, Locales: ntemoji.Locales()}
	// Prefer the unmodified CBDT file; Safari selects the sbix companion.
	for _, source := range []struct {
		files fs.FS
		face  fontFace
	}{
		{emoji.Files(), fontFace{File: cbdtFace.File, Weight: cbdtFace.Weight, Style: cbdtFace.Style, Tech: cbdtFace.Tech}},
		{sbix.Files(), fontFace{File: sbixFace.File, Weight: sbixFace.Weight, Style: sbixFace.Style, Tech: sbixFace.Tech}},
	} {
		data, err := fs.ReadFile(source.files, source.face.File)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(fonts, source.face.File), data, 0644); err != nil {
			return err
		}
		source.face.Bytes = len(data)
		m.Faces = append(m.Faces, source.face)
	}
	for _, source := range []struct {
		file string
		text string
	}{
		{"OFL.txt", emoji.Licenses()[0].Text},
		{"OFL-sbix.txt", sbix.Licenses()[0].Text},
		{"NOTICE.txt", "NotoColorEmoji.ttf: unmodified Noto Color Emoji v2.051 (Google), supplied by ntfonts/packs/emoji.\nNotoColorEmoji-sbix.ttf: the same artwork repackaged as sbix by ntfonts/packs/emoji/sbix; modified font name: Noto Color Emoji sbix.\nBoth fonts are licensed under SIL OFL-1.1. See OFL.txt and OFL-sbix.txt.\n"},
	} {
		if err := os.WriteFile(filepath.Join(fonts, source.file), []byte(source.text), 0644); err != nil {
			return err
		}
	}
	for _, text := range []string{"😀", "👍🏽", "👨‍👩‍👧‍👦", "🇺🇳", "🧑🏽‍🚀", "❤️", "⚙️", "1️⃣"} {
		s := specimen{Text: text, Names: map[string]string{}}
		for _, locale := range m.Locales {
			s.Names[locale] = ntemoji.EmojiNameFor(text, locale)
		}
		m.Specimens = append(m.Specimens, s)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(fonts, "specimens.json"), append(data, '\n'), 0644)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: demo-assets OUTPUT_DIRECTORY")
		os.Exit(2)
	}
	if err := stage(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
