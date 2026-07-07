// main.go — Entry point for voidmatrix
//
// Build:   go build -o voidmatrix
// Run:     ./voidmatrix [flags]
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/terminfo/extended"
)

func main() {
	// Check for subcommand update
	if len(os.Args) > 1 && (os.Args[1] == "update" || os.Args[1] == "--update" || os.Args[1] == "-update") {
		runSelfUpdate()
		return
	}

	// ── Flags ──────────────────────────────────────────────────────────────
	fSpeed   := flag.Float64("s", 0.45, "Speed multiplier (0.1–8.0)")
	fDensity := flag.Float64("density", 0.72, "Streams per column (0.1–3.0; >1 = multiple per col)")
	fColor   := flag.String("c", "green", "Color: green red blue white cyan magenta purple rainbow")
	fColorID := flag.Int("color", 0, "Color by number 1–9 (overrides -c if set)")
	fBg      := flag.String("g", "", "Background color: black red green blue white cyan magenta")
	fCharset := flag.String("l", "", "Character set spec, e.g. 'knnssss', 'naAS', 'gGcCj'")
	fCustom  := flag.String("u", "", "Custom characters (use with -l ...u...)")
	fAsync   := flag.Bool("a", true, "Async scroll: each stream moves at its own speed")
	fBold    := flag.Bool("b", false, "All-bold mode")
	fNoBold  := flag.Bool("n", false, "No-bold mode (overrides -b)")
	fFlash   := flag.Bool("f", false, "Enable flashers: random glowing characters")
	fIgnore  := flag.Bool("i", false, "Ignore keyboard input")
	fNoStatus:= flag.Bool("o", false, "Hide OSD status overlay")
	fWave    := flag.Bool("w", false, "Single-wave: one full pass then exit")
	fTime    := flag.Float64("t", 0, "Exit after this many seconds (0 = run forever)")
	fWind    := flag.String("wind", "none", "Wind direction: none left right")
	fSplash  := flag.Bool("splash", false, "Enable ground impact splashes")
	fGlitch  := flag.Bool("glitch", false, "Enable transmission glitches")
	fMsg     := flag.String("msg", "", "Custom text message to reveal on screen")
	fMode    := flag.String("mode", "", "Preset visual mode: hacker chill chaos cinematic")
	fScreensaver := flag.Bool("screensaver", false, "Screensaver mode: exit on any key, auto-exit timer")
	fDebug       := flag.Bool("d", false, "Show performance profiling metrics HUD")

	// Long-form aliases
	flag.Float64Var(fSpeed,    "speed",    0.45,  "")
	flag.StringVar(fColor,     "color-name","green","")
	flag.StringVar(fBg,        "bg",       "",   "")
	flag.StringVar(fCharset,   "charset",  "",   "")
	flag.StringVar(fCustom,    "custom",   "",   "")
	flag.BoolVar(fAsync,       "async",    false,"")
	flag.BoolVar(fBold,        "bold",     false,"")
	flag.BoolVar(fNoBold,      "no-bold",  false,"")
	flag.BoolVar(fFlash,       "flashers", false,"")
	flag.BoolVar(fIgnore,      "ignore-input", false, "")
	flag.BoolVar(fNoStatus,    "no-status",false, "")
	flag.BoolVar(fWave,        "wave",     false,"")
	flag.Float64Var(fTime,     "time",     0,    "")
	flag.StringVar(fMsg,      "message",  "",   "")
	flag.BoolVar(fDebug,      "debug",    false,"")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `voidmatrix — Matrix-style digital rain, advanced edition

Usage:  voidmatrix [flags]

  -s  SPEED     Speed multiplier 0.1–8.0 (default 0.45)
  -c  COLOR     Color: green red blue white cyan magenta purple rainbow
  -g  COLOR     Background color: black red green blue white cyan magenta
  -l  SPEC      Character sets, e.g. "knnssss" "kj" "gGcC" "naAS"
  -u  CHARS     Custom characters (pair with -l ...u...)
  -a            Async scroll — streams move at varied speeds
  -b            All bold characters
  -n            No bold (overrides -b)
  -f            Flashers — random glowing characters that pulse every frame
  -i            Ignore keyboard input (screensaver mode)
  -o            Hide OSD status overlay
  -w            Single-wave: one full pass then exit (great for .bashrc)
  -t  SECS      Exit after N seconds
  --density N   Streams per column, >1 = overlapping (default 0.72)
  --color N     Color by number 1–9
  --wind DIR    Wind direction: none, left, right (default none)
  --splash      Enable ground impact splashes
  --glitch      Enable transmission glitches
  -msg MSG      Custom text message to reveal on screen (alias --message)
  --mode MODE   Preset visual mode: hacker, chill, chaos, cinematic
  --screensaver Screensaver mode: exit on any key, default 5 min auto-exit
  -d, --debug   Show performance profiling metrics HUD

Character set codes for -l:
  k  Half-width Katakana    K  Full-width Katakana
  j  Hiragana               n  Numbers 0-9
  a  Lowercase ASCII        A  Uppercase ASCII
  s  Matrix symbols         S  All keyboard symbols
  g  Greek lowercase        G  Greek uppercase
  c  Cyrillic lowercase     C  Cyrillic uppercase
  e  Emoji subset           m  Matrix default (=knnssss)
  o  Old cmatrix style      u  Custom (via -u)

Keyboard controls (while running):
  ↑/W/+  Speed up      ↓/S/-  Speed down
  →/D    More columns  ←/A    Fewer columns
  1–9    Color theme   c/C    Cycle charset
  a      Async toggle  b      Bold cycle    f  Flashers
  o      Toggle OSD    Q/Space/Ctrl-C  Quit

Examples:
  voidmatrix -c cyan -l kj -f                  # Cyan, kana+hiragana, flashers
  voidmatrix -l gGcC -c purple -a              # Greek+Cyrillic, async
  voidmatrix -w -i -s 3 -c green              # Single wave, fast, green
  voidmatrix -l naAS -n -c white              # cmatrix style, no bold
  voidmatrix -l e -c rainbow -f --density 1.5 # Emoji rain, rainbow, flashers
`)
	}
	loadConfigFile()
	flag.Parse()

	// ── Apply Preset Modes ──────────────────────────────────────────────────
	if *fMode != "" {
		mode := strings.ToLower(*fMode)
		switch mode {
		case "hacker":
			if !isCLIExplicit("s") && !isCLIExplicit("speed") {
				*fSpeed = 1.5
			}
			if !isCLIExplicit("density") {
				*fDensity = 1.2
			}
			if !isCLIExplicit("glitch") {
				*fGlitch = true
			}
			if !isCLIExplicit("wind") {
				*fWind = "none"
			}
			if !isCLIExplicit("c") && !isCLIExplicit("color-name") && !isCLIExplicit("color") {
				*fColor = "green"
			}
			if !isCLIExplicit("splash") {
				*fSplash = true
			}
			if !isCLIExplicit("f") && !isCLIExplicit("flashers") {
				*fFlash = true
			}
		case "chill":
			if !isCLIExplicit("s") && !isCLIExplicit("speed") {
				*fSpeed = 0.25
			}
			if !isCLIExplicit("density") {
				*fDensity = 0.5
			}
			if !isCLIExplicit("glitch") {
				*fGlitch = false
			}
			if !isCLIExplicit("wind") {
				*fWind = "none"
			}
			if !isCLIExplicit("c") && !isCLIExplicit("color-name") && !isCLIExplicit("color") {
				*fColor = "cyan"
			}
			if !isCLIExplicit("splash") {
				*fSplash = false
			}
			if !isCLIExplicit("f") && !isCLIExplicit("flashers") {
				*fFlash = false
			}
			if !isCLIExplicit("n") && !isCLIExplicit("no-bold") {
				*fNoBold = true
			}
		case "chaos":
			if !isCLIExplicit("s") && !isCLIExplicit("speed") {
				*fSpeed = 3.0
			}
			if !isCLIExplicit("density") {
				*fDensity = 2.0
			}
			if !isCLIExplicit("glitch") {
				*fGlitch = true
			}
			if !isCLIExplicit("wind") {
				*fWind = "left"
			}
			if !isCLIExplicit("c") && !isCLIExplicit("color-name") && !isCLIExplicit("color") {
				*fColor = "rainbow"
			}
			if !isCLIExplicit("splash") {
				*fSplash = true
			}
			if !isCLIExplicit("f") && !isCLIExplicit("flashers") {
				*fFlash = true
			}
		case "cinematic":
			if !isCLIExplicit("s") && !isCLIExplicit("speed") {
				*fSpeed = 0.4
			}
			if !isCLIExplicit("density") {
				*fDensity = 0.75
			}
			if !isCLIExplicit("glitch") {
				*fGlitch = false
			}
			if !isCLIExplicit("wind") {
				*fWind = "right"
			}
			if !isCLIExplicit("c") && !isCLIExplicit("color-name") && !isCLIExplicit("color") {
				*fColor = "green"
			}
			if !isCLIExplicit("splash") {
				*fSplash = true
			}
			if !isCLIExplicit("f") && !isCLIExplicit("flashers") {
				*fFlash = false
			}
		default:
			fmt.Fprintf(os.Stderr, "Invalid mode: '%s'. Valid modes are: hacker, chill, chaos, cinematic\n", *fMode)
			os.Exit(1)
		}
	}

	// ── Map flags to Config ─────────────────────────────────────────────────
	speed := clampFloat(*fSpeed, 0.1, 8.0)
	density := clampFloat(*fDensity, 0.1, 3.0)

	// Theme
	themeID := ColorNameToTheme(*fColor)
	if *fColorID >= 1 && *fColorID <= 9 {
		themeID = ThemeID(*fColorID)
	}

	// Background color (packed RGB int32)
	var bgColor int32
	if *fBg != "" {
		c := parseBgColor(*fBg)
		r, g, b := c.RGB()
		bgColor = (int32(r) << 16) | (int32(g) << 8) | int32(b)
	}

	// Character pool
	custom := []rune(*fCustom)
	charPool := ParseCharList(*fCharset, custom)
	poolName := "Custom"
	if *fCharset == "" {
		poolName = charPresets[0].name
	}

	// Bold mode
	boldMode := BoldMixed
	if *fBold {
		boldMode = BoldAll
	}
	if *fNoBold {
		boldMode = BoldNone
	}

	// Exit timer
	var exitAfter time.Duration
	if *fTime > 0 {
		exitAfter = time.Duration(*fTime * float64(time.Second))
	} else if *fScreensaver {
		// Default auto-exit to 5 minutes for screensaver if not explicitly specified
		exitAfter = 5 * time.Minute
	}

	wind := *fWind
	if wind != "none" && wind != "left" && wind != "right" {
		fmt.Fprintf(os.Stderr, "Invalid wind value: must be 'none', 'left', or 'right'\n")
		os.Exit(1)
	}

	cfg := Config{
		AsyncMode:   *fAsync,
		Flashers:    *fFlash,
		BoldMode:    boldMode,
		BgColor:     bgColor,
		StatusOff:   *fNoStatus,
		IgnoreKeys:  *fIgnore,
		SingleWave:  *fWave,
		ExitAfter:   exitAfter,
		Wind:        wind,
		Splash:      *fSplash,
		Glitch:      *fGlitch,
		Message:     *fMsg,
		Screensaver: *fScreensaver,
		Debug:       *fDebug,
	}

	// ── Renderer ────────────────────────────────────────────────────────────
	renderer, err := NewRenderer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialise renderer: %v\n", err)
		os.Exit(1)
	}
	defer renderer.Fini()

	// ── State ───────────────────────────────────────────────────────────────
	w, h := renderer.Screen().Size()
	state := NewState(w, h, speed, density, themeID, charPool, poolName, cfg)

	// ── Goroutines & Channels ───────────────────────────────────────────────
	quit := make(chan struct{}, 1)
	redrawChan := make(chan struct{}, 1)
	go NewInputHandler(renderer.Screen(), state, quit, redrawChan).Run()

	// Optional exit timer.
	var exitTimer <-chan time.Time
	if exitAfter > 0 {
		exitTimer = time.After(exitAfter)
	}

	// ── Main loop (Rate-limited, event-driven) ─────────────────────────────
	const baseMs = 50
	interval := time.Duration(float64(baseMs*time.Millisecond) / speed)
	if interval < 5*time.Millisecond {
		interval = 5 * time.Millisecond
	}
	tickTicker := time.NewTicker(interval)
	defer tickTicker.Stop()

	renderTicker := time.NewTicker(16 * time.Millisecond) // ~60 FPS rate limiter
	defer renderTicker.Stop()

	drawPending := true // force initial frame draw

	for {
		select {
		case <-quit:
			return

		case <-exitTimer:
			return

		case <-tickTicker.C:
			if state.IsDone() {
				return
			}
			state.Tick()
			drawPending = true

			// Dynamic tick rate adaptation
			spd := state.GetSpeed()
			newInterval := time.Duration(float64(baseMs*time.Millisecond) / spd)
			if newInterval < 5*time.Millisecond {
				newInterval = 5 * time.Millisecond
			}
			if newInterval != interval {
				interval = newInterval
				tickTicker.Reset(interval)
			}

		case <-redrawChan:
			drawPending = true

		case <-renderTicker.C:
			// Only draw if there are visual changes (drawn cell updates)
			// or if the OSD overlay is active and fading.
			if drawPending || state.HasOSDActive() {
				renderer.Draw(state.Snapshot())
				drawPending = false
			}
		}
	}
}

// parseBgColor maps a named color to a tcell.Color for background use.
func parseBgColor(name string) tcell.Color {
	switch name {
	case "black":
		return tcell.NewRGBColor(0, 0, 0)
	case "red":
		return tcell.NewRGBColor(60, 0, 0)
	case "green":
		return tcell.NewRGBColor(0, 25, 0)
	case "blue":
		return tcell.NewRGBColor(0, 0, 60)
	case "white":
		return tcell.NewRGBColor(210, 210, 210)
	case "yellow":
		return tcell.NewRGBColor(50, 40, 0)
	case "cyan":
		return tcell.NewRGBColor(0, 40, 40)
	case "magenta", "purple":
		return tcell.NewRGBColor(40, 0, 50)
	default:
		return tcell.ColorDefault
	}
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func loadConfigFile() {
	var configPath string

	// 1. Try standard user config directory: ~/.config/voidmatrix/config.yaml
	if configDir, err := os.UserConfigDir(); err == nil {
		configPath = filepath.Join(configDir, "voidmatrix", "config.yaml")
	}

	// 2. Fallback check: if standard doesn't exist, check legacy ~/.voidmatrix/config.yaml
	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if home, err := os.UserHomeDir(); err == nil {
				legacyPath := filepath.Join(home, ".voidmatrix", "config.yaml")
				if _, err := os.Stat(legacyPath); err == nil {
					configPath = legacyPath
				}
			}
		}
	} else {
		// Fallback if os.UserConfigDir() failed
		if home, err := os.UserHomeDir(); err == nil {
			configPath = filepath.Join(home, ".voidmatrix", "config.yaml")
		}
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Warning: Failed to read config file: %v\n", err)
		return
	}

	keyToFlag := map[string]string{
		"speed":        "speed",
		"density":      "density",
		"theme":        "color-name",
		"color":        "color",
		"bg":           "bg",
		"charset":      "charset",
		"custom":       "custom",
		"async":        "async",
		"bold":         "bold",
		"no-bold":      "no-bold",
		"flashers":     "flashers",
		"ignore-input": "ignore-input",
		"no-status":    "no-status",
		"wave":         "wave",
		"time":         "time",
		"wind":         "wind",
		"splash":       "splash",
		"glitch":       "glitch",
		"message":      "message",
	}

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Warning: config.yaml line %d is invalid: %s\n", lineNum+1, line)
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		// Remove inline comments
		if idx := strings.Index(val, "#"); idx != -1 {
			val = strings.TrimSpace(val[:idx])
		}
		// Strip quotes
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			if len(val) >= 2 {
				val = val[1 : len(val)-1]
			}
		}

		flagName, exists := keyToFlag[key]
		if !exists {
			flagName = key
		}

		f := flag.Lookup(flagName)
		if f != nil {
			if err := flag.Set(flagName, val); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: config.yaml line %d: failed to set option '%s' to '%s': %v\n", lineNum+1, key, val, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: config.yaml line %d: unknown option '%s'\n", lineNum+1, key)
		}
	}
}

func isCLIExplicit(name string) bool {
	prefix1 := "-" + name
	prefix2 := "--" + name
	for _, arg := range os.Args[1:] {
		if arg == prefix1 || arg == prefix2 ||
			strings.HasPrefix(arg, prefix1+"=") || strings.HasPrefix(arg, prefix2+"=") {
			return true
		}
	}
	return false
}

func runSelfUpdate() {
	fmt.Println("⚡ Checking for voidmatrix updates from GitHub...")
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "go install github.com/abdulrahman-sec/voidmatrix@latest")
	} else {
		cmd = exec.Command("sh", "-c", "curl -sSL https://raw.githubusercontent.com/abdulrahman-sec/voidmatrix/main/install.sh | sh")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n❌ Update failed: %v\n", err)
		fmt.Println("Please make sure you have Go/curl and internet connection, or install manually:")
		fmt.Println("  go install github.com/abdulrahman-sec/voidmatrix@latest")
		os.Exit(1)
	}
	fmt.Println("\n✅ voidmatrix updated successfully!")
}
