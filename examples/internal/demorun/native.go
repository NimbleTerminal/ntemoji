//go:build !js || !wasm

// Package demorun starts examples in a native terminal or the browser.
package demorun

import tea "charm.land/bubbletea/v2"

func Run(model tea.Model) (tea.Model, error) {
	wrapped, err := withActions(model)
	if err != nil {
		return model, err
	}
	final, err := tea.NewProgram(wrapped).Run()
	return unwrap(final), err
}
