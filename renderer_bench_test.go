package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func BenchmarkRendererDraw(b *testing.B) {
	cfg := Config{
		AsyncMode:   true,
		Flashers:    true,
		BoldMode:    BoldMixed,
		BgColor:     0,
		StatusOff:   true,
		IgnoreKeys:  true,
		SingleWave:  false,
		ExitAfter:   0,
		Wind:        "none",
		Splash:      true,
		Glitch:      true,
		Message:     "",
		Screensaver: false,
		Debug:       false, // Test with Debug OFF as reported by user
	}
	w, h := 80, 24
	charPool := ParseCharList("", nil)
	s := NewState(w, h, 1.0, 0.72, ThemeGreen, charPool, "Mixed", cfg)

	// Setup simulation screen
	simScreen := tcell.NewSimulationScreen("")
	simScreen.SetSize(w, h)
	if err := simScreen.Init(); err != nil {
		b.Fatalf("failed to init screen: %v", err)
	}
	simScreen.SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault))
	simScreen.Clear()
	simScreen.HideCursor()

	r := &Renderer{screen: simScreen}
	defer r.Fini()

	// Warm up
	for i := 0; i < 100; i++ {
		s.Tick()
		snap := s.Snapshot()
		r.Draw(snap)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Tick()
		snap := s.Snapshot()
		r.Draw(snap)
	}
}
