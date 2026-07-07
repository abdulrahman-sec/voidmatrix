// renderer.go — tcell drawing logic for voidmatrix
package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Renderer owns the tcell screen and renders frames from StateSnapshots.
type Renderer struct {
	screen tcell.Screen

	// Both grids are pre-allocated once and reused every frame.
	// This eliminates the per-frame heap allocation that caused GC pauses
	// and stuttering in earlier versions.
	desired   [][]cellState
	prevCells [][]cellState
	prevW     int
	prevH     int
	prevBg    int32

	lastDrawTime time.Time
	frameDurs    []time.Duration
}

// cellState describes a single terminal cell.
type cellState struct {
	ch   rune
	fg   tcell.Color
	bg   tcell.Color
	bold bool
}

func NewRenderer() (*Renderer, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("tcell.NewScreen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("screen.Init: %w", err)
	}
	screen.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorDefault).
		Foreground(tcell.ColorDefault))
	screen.Clear()
	screen.HideCursor()
	return &Renderer{screen: screen}, nil
}

func (r *Renderer) Screen() tcell.Screen { return r.screen }
func (r *Renderer) Fini()                { r.screen.Fini() }

// Draw renders one frame from the snapshot.
func (r *Renderer) Draw(snap *StateSnapshot) {
	startTime := time.Now()
	w, h := snap.width, snap.height

	// Resolve background color.
	var bgColor tcell.Color
	if snap.bgColor != 0 {
		bgColor = tcell.NewRGBColor(
			(snap.bgColor>>16)&0xFF,
			(snap.bgColor>>8)&0xFF,
			snap.bgColor&0xFF,
		)
	} else {
		bgColor = tcell.ColorDefault
	}
	bgStyle := tcell.StyleDefault.Background(bgColor).Foreground(tcell.ColorDefault)

	// Re-allocate both grids only on resize or background change.
	if w != r.prevW || h != r.prevH || snap.bgColor != r.prevBg {
		r.screen.Sync() // Safe sync-on-resize in draw thread
		r.desired = make([][]cellState, h)
		r.prevCells = make([][]cellState, h)
		for row := 0; row < h; row++ {
			r.desired[row] = make([]cellState, w)
			r.prevCells[row] = make([]cellState, w)
		}
		r.prevW = w
		r.prevH = h
		r.prevBg = snap.bgColor
		r.screen.SetStyle(bgStyle)
		r.screen.Clear()
	}

	// Zero the desired grid by copying a zero struct into each cell.
	// This reuses the existing allocation — no GC pressure.
	var zero cellState
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			r.desired[row][col] = zero
		}
	}

	theme := themes[snap.theme]

	// Paint rain streams into desired.
	// Paint rain streams into desired, layered by depth (Far -> Med -> Near).
	// This naturally overwrites background streams with foreground streams in overlapping columns.
	for dLayer := 0; dLayer <= 2; dLayer++ {
		for _, st := range snap.streams {
			if st.depth != dLayer {
				continue
			}
			if st.col < 0 || st.col >= w {
				continue
			}
			for pos := 0; pos < st.length; pos++ {
				row := st.head - pos
				if row < 0 || row >= h {
					continue
				}
				if pos >= len(st.chars) {
					continue
				}

				isFlasher := st.flashers != nil &&
					pos < len(st.flashers) && st.flashers[pos]

				isGlitched := snap.glitchedCols != nil && st.col < len(snap.glitchedCols) && snap.glitchedCols[st.col]

				var fg tcell.Color
				if isGlitched {
					fg = tcell.NewRGBColor(255, 60, 100) // Vibrant hot-pink
					if snap.tick%2 == 0 {
						fg = tcell.NewRGBColor(255, 255, 255) // White strobe
					}
				} else if isFlasher {
					fg = theme.HeadColor
					if theme.ID == ThemeRainbow {
						fg = RainbowColor(st.col, 0, st.length, snap.tick)
					}
				} else {
					fadeBg := bgColor
					if fadeBg == tcell.ColorDefault {
						fadeBg = tcell.NewRGBColor(0, 0, 0)
					}
					fg = ColorForDepth(theme, st.col, pos, st.length, snap.tick, fadeBg, st.depth)
				}

				r.desired[row][st.col] = cellState{
					ch:   st.chars[pos],
					fg:   fg,
					bg:   bgColor,
					bold: isGlitched || isBold(snap.boldMode, pos, st.depth, isFlasher),
				}
			}
		}
	}

	// Paint splash particles on the bottom row.
	if snap.splashes != nil {
		row := h - 1
		for _, sp := range snap.splashes {
			var chLeft, chRight rune
			var color tcell.Color

			switch sp.Age {
			case 0:
				chLeft, chRight = '*', '*'
				color = theme.HeadColor
			case 1:
				chLeft, chRight = '+', '+'
				color = theme.BrightColor
			default:
				chLeft, chRight = '.', '.'
				color = theme.MidColor
			}

			// Draw left splash particle
			if sp.Col > 0 {
				r.desired[row][sp.Col-1] = cellState{
					ch:   chLeft,
					fg:   color,
					bg:   bgColor,
					bold: true,
				}
			}
			// Draw right splash particle
			if sp.Col < w-1 {
				r.desired[row][sp.Col+1] = cellState{
					ch:   chRight,
					fg:   color,
					bg:   bgColor,
					bold: true,
				}
			}
		}
	}

	// Paint decoded custom message cells on top of rain streams.
	if snap.msgActive && snap.msgCells != nil {
		for _, cell := range snap.msgCells {
			if cell.Col < 0 || cell.Col >= w || cell.Row < 0 || cell.Row >= h {
				continue
			}
			if !cell.Revealed {
				continue
			}

			var fgColor tcell.Color
			if cell.GlowTime > 0 {
				fgColor = tcell.NewRGBColor(255, 255, 255) // white glow
			} else {
				fgColor = theme.BrightColor // theme highlight color
			}

			r.desired[cell.Row][cell.Col] = cellState{
				ch:   cell.Char,
				fg:   fgColor,
				bg:   bgColor,
				bold: true,
			}
		}
	}

	// OSD overlay — centred at the very top row.
	if snap.osd != "" && w > 4 {
		runes := []rune(snap.osd)
		osdRow := 0
		osdStart := (w - len(runes)) / 2
		if osdStart < 0 {
			osdStart = 0
		}
		osdFg := tcell.NewRGBColor(240, 240, 240)
		osdBg := tcell.NewRGBColor(18, 18, 18)
		for i, ch := range runes {
			c := osdStart + i
			if c >= w {
				break
			}
			r.desired[osdRow][c] = cellState{ch: ch, fg: osdFg, bg: osdBg, bold: true}
		}
	}

	// Draw debug performance HUD if requested
	if snap.Debug {
		r.drawDebugHUD(w, h, time.Since(startTime))
	}

	// Diff against previous frame — only write cells that changed.
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			cur := r.desired[row][col]
			if cur == r.prevCells[row][col] {
				continue
			}
			if cur.ch == 0 {
				r.screen.SetContent(col, row, ' ', nil, bgStyle)
			} else {
				style := tcell.StyleDefault.Foreground(cur.fg).Background(cur.bg)
				if cur.bold {
					style = style.Bold(true)
				}
				r.screen.SetContent(col, row, cur.ch, nil, style)
			}
			r.prevCells[row][col] = cur
		}
	}

	r.screen.Show()
}

// isBold decides whether a cell should be bold given the current bold mode and 3D depth of the stream.
func isBold(mode BoldMode, pos int, streamDepth int, isFlasher bool) bool {
	if mode == BoldNone {
		return false
	}
	if streamDepth == 0 { // background is never bold
		return false
	}
	if streamDepth == 2 { // foreground is always bold
		return true
	}
	// Midground (streamDepth == 1) follows normal bold mode
	if mode == BoldAll {
		return true
	}
	// BoldMixed (default)
	return pos < 2 || isFlasher
}

func (r *Renderer) drawDebugHUD(w, h int, drawLatency time.Duration) {
	// 1. Calculate FPS
	now := time.Now()
	if !r.lastDrawTime.IsZero() {
		frameDur := now.Sub(r.lastDrawTime)
		r.frameDurs = append(r.frameDurs, frameDur)
		if len(r.frameDurs) > 60 {
			r.frameDurs = r.frameDurs[1:]
		}
	}
	r.lastDrawTime = now

	var total time.Duration
	for _, d := range r.frameDurs {
		total += d
	}
	fps := 0.0
	if len(r.frameDurs) > 0 {
		fps = float64(len(r.frameDurs)) / total.Seconds()
	}

	// 2. Read mem stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 3. Format stats string
	stats := fmt.Sprintf(" FPS: %.1f | Render: %dµs | Heap: %.2fMB | GCs: %d ",
		fps,
		drawLatency.Microseconds(),
		float64(m.Alloc)/(1024*1024),
		m.NumGC,
	)

	// 4. Render overlay on the bottom-right corner
	runes := []rune(stats)
	hudRow := h - 1
	hudStart := w - len(runes) - 1
	if hudStart < 0 {
		hudStart = 0
	}
	hudFg := tcell.NewRGBColor(0, 255, 0)
	hudBg := tcell.NewRGBColor(10, 10, 10)
	for i, ch := range runes {
		c := hudStart + i
		if c >= w {
			break
		}
		r.desired[hudRow][c] = cellState{
			ch:   ch,
			fg:   hudFg,
			bg:   hudBg,
			bold: true,
		}
	}
}
