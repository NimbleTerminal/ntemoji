package demorun

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/go-booba/actions"
	"github.com/NimbleMarkets/go-booba/actions/bubblehelp"
)

// Description is the current set of controls advertised to the browser.
type Description = bubblehelp.Description

type actionModel interface {
	DemoActions() Description
}

type helpKeys struct {
	short []key.Binding
	full  [][]key.Binding
}

func (k helpKeys) ShortHelp() []key.Binding  { return k.short }
func (k helpKeys) FullHelp() [][]key.Binding { return k.full }

func binding(k, label string) key.Binding {
	return key.NewBinding(key.WithKeys(k), key.WithHelp(k, label))
}

// PickerActions describes ordinary keyboard controls shared by both pickers.
func PickerActions() Description {
	selectEmoji := binding("enter", "Select")
	next := binding("tab", "Next section")
	cancel := binding("esc", "Cancel")
	return Description{
		Context: "picker",
		Status:  actions.ProgramStatus{App: "ntemoji", Summary: "Choose an emoji"},
		Keys: helpKeys{
			short: []key.Binding{selectEmoji, next, cancel},
			full: [][]key.Binding{
				{binding("left", "Left"), binding("right", "Right"), binding("up", "Up"), binding("down", "Down")},
				{selectEmoji, next, binding("shift+tab", "Previous section"), cancel, binding("ctrl+c", "Quit")},
			},
		},
	}
}

// SwatchActions offers shortcuts only when no picker is open.
func SwatchActions() Description {
	open := []key.Binding{binding("1", "Launcher"), binding("2", "Hotness"), binding("3", "Celebration"), binding("4", "Idea")}
	return Description{
		Context: "swatches",
		Status:  actions.ProgramStatus{App: "ntemoji", Summary: "Open an emoji swatch"},
		Keys:    helpKeys{short: open, full: [][]key.Binding{open, {binding("ctrl+c", "Quit")}}},
	}
}

func withActions(model tea.Model) (*bubblehelp.Model, error) {
	return bubblehelp.NewModel(model, func(inner tea.Model) Description {
		return inner.(actionModel).DemoActions()
	}, bubblehelp.ModelOptions{OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
}

func unwrap(model tea.Model) tea.Model {
	if wrapped, ok := model.(*bubblehelp.Model); ok {
		return wrapped.Inner
	}
	return model
}
