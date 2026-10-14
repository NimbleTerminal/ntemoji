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

	GridCols int // Grid columns; in list mode, controls width in four-cell units.
	GridRows int // Visible result rows in either mode.

	gridOffset int // First visible row; moving within the viewport does not scroll.

	zm *zone.Manager

	Theme       Theme
	frameStyle  lipgloss.Style
	customFrame bool
	locale      string
	listMode    bool
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
		listMode:    cfg.ListMode,
		locale:      cfg.Locale,
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
					cols := m.resultCols()
					m.GridCursorX = eIdx % cols
					m.GridCursorY = eIdx / cols
					m.Focus = FocusGrid
					m.SearchInput.Blur()
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

// gridDims normalizes unset dimensions for rendering and interaction.
func (m Model) gridDims() (cols, rows int) {
	cols, rows = m.GridCols, m.GridRows
	if cols <= 0 {
		cols = 8
	}
	if rows <= 0 {
		rows = 5
	}
	return
}

// gridScrollOffset keeps the cursor visible while preserving the viewport when
// the cursor moves within it, including through the exported cursor fields.
func (m Model) gridScrollOffset() int {
	_, rows := m.gridDims()
	offset := max(0, min(m.gridOffset, m.GridCursorY))
	return max(offset, m.GridCursorY-rows+1)
}

// SelectedEmoji returns the emoji under the grid cursor.
func (m Model) SelectedEmoji() string {
	return m.selectedEmoji(m.filteredEmojis())
}

func (m Model) selectedEmoji(emojis []string) string {
	if len(emojis) == 0 {
		return ""
	}
	cols := m.resultCols()
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
			if !seen[e] && MatchEmojiFor(e, query, m.locale) {
				seen[e] = true
				filtered = append(filtered, e)
			}
		}
	}
	return filtered
}

// ViewSize returns the size of the picker (width, height in cells).
func (m Model) ViewSize() (width, height int) {
	return m.viewSize(m.filteredEmojis())
}

func (m Model) viewSize(emojis []string) (width, height int) {
	return lipgloss.Size(m.view(emojis).Content)
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

	cols := m.resultCols()

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
	m.gridOffset = m.gridScrollOffset()
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width <= 0 || msg.Height <= 0 {
			return m, nil
		}
		// Preserve the selected emoji when the terminal changes the column count.
		cols := m.resultCols()
		selectedIndex := m.GridCursorY*cols + m.GridCursorX
		m.Width = msg.Width
		m.Height = msg.Height
		// Measure the rendered chrome, including custom frame padding and
		// wrapped sections, rather than duplicating View's layout arithmetic.
		emojis := m.filteredEmojis()
		measure := m
		measure.zm = nil
		measure.GridCols, measure.GridRows = 1, 1
		baseWidth, _ := measure.viewSize(emojis)
		measure.GridCols = max(1, min(12, 1+(msg.Width-baseWidth)/4))
		// Remeasure at the chosen width: presets and categories can wrap.
		_, baseHeight := measure.viewSize(emojis)
		measure.GridRows = max(1, min(8, 1+(msg.Height-baseHeight)/2))
		// Explicit frame dimensions can pad or clip content, so their size
		// is not linear in grid dimensions. Keep measured fitting for those.
		if m.customFrame && (m.frameStyle.GetWidth() > 0 || m.frameStyle.GetHeight() > 0 || m.frameStyle.GetMaxWidth() > 0 || m.frameStyle.GetMaxHeight() > 0) {
			measure.GridCols, measure.GridRows = 12, 1
			for measure.GridCols > 1 {
				w, _ := measure.viewSize(emojis)
				if w <= msg.Width {
					break
				}
				measure.GridCols--
			}
			measure.GridRows = 8
			for measure.GridRows > 1 {
				_, h := measure.viewSize(emojis)
				if h <= msg.Height {
					break
				}
				measure.GridRows--
			}
		}
		m.GridCols, m.GridRows = measure.GridCols, measure.GridRows

		m.GridCursorX = selectedIndex % m.resultCols()
		m.GridCursorY = selectedIndex / m.resultCols()
		m.gridOffset = 0
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
				relX, relY := z.Pos(msg)
				cols, _ := m.gridDims()
				idx := relY*cols + relX/4
				if relX >= 0 && relX < cols*4 && relY >= 0 && idx < len(m.Presets) {
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
			if m.listMode {
				cellX = 0
			}
			// Blank separator lines do not select an emoji.
			if relY%2 != 0 {
				return m, nil
			}
			cols := m.resultCols()
			cellY := relY/2 + m.gridOffset
			emojis := m.filteredEmojis()

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
	return m.view(m.filteredEmojis())
}

func (m Model) view(emojis []string) tea.View {
	cols, rows := m.gridDims()

	// Focus borders
	focusBorderFocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(m.Theme.BorderFocusedColor)

	focusBorderUnfocused := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).BorderRight(true).BorderTop(true).BorderBottom(true).
		BorderForeground(m.Theme.BorderUnfocusedColor)

	// The preview follows focus, independently of the committed selection.
	headerStr := m.Theme.TitleStyle.Render("Pick an Emoji")
	state := m.accessibilityState(emojis)
	preview := state.FocusedName
	if state.FocusedEmoji != "" {
		preview = state.FocusedEmoji + " " + preview
	}
	if state.Count > 0 {
		preview += fmt.Sprintf(" (%d/%d)", state.Position, state.Count)
	}
	if m.Focus == FocusSearch {
		preview = fmt.Sprintf("Search: %d results", state.ResultCount)
	} else if m.Focus == FocusGrid && state.ResultCount == 0 {
		preview = "No matching emoji"
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
	sections = append(sections, toCols(headerStr), toCols(m.Theme.HelpStyle.Render(preview)))

	cellStyle := lipgloss.NewStyle().Width(4).MaxWidth(4).Align(lipgloss.Center)

	// 2. Presets rows use the same fixed-width cells as their mouse targets.
	if len(m.Presets) > 0 {
		presetLine := ""
		for i, p := range m.Presets {
			if i > 0 && i%cols == 0 {
				presetLine += "\n"
			}
			cell := cellStyle.Render(p)
			if m.listMode && m.Focus == FocusPresets && i == m.PresetCursor {
				cell = cellStyle.Render(">" + p)
			}
			if m.Focus == FocusPresets && i == m.PresetCursor {
				presetLine += m.Theme.SelectedPresetStyle.Render(cell)
			} else {
				presetLine += m.Theme.UnselectedPresetStyle.Render(cell)
			}
		}
		// Keep the marked block rectangular, including a partial final row.
		if remainder := len(m.Presets) % cols; remainder != 0 {
			presetLine += strings.Repeat(" ", (cols-remainder)*4)
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
	if m.listMode {
		catLine = m.textCategories(innerW)
	} else {
		for i, p := range m.Palettes {
			if i > 0 && i%cols == 0 {
				catLine += "\n"
			}
			tabText := cellStyle.Render(p.Icon)
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
		if remainder := len(m.Palettes) % cols; remainder != 0 {
			catLine += strings.Repeat(" ", (cols-remainder)*4)
		}
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
	scrollOffset := m.gridScrollOffset()

	gridStr := ""
	if m.listMode {
		gridStr = m.labeledResults(emojis, innerW, rows)
	} else {
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
	help := "←↑↓→ move  hjkl navigates grid"
	if m.listMode {
		help = "↑↓ / jk results  ←→ tabs/presets"
	}
	footer2 := m.Theme.HelpStyle.Render(help)
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
