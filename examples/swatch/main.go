// Run with: go run ./examples/swatch
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"nimbleterminal.dev/ntemoji"
	"nimbleterminal.dev/ntemoji/internal/demorun"
)

const (
	labelLen = 10 // e.g., "Emoji 1: " length
	gap      = 4  // spaces between column 1 and column 2
)

func main() {
	final, err := demorun.Run(newApp())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if app, ok := final.(*appModel); ok && app.quitting {
		fmt.Println("Final Emojis:")
		fmt.Printf("  1: %s\n", app.swatches[0].Emoji())
		fmt.Printf("  2: %s\n", app.swatches[1].Emoji())
		fmt.Printf("  3: %s\n", app.swatches[2].Emoji())
		fmt.Printf("  4: %s\n", app.swatches[3].Emoji())
	}
}

type appModel struct {
	width    int
	height   int
	swatches [4]*ntemoji.SwatchPicker
	zm       *zone.Manager
	quitting bool
}

func newApp() *appModel {
	a := &appModel{
		zm: zone.New(),
		swatches: [4]*ntemoji.SwatchPicker{
			ntemoji.NewSwatchPicker("🚀", "Launcher"),
			ntemoji.NewSwatchPicker("🔥", "Hotness"),
			ntemoji.NewSwatchPicker("🎉", "Celebration"),
			ntemoji.NewSwatchPicker("💡", "Idea"),
		},
	}

	// Configure preset options for the modal pickers
	for _, s := range a.swatches {
		s.SetZoneManager(a.zm)
		s.SetPickerOptions(
			ntemoji.WithPresets([]string{"😀", "👍", "🔥", "✅", "❌", "🚨"}),
			ntemoji.WithShowSearch(true),
		)
	}
	return a
}

func (a *appModel) Init() tea.Cmd {
	return nil
}

func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.width <= 0 {
			a.width = 60
		}
		if a.height <= 0 {
			a.height = 20
		}

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			a.quitting = true
			return a, tea.Quit
		}

		// If no picker is open, pressing keys 1-4 opens the corresponding swatch picker
		anyOpen := false
		for _, s := range a.swatches {
			if s.Open() {
				anyOpen = true
				break
			}
		}

		if !anyOpen {
			if msg.String() == "q" {
				a.quitting = true
				return a, tea.Quit
			}
			var cmd tea.Cmd
			openPress := tea.MouseClickMsg{Button: tea.MouseLeft}
			switch msg.String() {
			case "1":
				a.swatches[0], cmd = a.swatches[0].Update(openPress)
				return a, cmd
			case "2":
				a.swatches[1], cmd = a.swatches[1].Update(openPress)
				return a, cmd
			case "3":
				a.swatches[2], cmd = a.swatches[2].Update(openPress)
				return a, cmd
			case "4":
				a.swatches[3], cmd = a.swatches[3].Update(openPress)
				return a, cmd
			}
		}
	}

	var cmd tea.Cmd
	// If a modal is open, route all keyboard/mouse messages solely to it
	openIdx := -1
	for i := range a.swatches {
		if a.swatches[i].Open() {
			openIdx = i
			break
		}
	}

	if openIdx >= 0 {
		a.swatches[openIdx], cmd = a.swatches[openIdx].Update(msg)
		return a, cmd
	}

	// Use bubblezone to route main-view mouse clicks to the correct swatch
	if m, ok := msg.(tea.MouseMsg); ok {
		mouse := m.Mouse()
		var isPress bool
		if _, ok := m.(tea.MouseClickMsg); ok {
			isPress = true
		}
		if isPress && mouse.Button == tea.MouseLeft {
			for i := range a.swatches {
				z := a.zm.Get(swatchZoneID(i))
				if z != nil && z.InBounds(m) {
					a.swatches[i], cmd = a.swatches[i].Update(msg)
					return a, cmd
				}
			}
			return a, nil
		}
	}

	// For all other messages, pass them to all swatches
	for i := range a.swatches {
		var c tea.Cmd
		a.swatches[i], c = a.swatches[i].Update(msg)
		if c != nil {
			cmd = c
		}
	}
	return a, cmd
}

func swatchZoneID(i int) string {
	return "swatch-" + strconv.Itoa(i)
}

func (a *appModel) View() tea.View {
	if a.width <= 0 {
		a.width = 60
	}
	if a.height <= 0 {
		a.height = 20
	}

	mainView := a.buildMainView()
	for _, s := range a.swatches {
		mainView = s.ViewWithOverlay(mainView, a.width, a.height)
	}
	v := tea.NewView(a.zm.Scan(mainView))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func (a *appModel) buildMainView() string {
	title := lipgloss.NewStyle().Bold(true).Render("Click an emoji swatch to change it, or press keys 1-4")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("q: quit")

	sw, sh := a.swatches[0].Size()
	col1 := labelLen
	col2 := col1 + sw + gap + labelLen

	// Position swatches in 2x2 grid
	a.swatches[0].SetBounds(2, col1, sw, sh)
	a.swatches[1].SetBounds(2, col2, sw, sh)
	a.swatches[2].SetBounds(3, col1, sw, sh)
	a.swatches[3].SetBounds(3, col2, sw, sh)

	label0 := fmt.Sprintf("%-*s", labelLen, "Emoji 1:")
	label1 := fmt.Sprintf("%-*s", labelLen, "Emoji 2:")
	label2 := fmt.Sprintf("%-*s", labelLen, "Emoji 3:")
	label3 := fmt.Sprintf("%-*s", labelLen, "Emoji 4:")

	row1 := a.zm.Mark(swatchZoneID(0), label0+a.swatches[0].SwatchView()) +
		strings.Repeat(" ", gap) +
		a.zm.Mark(swatchZoneID(1), label1+a.swatches[1].SwatchView())

	row2 := a.zm.Mark(swatchZoneID(2), label2+a.swatches[2].SwatchView()) +
		strings.Repeat(" ", gap) +
		a.zm.Mark(swatchZoneID(3), label3+a.swatches[3].SwatchView())

	lines := []string{title, "", row1, row2, "", help}
	return strings.Join(lines, "\n")
}

func (a *appModel) DemoActions() demorun.Description {
	for _, swatch := range a.swatches {
		if swatch.Open() {
			return demorun.PickerActions()
		}
	}
	return demorun.SwatchActions()
}
