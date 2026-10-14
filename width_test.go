package ntemoji

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type Cell struct {
	Char  string
	Width int
	Style string
}

func parseLine(line string) []Cell {
	var cells []Cell
	runes := []rune(line)
	n := len(runes)

	activeStyle := ""

	for i := 0; i < n; i++ {
		r := runes[i]
		if r == '\x1b' {
			styleSeq := string(r)
			i++
			if i < n && runes[i] == '[' {
				styleSeq += "["
				i++
				for i < n {
					styleSeq += string(runes[i])
					if runes[i] >= 0x40 && runes[i] <= 0x7E {
						break
					}
					i++
				}
			} else {
				if i < n {
					styleSeq += string(runes[i])
				}
			}
			if styleSeq == "\x1b[0m" || styleSeq == "\x1b[m" {
				activeStyle = ""
			} else {
				activeStyle += styleSeq
			}
			continue
		}

		w := runewidth.RuneWidth(r)
		if w <= 0 {
			if len(cells) > 0 {
				cells[len(cells)-1].Char += string(r)
			}
			continue
		}

		cells = append(cells, Cell{
			Char:  string(r),
			Width: w,
			Style: activeStyle,
		})

		if w == 2 {
			cells = append(cells, Cell{
				Char:  "",
				Width: 0,
				Style: activeStyle,
			})
		}
	}

	return cells
}

func renderCells(cells []Cell) string {
	var sb strings.Builder
	lastStyle := ""

	for _, c := range cells {
		if c.Char == "" && c.Width == 0 {
			continue
		}

		if c.Style != lastStyle {
			if lastStyle != "" {
				sb.WriteString("\x1b[0m")
			}
			if c.Style != "" {
				sb.WriteString(c.Style)
			}
			lastStyle = c.Style
		}
		sb.WriteString(c.Char)
	}

	if lastStyle != "" {
		sb.WriteString("\x1b[0m")
	}

	return sb.String()
}

func customOverlayView(bgView, modalView string, bgW, bgH, top, left int) string {
	bgLines := strings.Split(bgView, "\n")
	fgLines := strings.Split(modalView, "\n")

	for len(bgLines) < bgH {
		bgLines = append(bgLines, "")
	}

	for i, fgLine := range fgLines {
		y := top + i
		if y < 0 || y >= len(bgLines) {
			continue
		}

		bgLine := bgLines[y]
		bgCells := parseLine(bgLine)
		fgCells := parseLine(fgLine)

		// Overwrite cells with fgCells at left offset
		for j, fgCell := range fgCells {
			targetIdx := left + j
			if targetIdx < 0 {
				continue
			}

			// Ensure bgCells is large enough
			for len(bgCells) <= targetIdx {
				bgCells = append(bgCells, Cell{Char: " ", Width: 1, Style: ""})
			}

			// Boundary check 1: if we are overwriting a placeholder cell, clear the main character
			if bgCells[targetIdx].Width == 0 && bgCells[targetIdx].Char == "" && targetIdx > 0 {
				if targetIdx-1 < left || targetIdx-1 >= left+len(fgCells) {
					bgCells[targetIdx-1] = Cell{Char: " ", Width: 1, Style: bgCells[targetIdx-1].Style}
				}
			}

			// Boundary check 2: if we are overwriting a double-width char, clear its placeholder
			if bgCells[targetIdx].Width == 2 && targetIdx+1 < len(bgCells) {
				if targetIdx+1 < left || targetIdx+1 >= left+len(fgCells) {
					bgCells[targetIdx+1] = Cell{Char: " ", Width: 1, Style: bgCells[targetIdx+1].Style}
				}
			}

			bgCells[targetIdx] = fgCell
		}

		bgLines[y] = renderCells(bgCells)
	}

	return strings.Join(bgLines, "\n")
}

func TestRenderedRowWidths(t *testing.T) {
	picker := New(
		WithInitialEmoji("🚀"),
		WithPresets([]string{"😀", "👍", "🔥", "✅", "❌", "🚨"}),
		WithShowSearch(true),
	)

	view := picker.View().Content
	lines := strings.Split(view, "\n")

	t.Logf("Total lines: %d", len(lines))
	for i, line := range lines {
		w := lipgloss.Width(line)
		t.Logf("Line %d (width %d): %q", i, w, line)
	}
}

func TestOverlayWithAnsi(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	bgLine := bgStyle.Render("012345678901234567890123456789")
	bgView := bgLine + "\n" + bgLine + "\n" + bgLine

	modalView := "XXXX\nXXXX"

	result := customOverlayView(bgView, modalView, 30, 3, 1, 5)
	resultLines := strings.Split(result, "\n")

	for i, line := range resultLines {
		w := lipgloss.Width(line)
		t.Logf("Result Line %d (visual width %d): %q", i, w, line)
		if w != 30 {
			t.Errorf("Expected line %d to have visual width 30, got %d", i, w)
		}
	}
}

func TestAllDefaultEmojisWidths(t *testing.T) {
	for _, p := range DefaultPalettes {
		for _, e := range p.Emojis {
			rw := runewidth.StringWidth(e)
			lw := lipgloss.Width(e)
			if lw != 2 {
				t.Errorf("Palette %s emoji %s: runewidth=%d, lipgloss=%d, bytes=%d", p.Name, e, rw, lw, len(e))
			}
		}
	}
}

func TestSearchW(t *testing.T) {
	picker := New(
		WithShowSearch(true),
	)
	picker.Focus = FocusSearch
	for _, char := range "w" {
		m, _ := picker.Update(tea.KeyPressMsg{Text: string(char), Code: char})
		picker = m.(Model)
	}

	filtered := picker.filteredEmojis()
	t.Logf("Filtered emojis count for 'w': %d", len(filtered))
	for i, e := range filtered {
		rw := runewidth.StringWidth(e)
		lw := lipgloss.Width(e)
		t.Logf("  %d: %s (runewidth=%d, lipgloss=%d)", i, e, rw, lw)
	}
}

func TestExploreAlternativeWidths(t *testing.T) {
	alts := []string{
		"🌧", // cloud with rain without variation selector
		"☔", // umbrella with rain drops
		"💧", // droplet
		"❄", // snowflake without variation selector
		"⛄", // snowman without snow
		"☃", // snowman
		"🏔", // snow-capped mountain
		"🌬", // wind face
	}
	for _, e := range alts {
		rw := runewidth.StringWidth(e)
		lw := lipgloss.Width(e)
		t.Logf("Emoji %s: runewidth=%d, lipgloss=%d, bytes=%d", e, rw, lw, len(e))
	}
}

func TestExploreMoreAlternativeWidths(t *testing.T) {
	alts := []string{
		"🤞", // crossed fingers
		"🤘", // horn sign
		"👊", // oncoming fist
		"✊", // raised fist
		"🤝", // handshake
		"🌈", // rainbow
		"🍂", // fallen leaf
		"🍃", // leaf fluttering
		"🛠", // hammer and wrench
		"🧰", // toolbox
	}
	for _, e := range alts {
		rw := runewidth.StringWidth(e)
		lw := lipgloss.Width(e)
		t.Logf("Emoji %s: runewidth=%d, lipgloss=%d, bytes=%d", e, rw, lw, len(e))
	}
}
