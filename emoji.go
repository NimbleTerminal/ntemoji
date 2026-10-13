package ntemoji

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Palette represents a category of emojis for the picker.
type Palette struct {
	Name   string   // Display name of the category (e.g. "Smileys")
	Icon   string   // Single emoji representing the category tab
	Emojis []string // Slice of emojis in this palette
}

// DefaultPalettes is a list of curated TUI-safe emojis grouped by category.
var DefaultPalettes = []Palette{
	{
		Name: "Smileys",
		Icon: EmojiGrinningFace,
		Emojis: []string{
			EmojiGrinningFace, EmojiGrinningWithEyes, EmojiSmilingWithOpen, EmojiGrinningSquinting,
			EmojiSmilingSweat, EmojiTearsOfJoy, EmojiRollingOnFloor, EmojiSmilingFace,
			EmojiAngelFace, EmojiSlightlySmiling, EmojiUpsideDown, EmojiWinkingFace,
			EmojiRelievedFace, EmojiHeartEyes, EmojiSmilingWithHearts, EmojiKissingFace,
			EmojiThinkingFace, EmojiShushingFace, EmojiGrimacingFace, EmojiZippedMouth,
			EmojiSweatFace, EmojiCryingFace, EmojiLoudlyCrying, EmojiAngryFace,
			EmojiRedAngryFace, EmojiExplodingHead, EmojiHotFace, EmojiColdFace,
			EmojiSleepingFace, EmojiNerdFace, EmojiCoolFace, EmojiPartyFace,
		},
	},
	{
		Name: "Gestures",
		Icon: EmojiThumbsUp,
		Emojis: []string{
			EmojiWave, EmojiRaisedHand, EmojiVulcanSalute, EmojiOkHand,
			EmojiCrossedFingers, EmojiThumbsUp, EmojiThumbsDown, EmojiClappingHands,
			EmojiRaisedHands, EmojiFoldedHands, EmojiFlexedBiceps,
		},
	},
	{
		Name: "Animals & Nature",
		Icon: EmojiCatFace,
		Emojis: []string{
			EmojiCatFace, EmojiCat, EmojiDogFace, EmojiDog,
			EmojiFox, EmojiLion, EmojiTigerFace, EmojiMonkeyFace,
			EmojiCowFace, EmojiPigFace, EmojiFrog, EmojiOctopus,
			EmojiBee, EmojiBug, EmojiPenguin, EmojiOwl,
			EmojiTurtle, EmojiSnake, EmojiDolphin, EmojiTree,
			EmojiDeciduousTree, EmojiCactus, EmojiMapleLeaf, EmojiSun,
			EmojiRainbow, EmojiSnowman, EmojiLightning, EmojiFire,
			EmojiWaterWave, EmojiGlobe,
		},
	},
	{
		Name: "Food & Drink",
		Icon: EmojiCoffee,
		Emojis: []string{
			EmojiGreenApple, EmojiRedApple, EmojiBanana, EmojiWatermelon,
			EmojiGrapes, EmojiStrawberry, EmojiPizza, EmojiHamburger,
			EmojiFrenchFries, EmojiHotDog, EmojiTaco, EmojiPopcorn,
			EmojiRiceBall, EmojiSushi, EmojiIceCream, EmojiDoughnut,
			EmojiCookie, EmojiBirthdayCake, EmojiCoffee, EmojiTea,
			EmojiBeerMug, EmojiWineGlass, EmojiCocktail,
		},
	},
	{
		Name: "Status & Shapes",
		Icon: EmojiGreenCircle,
		Emojis: []string{
			EmojiRedCircle, EmojiYellowCircle, EmojiGreenCircle, EmojiBlueCircle, EmojiPurpleCircle,
			EmojiBlackSquare, EmojiWhiteSquare, EmojiStopSign, EmojiConstruction,
			EmojiChequeredFlag, EmojiTriangularFlag, EmojiBell, EmojiMutedBell,
			EmojiMegaphone, EmojiHourglass, EmojiHourglassEmpty, EmojiSpeechBalloon,
			EmojiHeart, EmojiBrokenHeart, EmojiSparkles, EmojiCollision,
			EmojiRocket, EmojiCheckMarkButton, EmojiCrossMark, EmojiWarning, EmojiInfo,
		},
	},
	{
		Name: "Symbols & Tools",
		Icon: EmojiGear,
		Emojis: []string{
			EmojiGear, EmojiWrench, EmojiHammer, EmojiToolbox,
			EmojiKey, EmojiLock, EmojiUnlock, EmojiPackage,
			EmojiFolder, EmojiOpenFolder, EmojiCalendar, EmojiBarChart,
			EmojiChartUp, EmojiChartDown, EmojiMemo, EmojiEnvelope,
			EmojiIncomingMail, EmojiPaperclip, EmojiLightBulb, EmojiMagnifyingGlass,
		},
	},
	{
		Name: "Navigation",
		Icon: EmojiArrowRight,
		Emojis: []string{
			EmojiArrowUp, EmojiArrowDown, EmojiArrowLeft, EmojiArrowRight, EmojiRefresh,
		},
	},
}

// Theme defines the visual styles for the emoji picker.
type Theme struct {
	// Outer frame styling
	FrameBorder      lipgloss.Border
	FrameBorderColor color.Color

	// Inner focus borders
	BorderFocusedColor   color.Color
	BorderUnfocusedColor color.Color

	// Text styling
	TitleStyle lipgloss.Style
	HelpStyle  lipgloss.Style

	// Tabs/Categories styling
	ActiveTabStyle   lipgloss.Style
	InactiveTabStyle lipgloss.Style

	// Emoji grid cell styling
	SelectedCellStyle          lipgloss.Style
	SelectedUnfocusedCellStyle lipgloss.Style
	UnselectedCellStyle        lipgloss.Style

	// Preset cell styling
	SelectedPresetStyle   lipgloss.Style
	UnselectedPresetStyle lipgloss.Style

	// Search text input styling
	SearchPromptStyle      lipgloss.Style
	SearchTextStyle        lipgloss.Style
	SearchPlaceholderStyle lipgloss.Style
	SearchCursorStyle      lipgloss.Style
}

// DefaultTheme provides a dark, purple-accented modern theme.
var DefaultTheme = Theme{
	FrameBorder:      lipgloss.DoubleBorder(),
	FrameBorderColor: lipgloss.Color("93"), // Purple

	BorderFocusedColor:   lipgloss.Color("255"), // White
	BorderUnfocusedColor: lipgloss.Color("240"), // Dark Gray

	TitleStyle: lipgloss.NewStyle().Bold(true),
	HelpStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("241")),

	ActiveTabStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
	InactiveTabStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("246")),

	SelectedCellStyle:          lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("255")).Bold(true),
	SelectedUnfocusedCellStyle: lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("255")),
	UnselectedCellStyle:        lipgloss.NewStyle(),
	SelectedPresetStyle:        lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("255")).Bold(true),
	UnselectedPresetStyle:      lipgloss.NewStyle(),

	SearchPromptStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("93")),
	SearchTextStyle:        lipgloss.NewStyle(),
	SearchPlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	SearchCursorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("93")),
}

// Config holds the configuration options for the emoji picker.
type Config struct {
	InitialEmoji string
	Presets      []string
	Palettes     []Palette
	FrameStyle   lipgloss.Style
	CustomFrame  bool
	AutoDismiss  bool
	ShowSearch   bool
	Theme        Theme
}

// Option configures an emoji picker.
type Option func(*Config)

// WithInitialEmoji sets the starting emoji selection.
func WithInitialEmoji(emoji string) Option {
	return func(c *Config) {
		c.InitialEmoji = emoji
	}
}

// WithPresets configures a list of preset/recent emojis displayed at the top.
func WithPresets(emojis []string) Option {
	return func(c *Config) {
		c.Presets = append([]string(nil), emojis...)
	}
}

// WithPalettes sets custom emoji palettes for the picker.
func WithPalettes(palettes []Palette) Option {
	return func(c *Config) {
		c.Palettes = append([]Palette(nil), palettes...)
	}
}

// WithStyle sets the style of the outer frame (border, padding, etc.).
func WithStyle(s lipgloss.Style) Option {
	return func(c *Config) {
		c.FrameStyle = s
		c.CustomFrame = true
	}
}

// WithAutoDismiss sets whether the picker emits EmojiChangedMsg with Dismiss=true
// to let the host immediately close/dismiss the picker.
func WithAutoDismiss(dismiss bool) Option {
	return func(c *Config) {
		c.AutoDismiss = dismiss
	}
}

// WithShowSearch toggles the search input bar.
func WithShowSearch(show bool) Option {
	return func(c *Config) {
		c.ShowSearch = show
	}
}

// WithTheme configures a custom visual theme for the picker.
func WithTheme(theme Theme) Option {
	return func(c *Config) {
		c.Theme = theme
	}
}

// EmojiChangedMsg is emitted when the user confirms their emoji selection.
type EmojiChangedMsg struct {
	Emoji   string // The selected emoji
	Dismiss bool   // Whether the picker should be dismissed automatically
}

// EmojiChosenMsg is an alias for EmojiChangedMsg for compatibility.
type EmojiChosenMsg = EmojiChangedMsg

// EmojiCanceledMsg is emitted when the user cancels the picker (e.g. presses Esc).
type EmojiCanceledMsg struct{}

// MatchEmoji returns true if the emoji matches the query based on tags.
func MatchEmoji(emoji, query string) bool {
	if query == "" {
		return true
	}
	importStr := strings.ToLower(query)
	if strings.Contains(emoji, importStr) {
		return true
	}
	tags, ok := EmojiTags[emoji]
	if !ok {
		return false
	}
	for _, t := range tags {
		if strings.Contains(t, importStr) {
			return true
		}
	}
	return false
}

// EmojiTags maps common emojis to searchable keywords.
var EmojiTags = map[string][]string{
	// Smileys
	"😀":  {"smile", "happy", "face", "grinning", "laugh"},
	"😃":  {"smile", "happy", "face", "grinning", "laugh", "open"},
	"😄":  {"smile", "happy", "face", "grinning", "laugh", "eyes"},
	"😁":  {"smile", "happy", "face", "grinning", "laugh", "teeth"},
	"😆":  {"smile", "happy", "face", "grinning", "laugh", "closed"},
	"😅":  {"smile", "happy", "face", "sweat", "laugh"},
	"😂":  {"smile", "happy", "face", "tears", "joy", "laugh", "crying"},
	"🤣":  {"smile", "happy", "face", "tears", "joy", "laugh", "rofl", "floor"},
	"😊":  {"smile", "happy", "face", "blush", "nice"},
	"😇":  {"smile", "happy", "face", "angel", "halo"},
	"🙂":  {"smile", "happy", "face", "slight"},
	"🙃":  {"smile", "happy", "face", "upside", "down"},
	"😉":  {"smile", "happy", "face", "wink"},
	"😌":  {"smile", "happy", "face", "relieved", "calm"},
	"😍":  {"smile", "happy", "face", "love", "heart", "eyes"},
	"🥰":  {"smile", "happy", "face", "love", "hearts", "affection"},
	"😘":  {"smile", "happy", "face", "love", "kiss", "blowing"},
	"🤔":  {"think", "face", "pondering", "question", "hm"},
	"👍":  {"thumbs", "up", "like", "yes", "agree", "ok", "good"},
	"👎":  {"thumbs", "down", "dislike", "no", "bad", "disagree"},
	"🤞":  {"crossed", "fingers", "luck", "hope", "hand", "gesture"},
	"🎉":  {"party", "celebration", "popper", "congrats", "hooray"},
	"🚀":  {"rocket", "space", "launch", "fast", "ship"},
	"🔥":  {"fire", "hot", "burn", "lit", "flame"},
	"💡":  {"light", "bulb", "idea", "inspiration", "brain"},
	"🌟":  {"star", "shine", "sparkle", "gold"},
	"💥":  {"collision", "explosion", "boom", "bang"},
	"💖":  {"heart", "love", "red"},
	"✅":  {"check", "mark", "correct", "yes", "green", "done", "success"},
	"❌":  {"cross", "mark", "wrong", "no", "red", "fail", "error"},
	"🚨":  {"warning", "alert", "danger", "yellow", "caution"},
	"📢":  {"info", "information", "blue", "about"},
	"⚙":  {"gear", "settings", "cog", "tool", "config"},
	"🔧":  {"wrench", "spanner", "tool", "repair", "fix"},
	"🟢":  {"green", "circle", "status", "go", "dot"},
	"🔴":  {"red", "circle", "status", "stop", "dot"},
	"🟡":  {"yellow", "circle", "status", "wait", "dot"},
	"🔵":  {"blue", "circle", "status", "dot"},
	"⏰":  {"alarm", "clock", "time", "wake"},
	"⏳":  {"hourglass", "time", "wait", "loading"},
	"📝":  {"memo", "write", "note", "paper", "pencil"},
	"💻":  {"computer", "laptop", "pc", "tech", "screen"},
	"📁":  {"folder", "file", "directory", "save"},
	"📦":  {"package", "box", "delivery", "shipping"},
	"🛠️": {"tools", "hammer", "wrench", "build", "dev", "fix"},
	"📩":  {"envelope", "mail", "letter", "email"},
	"📈":  {"chart", "graph", "increase", "up", "trend"},
	"📉":  {"chart", "graph", "decrease", "down", "trend"},
	"📊":  {"chart", "graph", "bar", "stats"},
	"🏁":  {"flag", "checkered", "finish", "done"},
	"🧰":  {"toolbox", "box", "tools", "kit", "config"},
	"🔑":  {"key", "lock", "access", "secret"},
	"🔒":  {"lock", "secure", "closed", "security"},
	"🔓":  {"lock", "unlock", "open", "access"},
	// Animals & Nature
	"🐱": {"cat", "feline", "animal", "pet"},
	"🐈": {"cat", "feline", "animal", "pet", "walking"},
	"🐶": {"dog", "canine", "animal", "pet", "puppy"},
	"🐕": {"dog", "canine", "animal", "pet", "walking"},
	"🦊": {"fox", "animal", "wild"},
	"🦁": {"lion", "animal", "wild", "cat"},
	"🐯": {"tiger", "animal", "wild", "cat"},
	"🐵": {"monkey", "animal"},
	"🐮": {"cow", "animal", "farm"},
	"🐷": {"pig", "animal", "farm"},
	"🐸": {"frog", "animal", "amphibian"},
	"🐙": {"octopus", "animal", "sea"},
	"🐝": {"bee", "bug", "insect", "honey"},
	"🐛": {"bug", "insect", "caterpillar"},
	"🐧": {"penguin", "bird", "animal", "cold"},
	"🦉": {"owl", "bird", "animal"},
	"🐢": {"turtle", "animal", "reptile"},
	"🐍": {"snake", "animal", "reptile"},
	"🐬": {"dolphin", "animal", "sea"},
	"🌲": {"tree", "pine", "forest", "nature"},
	"🌳": {"tree", "forest", "nature"},
	"🌵": {"cactus", "plant", "desert"},
	"🍁": {"maple", "leaf", "canada", "fall"},
	"🌞": {"sun", "sunny", "weather", "hot"},
	"🌈": {"rainbow", "color", "sky", "weather", "nature"},
	"⛄": {"snowman", "snow", "cold", "weather", "winter"},
	"⚡": {"lightning", "thunder", "storm", "flash"},
	"🌊": {"wave", "water", "sea", "ocean"},
	"🌍": {"globe", "earth", "world", "map"},

	// Food & Drink
	"🍏": {"apple", "green", "fruit", "food"},
	"🍎": {"apple", "red", "fruit", "food"},
	"🍌": {"banana", "yellow", "fruit", "food"},
	"🍉": {"watermelon", "fruit", "food"},
	"🍇": {"grapes", "fruit", "food"},
	"🍓": {"strawberry", "fruit", "food"},
	"🍕": {"pizza", "food", "italian"},
	"🍔": {"burger", "hamburger", "food"},
	"🍟": {"fries", "french fries", "food"},
	"🌭": {"hotdog", "hot dog", "food"},
	"🌮": {"taco", "food", "mexican"},
	"🍿": {"popcorn", "food", "snack"},
	"🍙": {"riceball", "rice", "food", "japanese"},
	"🍣": {"sushi", "fish", "food", "japanese"},
	"🍦": {"icecream", "ice cream", "sweet", "dessert"},
	"🍩": {"donut", "doughnut", "sweet", "dessert"},
	"🍪": {"cookie", "sweet", "dessert", "snack"},
	"🎂": {"cake", "birthday", "sweet", "dessert"},
	"☕": {"coffee", "cafe", "drink", "hot"},
	"🍵": {"tea", "drink", "hot", "green"},
	"🍺": {"beer", "alcohol", "drink", "bar"},
	"🍷": {"wine", "alcohol", "drink"},
	"🍸": {"cocktail", "alcohol", "drink"},
}
