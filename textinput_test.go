package ntemoji

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSearchReceivesEditingKeys(t *testing.T) {
	picker := New(WithShowSearch(true))
	for _, r := range "heart hjkl" {
		m, _ := picker.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		picker = m.(Model)
	}
	if got := picker.SearchInput.Value(); got != "heart hjkl" {
		t.Fatalf("search swallowed letters: %q", got)
	}
	for _, msg := range []tea.Msg{
		tea.KeyPressMsg{Code: tea.KeyHome},
		tea.KeyPressMsg{Code: tea.KeyRight},
		tea.KeyPressMsg{Code: tea.KeyRight},
		tea.KeyPressMsg{Code: tea.KeyLeft},
		tea.KeyPressMsg{Code: 'X', Text: "X"},
	} {
		m, _ := picker.Update(msg)
		picker = m.(Model)
	}
	if got := picker.SearchInput.Value(); got != "hXeart hjkl" {
		t.Fatalf("cursor editing failed: %q", got)
	}
	m, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.(Model).Focus == FocusSearch {
		t.Fatal("Tab must still leave search")
	}
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Escape must still cancel")
	}
	if _, ok := cmd().(EmojiCanceledMsg); !ok {
		t.Fatal("Escape emitted wrong message")
	}
}

func TestPickerIgnoresKeyReleases(t *testing.T) {
	for _, code := range []rune{tea.KeyRight, tea.KeyDown, tea.KeyTab, tea.KeyEnter, tea.KeyEscape, 'h', 'j', 'k', 'l'} {
		picker := New()
		picker.Focus = FocusGrid
		before := picker.View().Content
		m, cmd := picker.Update(tea.KeyReleaseMsg{Code: code})
		if cmd != nil || m.(Model).View().Content != before || m.(Model).Focus != picker.Focus {
			t.Fatalf("key release %d changed picker", code)
		}
	}
}

func TestInputPasteAndValueSemantics(t *testing.T) {
	input := NewTextInput()
	input.SetValue("abc")
	input.Focus()
	input, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	updated, _ := input.Update(tea.PasteMsg{Content: "👩‍💻\r\n\t\x1b[31mred\x1b[0m\x00"})
	if got := updated.Value(); got != "ab👩‍💻  redc" {
		t.Fatalf("paste result %q", got)
	}
	if input.Value() != "abc" {
		t.Fatal("paste mutated the original model")
	}
	updated, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if input.Value() != "abc" || updated.Value() != "ac" {
		t.Fatal("deletion must preserve the original model")
	}
	input.Blur()
	updated, _ = input.Update(tea.PasteMsg{Content: "ignored"})
	if updated.Value() != input.Value() {
		t.Fatal("unfocused paste changed input")
	}
}

func TestInputWidthAndCursorViewport(t *testing.T) {
	input := NewTextInput()
	input.SetValue("abcdefghi")
	input.SetWidth(5)
	input.Focus()
	if got := ansi.Strip(input.View()); got != "fghi " {
		t.Fatalf("end cursor not visible: %q", got)
	}
	input, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if got := ansi.Strip(input.View()); got != "abcde" {
		t.Fatalf("home cursor not visible: %q", got)
	}
	for _, width := range []int{-1, 0, 1, 3, 5, 15} {
		for _, value := range []string{"", "a very long query", "👩‍💻👍🏽🇺🇸e\u0301"} {
			input.SetWidth(width)
			input.Prompt = "🔍 "
			input.Placeholder = "Long placeholder"
			input.SetValue(value)
			if got := lipgloss.Width(input.View()); got > max(0, width) {
				t.Fatalf("width %d, value %q rendered %d cells", width, value, got)
			}
		}
	}
}

func TestInputEditsWholeGraphemes(t *testing.T) {
	for _, cluster := range []string{"👩‍💻", "👍🏽", "🇺🇸", "e\u0301"} {
		input := NewTextInput()
		input.SetValue("a" + cluster + "b")
		input.Focus()
		input, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		deleted, _ := input.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if deleted.Value() != "ab" {
			t.Fatalf("backspace split %q: %q", cluster, deleted.Value())
		}
		input, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		deleted, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
		if deleted.Value() != "ab" {
			t.Fatalf("delete split %q: %q", cluster, deleted.Value())
		}
	}
}

func TestSearchPasteFitsPickerAndResetsGrid(t *testing.T) {
	picker := New(WithShowSearch(true))
	picker.GridCursorX, picker.GridCursorY = 2, 2
	before, _ := picker.ViewSize()
	m, _ := picker.Update(tea.PasteMsg{Content: strings.Repeat("heart ", 30)})
	picker = m.(Model)
	if picker.SearchInput.Value() == "" || picker.GridCursorX != 0 || picker.GridCursorY != 0 {
		t.Fatal("paste did not update search and reset selection")
	}
	after, _ := picker.ViewSize()
	if before != after {
		t.Fatalf("long search expanded picker from %d to %d", before, after)
	}
	// Paste must also reach a search field inside an open swatch.
	swatch := NewSwatchPicker("😀", "test")
	swatch.open = true
	swatch.picker = New(WithShowSearch(true))
	swatch, _ = swatch.Update(tea.PasteMsg{Content: "heart"})
	if swatch.picker.SearchInput.Value() != "heart" {
		t.Fatal("swatch did not forward paste")
	}
}

func TestNarrowInputKeepsCursorGraphemeVisible(t *testing.T) {
	input := NewTextInput()
	input.Prompt = "> "
	input.SetValue("👩‍💻x")
	input.Focus()
	input, _ = input.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	for _, width := range []int{2, 3, 4, 5} {
		input.SetWidth(width)
		view := ansi.Strip(input.View())
		if !strings.Contains(view, "👩‍💻") || lipgloss.Width(view) > width {
			t.Fatalf("width %d: %q", width, view)
		}
	}
	input.SetWidth(1)
	if got := ansi.Strip(input.View()); got != "…" {
		t.Fatalf("want clipping indicator, got %q", got)
	}
	if input.Value() != "👩‍💻x" {
		t.Fatal("render changed input")
	}
}
