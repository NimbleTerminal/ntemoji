package ntemoji

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

type overlayCell struct {
	Char   string
	Width  int
	Style  string
	Prefix string
	Suffix string
}

func parseOverlayLine(line string) []overlayCell {
	var cells []overlayCell
	activeStyle := ""
	pending := ""
	var state byte
	for len(line) > 0 {
		cluster, width, n, next := ansi.DecodeSequence(line, state, nil)
		if width > 0 {
			// DecodeSequence fast-paths ASCII; a digit may start a keycap
			// grapheme, so segment all printable text before measuring it.
			cluster, width = ansi.FirstGraphemeCluster(line, ansi.GraphemeWidth)
			n = len(cluster)
		}
		line, state = line[n:], next
		if width == 0 {
			// Only SGR changes persistent style. Zone markers and other
			// controls belong to this boundary and must never be replayed.
			if (strings.HasPrefix(cluster, "\x1b[") || strings.HasPrefix(cluster, "\x9b")) && strings.HasSuffix(cluster, "m") {
				if cluster == "\x1b[0m" || cluster == "\x1b[m" {
					activeStyle = ""
				} else {
					activeStyle += cluster
				}
			} else {
				pending += cluster
			}
			continue
		}
		cells = append(cells, overlayCell{Char: cluster, Width: width, Style: activeStyle, Prefix: pending})
		pending = ""
		for i := 1; i < width; i++ {
			cells = append(cells, overlayCell{Style: activeStyle})
		}
	}

	if pending != "" {
		if len(cells) == 0 {
			cells = append(cells, overlayCell{Prefix: pending})
		} else {
			cells[len(cells)-1].Suffix = pending
		}
	}
	return cells
}

func renderOverlayCells(cells []overlayCell) string {
	var sb strings.Builder
	lastStyle := ""

	for _, c := range cells {
		sb.WriteString(c.Prefix)
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
		sb.WriteString(c.Suffix)
	}

	if lastStyle != "" {
		sb.WriteString("\x1b[0m")
	}

	return sb.String()
}

// ModalCellSize calculates the cell-width and cell-height of the modal content string.
func ModalCellSize(s string) (width, height int) {
	lines := strings.Split(s, "\n")
	h := len(lines)
	w := 0
	for _, l := range lines {
		cells := parseOverlayLine(l)
		lw := 0
		for _, cell := range cells {
			lw += cell.Width
		}
		if lw > w {
			w = lw
		}
	}
	return w, h
}

// OverlayView overlays the modalView content on top of the bgView at coordinates (top, left).
func OverlayView(bgView, modalView string, bgW, bgH, top, left int) string {
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
		bgCells := parseOverlayLine(bgLine)
		fgCells := parseOverlayLine(fgLine)

		for j, fgCell := range fgCells {
			targetIdx := left + j
			if targetIdx < 0 {
				continue
			}

			// Ensure bgCells is large enough
			for len(bgCells) <= targetIdx {
				bgCells = append(bgCells, overlayCell{Char: " ", Width: 1, Style: ""})
			}

			// Boundary check 1: if we are overwriting a placeholder cell, clear the main character
			if bgCells[targetIdx].Width == 0 && bgCells[targetIdx].Char == "" && targetIdx > 0 {
				if targetIdx-1 < left || targetIdx-1 >= left+len(fgCells) {
					bgCells[targetIdx-1] = overlayCell{Char: " ", Width: 1, Style: bgCells[targetIdx-1].Style}
				}
			}

			// Boundary check 2: if we are overwriting a double-width char, clear its placeholder
			if bgCells[targetIdx].Width == 2 && targetIdx+1 < len(bgCells) {
				if targetIdx+1 < left || targetIdx+1 >= left+len(fgCells) {
					bgCells[targetIdx+1] = overlayCell{Char: " ", Width: 1, Style: bgCells[targetIdx+1].Style}
				}
			}

			if len(fgCells) == 1 && fgCell.Width == 0 {
				bgCells[targetIdx].Prefix = fgCell.Prefix + fgCell.Suffix + bgCells[targetIdx].Prefix
			} else {
				bgCells[targetIdx] = fgCell
			}
		}

		bgLines[y] = renderOverlayCells(bgCells)
	}

	return strings.Join(bgLines, "\n")
}

// Placement represents the position of the overlay.
type Placement struct {
	top  int
	left int
}

// Fixed returns a Placement configuration.
func Fixed(top, left int) Placement {
	return Placement{top: top, left: left}
}

// ClampedOrigin returns topPad, leftPad centering/clamping the overlay on the background space.
func (p Placement) ClampedOrigin(modalW, modalH, bgW, bgH int) (top, left int) {
	top = p.top
	left = p.left

	if top < 0 {
		top = 0
	}
	if top+modalH > bgH {
		top = bgH - modalH
	}
	if top < 0 {
		top = 0
	}

	if left < 0 {
		left = 0
	}
	if left+modalW > bgW {
		left = bgW - modalW
	}
	if left < 0 {
		left = 0
	}

	return top, left
}

// CellInModal returns true if the coordinate (x, y) falls inside the modal bounds.
func CellInModal(x, y, top, left, width, height int) bool {
	return x >= left && x < left+width && y >= top && y < top+height
}
