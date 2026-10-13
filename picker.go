package ntemoji

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

// Focus indicates which part of the picker is focused.
type Focus int

const (
	FocusPresets Focus = iota
	FocusSearch
	FocusCategories
	FocusGrid
)

const (
	ZonePresets    = "picker-presets"
	ZoneCategories = "picker-categories"
	ZoneGrid       = "picker-grid"
	ZoneSearch     = "picker-search"
)

// Model represents the emoji picker component state.
type Model struct {
	Palettes     []Palette
	Presets      []string
	ActiveTab    int
	GridCursorX  int
	GridCursorY  int
	PresetCursor int
	Focus        Focus
	SearchInput  TextInput
	ShowSearch   bool
	AutoDismiss  bool
	Selected     string

	Width  int
	Height int

	GridCols int
	GridRows int

	zm *zone.Manager

	Theme       Theme
	frameStyle  lipgloss.Style
	customFrame bool
}

// New constructs a new EmojiPicker Model with the given options.
func New(opts ...Option) Model {
	cfg := Config{
		Palettes: DefaultPalettes,
		Theme:    DefaultTheme,
	}
	for _, o := range opts {
		o(&cfg)
	}

	ti := NewTextInput()
	ti.Placeholder = "Search..."
	ti.Prompt = "🔍 "

	ti.PromptStyle = cfg.Theme.SearchPromptStyle
	ti.TextStyle = cfg.Theme.SearchTextStyle
	ti.PlaceholderStyle = cfg.Theme.SearchPlaceholderStyle
	ti.CursorStyle = cfg.Theme.SearchCursorStyle

	m := Model{
		Palettes:    cfg.Palettes,
		Presets:     cfg.Presets,
		Focus:       FocusGrid,
		ShowSearch:  cfg.ShowSearch,
		AutoDismiss: cfg.AutoDismiss,
		GridCols:    8,
		GridRows:    5,
		SearchInput: ti,

		Theme:       cfg.Theme,
		frameStyle:  cfg.FrameStyle,
		customFrame: cfg.CustomFrame,
	}

	if len(m.Presets) > 0 {
		m.Focus = FocusPresets
	} else if m.ShowSearch {
		m.Focus = FocusSearch
		m.SearchInput.Focus()
	} else {
		m.Focus = FocusCategories
	}

	if cfg.InitialEmoji != "" {
		m.Selected = cfg.InitialEmoji
		// Find in presets
		for i, p := range m.Presets {
			if p == cfg.InitialEmoji {
				m.PresetCursor = i
				m.Focus = FocusPresets
				return m
			}
		}
		// Find in palettes
		for pIdx, p := range m.Palettes {
			for eIdx, e := range p.Emojis {
				if e == cfg.InitialEmoji {
					m.ActiveTab = pIdx
					cols := m.GridCols
					m.GridCursorX = eIdx % cols
					m.GridCursorY = eIdx / cols
					m.Focus = FocusGrid
					return m
				}
			}
		}
	}

	return m
}

// SetZoneManager configures the zone manager for mouse hit testing.
func (m *Model) SetZoneManager(zm *zone.Manager) {
	m.zm = zm
}

// Value returns the currently selected emoji.
func (m Model) Value() string {
	if m.Selected != "" {
		return m.Selected
	}
	return m.SelectedEmoji()
}

// SelectedEmoji returns the emoji under the grid cursor.
func (m Model) SelectedEmoji() string {
	emojis := m.filteredEmojis()
	if len(emojis) == 0 {
		return ""
	}
	cols := m.GridCols
	if cols <= 0 {
		cols = 8
	}
	idx := m.GridCursorY*cols + m.GridCursorX
	if idx >= 0 && idx < len(emojis) {
		return emojis[idx]
	}
	return ""
}

// filteredEmojis returns the list of emojis after applying search filters.
func (m Model) filteredEmojis() []string {
	query := ""
	if m.ShowSearch {
		query = m.SearchInput.Value()
	}
	if query == "" {
		if m.ActiveTab >= 0 && m.ActiveTab < len(m.Palettes) {
			return m.Palettes[m.ActiveTab].Emojis
		}
		return nil
	}
	var filtered []string
	seen := make(map[string]bool)
	for _, p := range m.Palettes {
		for _, e := range p.Emojis {
			if !seen[e] && MatchEmoji(e, query) {
				seen[e] = true
				filtered = append(filtered, e)
			}
		}
	}
	return filtered
}

// ViewSize returns the size of the picker (width, height in cells).
func (m Model) ViewSize() (width, height int) {
	return lipgloss.Size(m.View().Content)
}

func (m Model) cycleFocus(next bool) Focus {
	var sections []Focus
	if len(m.Presets) > 0 {
		sections = append(sections, FocusPresets)
	}
	if m.ShowSearch {
		sections = append(sections, FocusSearch)
	}
	sections = append(sections, FocusCategories, FocusGrid)

	idx := -1
	for i, f := range sections {
		if f == m.Focus {
			idx = i
			break
		}
	}
	if idx == -1 {
		return FocusGrid
	}

	if next {
		idx = (idx + 1) % len(sections)
	} else {
		idx = (idx - 1 + len(sections)) % len(sections)
	}
	return sections[idx]
}

func (m *Model) clampCursor(totalEmojis int) {
	if totalEmojis == 0 {
		m.GridCursorX = 0
		m.GridCursorY = 0
		return
	}

	cols := m.GridCols
	if cols <= 0 {
		cols = 8
	}

	totalRows := (totalEmojis + cols - 1) / cols

	if m.GridCursorY >= totalRows {
		m.GridCursorY = totalRows - 1
	}
	if m.GridCursorY < 0 {
		m.GridCursorY = 0
	}

	emojisInRow := cols
	if m.GridCursorY == totalRows-1 {
		rem := totalEmojis % cols
		if rem > 0 {
			emojisInRow = rem
		}
	}

	if m.GridCursorX >= emojisInRow {
		m.GridCursorX = emojisInRow - 1
	}
	if m.GridCursorX < 0 {
		m.GridCursorX = 0
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width <= 0 || msg.Height <= 0 {
			return m, nil
		}
		// Preserve the selected emoji when the terminal changes the column count.
		cols := m.GridCols
		if cols <= 0 {
			cols = 8
		}
		selectedIndex := m.GridCursorY*cols + m.GridCursorX
		m.Width = msg.Width
		m.Height = msg.Height
		// Measure the rendered chrome, including custom frame padding and
		// wrapped sections, rather than duplicating View's layout arithmetic.
		measure := m
		measure.zm = nil
		measure.GridCols, measure.GridRows = 12, 1
		for measure.GridCols > 1 {
			w, _ := measure.ViewSize()
			if w <= msg.Width {
				break
			}
			measure.GridCols--
		}
		measure.GridRows = 8
		for measure.GridRows > 1 {
			_, h := measure.ViewSize()
			if h <= msg.Height {
				break
			}
			measure.GridRows--
		}
		m.GridCols, m.GridRows = measure.GridCols, measure.GridRows

		m.GridCursorX = selectedIndex % m.GridCols
		m.GridCursorY = selectedIndex / m.GridCols
		m.clampCursor(len(m.filteredEmojis()))
		return m, nil

	case tea.KeyReleaseMsg:
		return m, nil

	case tea.KeyPressMsg:
		if m.Focus == FocusSearch && m.ShowSearch {
			switch msg.String() {
			case "tab", "shift+tab", "enter", "esc":
				// Picker-level focus, selection and cancellation shortcuts.
			default:
				return m.updateSearch(msg)
			}
		}
		switch msg.String() {
		case "enter":
			var chosen string
			if m.Focus == FocusPresets && len(m.Presets) > 0 {
				chosen = m.Presets[m.PresetCursor]
			} else {
				chosen = m.SelectedEmoji()
			}
			if chosen != "" {
				m.Selected = chosen
				return m, func() tea.Msg {
					return EmojiChangedMsg{Emoji: chosen, Dismiss: m.AutoDismiss}
				}
			}
			return m, nil

		case "esc":
			return m, func() tea.Msg {
				return EmojiCanceledMsg{}
			}

		case "tab":
			m.Focus = m.cycleFocus(true)
			if m.Focus == FocusSearch {
				cmd = m.SearchInput.Focus()
			} else {
				m.SearchInput.Blur()
			}
			return m, cmd

		case "shift+tab":
			m.Focus = m.cycleFocus(false)
			if m.Focus == FocusSearch {
				cmd = m.SearchInput.Focus()
			} else {
				m.SearchInput.Blur()
			}
			return m, cmd

		case "left", "h":
			switch m.Focus {
			case FocusPresets:
				if m.PresetCursor > 0 {
					m.PresetCursor--
				}
			case FocusCategories:
				if m.ActiveTab > 0 {
					m.ActiveTab--
					m.GridCursorX = 0
					m.GridCursorY = 0
				}
			case FocusGrid:
				m.GridCursorX--
				m.clampCursor(len(m.filteredEmojis()))
			}
			return m, nil

		case "right", "l":
			switch m.Focus {
			case FocusPresets:
				if m.PresetCursor < len(m.Presets)-1 {
					m.PresetCursor++
				}
			case FocusCategories:
				if m.ActiveTab < len(m.Palettes)-1 {
					m.ActiveTab++
					m.GridCursorX = 0
					m.GridCursorY = 0
				}
			case FocusGrid:
				m.GridCursorX++
				m.clampCursor(len(m.filteredEmojis()))
			}
			return m, nil

		case "up", "k":
			if m.Focus == FocusGrid {
				m.GridCursorY--
				m.clampCursor(len(m.filteredEmojis()))
			}
			return m, nil

		case "down", "j":
			if m.Focus == FocusGrid {
				m.GridCursorY++
				m.clampCursor(len(m.filteredEmojis()))
			}
			return m, nil
		}

	case tea.MouseMsg:
		if m.zm == nil {
			return m, nil
		}
		mouse := msg.Mouse()
		var isClick, isRelease, isMotion bool
		switch msg.(type) {
		case tea.MouseClickMsg:
			isClick = true
		case tea.MouseReleaseMsg:
			isClick = true
			isRelease = true
		case tea.MouseMotionMsg:
			isMotion = true
		}
		if !isClick && !isMotion {
			return m, nil
		}

		if len(m.Presets) > 0 {
			if z := m.zm.Get(ZonePresets); z != nil && z.InBounds(msg) {
				relX, _ := z.Pos(msg)
				cellW := 4
				idx := relX / cellW
				if idx >= 0 && idx < len(m.Presets) {
					m.PresetCursor = idx
					m.Focus = FocusPresets
					m.SearchInput.Blur()
					if isRelease && mouse.Button == tea.MouseLeft {
						chosen := m.Presets[idx]
						m.Selected = chosen
						return m, func() tea.Msg {
							return EmojiChangedMsg{Emoji: chosen, Dismiss: m.AutoDismiss}
						}
					}
				}
				return m, nil
			}
		}

		if m.ShowSearch {
			if z := m.zm.Get(ZoneSearch); z != nil && z.InBounds(msg) {
				m.Focus = FocusSearch
				cmd = m.SearchInput.Focus()
				return m, cmd
			}
		}

		for i := range m.Palettes {
			tabZone := fmt.Sprintf("picker-tab-%d", i)
			if zTab := m.zm.Get(tabZone); zTab != nil && zTab.InBounds(msg) {
				m.ActiveTab = i
				m.Focus = FocusCategories
				m.SearchInput.Blur()
				m.GridCursorX = 0
				m.GridCursorY = 0
				return m, nil
			}
		}

		if z := m.zm.Get(ZoneGrid); z != nil && z.InBounds(msg) {
			relX, relY := z.Pos(msg)
			cellX := relX / 4
			// Blank separator lines do not select an emoji.
			if relY%2 != 0 {
				return m, nil
			}
			rows := m.GridRows
			if rows <= 0 {
				rows = 5
			}
			scrollOffset := max(0, m.GridCursorY-rows+1)
			cellY := relY/2 + scrollOffset
			emojis := m.filteredEmojis()
			cols := m.GridCols
			if cols <= 0 {
				cols = 8
			}

			if cellX >= 0 && cellX < cols && cellY*cols+cellX < len(emojis) {
				m.GridCursorX = cellX
				m.GridCursorY = cellY
				m.Focus = FocusGrid
				m.SearchInput.Blur()
				m.clampCursor(len(emojis))

				if isRelease && mouse.Button == tea.MouseLeft {
					chosen := m.SelectedEmoji()
					if chosen != "" {
						m.Selected = chosen
						return m, func() tea.Msg {
							return EmojiChangedMsg{Emoji: chosen, Dismiss: m.AutoDismiss}
						}
					}
				}
			}
			return m, nil
		}
	}

	if m.Focus == FocusSearch && m.ShowSearch {
		return m.updateSearch(msg)
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	previous := m.SearchInput.Value()
	input, cmd := m.SearchInput.Update(msg)
	if input.Value() != previous {
		m.GridCursorX, m.GridCursorY = 0, 0
	}
	m.SearchInput = input
	return m, cmd
}

// View renders the emoji picker UI.
func (m Model) View() tea.View {
	cols := m.GridCols
	if cols <= 0 {
		cols = 8
	}
	rows := m.GridRows
	if rows <= 0 {
		rows = 5
	}

	// Focus borders
	focusBorderFocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(m.Theme.BorderFocusedColor)

	focusBorderUnfocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(m.Theme.BorderUnfocusedColor)

	// 1. Header (Title + Value Preview)
	headerStr := m.Theme.TitleStyle.Render("Pick an Emoji")
	activeVal := m.Value()
	if activeVal != "" {
		tagsStr := ""
		if tags, ok := EmojiTags[activeVal]; ok && len(tags) > 0 {
			tagsStr = fmt.Sprintf(" (%s)", strings.Join(tags[:min(2, len(tags))], ", "))
		}
		headerStr += fmt.Sprintf(": %s%s", activeVal, m.Theme.HelpStyle.Render(tagsStr))
	}

	// Helper to center text in the inner content area
	innerW := cols * 4
	outerW := innerW + 2
	trunc := lipgloss.NewStyle().MaxWidth(outerW)
	toCols := func(content string) string {
		content = trunc.Render(content)
		w := lipgloss.Width(content)
		if w < outerW {
			return content + strings.Repeat(" ", outerW-w)
		}
		return content
	}

	var sections []string
	sections = append(sections, toCols(headerStr))

	// 2. Presets Row
	if len(m.Presets) > 0 {
		presetLine := ""
		for i, p := range m.Presets {
			cell := fmt.Sprintf(" %s ", p)
			if m.Focus == FocusPresets && i == m.PresetCursor {
				presetLine += m.Theme.SelectedPresetStyle.Render(cell)
			} else {
				presetLine += m.Theme.UnselectedPresetStyle.Render(cell)
			}
		}
		presetsBlock := presetLine
		if m.zm != nil {
			presetsBlock = m.zm.Mark(ZonePresets, presetLine)
		}
		presetsBorder := focusBorderUnfocused
		if m.Focus == FocusPresets {
			presetsBorder = focusBorderFocused
		}
		sections = append(sections, presetsBorder.Width(outerW).Render(presetsBlock))
	}

	// 3. Search Box
	if m.ShowSearch {
		sbBorder := focusBorderUnfocused
		if m.Focus == FocusSearch {
			sbBorder = focusBorderFocused
		}
		search := m.SearchInput
		search.SetWidth(innerW)
		sbView := search.View()
		if m.zm != nil {
			sbView = m.zm.Mark(ZoneSearch, sbView)
		}
		sections = append(sections, sbBorder.Width(outerW).Render(sbView))
	}

	// 4. Categories Selector
	catLine := ""
	for i, p := range m.Palettes {
		tabText := fmt.Sprintf(" %s ", p.Icon)
		var style lipgloss.Style
		if i == m.ActiveTab {
			style = m.Theme.ActiveTabStyle
		} else {
			style = m.Theme.InactiveTabStyle
		}
		tabRendered := style.Render(tabText)
		if m.zm != nil {
			tabRendered = m.zm.Mark(fmt.Sprintf("picker-tab-%d", i), tabRendered)
		}
		catLine += tabRendered
	}
	catBlock := catLine
	if m.zm != nil {
		catBlock = m.zm.Mark(ZoneCategories, catLine)
	}
	catBorder := focusBorderUnfocused
	if m.Focus == FocusCategories {
		catBorder = focusBorderFocused
	}
	sections = append(sections, catBorder.Width(outerW).Render(catBlock))

	// 5. Grid View (filtered & scrolled)
	emojis := m.filteredEmojis()
	scrollOffset := 0
	if m.GridCursorY >= rows {
		scrollOffset = m.GridCursorY - rows + 1
	}

	cellStyle := lipgloss.NewStyle().Width(4).MaxWidth(4).Align(lipgloss.Center)
	gridStr := ""
	for r := 0; r < rows; r++ {
		rowIdx := r + scrollOffset
		line := ""
		for c := 0; c < cols; c++ {
			idx := rowIdx*cols + c
			if idx < len(emojis) {
				emojiCell := cellStyle.Render(emojis[idx])
				isSelected := c == m.GridCursorX && rowIdx == m.GridCursorY
				if isSelected && m.Focus == FocusGrid {
					line += m.Theme.SelectedCellStyle.Render(emojiCell)
				} else if isSelected {
					// selected but grid is not focused
					line += m.Theme.SelectedUnfocusedCellStyle.Render(emojiCell)
				} else {
					line += m.Theme.UnselectedCellStyle.Render(emojiCell)
				}
			} else {
				line += "    " // 4 spaces
			}
		}
		gridStr += line + "\n"
		if r < rows-1 {
			gridStr += strings.Repeat(" ", innerW) + "\n"
		}
	}
	gridBlock := strings.TrimSuffix(gridStr, "\n")
	if m.zm != nil {
		gridBlock = m.zm.Mark(ZoneGrid, gridBlock)
	}
	gridBorder := focusBorderUnfocused
	if m.Focus == FocusGrid {
		gridBorder = focusBorderFocused
	}
	sections = append(sections, gridBorder.Width(outerW).Render(gridBlock))

	// 6. Help Footer
	footer1 := m.Theme.HelpStyle.Render("↵ pick  ⎋ close  ⇥ switch focus")
	footer2 := m.Theme.HelpStyle.Render("←↑↓→ move  hjkl navigates grid")
	sections = append(sections, toCols(footer1))
	sections = append(sections, toCols(footer2))

	// Join all inner elements
	innerContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Outer frame border
	frame := lipgloss.NewStyle().
		BorderStyle(m.Theme.FrameBorder).
		BorderForeground(m.Theme.FrameBorderColor).
		Padding(1, 2)

	if m.customFrame {
		frame = m.frameStyle
	}

	return tea.NewView(frame.Render(innerContent))
}
