package ntemoji

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestLocalizedMetadata(t *testing.T) {
	for _, tc := range []struct{ locale, want string }{
		{"en", "rocket"}, {"es", "cohete"}, {"ES-mx", "cohete"},
		{"fr_CA", "fusée"}, {"de-DE", "Rakete"}, {"ja-JP", "ロケット"},
		{"zh", "火箭"}, {"zh-Hans-CN", "火箭"}, {"zh-Hans-TW", "火箭"},
		{"zh-TW", "rocket"}, {"zh-Hant", "rocket"}, {"es-Cyrl", "rocket"},
		{"zh-x-hant", "火箭"}, {"fr-u-nu-latn", "fusée"},
		{"", "rocket"}, {"unknown", "rocket"},
	} {
		t.Run(tc.locale, func(t *testing.T) {
			if got := EmojiNameFor("🚀", tc.locale); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if !MatchEmojiFor("🚀", strings.ToUpper(tc.want), tc.locale) || !MatchEmojiFor("🚀", "rocket", tc.locale) {
				t.Fatal("localized or English search failed")
			}
		})
	}
	for _, locale := range Locales() {
		if EmojiNameFor("❤️", locale) != EmojiNameFor("❤", locale) {
			t.Fatal("presentation selector differs")
		}
		if EmojiNameFor("unknown", locale) != "unknown" || TagsFor("unknown", locale) != nil {
			t.Fatal("unknown fallback failed")
		}
		if EmojiNameFor("", locale) != "" {
			t.Fatal("empty name changed")
		}
		tags := TagsFor("🚀", locale)
		tags[0] = "mutated"
		if TagsFor("🚀", locale)[0] == "mutated" {
			t.Fatal("tags alias internal storage")
		}
	}
	locales := Locales()
	if !slices.Equal(locales, []string{"de", "en", "es", "fr", "ja", "zh"}) {
		t.Fatalf("locales %v", locales)
	}
	locales[0] = "changed"
	if Locales()[0] != "de" {
		t.Fatal("locales aliases internal storage")
	}
}

func TestLocalizedPickerAndSwatch(t *testing.T) {
	m := New(WithLocale("es"), WithListMode(true), WithShowSearch(true), WithInitialEmoji("🚀"))
	if s := m.AccessibilityState(); s.FocusedName != "cohete" || s.SelectedName != "cohete" {
		t.Fatalf("state: %+v", s)
	}
	if !strings.Contains(m.View().Content, "cohete") {
		t.Fatal("view lacks localized name")
	}
	m.Focus = FocusSearch
	m.SearchInput.SetValue("cohete")
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = next.(Model)
	if !slices.Contains(m.filteredEmojis(), "🚀") {
		t.Fatal("localized query failed")
	}
	swatch := NewSwatchPicker("🚀", "Launch")
	swatch.SetPickerOptions(WithLocale("fr"))
	if swatch.AccessibilityState().Name != "fusée" {
		t.Fatal("closed swatch locale ignored")
	}
	swatch.OpenPicker()
	if swatch.AccessibilityState().Picker.FocusedName != "fusée" {
		t.Fatal("open swatch locale ignored")
	}
}

func TestTranslationsContainNoUnresolvedLabels(t *testing.T) {
	for locale, entries := range emojiLocales {
		for emoji, e := range entries {
			if e.Name == "" || strings.Contains(e.Name, "↑↑↑") || strings.Contains(e.Name, "∅∅∅") {
				t.Fatalf("%s %s unresolved", locale, emoji)
			}
		}
		// Missing localized symbols still have usable English labels.
		for emoji, name := range emojiNames {
			if _, ok := entries[emoji]; !ok && EmojiNameFor(emoji, locale) != name {
				t.Fatalf("%s %s fallback failed", locale, emoji)
			}
		}
	}
}
