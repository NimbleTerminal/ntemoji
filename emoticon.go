package ntemoji

// EmoticonToEmoji converts a single recognized emoticon to an emoji. Matching
// is exact: surrounding text and whitespace are not replaced or trimmed.
// Unknown values, including the empty string, are returned unchanged.
func EmoticonToEmoji(emoticon string) string {
	for _, entry := range emoticons {
		for _, spelling := range entry.spellings {
			if spelling == emoticon {
				return entry.emoji
			}
		}
	}
	return emoticon
}

// EmojiToEmoticon converts a single recognized emoji to its preferred ASCII
// emoticon. Unknown values are returned unchanged. Conversion is intentionally
// lossy: multiple emoticon spellings share one preferred reverse spelling.
func EmojiToEmoticon(emoji string) string {
	for _, entry := range emoticons {
		if entry.emoji == emoji {
			return entry.spellings[0]
		}
	}
	return emoji
}

// The first spelling is the preferred reverse conversion. Keep both directions
// in one private table so their mappings cannot drift apart.
var emoticons = []struct {
	emoji     string
	spellings []string
}{
	{EmojiSlightlySmiling, []string{":)", ":-)"}},
	{"🙁", []string{":(", ":-("}},
	{EmojiGrinningWithEyes, []string{":D", ":-D", ":d", ":-d"}},
	{EmojiWinkingFace, []string{";)", ";-)"}},
	{EmojiGrinningSquinting, []string{"XD", "xD"}},
	{"😛", []string{":P", ":-P", ":p", ":-p"}},
	{"😮", []string{":O", ":-O", ":o", ":-o"}},
	{"😕", []string{":/", ":-/", ":\\", ":-\\"}},
	{"😐", []string{":|", ":-|"}},
	{EmojiCryingFace, []string{":'(", ":'-("}},
	{EmojiAngryFace, []string{">:(", ">:-("}},
	{EmojiKissingFace, []string{":*", ":-*"}},
	{EmojiCoolFace, []string{"B)", "B-)", "8)", "8-)"}},
	{"❤️", []string{"<3"}},
	{EmojiBrokenHeart, []string{"</3"}},
}
