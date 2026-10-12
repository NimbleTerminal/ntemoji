//go:build js && wasm

package demorun

import (
	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/go-booba/wasm"
)

func Run(model tea.Model) (tea.Model, error) {
	wrapped, err := withActions(model)
	if err != nil {
		return model, err
	}
	final, err := wasm.NewProgram(wrapped).Run()
	return unwrap(final), err
}
