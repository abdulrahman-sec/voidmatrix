// input.go — Keyboard input handling for voidmatrix
package main

import (
	"github.com/gdamore/tcell/v2"
)

// InputHandler reads keyboard events from the tcell screen.
type InputHandler struct {
	screen tcell.Screen
	state  *State
	quit   chan<- struct{}
	redraw chan<- struct{}
}

func NewInputHandler(screen tcell.Screen, state *State, quit chan<- struct{}, redraw chan<- struct{}) *InputHandler {
	return &InputHandler{screen: screen, state: state, quit: quit, redraw: redraw}
}

func (h *InputHandler) Run() {
	for {
		ev := h.screen.PollEvent()
		if ev == nil {
			h.signalQuit()
			return
		}
		switch e := ev.(type) {
		case *tcell.EventResize:
			w, ht := h.screen.Size()
			h.state.Resize(w, ht)
			h.triggerRedraw()
		case *tcell.EventKey:
			// Screensaver mode: exit immediately on any key press.
			if h.state.cfg.Screensaver {
				h.signalQuit()
				return
			}
			// Always allow Ctrl-C to quit even in ignore-input mode.
			if h.state.cfg.IgnoreKeys && e.Key() != tcell.KeyCtrlC {
				continue
			}
			if h.handleKey(e) {
				return
			}
			h.triggerRedraw()
		}
	}
}

func (h *InputHandler) handleKey(e *tcell.EventKey) bool {
	switch {
	// ── Quit ────────────────────────────────────────────────────────────────
	case e.Key() == tcell.KeyCtrlC,
		e.Key() == tcell.KeyEscape,
		e.Rune() == 'q', e.Rune() == 'Q',
		e.Rune() == ' ':
		h.signalQuit()
		return true

	// ── Speed up ────────────────────────────────────────────────────────────
	// Arrow UP, letter W (both cases), plus/equals
	case e.Key() == tcell.KeyUp,
		e.Rune() == 'w', e.Rune() == 'W',
		e.Rune() == '+', e.Rune() == '=':
		h.state.IncreaseSpeed()

	// ── Speed down ──────────────────────────────────────────────────────────
	// Arrow DOWN, letter S (both cases), minus/underscore
	case e.Key() == tcell.KeyDown,
		e.Rune() == 's', e.Rune() == 'S',
		e.Rune() == '-', e.Rune() == '_':
		h.state.DecreaseSpeed()

	// ── Density up ──────────────────────────────────────────────────────────
	// Arrow RIGHT, letter D (both cases)
	case e.Key() == tcell.KeyRight,
		e.Rune() == 'd', e.Rune() == 'D':
		h.state.IncreaseDensity()

	// ── Density down ────────────────────────────────────────────────────────
	// Arrow LEFT, letter A (both cases)
	case e.Key() == tcell.KeyLeft,
		e.Rune() == 'a', e.Rune() == 'A':
		h.state.DecreaseDensity()

	// ── Themes 1–9 ──────────────────────────────────────────────────────────
	case e.Rune() == '1':
		h.state.SetTheme(ThemeGreen)
	case e.Rune() == '2':
		h.state.SetTheme(ThemeRed)
	case e.Rune() == '3':
		h.state.SetTheme(ThemeBlue)
	case e.Rune() == '4':
		h.state.SetTheme(ThemeWhite)
	case e.Rune() == '5':
		h.state.SetTheme(ThemeRainbow)
	case e.Rune() == '6':
		h.state.SetTheme(ThemePurple)
	case e.Rune() == '7':
		h.state.SetTheme(ThemeCyan)
	case e.Rune() == '8':
		h.state.SetTheme(ThemeOrange)
	case e.Rune() == '9':
		h.state.SetTheme(ThemeGold)

	// ── Feature toggles (single unique letters, no conflicts) ────────────────
	// [ = async scroll toggle
	case e.Rune() == '[':
		h.state.ToggleAsync()

	// b = bold cycle (mixed → all → none)
	case e.Rune() == 'b', e.Rune() == 'B':
		h.state.CycleBold()

	// f = flashers toggle
	case e.Rune() == 'f', e.Rune() == 'F':
		h.state.ToggleFlashers()

	// o = OSD toggle
	case e.Rune() == 'o', e.Rune() == 'O':
		h.state.ToggleStatus()

	// c = cycle charset
	case e.Rune() == 'c', e.Rune() == 'C':
		h.state.CycleCharSet()
	}

	return false
}

func (h *InputHandler) signalQuit() {
	select {
	case h.quit <- struct{}{}:
	default:
	}
}

func (h *InputHandler) triggerRedraw() {
	select {
	case h.redraw <- struct{}{}:
	default:
	}
}
