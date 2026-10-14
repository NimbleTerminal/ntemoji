package ntemoji

// AccessibilityState is a plain, comparable snapshot for host accessibility
// adapters. Names are localized CLDR labels (unknown emoji fall back to the glyph).
// Reading it does not move focus or confirm a selection.
type AccessibilityState struct {
	Focus         Focus
	FocusedEmoji  string // Empty when focus is on search or a category.
	FocusedName   string // Emoji name, category name, or "Search".
	Position      int    // One-based position within the focused section; zero if absent.
	Count         int    // Items in that section; zero for search.
	Category      string // Palette of the focused item or active category; empty if unknown.
	Query         string
	ResultCount   int
	SelectedEmoji string // Configured initial value or last confirmed selection.
	SelectedName  string
}

// AccessibilityState returns the full names even when the view truncates them.
// Hosts can compare snapshots after Update and announce meaningful changes.
// This API does not itself send screen-reader announcements.
func (m Model) AccessibilityState() AccessibilityState {
	return m.accessibilityState(m.filteredEmojis())
}

func (m Model) accessibilityState(emojis []string) AccessibilityState {
	s := AccessibilityState{Focus: m.Focus, ResultCount: len(emojis),
		SelectedEmoji: m.Selected, SelectedName: EmojiNameFor(m.Selected, m.locale)}
	if m.ShowSearch {
		s.Query = m.SearchInput.Value()
	}
	if s.Query == "" && m.ActiveTab >= 0 && m.ActiveTab < len(m.Palettes) {
		s.Category = m.Palettes[m.ActiveTab].Name
	}
	switch m.Focus {
	case FocusPresets:
		s.Count = len(m.Presets)
		if m.PresetCursor >= 0 && m.PresetCursor < s.Count {
			s.Position, s.FocusedEmoji = m.PresetCursor+1, m.Presets[m.PresetCursor]
		}
	case FocusGrid:
		s.Count = s.ResultCount
		s.FocusedEmoji = m.selectedEmoji(emojis)
		if s.FocusedEmoji != "" {
			s.Position = m.GridCursorY*m.resultCols() + m.GridCursorX + 1
		}
	case FocusCategories:
		s.Count = len(m.Palettes)
		if m.ActiveTab >= 0 && m.ActiveTab < s.Count {
			s.Position = m.ActiveTab + 1
			s.FocusedName = m.Palettes[m.ActiveTab].Name
			s.Category = s.FocusedName
		}
	case FocusSearch:
		s.FocusedName = "Search"
	}
	if s.FocusedEmoji != "" {
		s.FocusedName = EmojiNameFor(s.FocusedEmoji, m.locale)
		if m.Focus == FocusPresets || s.Query != "" {
			s.Category = ""
			for _, p := range m.Palettes {
				for _, e := range p.Emojis {
					if e == s.FocusedEmoji {
						s.Category = p.Name
						return s
					}
				}
			}
		}
	}
	return s
}

// SwatchAccessibilityState describes the swatch control and its open picker.
type SwatchAccessibilityState struct {
	Label    string
	Emoji    string
	Name     string
	Focused  bool
	Expanded bool
	Picker   AccessibilityState // Zero value while closed.
}

// AccessibilityState includes the label supplied to NewSwatchPicker.
func (s *SwatchPicker) AccessibilityState() SwatchAccessibilityState {
	state := SwatchAccessibilityState{Label: s.label, Emoji: s.emoji,
		Name: EmojiNameFor(s.emoji, s.locale), Focused: s.focused, Expanded: s.open}
	if s.open {
		state.Picker = s.picker.AccessibilityState()
	}
	return state
}
