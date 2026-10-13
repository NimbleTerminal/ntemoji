// Run with: go run ./examples/simple
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"nimbleterminal.dev/ntemoji"
	"nimbleterminal.dev/ntemoji/internal/demorun"
)

func main() {
	zm := zone.New()
	// Define a custom forest/matrix theme with green accents
	forestTheme := ntemoji.DefaultTheme
	forestTheme.FrameBorderColor = lipgloss.Color("42")                                                                               // Green accent border
	forestTheme.BorderFocusedColor = lipgloss.Color("84")                                                                             // Bright green focus
	forestTheme.BorderUnfocusedColor = lipgloss.Color("22")                                                                           // Dark green unfocus
	forestTheme.ActiveTabStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("28"))     // Green active tab
	forestTheme.SelectedCellStyle = lipgloss.NewStyle().Background(lipgloss.Color("28")).Foreground(lipgloss.Color("255")).Bold(true) // Green active cell
	forestTheme.SearchPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	forestTheme.SearchCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	picker := ntemoji.New(
		ntemoji.WithInitialEmoji("🚀"),
		ntemoji.WithPresets([]string{"😀", "👍", "🔥", "✅", "❌", "🚨"}),
		ntemoji.WithShowSearch(true),
		ntemoji.WithListMode(os.Getenv("NTEMOJI_LIST") == "1"),
		ntemoji.WithLocale(os.Getenv("NTEMOJI_LOCALE")),
		ntemoji.WithTheme(forestTheme),
	)
	picker.SetZoneManager(zm)

	app := &simpleModel{picker: picker, zm: zm}
	model, err := demorun.Run(app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := model.(*simpleModel); ok && m.chosenEmoji != "" {
		fmt.Printf("Selected Emoji: %s\n", m.chosenEmoji)
	} else {
		fmt.Println("No emoji selected.")
	}
}

type simpleModel struct {
	picker      ntemoji.Model
	zm          *zone.Manager
	chosenEmoji string
}

func (m *simpleModel) Init() tea.Cmd {
	return nil
}

func (m *simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case ntemoji.EmojiChangedMsg:
		m.chosenEmoji = msg.Emoji
		return m, tea.Quit
	case ntemoji.EmojiCanceledMsg:
		return m, tea.Quit
	}

	updated, cmd := m.picker.Update(msg)
	m.picker = updated.(ntemoji.Model)
	return m, cmd
}

func (m *simpleModel) View() tea.View {
	// Center the picker view in terminal window
	pickerView := m.picker.View().Content

	// Lipgloss helper to align/center
	centered := lipgloss.Place(
		m.picker.Width, m.picker.Height,
		lipgloss.Center, lipgloss.Center,
		pickerView,
	)

	v := tea.NewView(m.zm.Scan(centered))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func (m *simpleModel) DemoActions() demorun.Description {
	description := demorun.PickerActions()
	state := m.picker.AccessibilityState()
	// Hosts can use this same snapshot in a status line or accessibility adapter.
	// Publishing browser action status alone does not announce it to a reader.
	description.Status.Summary = state.FocusedName
	if state.Count > 0 {
		description.Status.Summary += fmt.Sprintf(" (%d of %d)", state.Position, state.Count)
	} else if state.Focus == ntemoji.FocusSearch {
		description.Status.Summary = fmt.Sprintf("Search: %d results", state.ResultCount)
	}
	return description
}
