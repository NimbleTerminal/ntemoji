package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntemoji"
)

func TestKeyboardOpenSelectsOnFirstMouseRelease(t *testing.T) {
	for _, key := range []rune{'1', '2', '3', '4'} {
		t.Run(string(key), func(t *testing.T) {
			a := newApp()
			defer a.zm.Close()
			a.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			a.Update(tea.KeyPressMsg{Code: key})
			if !a.swatches[int(key-'1')].Open() {
				t.Fatal("key did not open picker")
			}
			a.View()
			deadline := time.Now().Add(time.Second)
			for a.zm.Get(ntemoji.ZoneGrid) == nil && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			grid := a.zm.Get(ntemoji.ZoneGrid)
			if grid == nil {
				t.Fatal("missing grid zone")
			}
			a.Update(tea.MouseClickMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
			_, cmd := a.Update(tea.MouseReleaseMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
			if cmd == nil {
				t.Fatal("first click swallowed after keyboard open")
			}
			selected, ok := cmd().(ntemoji.EmojiChangedMsg)
			if !ok || selected.Emoji == "" {
				t.Fatal("click did not select emoji")
			}
		})
	}
}
