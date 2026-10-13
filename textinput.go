package ntemoji

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TextInput is a single-line input with terminal-cell width limits.
// It intentionally avoids bubbles/textinput and its system clipboard dependency:
// paste arrives through Bubble Tea bracketed-paste events instead.
type TextInput struct {
	Placeholder string
	Prompt      string

	PromptStyle      lipgloss.Style
	TextStyle        lipgloss.Style
	PlaceholderStyle lipgloss.Style
	CursorStyle      lipgloss.Style

	value   []rune
	cursor  int
	width   int
	focused bool
}

// NewTextInput constructs an input with a 15-cell width.
func NewTextInput() TextInput { return TextInput{width: 15} }

// SetWidth limits the whole rendered input, including prompt and cursor, in
// terminal cells. Zero and negative widths hide the input without losing text.
func (t *TextInput) SetWidth(w int) { t.width = max(0, w) }

// Focus focuses the input.
func (t *TextInput) Focus() tea.Cmd {
	t.focused = true
	return nil
}

// Blur removes input focus.
func (t *TextInput) Blur() { t.focused = false }

// Focused reports whether the input accepts keyboard and paste events.
func (t *TextInput) Focused() bool { return t.focused }

// Value returns the complete text, including text outside the visible viewport.
func (t *TextInput) Value() string { return string(t.value) }

// SetValue replaces the input, normalizes it to one line, and moves to its end.
func (t *TextInput) SetValue(s string) {
	t.value = []rune(singleLine(s))
	t.cursor = len(t.value)
}

func singleLine(s string) string {
	s = strings.ReplaceAll(ansi.Strip(s), "\r\n", "\n")
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', '\u2028', '\u2029':
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// boundaries returns rune offsets so editing never splits a grapheme cluster.
func (t TextInput) boundaries() []int {
	result := []int{0}
	text, offset := string(t.value), 0
	for len(text) > 0 {
		cluster, _ := ansi.FirstGraphemeCluster(text, ansi.GraphemeWidth)
		offset += utf8.RuneCountInString(cluster)
		result = append(result, offset)
		text = text[len(cluster):]
	}
	return result
}

func (t TextInput) previous() int {
	previous := 0
	for _, boundary := range t.boundaries() {
		if boundary >= t.cursor {
			break
		}
		previous = boundary
	}
	return previous
}

func (t TextInput) next() int {
	for _, boundary := range t.boundaries() {
		if boundary > t.cursor {
			return boundary
		}
	}
	return len(t.value)
}

func (t *TextInput) insert(s string) {
	runes := []rune(singleLine(s))
	value := make([]rune, 0, len(t.value)+len(runes))
	value = append(value, t.value[:t.cursor]...)
	value = append(value, runes...)
	value = append(value, t.value[t.cursor:]...)
	t.value = value
	t.cursor += len(runes)
	// Insertion can join a combining sequence with the following text.
	for _, boundary := range t.boundaries() {
		if boundary >= t.cursor {
			t.cursor = boundary
			break
		}
	}
}

// Update handles key presses and bracketed-paste events while focused.
func (t TextInput) Update(msg tea.Msg) (TextInput, tea.Cmd) {
	if !t.focused {
		return t, nil
	}
	switch msg := msg.(type) {
	case tea.PasteMsg:
		t.insert(msg.Content)
	case tea.KeyPressMsg:
		// Preserve value semantics when models are copied by their parents.
		t.value = slices.Clone(t.value)
		switch msg.String() {
		case "backspace":
			previous := t.previous()
			t.value = append(t.value[:previous], t.value[t.cursor:]...)
			t.cursor = previous
		case "delete":
			t.value = append(t.value[:t.cursor], t.value[t.next():]...)
		case "left":
			t.cursor = t.previous()
		case "right":
			t.cursor = t.next()
		case "home", "ctrl+a":
			t.cursor = 0
		case "end", "ctrl+e":
			t.cursor = len(t.value)
		default:
			text := msg.Text
			if text == "" && msg.Code == ' ' && msg.Mod == 0 {
				text = " "
			}
			if text != "" {
				t.insert(text)
			}
		}
	}
	return t, nil
}

// View keeps the cursor visible by scrolling over whole grapheme clusters.
func (t TextInput) View() string {
	if t.width <= 0 {
		return ""
	}
	promptWidth := t.width
	if t.focused && t.cursor < len(t.value) {
		_, width := ansi.FirstGraphemeCluster(string(t.value[t.cursor:]), ansi.GraphemeWidth)
		// Give the cursor grapheme priority over the prompt in narrow inputs.
		promptWidth = max(0, t.width-width)
	}
	prompt := ansi.Truncate(t.PromptStyle.Render(t.Prompt), promptWidth, "")
	available := t.width - lipgloss.Width(prompt)
	if available <= 0 {
		return prompt
	}
	if len(t.value) == 0 {
		placeholder := t.PlaceholderStyle.Render(singleLine(t.Placeholder))
		if t.focused {
			placeholder = t.CursorStyle.Render(" ") + placeholder
		}
		return prompt + ansi.Truncate(placeholder, available, "")
	}
	bounds := t.boundaries()
	clusters := make([]string, 0, len(bounds))
	cursor := 0
	for i := 0; i < len(bounds)-1; i++ {
		if bounds[i] < t.cursor {
			cursor = i + 1
		}
		clusters = append(clusters, string(t.value[bounds[i]:bounds[i+1]]))
	}
	clusters = append(clusters, " ") // Cursor at the end of the value.
	start := 0
	if t.focused {
		if lipgloss.Width(clusters[cursor]) > available {
			clusters[cursor] = "…" // Indicate clipping when even the full input cannot fit it.
		}
		used := lipgloss.Width(strings.Join(clusters[:cursor+1], ""))
		for used > available && start < cursor {
			used -= lipgloss.Width(clusters[start])
			start++
		}
	}
	// Style contiguous runs once, rather than emitting escapes per grapheme.
	end := len(clusters) - 1
	var rendered string
	if t.focused {
		rendered = t.TextStyle.Render(strings.Join(clusters[start:cursor], "")) +
			t.CursorStyle.Render(clusters[cursor])
		if cursor < end {
			rendered += t.TextStyle.Render(strings.Join(clusters[cursor+1:end], ""))
		}
	} else {
		rendered = t.TextStyle.Render(strings.Join(clusters[:end], ""))
	}
	return prompt + ansi.Truncate(rendered, available, "")
}
