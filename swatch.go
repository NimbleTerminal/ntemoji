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

	locale     string
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
	cfg := Config{}
	for _, option := range s.pickerOpts {
		option(&cfg)
	}
	s.locale = cfg.Locale
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

// OpenPicker opens the modal for keyboard or programmatic activation.
// Forward WindowSizeMsg through Update before opening to fit the available space.
// Mouse selection inside the modal requires a zone manager and scanning the view.
func (s *SwatchPicker) OpenPicker() tea.Cmd {
	if s.open {
		return nil
	}
	s.picker = s.newPickerModel()
	updated, cmd := s.picker.Update(tea.WindowSizeMsg{Width: s.lastViewWidth, Height: s.lastViewHeight})
	s.picker = updated.(Model)
	s.open = true
	s.ignoreNextRelease = false
	s.updateOverlayBounds()
	return cmd
}

func (s *SwatchPicker) updateOverlayBounds() {
	if !s.open {
		return
	}
	s.lastModalW, s.lastOverlayHeight = s.picker.ViewSize()
	s.lastOverlayTop, s.lastOverlayLeft = Fixed(s.row+s.h/2-s.lastOverlayHeight/2, s.col+s.w/2-s.lastModalW/2).
		ClampedOrigin(s.lastModalW, s.lastOverlayHeight, s.lastViewWidth, s.lastViewHeight)
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
	defer s.updateOverlayBounds()
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
// WindowSizeMsg controls picker size; rendering records the actual hit rectangle.
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
	s.lastOverlayTop, s.lastOverlayLeft = topPad, leftPad
	s.lastModalW = min(modalW, max(0, viewWidth-leftPad))
	s.lastOverlayHeight = min(overlayHeight, max(0, viewHeight-topPad))
	return OverlayView(mainView, modalContent, viewWidth, viewHeight, topPad, leftPad)
}

// Update forwards tea.Msg events to the swatch picker.
func (s *SwatchPicker) Update(msg tea.Msg) (*SwatchPicker, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.lastViewWidth = m.Width
		s.lastViewHeight = m.Height
		if s.open {
			picker, cmd := s.picker.Update(m)
			s.picker = picker.(Model)
			s.updateOverlayBounds()
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
			return s, nil
		}

		if isPress && mouse.Button == tea.MouseLeft {
			inBounds := s.zoneManager != nil ||
				(mouse.X >= s.col && mouse.X < s.col+s.w && mouse.Y >= s.row && mouse.Y < s.row+s.h)
			if inBounds {
				next := *s
				cmd := next.OpenPicker()
				next.ignoreNextRelease = true

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
