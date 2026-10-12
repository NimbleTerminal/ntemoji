package ntemoji

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

// SwatchPicker represents a clickable UI element (emoji + arrow)
// that opens the emoji picker overlay modal when activated.
type SwatchPicker struct {
	emoji   string
	label   string
	row     int
	col     int
	w       int
	h       int
	picker  Model
	open    bool
	focused bool

	ignoreNextRelease bool
	zoneManager       *zone.Manager

	lastOverlayLeft   int
	lastOverlayTop    int
	lastModalW        int
	lastOverlayHeight int
	lastViewWidth     int
	lastViewHeight    int

	pickerOpts []Option
}

const swatchPickerSymbol = "▼"

// NewSwatchPicker returns a new SwatchPicker instance.
func NewSwatchPicker(initialEmoji, label string) *SwatchPicker {
	if initialEmoji == "" {
		initialEmoji = "😀"
	}
	return &SwatchPicker{
		emoji: initialEmoji,
		label: label,
	}
}

// SetPickerOptions configures options that are passed to the EmojiPicker model when it opens.
func (s *SwatchPicker) SetPickerOptions(opts ...Option) {
	s.pickerOpts = append([]Option(nil), opts...)
}

func (s *SwatchPicker) newPickerModel() Model {
	opts := make([]Option, 0, 1+len(s.pickerOpts))
	opts = append(opts, WithInitialEmoji(s.emoji))
	opts = append(opts, s.pickerOpts...)
	p := New(opts...)
	if s.zoneManager != nil {
		p.SetZoneManager(s.zoneManager)
	}
	return p
}

// SwatchView renders the swatch preview (emoji + down arrow symbol).
func (s *SwatchPicker) SwatchView() string {
	emojiBlock := lipgloss.NewStyle().Render(s.emoji)
	symbol := swatchPickerSymbol
	if s.focused {
		symbol = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true).Render(symbol)
	}
	return emojiBlock + " " + symbol
}

// SetFocused toggles highlight focus status.
func (s *SwatchPicker) SetFocused(f bool) {
	s.focused = f
}

// Focused returns whether the swatch is currently focused.
func (s *SwatchPicker) Focused() bool {
	return s.focused
}

// Open returns whether the modal is currently displayed.
func (s *SwatchPicker) Open() bool {
	return s.open
}

// SetZoneManager configures the bubblezone manager.
func (s *SwatchPicker) SetZoneManager(zm *zone.Manager) {
	s.zoneManager = zm
}

// Size returns the display size of the swatch (width, height in cells).
func (s *SwatchPicker) Size() (width, height int) {
	return 4, 1
}

// SetBounds sets where the swatch is positioned in terminal cells.
func (s *SwatchPicker) SetBounds(row, col, w, h int) {
	s.row = row
	s.col = col
	if w <= 0 || h <= 0 {
		pw, ph := s.Size()
		if w <= 0 {
			s.w = pw
		} else {
			s.w = w
		}
		if h <= 0 {
			s.h = ph
		} else {
			s.h = h
		}
	} else {
		s.w = w
		s.h = h
	}
}

// SetEmoji sets the active emoji value.
func (s *SwatchPicker) SetEmoji(e string) {
	s.emoji = e
}

// Emoji returns the active emoji value.
func (s *SwatchPicker) Emoji() string {
	return s.emoji
}

// ViewWithOverlay renders the overlay modal on top of mainView when the picker is open.
func (s *SwatchPicker) ViewWithOverlay(mainView string, viewWidth, viewHeight int) string {
	if !s.open {
		return mainView
	}
	modalContent := s.picker.View().Content
	modalW, overlayHeight := ModalCellSize(modalContent)
	centerRow := s.row + s.h/2
	centerCol := s.col + s.w/2
	topPad, leftPad := Fixed(centerRow-overlayHeight/2, centerCol-modalW/2).
		ClampedOrigin(modalW, overlayHeight, viewWidth, viewHeight)
	s.lastOverlayLeft = leftPad
	s.lastOverlayTop = topPad
	s.lastModalW = modalW
	s.lastOverlayHeight = overlayHeight
	s.lastViewWidth = viewWidth
	s.lastViewHeight = viewHeight
	return OverlayView(mainView, modalContent, viewWidth, viewHeight, topPad, leftPad)
}

// Update forwards tea.Msg events to the swatch picker.
func (s *SwatchPicker) Update(msg tea.Msg) (*SwatchPicker, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.lastViewWidth = m.Width
		s.lastViewHeight = m.Height
		if s.open {
			pickerW, pickerH := s.picker.ViewSize()
			picker, cmd := s.picker.Update(tea.WindowSizeMsg{Width: pickerW, Height: pickerH})
			s.picker = picker.(Model)
			if s.lastOverlayHeight > 0 && s.lastModalW > 0 {
				centerRow := s.row + s.h/2
				centerCol := s.col + s.w/2
				topPad, leftPad := Fixed(centerRow-s.lastOverlayHeight/2, centerCol-s.lastModalW/2).
					ClampedOrigin(s.lastModalW, s.lastOverlayHeight, s.lastViewWidth, s.lastViewHeight)
				s.lastOverlayLeft = leftPad
				s.lastOverlayTop = topPad
			}
			return s, cmd
		}
		return s, nil

	case tea.KeyReleaseMsg:
		return s, nil

	case tea.KeyPressMsg:
		if s.open {
			updated, cmd := s.picker.Update(m)
			s.picker = updated.(Model)
			return s, cmd
		}
		return s, nil

	case tea.MouseMsg:
		mouse := m.Mouse()
		var isPress, isRelease bool
		switch m.(type) {
		case tea.MouseClickMsg:
			isPress = true
		case tea.MouseReleaseMsg:
			isRelease = true
		}

		if s.open {
			if s.ignoreNextRelease && isRelease && mouse.Button == tea.MouseLeft {
				next := *s
				next.ignoreNextRelease = false
				return &next, nil
			}
			leftPad := s.lastOverlayLeft
			topPad := s.lastOverlayTop
			if s.lastModalW <= 0 && s.lastViewWidth > 0 {
				pw, _ := s.picker.ViewSize()
				leftPad = max((s.lastViewWidth-pw)/2, 0)
			}
			if s.lastOverlayHeight <= 0 && s.lastViewHeight > 0 {
				_, ph := s.picker.ViewSize()
				topPad = max((s.lastViewHeight-ph)/2, 0)
			}
			inModal := CellInModal(mouse.X, mouse.Y, topPad, leftPad, s.lastModalW, s.lastOverlayHeight)
			if !inModal {
				if isPress && mouse.Button == tea.MouseLeft {
					next := *s
					next.open = false
					return &next, func() tea.Msg { return EmojiCanceledMsg{} }
				}
				return s, nil
			}
			if s.zoneManager != nil {
				updated, cmd := s.picker.Update(m)
				s.picker = updated.(Model)
				return s, cmd
			}
			relMsg := offsetMouseMsg(m, 2-leftPad, 1-topPad)
			updated, cmd := s.picker.Update(relMsg)
			s.picker = updated.(Model)
			return s, cmd
		}

		if isPress && mouse.Button == tea.MouseLeft {
			inBounds := s.zoneManager != nil ||
				(mouse.X >= s.col && mouse.X < s.col+s.w && mouse.Y >= s.row && mouse.Y < s.row+s.h)
			if inBounds {
				next := *s
				next.picker = next.newPickerModel()
				pickerW, pickerH := next.picker.ViewSize()
				picker, cmd := next.picker.Update(tea.WindowSizeMsg{Width: pickerW, Height: pickerH})
				next.picker = picker.(Model)
				next.open = true
				next.ignoreNextRelease = true

				modalW, overlayHeight := next.picker.ViewSize()
				centerRow := next.row + next.h/2
				centerCol := next.col + next.w/2
				topPad, leftPad := Fixed(centerRow-overlayHeight/2, centerCol-modalW/2).
					ClampedOrigin(modalW, overlayHeight, next.lastViewWidth, next.lastViewHeight)
				next.lastOverlayLeft = leftPad
				next.lastOverlayTop = topPad
				next.lastModalW = modalW
				next.lastOverlayHeight = overlayHeight
				return &next, cmd
			}
		}
		return s, nil

	case EmojiChangedMsg:
		if !s.open {
			return s, nil
		}
		next := *s
		next.emoji = m.Emoji
		next.open = false
		return &next, nil

	case EmojiCanceledMsg:
		if !s.open {
			return s, nil
		}
		next := *s
		next.open = false
		return &next, nil
	}

	if s.open {
		updated, cmd := s.picker.Update(msg)
		s.picker = updated.(Model)
		return s, cmd
	}
	return s, nil
}

// MouseToModalCoords converts screen coordinates to modal-relative coordinates.
func MouseToModalCoords(screenX, screenY, overlayLeft, overlayTop int) (relX, relY int) {
	relY = screenY - overlayTop + 1
	relX = screenX - overlayLeft + 2
	return relX, relY
}

func offsetMouseMsg(m tea.MouseMsg, dx, dy int) tea.MouseMsg {
	switch concrete := m.(type) {
	case tea.MouseClickMsg:
		return tea.MouseClickMsg{
			X:      concrete.X + dx,
			Y:      concrete.Y + dy,
			Button: concrete.Button,
			Mod:    concrete.Mod,
		}
	case tea.MouseReleaseMsg:
		return tea.MouseReleaseMsg{
			X:      concrete.X + dx,
			Y:      concrete.Y + dy,
			Button: concrete.Button,
			Mod:    concrete.Mod,
		}
	case tea.MouseMotionMsg:
		return tea.MouseMotionMsg{
			X:      concrete.X + dx,
			Y:      concrete.Y + dy,
			Button: concrete.Button,
			Mod:    concrete.Mod,
		}
	default:
		return m
	}
}
