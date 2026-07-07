<div align="center">

```
              _     __                __       _
 _   ______  (_)___/ /___ ___  ____ _/ /______(_)  __
| | / / __ \/ / __  / __ `__ \/ __ `/ __/ ___/ / |/_/
| |/ / /_/ / / /_/ / / / / / / /_/ / /_/ /  / />  <
|___/\____/_/\__,_/_/ /_/ /_/\__,_/\__/_/  /_/_/|_|
```

**A 3D, physically-layered Matrix digital rain simulator for your terminal — written in Go, tuned for zero-allocation 60 FPS.**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey)](#-installation)

</div>

---

## 🎬 Showcase

![Showcase](showcase.gif)

---

## 🤔 Why `voidmatrix`, not `cmatrix` or `unimatrix`?

`cmatrix` and `unimatrix` are great, but both draw a single flat plane of characters with no sense of depth, and both allocate memory inside the render loop — meaning long sessions on a low-power machine can visibly stutter.

`voidmatrix` was built from scratch to fix both of those:

| | `cmatrix` | `unimatrix` | `voidmatrix` |
|---|:---:|:---:|:---:|
| 3D parallax depth layers | ❌ | ❌ | ✅ |
| Zero heap allocations at steady-state | ❌ | ❌ | ✅ |
| Live text decoder reveal | ❌ | ✅ (partial) | ✅ |
| Wind skew / ground splash physics | ❌ | ❌ | ✅ |
| Config file + preset modes | ❌ | ❌ | ✅ |
| Live perf/debug dashboard | ❌ | ❌ | ✅ |

---

## 🌌 Core Features

- **3D Parallax Depth** — rain is distributed across Far (30%), Med (50%), and Near (20%) layers, each with independent speed, brightness, and stream length, drawn back-to-front so near streams correctly occlude far ones.
- **Decoder Reveal Engine** — pass `-msg "HELLO"` and watch the message decode letter-by-letter as streams sweep over it, with a guaranteed auto-reveal fallback so the message always finishes.
- **Weather Dynamics** — wind drift skew (`--wind left|right`) and floor-impact splash particles (`--splash`).
- **Transmission Glitches** — a small per-tick chance any column strobes white-hot/hot-pink for a single frame, like a corrupted signal.
- **Zero-Allocation Render Loop** — pre-allocated double-buffered grids, an inline lock-free Xorshift32 PRNG, and diff-only terminal writes. No GC pauses during steady-state animation.
- **Preset Modes** — `hacker`, `chill`, `chaos`, and `cinematic`, each a hand-tuned combination of every parameter below.

---

## 📋 Requirements

- Go **1.21+** (only needed if building from source or using `go install`)
- A terminal emulator with 256-color or true-color support for full gradient fidelity (most modern terminals — Alacritty, kitty, WezTerm, Windows Terminal, iTerm2 — qualify)
- [`tcell/v2`](https://github.com/gdamore/tcell) (pulled automatically as a Go module dependency)

---

## 🚀 Installation

### One-line install (macOS / Linux)
```bash
curl -sSL https://raw.githubusercontent.com/abdulrahman-sec/voidmatrix/main/install.sh | sh
```

### Go install
```bash
go install github.com/abdulrahman-sec/voidmatrix@latest
```

### Arch Linux (AUR)
```bash
git clone https://github.com/abdulrahman-sec/voidmatrix.git
cd voidmatrix && makepkg -si
```

### Manual build
```bash
git clone https://github.com/abdulrahman-sec/voidmatrix.git
cd voidmatrix
go build -o voidmatrix .
./voidmatrix
```

### Windows Installation (CMD / PowerShell)

#### Option 1: Direct Download (Easiest)
1. Go to the [Releases](https://github.com/abdulrahman-sec/voidmatrix/releases) page.
2. Download the pre-compiled `voidmatrix.exe` binary.
3. Open CMD/PowerShell in your download folder and run it:
   ```powershell
   .\voidmatrix.exe
   ```

#### Option 2: Compile from Source (Requires Go)
1. Download and install Go from [go.dev/dl](https://go.dev/dl/).
2. Open PowerShell or Command Prompt, clone the repository, compile and run:
   ```powershell
   git clone https://github.com/abdulrahman-sec/voidmatrix.git
   cd voidmatrix
   go build -o voidmatrix.exe .
   .\voidmatrix.exe
   ```

### Makefile
```bash
make          # build locally
make install  # install to $PREFIX (default: /usr/local)
make clean    # remove binary
```

### Self-update
Once installed, update to the latest release directly:
```bash
voidmatrix update
```

---

## 🕹️ Interactive Controls

All controls apply live, no restart needed.

| Key | Action |
|---|---|
| `↑` / `W` | Speed up (fast, +3) |
| `↓` / `S` | Slow down (fast, −3) |
| `→` / `+` / `=` | Speed up (slow, +1) |
| `←` / `-` | Slow down (slow, −1) |
| `D` | Density up |
| `A` | Density down |
| `[` | Toggle async stream speeds |
| `F` | Toggle glowing flashers |
| `B` | Cycle bold mode (Mixed / All / None) |
| `C` | Cycle character sets |
| `1`–`9` | Color theme: Green, Red, Blue, White, Rainbow, Purple, Cyan, Amber, Gold |
| `O` | Toggle metrics HUD overlay |
| `Space` / `Q` | Quit safely |

---

## ⚡ Visual Presets

```bash
voidmatrix --mode hacker      # fast, dense, glitching, green — classic terminal-hacker feel
voidmatrix --mode chill       # slow, dim, cyan, no bold — background ambience
voidmatrix --mode chaos       # rainbow storm, heavy left wind skew, high speed
voidmatrix --mode cinematic   # slow green rain with a light right-hand drift, movie-accurate
```

Preset values are overridden by any CLI flag you pass explicitly, so `--mode chill -c purple` gives you the chill preset but in purple.

---

## 📖 CLI Flags Reference

| Short | Long | Value | Description |
|---|---|---|---|
| `-s` | `--speed` | `float` | Speed multiplier (default `0.45`, range `0.1`–`8.0`) |
| — | `--density` | `float` | Streams per column (default `0.72`, range `0.1`–`3.0`) |
| `-c` | `--color-name` | `string` | Named theme: `green`, `red`, `blue`, `white`, `yellow`, `cyan`, `magenta`, `purple`, `rainbow` |
| — | `--color` | `int` | Theme by numeric ID (`1`–`9`) |
| `-g` | `--bg` | `string` | Background color |
| `-l` | `--charset` | `string` | Character set: single codes or aliases (`japanese`, `greek`, `binary`, `emoji`, `classic`, `cyrillic`) |
| `-u` | `--custom` | `string` | Custom character sequence (used with `-l ...u...`) |
| `-f` | `--flashers` | `bool` | Enable glowing pulse characters |
| `-a` | `--async` | `bool` | Enable asynchronous stream scrolling |
| `-b` | `--bold` | `bool` | Force bold rendering |
| `-n` | `--no-bold` | `bool` | Disable bold entirely |
| `-o` | `--no-status` | `bool` | Hide OSD overlay |
| `-i` | `--ignore-input` | `bool` | Ignore keyboard input (screensaver mode) |
| `-w` | `--wave` | `bool` | Single wave pass, exits after one sweep |
| `-t` | `--time` | `float` | Auto-exit after N seconds |
| `-d` | `--debug` | `bool` | Show live performance dashboard |
| — | `--wind` | `string` | Wind skew: `none`, `left`, `right` |
| — | `--splash` | `bool` | Enable ground splash particles |
| — | `--glitch` | `bool` | Enable transmission glitch strobes |
| `-msg` | `--message` | `string` | Centered text to reveal |
| — | `--mode` | `string` | Preset: `hacker`, `chill`, `chaos`, `cinematic` |

---

## ⚙️ Configuration

Config is loaded automatically, checked in this order (later entries override earlier ones):

1. **Built-in hardcoded defaults**
2. **Config file** (first path found wins):
   - XDG standard — `~/.config/voidmatrix/config.yaml` (Unix/macOS) or `%APPDATA%\voidmatrix\config.yaml` (Windows)
   - Legacy fallback — `~/.voidmatrix/config.yaml` (Unix/macOS) or `%USERPROFILE%\.voidmatrix\config.yaml` (Windows)
3. **`--mode` preset**, if specified
4. **Explicit CLI flags** — always win, and can be combined with a preset (`--mode chill -c purple`)

```yaml
speed: 0.45
density: 0.72
theme: green
glitch: false
wind: none
splash: false
```

---

## 📊 Live Metrics Dashboard

```bash
voidmatrix --debug
```
Shows rolling FPS, per-frame draw latency (µs), live Go heap usage, and cumulative GC cycle count.

---

## ✅ Verifying the Zero-Allocation Claim Yourself

The core simulation loop (stream ticking + snapshotting) is benchmarked directly rather than just asserted. Run it yourself:

```bash
go test -bench=BenchmarkSnapshot -benchmem -race .
```

Measured result on the reference machine:

```
goos: linux
goarch: amd64
BenchmarkSnapshot-4       415460              3002 ns/op               0 B/op           0 allocs/op
PASS
ok      github.com/austin/voidmatrix    1.310s
```

`0 B/op` / `0 allocs/op` confirms steady-state stream ticking and snapshotting do not touch the heap — including on stream respawn, where an earlier version of `newStream` was reallocating a struct and two slices every time a column died. That path was refactored into an in-place `resetStream` that reslices existing backing arrays instead.

This benchmark covers the simulation/snapshot path specifically. If you're extending the renderer or decoder logic, consider adding benchmarks for those paths too before relying on the same guarantee there.

---

## 🏗️ Architecture

```text
   ┌───────────────────────────────────────────────┐
   │ main.go (CLI flags, Config, Dual-Ticker loop) │
   └───────────────────────┬───────────────────────┘
                           │
                 Ticks     │     Polls
                 State     ▼     Input
   ┌───────────────────────────────────────────────┐
   │          state.go (Physics engine)            │◄───────┐
   └───────────────────────┬───────────────────────┘        │
                           │                                │
                 Generates │                                │ Mutates
                 Snapshot  ▼                                │ Config
   ┌───────────────────────────────────────────────┐        │
   │  renderer.go (Double-buffered cell matrices)  │        │
   └───────────────────────┬───────────────────────┘        │
                           │                                │
                 Pushes    │                                │
                 Diffs     ▼                                │
   ┌───────────────────────────────────────────────┐        │
   │               tcell/v2 Library                │────────┘
   └───────────────────────────────────────────────┘
```

| File | Responsibility |
|---|---|
| `main.go` | CLI flags, config loading, dual-ticker event loop (simulation vs. 60 FPS render) |
| `state.go` | Column/stream tracking, speed & density modifiers, splashes, Xorshift32 PRNG, snapshot pool |
| `renderer.go` | Pre-allocated cell buffers, diff rendering, resize handling, debug dashboard |
| `theme.go` | Color palettes, RGB linear interpolation for gradient fades |
| `charset.go` | Character sets (Hiragana, Katakana, Cyrillic, emoji, symbols) and aliases |
| `input.go` | Non-blocking keyboard capture, triggers redraws |

For a full technical breakdown of the rendering pipeline, PRNG, and zero-allocation snapshot pool, see [`DOCUMENTATION.md`](DOCUMENTATION.md).

---

## 🤝 Contributing

Issues and PRs are welcome. If you're adding a feature, please check `DOCUMENTATION.md` first to keep the architecture consistent with the existing separation between state, rendering, and input.

---

## 📄 License

MIT — see [LICENSE](LICENSE) for details.
