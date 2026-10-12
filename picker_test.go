package ntemoji

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPickerInitialState(t *testing.T) {
	picker := New(
		WithInitialEmoji("😀"),
		WithPresets([]string{"😀", "👍"}),
		WithShowSearch(true),
	)

	if picker.Selected != "😀" {
		t.Errorf("Expected initial emoji to be 😀, got %s", picker.Selected)
	}

	if len(picker.Presets) != 2 {
		t.Errorf("Expected 2 presets, got %d", len(picker.Presets))
	}

	if picker.Focus != FocusPresets {
		t.Errorf("Expected initial focus to be FocusPresets, got %v", picker.Focus)
	}
}

func TestPickerFocusCycling(t *testing.T) {
	picker := New(
		WithPresets([]string{"😀"}),
		WithShowSearch(true),
	)

	// Focus sequence: Presets -> Search -> Categories -> Grid
	if picker.Focus != FocusPresets {
		t.Errorf("Expected FocusPresets, got %v", picker.Focus)
	}

	m, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	picker = m.(Model)
	if picker.Focus != FocusSearch {
		t.Errorf("Expected FocusSearch after Tab, got %v", picker.Focus)
	}

	m, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	picker = m.(Model)
	if picker.Focus != FocusCategories {
		t.Errorf("Expected FocusCategories after Tab, got %v", picker.Focus)
	}

	m, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	picker = m.(Model)
	if picker.Focus != FocusGrid {
		t.Errorf("Expected FocusGrid after Tab, got %v", picker.Focus)
	}

	m, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	picker = m.(Model)
	if picker.Focus != FocusPresets {
		t.Errorf("Expected FocusPresets after Tab, got %v", picker.Focus)
	}
}

func TestPickerGridNavigation(t *testing.T) {
	// Create a picker, force Grid focus
	picker := New()
	picker.Focus = FocusGrid
	picker.ActiveTab = 0 // Smileys
	picker.GridCursorX = 0
	picker.GridCursorY = 0

	// Check standard emojis lists are populated
	emojis := picker.filteredEmojis()
	if len(emojis) == 0 {
		t.Fatal("Expected default Smileys palette to contain emojis")
	}

	// Move right
	m, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	picker = m.(Model)
	if picker.GridCursorX != 1 {
		t.Errorf("Expected GridCursorX to be 1, got %d", picker.GridCursorX)
	}

	// Move down
	m, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	picker = m.(Model)
	if picker.GridCursorY != 1 {
		t.Errorf("Expected GridCursorY to be 1, got %d", picker.GridCursorY)
	}
}

func TestPickerSearch(t *testing.T) {
	picker := New(
		WithShowSearch(true),
	)
	picker.Focus = FocusSearch

	// Type "green"
	for _, char := range "green" {
		m, _ := picker.Update(tea.KeyPressMsg{Text: string(char), Code: char})
		picker = m.(Model)
	}

	if picker.SearchInput.Value() != "green" {
		t.Errorf("Expected search input value to be 'green', got '%s'", picker.SearchInput.Value())
	}

	// Filtered list should only contain emojis matching green (like 🟢, 🟩)
	filtered := picker.filteredEmojis()
	if len(filtered) == 0 {
		t.Error("Expected to find emojis matching query 'green'")
	}

	for _, e := range filtered {
		matched := MatchEmoji(e, "green")
		if !matched {
			t.Errorf("Emoji %s did not match query 'green'", e)
		}
	}
}

func TestPickerSelection(t *testing.T) {
	picker := New()
	picker.Focus = FocusGrid
	picker.GridCursorX = 0
	picker.GridCursorY = 0

	expectedEmoji := picker.SelectedEmoji()

	// Press Enter
	m, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	picker = m.(Model)

	if cmd == nil {
		t.Fatal("Expected non-nil cmd on Enter")
	}

	msg := cmd()
	changedMsg, ok := msg.(EmojiChangedMsg)
	if !ok {
		t.Fatalf("Expected EmojiChangedMsg, got %T", msg)
	}

	if changedMsg.Emoji != expectedEmoji {
		t.Errorf("Expected selected emoji to be %s, got %s", expectedEmoji, changedMsg.Emoji)
	}
}

func TestPickerCancellation(t *testing.T) {
	picker := New()
	m, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	picker = m.(Model)

	if cmd == nil {
		t.Fatal("Expected non-nil cmd on Esc")
	}

	msg := cmd()
	_, ok := msg.(EmojiCanceledMsg)
	if !ok {
		t.Fatalf("Expected EmojiCanceledMsg, got %T", msg)
	}
}

func TestPickerViewRender(t *testing.T) {
	picker := New()
	view := picker.View().Content
	if !strings.Contains(view, "Pick an Emoji") {
		t.Error("Expected view to contain header 'Pick an Emoji'")
	}
}

func TestPickerNewCategoriesAndConstants(t *testing.T) {
	// Verify constants exist and match expected values
	if EmojiCatFace != "🐱" {
		t.Errorf("Expected EmojiCatFace to be 🐱, got %s", EmojiCatFace)
	}
	if EmojiCat != "🐈" {
		t.Errorf("Expected EmojiCat to be 🐈, got %s", EmojiCat)
	}

	picker := New(WithShowSearch(true))

	// Test that we can search for cat and find both emojis
	picker.Focus = FocusSearch
	for _, char := range "cat" {
		m, _ := picker.Update(tea.KeyPressMsg{Text: string(char), Code: char})
		picker = m.(Model)
	}

	filtered := picker.filteredEmojis()
	foundFace := false
	foundWalk := false
	for _, e := range filtered {
		if e == EmojiCatFace {
			foundFace = true
		}
		if e == EmojiCat {
			foundWalk = true
		}
	}

	if !foundFace {
		t.Error("Expected to find EmojiCatFace in search results for 'cat'")
	}
	if !foundWalk {
		t.Error("Expected to find EmojiCat in search results for 'cat'")
	}
}

func TestPickerResizePreservesSelection(t *testing.T) {
	picker := New(WithInitialEmoji("🚀"))
	for _, width := range []int{100, 32, 80} {
		updated, _ := picker.Update(tea.WindowSizeMsg{Width: width, Height: 24})
		picker = updated.(Model)
		if got := picker.SelectedEmoji(); got != "🚀" {
			t.Fatalf("width %d: selected %q, want rocket", width, got)
		}
		_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if cmd == nil {
			t.Fatalf("width %d: Enter did not select an emoji", width)
		}
		if msg, ok := cmd().(EmojiChangedMsg); !ok || msg.Emoji != "🚀" {
			t.Fatalf("width %d: unexpected selection %#v", width, msg)
		}
	}
}
