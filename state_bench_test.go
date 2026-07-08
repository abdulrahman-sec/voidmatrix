package main

import (
	"testing"
)

func BenchmarkStateTickAndSnapshot(b *testing.B) {
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
	charPool := ParseCharList("", nil)
	s := NewState(80, 24, 1.0, 0.72, ThemeGreen, charPool, "Mixed", cfg)

	// Warm up
	for i := 0; i < 100; i++ {
		s.Tick()
		_ = s.Snapshot()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Tick()
		snap := s.Snapshot()
		_ = snap
	}
}
