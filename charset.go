// charset.go — Character set registry and -l spec parser for voidmatrix
package main

import "strings"

// ---------------------------------------------------------------------------
// Character set definitions (one-letter codes, matching unimatrix where possible)
// ---------------------------------------------------------------------------

var charDefs = map[byte][]rune{
	// ASCII
	'a': []rune("abcdefghijklmnopqrstuvwxyz"),
	'A': []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"),

	// Russian Cyrillic
	'c': runeRange(0x0430, 0x044F), // а–я
	'C': runeRange(0x0410, 0x042F), // А–Я

	// Emoji (single-width safe subset)
	'e': []rune("☺☻✌♡♥❤⚘❀❃✼☀♫♪☃❄❆☕☂★✦✧❋✿❦♠♣♦"),

	// Greek
	'g': runeRange(0x03B1, 0x03C9), // α–ω
	'G': runeRange(0x0391, 0x03A9), // Α–Ω

	// Japanese
	'j': runeRange(0x3041, 0x3096), // Hiragana
	'k': []rune("ｦｧｨｩｪｫｬｭｮｯｰｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ"), // Half-width Katakana
	'K': runeRange(0x30A1, 0x30F6), // Full-width Katakana

	// Numbers
	'n': []rune("0123456789"),

	// Symbols
	's': []rune(`-=*_+|:<>"`),                            // Matrix film subset
	'S': []rune("`-=~!@#$%^&*()_+[]{}|;':\",./<>?\\"), // All keyboard symbols
}

// 'm' and 'o' are composite — expanded in ParseCharList.
// 'u' is the custom characters passed via --custom.

// runeRange creates a []rune from Unicode codepoint start to end inclusive.
func runeRange(start, end rune) []rune {
	out := make([]rune, 0, int(end-start+1))
	for r := start; r <= end; r++ {
		out = append(out, r)
	}
	return out
}

// ParseCharList converts a -l spec string into a combined rune pool.
//
//	'm'  →  Matrix default  (kkknnssss)
//	'o'  →  Old cmatrix     (AaSn)
//	'u'  →  custom chars supplied via --custom
//	else →  looked up in charDefs
//
// If spec is empty, returns the built-in Mixed preset.
func ParseCharList(spec string, custom []rune) []rune {
	if spec == "" {
		return charPresets[0].pool // Mixed default
	}

	// Support human-readable aliases for -l:
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "classic", "matrix":
		spec = "m"
	case "cmatrix", "standard":
		spec = "o"
	case "japanese", "kana":
		spec = "Kkj"
	case "hiragana":
		spec = "j"
	case "katakana":
		spec = "k"
	case "greek":
		spec = "gG"
	case "cyrillic", "russian":
		spec = "cC"
	case "emoji":
		spec = "e"
	case "binary":
		return []rune("01")
	case "numbers", "digits":
		spec = "n"
	case "letters", "alphabet":
		spec = "aA"
	case "all":
		spec = "KkjgGcCaeS"
	}

	var pool []rune
	for i := 0; i < len(spec); i++ {
		ch := spec[i]
		switch ch {
		case 'm':
			for j := 0; j < 3; j++ {
				pool = append(pool, charDefs['k']...)
			}
			for j := 0; j < 2; j++ {
				pool = append(pool, charDefs['n']...)
			}
			for j := 0; j < 4; j++ {
				pool = append(pool, charDefs['s']...)
			}
		case 'o':
			pool = append(pool, charDefs['A']...)
			pool = append(pool, charDefs['a']...)
			pool = append(pool, charDefs['S']...)
			pool = append(pool, charDefs['n']...)
		case 'u':
			pool = append(pool, custom...)
		default:
			if rs, ok := charDefs[ch]; ok {
				pool = append(pool, rs...)
			}
		}
	}

	if len(pool) == 0 {
		return charPresets[0].pool
	}
	return pool
}

// ---------------------------------------------------------------------------
// Presets — cycled with the C key at runtime
// ---------------------------------------------------------------------------

type charPreset struct {
	name string
	pool []rune
}

var charPresets []charPreset

func init() {
	mk := func(codes ...byte) []rune {
		var p []rune
		for _, c := range codes {
			p = append(p, charDefs[c]...)
		}
		return p
	}

	charPresets = []charPreset{
		{"Mixed  (kana+ASCII)", mk('k', 'n', 's', 'a', 'A')},
		{"Katakana", charDefs['k']},
		{"Hiragana", charDefs['j']},
		{"Full Katakana", charDefs['K']},
		{"Cyrillic", mk('c', 'C')},
		{"Greek", mk('g', 'G')},
		{"ASCII", mk('a', 'A', 'n', 'S')},
		{"Numbers", charDefs['n']},
		{"Emoji", charDefs['e']},
	}
}

// ---------------------------------------------------------------------------
// Color name helpers (used by -c and -g flags)
// ---------------------------------------------------------------------------

// ColorNameToTheme maps a named color to a ThemeID.
func ColorNameToTheme(name string) ThemeID {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "red":
		return ThemeRed
	case "blue":
		return ThemeBlue
	case "white":
		return ThemeWhite
	case "rainbow":
		return ThemeRainbow
	case "purple", "magenta":
		return ThemePurple
	case "cyan":
		return ThemeCyan
	case "orange", "amber":
		return ThemeOrange
	case "gold", "yellow":
		return ThemeGold
	default:
		return ThemeGreen
	}
}
