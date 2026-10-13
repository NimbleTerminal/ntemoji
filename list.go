package ntemoji

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Result columns are independent of the width budget in list mode.
func (m Model) resultCols() int {
	if m.listMode {
		return 1
	}
	cols, _ := m.gridDims()
	return cols
}

func fitLabel(text string, width int) string {
	text = ansi.Truncate(text, width, "…")
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

// Pack complete text tabs into rows before inserting zone markers.
func (m Model) textCategories(width int) string {
	var out strings.Builder
	used := 0
	for i, p := range m.Palettes {
		prefix := "  "
		if i == m.ActiveTab {
			prefix = "* "
			if m.Focus == FocusCategories {
				prefix = "> "
			}
		}
		label := ansi.Truncate(prefix+p.Name+" ", width, "…")
		w := lipgloss.Width(label)
		if used > 0 && used+w > width {
			out.WriteString(strings.Repeat(" ", width-used) + "\n")
			used = 0
		}
		style := m.Theme.InactiveTabStyle
		if i == m.ActiveTab {
			style = m.Theme.ActiveTabStyle
		}
		label = style.Render(label)
		if m.zm != nil {
			label = m.zm.Mark(fmt.Sprintf("picker-tab-%d", i), label)
		}
		out.WriteString(label)
		used += w
	}
	out.WriteString(strings.Repeat(" ", width-used))
	return out.String()
}

func (m Model) labeledResults(width, rows int) string {
	emojis := m.filteredEmojis()
	lines := make([]string, 0, rows*2-1)
	for r := 0; r < rows; r++ {
		idx := m.gridScrollOffset() + r
		line := strings.Repeat(" ", width)
		if idx < len(emojis) {
			e := emojis[idx]
			prefix := "  "
			style := m.Theme.UnselectedCellStyle
			if idx == m.GridCursorY && m.Focus == FocusGrid {
				prefix = "> "
				style = m.Theme.SelectedCellStyle
			}
			line = style.Render(fitLabel(prefix+e+" "+EmojiNameFor(e, m.locale), width))
		}
		lines = append(lines, line)
		if r < rows-1 {
			lines = append(lines, strings.Repeat(" ", width))
		}
	}
	return strings.Join(lines, "\n")
}
