// state.go — Animation state, stream logic, runtime controls
package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// fastRand is a lock-free, ultra-fast Xorshift32 PRNG.
type fastRand uint32

func (r *fastRand) next() uint32 {
	x := *r
	if x == 0 {
		x = 1 // Xorshift seed must not be 0
	}
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*r = x
	return uint32(x)
}

func (r *fastRand) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint32(n))
}

func (r *fastRand) float64() float64 {
	return float64(r.next()) / 4294967296.0
}

func (r *fastRand) float32() float32 {
	return float32(r.next()) / 4294967296.0
}

func (r *fastRand) perm(n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := r.intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

type SnapshotPool struct {
	snaps [2]StateSnapshot
	idx   int
}

func (p *SnapshotPool) getNext() *StateSnapshot {
	p.idx = 1 - p.idx
	return &p.snaps[p.idx]
}

// ---------------------------------------------------------------------------
// Bold modes
// ---------------------------------------------------------------------------

// BoldMode controls which characters are rendered bold.
type BoldMode int

const (
	BoldMixed BoldMode = iota // leading chars bold (default)
	BoldAll                   // every char bold
	BoldNone                  // no bold anywhere
)

func (b BoldMode) String() string {
	switch b {
	case BoldAll:
		return "All"
	case BoldNone:
		return "None"
	default:
		return "Mixed"
	}
}

// ---------------------------------------------------------------------------
// Config — immutable settings (set at startup, toggled at runtime)
// ---------------------------------------------------------------------------

// Config holds all the feature flags that can come from CLI or keyboard.
type Config struct {
	AsyncMode   bool          // streams scroll at varied independent speeds
	Flashers    bool          // random cells glow and change every frame
	BoldMode    BoldMode      // mixed / all / none
	BgColor     int32         // packed RGB background (0 = terminal default)
	StatusOff   bool          // hide OSD indicator
	IgnoreKeys  bool          // ignore keyboard input
	SingleWave  bool          // one-pass mode: scroll everything off and exit
	ExitAfter   time.Duration // 0 = run forever
	Wind        string        // "none", "left", "right"
	Splash      bool          // true to enable ground splashes
	Glitch      bool          // true to enable random column transmission glitches
	Message     string        // custom message to decode on screen
	Screensaver bool          // true to enable screensaver mode (exit on any key)
	Debug       bool          // show performance metrics HUD
}

// ---------------------------------------------------------------------------
// Stream
// ---------------------------------------------------------------------------

// Stream is a single falling column of characters.
type Stream struct {
	col    int
	head   int    // row of the leading character
	length int    // total visible characters
	chars  []rune // character buffer
	// flashers[i] == true means chars[i] changes every Tick
	flashers []bool

	// async mode: only advance every tickPeriod global ticks
	tickPeriod int
	localTick  int

	depth      int // 3D depth layer: 0 = Far (background), 1 = Med (midground), 2 = Near (foreground)
	driftCounter int

	// single-wave: set true once the stream exits the bottom
	completed bool
}

// ---------------------------------------------------------------------------
// OSD
// ---------------------------------------------------------------------------

type osdEntry struct {
	text    string
	expires time.Time
}

func (o *osdEntry) active() bool {
	return o.text != "" && time.Now().Before(o.expires)
}

func (o *osdEntry) set(text string) {
	o.text = text
	o.expires = time.Now().Add(2 * time.Second)
}

// ---------------------------------------------------------------------------
type Splash struct {
	Col int
	Age int // ticks alive: 0, 1, 2
}

type DecodeCell struct {
	Char     rune
	Revealed bool
	GlowTime int // ticks of white-hot glow remaining
	Row, Col int // coordinate on screen
}

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

type State struct {
	mu sync.RWMutex

	width   int
	height  int
	streams []*Stream

	speed      float64  // multiplier driving the polling interval
	density    float64  // streams per column
	theme      ThemeID
	charPool   []rune   // active character pool
	poolName   string   // display name for OSD
	presetIdx  int      // index into charPresets for C-key cycling

	splashes []Splash // ground impact splash particles
	glitchActive []bool // tracks column glitch state per tick

	msgCells        []DecodeCell // message characters to reveal
	msgActive       bool         // true if custom message is currently decoding/showing
	msgCompleteTick int          // tick count when full reveal was reached

	cfg  Config
	tick int
	rng  fastRand
	osd  osdEntry
	done bool // set true when single-wave completes

	pool SnapshotPool // snapshot double buffering pool
}

// NewState creates the initial animation state.
func NewState(
	width, height int,
	speed, density float64,
	themeID ThemeID,
	charPool []rune,
	poolName string,
	cfg Config,
) *State {
	s := &State{
		width:    width,
		height:   height,
		speed:    speed,
		density:  density,
		theme:    themeID,
		charPool: charPool,
		poolName: poolName,
		cfg:      cfg,
		rng:      fastRand(time.Now().UnixNano()),
	}
	s.spawnStreams()
	return s
}

// spawnStreams creates the initial stream pool.
func (s *State) spawnStreams() {
	s.initMessageGrid()
	if s.cfg.SingleWave {
		// Fill every column, all starting just above the top.
		s.streams = make([]*Stream, s.width)
		for col := range s.streams {
			s.streams[col] = s.newStream(col, true, true)
		}
		return
	}
	count := int(float64(s.width) * s.density)
	if count < 1 {
		count = 1
	}
	if count > s.width {
		count = s.width * 3 // cap to prevent absurd overlap
	}
	s.streams = make([]*Stream, count)

	if s.density <= 1.0 {
		// Unique columns — no two streams in the same column.
		// This prevents the flickering caused by overlapping streams
		// fighting over the same cells.
		perm := s.rng.perm(s.width)
		for i := range s.streams {
			s.streams[i] = s.newStream(perm[i], false, true)
		}
	} else {
		// density > 1.0: intentional overlap allowed.
		for i := range s.streams {
			s.streams[i] = s.newStream(s.rng.intn(s.width), false, true)
		}
	}
}

// initMessageGrid initializes the centered text decode overlay.
func (s *State) initMessageGrid() {
	if s.cfg.Message == "" {
		return
	}
	s.msgActive = true
	s.msgCompleteTick = 0

	row := s.height / 2
	runes := []rune(s.cfg.Message)
	startCol := (s.width - len(runes)) / 2
	if startCol < 0 {
		startCol = 0
	}

	s.msgCells = make([]DecodeCell, 0, len(runes))
	for i, ch := range runes {
		col := startCol + i
		if col >= s.width {
			break
		}
		// Spaces are automatically marked as revealed.
		revealed := (ch == ' ')
		s.msgCells = append(s.msgCells, DecodeCell{
			Char:     ch,
			Row:      row,
			Col:      col,
			Revealed: revealed,
			GlowTime: 0,
		})
	}
}

// newStream allocates a stream for the given column.
// If wave=true (single-wave mode) the stream starts just above row 0.
func (s *State) newStream(col int, wave bool, init bool) *Stream {
	// 3D parallax depth selection: 30% background (far), 50% midground, 20% foreground (near)
	depthRoll := s.rng.float64()
	var depth int
	var tickPeriod int

	switch {
	case depthRoll < 0.30:
		depth = 0
		if s.cfg.AsyncMode {
			tickPeriod = 3 + s.rng.intn(2) // 3 or 4
		} else {
			tickPeriod = 3
		}
	case depthRoll < 0.80:
		depth = 1
		if s.cfg.AsyncMode {
			tickPeriod = 2
		} else {
			tickPeriod = 2
		}
	default:
		depth = 2
		tickPeriod = 1
	}

	minLen := s.height * 2 / 10
	if minLen < 3 {
		minLen = 3
	}
	var length int
	switch depth {
	case 0: // Far
		length = s.rng.intn(s.height*3/10) + minLen
	case 1: // Med
		length = s.rng.intn(s.height*5/10) + minLen + 1
	case 2: // Near
		length = s.rng.intn(s.height*7/10) + minLen + 3
	}

	var startRow int
	if wave {
		startRow = -1
	} else if init {
		// Startup/resize initialization: distribute streams uniformly on and off screen
		startRow = s.rng.intn(s.height+length) - length
	} else {
		// Respawn during run: start just above the top so it flows onto screen immediately
		startRow = -1 - s.rng.intn(3)
	}

	chars := make([]rune, length)
	for i := range chars {
		chars[i] = s.randRune()
	}

	// Flashers: ~6% of cells glow and change every frame.
	var flashers []bool
	if s.cfg.Flashers {
		flashers = make([]bool, length)
		for i := range flashers {
			flashers[i] = s.rng.float64() < 0.06
		}
	}

	return &Stream{
		col:        col,
		head:       startRow,
		length:     length,
		chars:      chars,
		flashers:   flashers,
		tickPeriod: tickPeriod,
		depth:      depth,
	}
}

// Resize adapts to a new terminal size.
func (s *State) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.width = width
	s.height = height
	s.spawnStreams()
}

// Tick advances the animation by one logical step.
func (s *State) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}
	s.tick++

	// Reset glitch states
	if len(s.glitchActive) != s.width {
		s.glitchActive = make([]bool, s.width)
	} else {
		for i := range s.glitchActive {
			s.glitchActive[i] = false
		}
	}

	// Trigger random glitch: 1.5% chance per frame that a random column glitches out
	if s.cfg.Glitch && s.rng.float64() < 0.015 {
		glitchCol := s.rng.intn(s.width)
		s.glitchActive[glitchCol] = true
		// Instantly mutate all characters of any stream in this column!
		for _, st := range s.streams {
			if st.col == glitchCol {
				for j := range st.chars {
					st.chars[j] = s.randRune()
				}
			}
		}
	}

	allDone := s.cfg.SingleWave // flipped to false below if any still running

	for i, st := range s.streams {
		// ── Flashers: mutate glowing cells every tick unconditionally ─────
		if s.cfg.Flashers {
			for j, flash := range st.flashers {
				if flash {
					st.chars[j] = s.randRune()
				}
			}
		}

		// ── Async: only advance this stream when its local timer fires ────
		shouldAdvance := true
		if s.cfg.AsyncMode {
			st.localTick++
			if st.localTick < st.tickPeriod {
				shouldAdvance = false
			} else {
				st.localTick = 0
			}
		}

		if !shouldAdvance {
			if s.cfg.SingleWave && !st.completed {
				allDone = false
			}
			continue
		}

		// ── Advance the stream ────────────────────────────────────────────────
		st.head++

		// Wind drift (parallax-aware): shifts column index.
		if s.cfg.Wind == "left" || s.cfg.Wind == "right" {
			st.driftCounter++
			threshold := 3 // Med (depth 1)
			if st.depth == 0 {
				threshold = 4 // Far
			} else if st.depth == 2 {
				threshold = 2 // Near
			}
			if st.driftCounter >= threshold {
				st.driftCounter = 0
				if s.cfg.Wind == "right" {
					st.col = (st.col + 1) % s.width
				} else {
					st.col = (st.col - 1 + s.width) % s.width
				}
			}
		}

		// Trigger ground splash if stream head hits the bottom of the screen.
		if s.cfg.Splash && st.head == s.height-1 {
			s.splashes = append(s.splashes, Splash{Col: st.col, Age: 0})
		}

		// Mutate at most one non-flasher cell per advance, and only 25% of
		// the time. Too-frequent mutations create eye-straining visual noise.
		n := len(st.chars)
		if n > 0 && s.rng.float32() < 0.25 {
			idx := s.rng.intn(n)
			if st.flashers == nil || !st.flashers[idx] {
				st.chars[idx] = s.randRune()
			}
		}

		dead := st.head-st.length > s.height

		if s.cfg.SingleWave {
			if dead {
				st.completed = true
			} else {
				allDone = false
			}
		} else if dead {
			s.streams[i] = s.newStream(s.rng.intn(s.width), false, false)
		}
	}

	if s.cfg.SingleWave && allDone {
		s.done = true
		return
	}

	if !s.cfg.SingleWave {
		// Reconcile stream count vs current density.
		want := int(float64(s.width) * s.density)
		if want < 1 {
			want = 1
		}
		for len(s.streams) < want {
			s.streams = append(s.streams, s.newStream(s.rng.intn(s.width), false, false))
		}
		if len(s.streams) > want {
			s.streams = s.streams[:want]
		}
	}

	// Update splash particles (in-place filtering to avoid allocations)
	if s.cfg.Splash {
		n := 0
		for _, sp := range s.splashes {
			sp.Age++
			if sp.Age <= 2 {
				s.splashes[n] = sp
				n++
			}
		}
		s.splashes = s.splashes[:n]
	}

	// Update custom message decoding
	if s.msgActive && len(s.msgCells) > 0 {
		// Check stream head positions to decode characters
		for _, st := range s.streams {
			for idx := range s.msgCells {
				cell := &s.msgCells[idx]
				if !cell.Revealed && st.col == cell.Col && st.head == cell.Row {
					cell.Revealed = true
					cell.GlowTime = 8 // glow white-hot for 8 frames
				}
			}
		}

		// Decay glow times and check completion
		allRevealed := true
		for idx := range s.msgCells {
			cell := &s.msgCells[idx]
			if cell.GlowTime > 0 {
				cell.GlowTime--
			}
			if !cell.Revealed {
				allRevealed = false
			}
		}

		// Emergency reveal: if 120 ticks pass and some cells are still hidden, reveal one random cell every 5 ticks
		if s.tick > 120 && s.tick%5 == 0 {
			var unrevealed []*DecodeCell
			for idx := range s.msgCells {
				cell := &s.msgCells[idx]
				if !cell.Revealed {
					unrevealed = append(unrevealed, cell)
				}
			}
			if len(unrevealed) > 0 {
				target := unrevealed[s.rng.intn(len(unrevealed))]
				target.Revealed = true
				target.GlowTime = 8
			}
		}

		if allRevealed {
			if s.msgCompleteTick == 0 {
				s.msgCompleteTick = s.tick
			} else if s.tick-s.msgCompleteTick > 120 {
				// Completed duration elapsed: dissolve message
				s.msgActive = false
				s.msgCells = nil
			}
		}
	}
}

// Snapshot returns a rendering-safe copy of the current state.
func (s *State) Snapshot() *StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := s.pool.getNext()
	snap.width = s.width
	snap.height = s.height
	snap.theme = s.theme
	snap.tick = s.tick
	snap.boldMode = s.cfg.BoldMode
	snap.bgColor = s.cfg.BgColor
	snap.done = s.done
	snap.msgActive = s.msgActive
	snap.Debug = s.cfg.Debug

	osdText := ""
	if !s.cfg.StatusOff && s.osd.active() {
		osdText = s.osd.text
	}
	snap.osd = osdText

	numActive := 0
	for _, st := range s.streams {
		if s.cfg.SingleWave && st.completed {
			continue
		}
		numActive++
	}

	if cap(snap.streams) < numActive {
		snap.streams = make([]StreamSnapshot, numActive)
	} else {
		snap.streams = snap.streams[:numActive]
	}

	idx := 0
	for _, st := range s.streams {
		if s.cfg.SingleWave && st.completed {
			continue
		}

		snap.streams[idx].col = st.col
		snap.streams[idx].head = st.head
		snap.streams[idx].length = st.length
		snap.streams[idx].depth = st.depth

		// Copy chars
		if cap(snap.streams[idx].chars) < len(st.chars) {
			snap.streams[idx].chars = make([]rune, len(st.chars))
		} else {
			snap.streams[idx].chars = snap.streams[idx].chars[:len(st.chars)]
		}
		copy(snap.streams[idx].chars, st.chars)

		// Copy flashers
		if st.flashers != nil {
			if cap(snap.streams[idx].flashers) < len(st.flashers) {
				snap.streams[idx].flashers = make([]bool, len(st.flashers))
			} else {
				snap.streams[idx].flashers = snap.streams[idx].flashers[:len(st.flashers)]
			}
			copy(snap.streams[idx].flashers, st.flashers)
		} else {
			snap.streams[idx].flashers = nil
		}

		idx++
	}

	// Copy splashes
	if cap(snap.splashes) < len(s.splashes) {
		snap.splashes = make([]Splash, len(s.splashes))
	} else {
		snap.splashes = snap.splashes[:len(s.splashes)]
	}
	copy(snap.splashes, s.splashes)

	// Copy glitchedCols
	if cap(snap.glitchedCols) < len(s.glitchActive) {
		snap.glitchedCols = make([]bool, len(s.glitchActive))
	} else {
		snap.glitchedCols = snap.glitchedCols[:len(s.glitchActive)]
	}
	copy(snap.glitchedCols, s.glitchActive)

	// Copy msgCells
	if cap(snap.msgCells) < len(s.msgCells) {
		snap.msgCells = make([]DecodeCell, len(s.msgCells))
	} else {
		snap.msgCells = snap.msgCells[:len(s.msgCells)]
	}
	copy(snap.msgCells, s.msgCells)

	return snap
}

// StreamSnapshot is a rendering-safe copy of one stream.
type StreamSnapshot struct {
	col      int
	head     int
	length   int
	chars    []rune
	flashers []bool // nil if flashers disabled
	depth    int
}

// StateSnapshot is a rendering-safe copy of the full frame.
type StateSnapshot struct {
	width        int
	height       int
	streams      []StreamSnapshot
	theme        ThemeID
	tick         int
	osd          string
	boldMode     BoldMode
	bgColor      int32 // packed RGB
	done         bool
	splashes     []Splash
	glitchedCols []bool
	msgCells     []DecodeCell
	msgActive    bool
	Debug        bool
}

// HasOSDActive returns true if the OSD overlay is currently active.
func (s *State) HasOSDActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.cfg.StatusOff && s.osd.active()
}

// ---------------------------------------------------------------------------
// Runtime controls
// ---------------------------------------------------------------------------

const speedStep = 1.5
const speedMin = 0.1
const speedMax = 8.0

func (s *State) IncreaseSpeed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed = math.Min(s.speed*speedStep, speedMax)
	s.osd.set(fmt.Sprintf("  ⚡ Speed  %s  %.2fx  ", logBar(s.speed, speedMin, speedMax, 10), s.speed))
}

func (s *State) DecreaseSpeed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed = math.Max(s.speed/speedStep, speedMin)
	s.osd.set(fmt.Sprintf("  ⚡ Speed  %s  %.2fx  ", logBar(s.speed, speedMin, speedMax, 10), s.speed))
}

func (s *State) IncreaseDensity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.density = math.Min(s.density+0.15, 3.0)
	s.osd.set(fmt.Sprintf("  ⣿ Density  %s  %.0f%%  ", linBar(s.density, 0.1, 3.0, 10), s.density/3.0*100))
}

func (s *State) DecreaseDensity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.density = math.Max(s.density-0.15, 0.1)
	s.osd.set(fmt.Sprintf("  ⣿ Density  %s  %.0f%%  ", linBar(s.density, 0.1, 3.0, 10), s.density/3.0*100))
}

func (s *State) SetTheme(id ThemeID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.theme = id
	if t, ok := themes[id]; ok {
		s.osd.set(fmt.Sprintf("  ◆ Theme  %s  ", t.Name))
	}
}

// CycleCharSet rotates through charPresets on each C key press.
func (s *State) CycleCharSet() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presetIdx = (s.presetIdx + 1) % len(charPresets)
	p := charPresets[s.presetIdx]
	s.charPool = p.pool
	s.poolName = p.name
	for _, st := range s.streams {
		for i := range st.chars {
			st.chars[i] = s.randRune()
		}
	}
	s.osd.set(fmt.Sprintf("  ✦ Charset  %s  ", s.poolName))
}

// ToggleAsync toggles asynchronous scroll mode.
func (s *State) ToggleAsync() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.AsyncMode = !s.cfg.AsyncMode
	// Re-assign tick periods now that mode changed.
	for _, st := range s.streams {
		if s.cfg.AsyncMode {
			st.tickPeriod = s.rng.intn(3) + 1
		} else {
			st.tickPeriod = 1
		}
		st.localTick = 0
	}
	state := "OFF"
	if s.cfg.AsyncMode {
		state = "ON"
	}
	s.osd.set(fmt.Sprintf("  ≋ Async  %s  ", state))
}

// CycleBold cycles through BoldMixed → BoldAll → BoldNone → BoldMixed.
func (s *State) CycleBold() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.BoldMode = (s.cfg.BoldMode + 1) % 3
	s.osd.set(fmt.Sprintf("  ⬛ Bold  %s  ", s.cfg.BoldMode))
}

// ToggleFlashers enables or disables the flasher effect.
func (s *State) ToggleFlashers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Flashers = !s.cfg.Flashers
	// Assign or clear flasher masks on existing streams.
	for _, st := range s.streams {
		if s.cfg.Flashers {
			if st.flashers == nil {
				st.flashers = make([]bool, len(st.chars))
			}
			for i := range st.flashers {
				st.flashers[i] = s.rng.float64() < 0.06
			}
		} else {
			st.flashers = nil
		}
	}
	state := "OFF"
	if s.cfg.Flashers {
		state = "ON"
	}
	s.osd.set(fmt.Sprintf("  ✦ Flashers  %s  ", state))
}

// ToggleStatus shows/hides the OSD overlay.
func (s *State) ToggleStatus() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.StatusOff = !s.cfg.StatusOff
	// Show confirmation only when turning it back on.
	if !s.cfg.StatusOff {
		s.osd.set("  ◉ Status  ON  ")
	}
}

func (s *State) GetSpeed() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.speed
}

func (s *State) IsDone() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.done
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *State) randRune() rune {
	return s.charPool[s.rng.intn(len(s.charPool))]
}

func logBar(val, minVal, maxVal float64, width int) string {
	fraction := (math.Log(val) - math.Log(minVal)) / (math.Log(maxVal) - math.Log(minVal))
	return renderBar(fraction, width)
}

func linBar(val, minVal, maxVal float64, width int) string {
	fraction := (val - minVal) / (maxVal - minVal)
	return renderBar(fraction, width)
}

func renderBar(fraction float64, width int) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(fraction*float64(width) + 0.5)
	b := make([]rune, width)
	for i := range b {
		if i < filled {
			b[i] = '█'
		} else {
			b[i] = '░'
		}
	}
	return string(b)
}
