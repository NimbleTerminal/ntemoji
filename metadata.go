package ntemoji

import (
	"slices"
	"strings"
)

//go:generate go run ./internal/cmd/gencldr

// metadataKey follows CLDR annotation keys, which omit emoji presentation selectors.
func metadataKey(emoji string) string { return strings.ReplaceAll(emoji, "\ufe0f", "") }

// EmojiName returns a descriptive English name from Unicode CLDR for an emoji or annotated symbol.
// Unknown values are returned unchanged, including the empty string. Callers
// using custom palettes can supply their own labels before falling back here.
// Names are intended for display and accessibility labels, not stable IDs.
func EmojiName(emoji string) string {
	if name, ok := emojiNames[metadataKey(emoji)]; ok {
		return name
	}
	return emoji
}

// Tags returns an independent copy of an emoji's search keywords. Built-in
// emojis without extra keywords return their descriptive name as the sole tag.
// Unknown values return nil. Changing the result never changes package metadata.
func Tags(emoji string) []string {
	if tags, ok := emojiTags[metadataKey(emoji)]; ok {
		return slices.Clone(tags)
	}
	if name, ok := emojiNames[metadataKey(emoji)]; ok {
		return []string{name}
	}
	return nil
}
