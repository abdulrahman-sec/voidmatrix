// theme.go — Color theme definitions for voidmatrix
package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

// ThemeID identifies a named color theme.
type ThemeID int

const (
	ThemeGreen   ThemeID = iota + 1 // 1 — Classic Matrix green
	ThemeRed                        // 2 — Blood red
	ThemeBlue                       // 3 — Cyber blue
	ThemeWhite                      // 4 — Ghost white
	ThemeRainbow                    // 5 — Cycling hue spectrum
	ThemePurple                     // 6 — Deep purple / violet
	ThemeCyan                       // 7 — Neon cyan / teal
	ThemeOrange                     // 8 — Amber / orange
	ThemeGold                       // 9 — Golden yellow
)

// Theme holds the color palette for a single visual theme.
type Theme struct {
	ID   ThemeID
	Name string

	// HeadColor is the bright leading character.
	HeadColor tcell.Color
	// BrightColor is the top few characters below the head.
	BrightColor tcell.Color
	// MidColor is the middle section of the stream.
	MidColor tcell.Color
	// DimColor is the fading tail.
	DimColor tcell.Color
}

// themes maps ThemeID → Theme for quick lookup.
var themes = map[ThemeID]Theme{
	// ── 1 ── Classic Matrix green
	ThemeGreen: {
		ID:          ThemeGreen,
		Name:        "Matrix Green",
		HeadColor:   tcell.NewRGBColor(225, 255, 225),
		BrightColor: tcell.NewRGBColor(20, 255, 80),
		MidColor:    tcell.NewRGBColor(0, 200, 50),
		DimColor:    tcell.NewRGBColor(0, 85, 20),
	},
	// ── 2 ── Blood red
	ThemeRed: {
		ID:          ThemeRed,
		Name:        "Blood Red",
		HeadColor:   tcell.NewRGBColor(255, 225, 225),
		BrightColor: tcell.NewRGBColor(255, 30, 70),
		MidColor:    tcell.NewRGBColor(190, 10, 30),
		DimColor:    tcell.NewRGBColor(85, 0, 10),
	},
	// ── 3 ── Cyber blue
	ThemeBlue: {
		ID:          ThemeBlue,
		Name:        "Cyber Blue",
		HeadColor:   tcell.NewRGBColor(225, 245, 255),
		BrightColor: tcell.NewRGBColor(0, 190, 255),
		MidColor:    tcell.NewRGBColor(0, 95, 230),
		DimColor:    tcell.NewRGBColor(0, 40, 110),
	},
	// ── 4 ── Ghost white
	ThemeWhite: {
		ID:          ThemeWhite,
		Name:        "Ghost White",
		HeadColor:   tcell.NewRGBColor(255, 255, 255),
		BrightColor: tcell.NewRGBColor(220, 225, 230),
		MidColor:    tcell.NewRGBColor(140, 145, 150),
		DimColor:    tcell.NewRGBColor(65, 70, 75),
	},
	// ── 5 ── Rainbow (computed dynamically per frame)
	ThemeRainbow: {
		ID:   ThemeRainbow,
		Name: "Rainbow",
	},
	// ── 6 ── Deep purple / violet
	ThemePurple: {
		ID:          ThemePurple,
		Name:        "Deep Purple",
		HeadColor:   tcell.NewRGBColor(250, 225, 255),
		BrightColor: tcell.NewRGBColor(210, 60, 255),
		MidColor:    tcell.NewRGBColor(135, 25, 200),
		DimColor:    tcell.NewRGBColor(60, 0, 100),
	},
	// ── 7 ── Neon cyan / teal
	ThemeCyan: {
		ID:          ThemeCyan,
		Name:        "Neon Cyan",
		HeadColor:   tcell.NewRGBColor(220, 255, 255),
		BrightColor: tcell.NewRGBColor(0, 245, 225),
		MidColor:    tcell.NewRGBColor(0, 175, 160),
		DimColor:    tcell.NewRGBColor(0, 80, 70),
	},
	// ── 8 ── Amber / orange
	ThemeOrange: {
		ID:          ThemeOrange,
		Name:        "Amber",
		HeadColor:   tcell.NewRGBColor(255, 245, 220),
		BrightColor: tcell.NewRGBColor(255, 145, 0),
		MidColor:    tcell.NewRGBColor(195, 90, 0),
		DimColor:    tcell.NewRGBColor(90, 40, 0),
	},
	// ── 9 ── Golden yellow
	ThemeGold: {
		ID:          ThemeGold,
		Name:        "Gold",
		HeadColor:   tcell.NewRGBColor(255, 255, 210),
		BrightColor: tcell.NewRGBColor(255, 215, 0),
		MidColor:    tcell.NewRGBColor(185, 145, 0),
		DimColor:    tcell.NewRGBColor(80, 60, 0),
	},
}

// RainbowColor returns an HSL-derived tcell color for a given column and
// depth position, producing a smooth animated rainbow waterfall.
func RainbowColor(col, pos, length, tick int) tcell.Color {
	// Each column gets a distinct base hue, shifted over time.
	hue := math.Mod(float64(col*17+tick*2)/360.0, 1.0)

	// Lightness fades toward the tail; head is near-white.
	lightness := 0.78 - (float64(pos)/float64(length+1))*0.55
	if pos == 0 {
		lightness = 1.0
	}

	r, g, b := hslToRGB(hue, 1.0, lightness)
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// hslToRGB converts HSL (each in [0,1]) to an RGB triple ([0,255] each).
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hueToRGB(p, q, h+1.0/3)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3)
	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}

// ColorForDepth returns the appropriate color for a character at a given
// depth within a stream, respecting the current theme, background color, and 3D depth layer.
func ColorForDepth(theme Theme, col, pos, length, tick int, bg tcell.Color, depth int) tcell.Color {
	if theme.ID == ThemeRainbow {
		return RainbowColor(col, pos, length, tick)
	}

	if depth == 0 {
		// Far (background): always dim, blend from MidColor -> DimColor -> bg
		ratio := float64(pos) / float64(length)
		if ratio < 0.4 {
			return blendColors(theme.MidColor, theme.DimColor, ratio/0.4)
		} else {
			return blendColors(theme.DimColor, bg, (ratio-0.4)/0.6)
		}
	}

	if depth == 2 {
		// Near (foreground): extra bright, first 10% stays at HeadColor, then blends slower
		ratio := float64(pos) / float64(length)
		if ratio < 0.10 {
			return theme.HeadColor
		} else if ratio < 0.40 {
			return blendColors(theme.HeadColor, theme.BrightColor, (ratio-0.10)/0.30)
		} else if ratio < 0.80 {
			return blendColors(theme.BrightColor, theme.MidColor, (ratio-0.40)/0.40)
		} else {
			return blendColors(theme.MidColor, bg, (ratio-0.80)/0.20)
		}
	}

	// Medium (depth 1): standard blending
	ratio := float64(pos) / float64(length)
	if ratio < 0.05 {
		return blendColors(theme.HeadColor, theme.BrightColor, ratio/0.05)
	} else if ratio < 0.35 {
		return blendColors(theme.BrightColor, theme.MidColor, (ratio-0.05)/0.30)
	} else if ratio < 0.70 {
		return blendColors(theme.MidColor, theme.DimColor, (ratio-0.35)/0.35)
	} else {
		return blendColors(theme.DimColor, bg, (ratio-0.70)/0.30)
	}
}

func blendColors(c1, c2 tcell.Color, t float64) tcell.Color {
	if t <= 0 {
		return c1
	}
	if t >= 1 {
		return c2
	}
	r1, g1, b1 := c1.RGB()
	r2, g2, b2 := c2.RGB()
	r := int32(float64(r1) + float64(r2-r1)*t)
	g := int32(float64(g1) + float64(g2-g1)*t)
	b := int32(float64(b1) + float64(b2-b1)*t)
	return tcell.NewRGBColor(r, g, b)
}
