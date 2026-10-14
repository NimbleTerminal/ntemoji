package ntemoji

import (
	"slices"
	"strings"
)

type localizedMetadata struct {
	Name string
	Tags []string
}

// Locales returns an independent, sorted list of bundled metadata locales.
// zh contains Simplified Chinese. UI chrome and category names are not translated.
func Locales() []string {
	locales := []string{"en"}
	for locale := range emojiLocales {
		locales = append(locales, locale)
	}
	slices.Sort(locales)
	return locales
}

// metadataLocale matches supported base languages and their regional variants.
// Explicit unsupported scripts fall back to English, rather than silently using
// a different writing system. This is a limited bundle matcher, not CLDR's full
// locale inheritance algorithm.
func metadataLocale(locale string) string {
	parts := strings.Split(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "_", "-")), "-")
	base := parts[0]
	script, region := "", ""
	for _, part := range parts[1:] {
		if len(part) == 1 {
			break
		} // Extensions and private-use do not select a script.
		if len(part) == 4 {
			script = part
		}
		if len(part) == 2 || len(part) == 3 {
			region = part
		}
	}
	expectedScript := "latn"
	if base == "zh" {
		expectedScript = "hans"
	}
	if base == "ja" {
		expectedScript = "jpan"
	}
	if script != "" && script != expectedScript {
		return "en"
	}
	if base == "zh" && script == "" && (region == "tw" || region == "hk" || region == "mo") {
		return "en"
	}

	if _, ok := emojiLocales[base]; ok {
		return base
	}
	return "en"
}

// EmojiNameFor returns an emoji's name in the requested locale, falling back to
// English and then the original input. Locale matching is case-insensitive and
// accepts hyphens or underscores (for example, es-MX and fr_CA). Unsupported
// scripts, including Traditional Chinese, fall back to English.
func EmojiNameFor(emoji, locale string) string {
	if e, ok := emojiLocales[metadataLocale(locale)][metadataKey(emoji)]; ok && e.Name != "" {
		return e.Name
	}
	return EmojiName(emoji)
}

// TagsFor returns an independent copy of localized keywords. Entries without
// keywords return their localized name. Missing entries fall back to English;
// unknown emoji return nil.
func TagsFor(emoji, locale string) []string {
	if e, ok := emojiLocales[metadataLocale(locale)][metadataKey(emoji)]; ok {
		if len(e.Tags) > 0 {
			return slices.Clone(e.Tags)
		}
		if e.Name != "" {
			return []string{e.Name}
		}
	}
	return Tags(emoji)
}

// MatchEmojiFor matches the glyph, localized name or keywords, and English
// metadata. Matching is case-insensitive; it does not transliterate or fold accents.
func MatchEmojiFor(emoji, query, locale string) bool {
	if MatchEmoji(emoji, query) {
		return true
	}
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(EmojiNameFor(emoji, locale)), query) {
		return true
	}
	// Internal search only reads metadata; the public TagsFor API still clones.
	for _, tag := range emojiLocales[metadataLocale(locale)][metadataKey(emoji)].Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// WithLocale selects the metadata language for labels, search and accessibility
// state. It does not translate UI instructions or caller-supplied category names.
func WithLocale(locale string) Option {
	return func(c *Config) { c.Locale = locale }
}
