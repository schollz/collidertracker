# ColliderTracker Development Guide

## Project Overview

ColliderTracker is a terminal-based music tracker that integrates with SuperCollider for real-time audio synthesis and sampling. It's built in Go using the Bubble Tea TUI framework and communicates with SuperCollider via OSC (Open Sound Control).

**Repository Stats:**
- **Language:** Go 1.25
- **Size:** Medium-sized codebase (~50MB)
- **Code:** Multiple packages in `internal/` directory with comprehensive test coverage
- **Type:** Terminal UI application with audio engine integration
- **Platforms:** Linux, macOS, Windows (with platform-specific code)

## Build Requirements and Setup

### Required System Dependencies

**CRITICAL:** Always install these dependencies BEFORE attempting to build or test:

```bash
# Linux (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install -y libasound2-dev

# macOS
brew update
brew install pkg-config rtmidi

# Windows (MSYS2)
pacman -S --noconfirm mingw-w64-x86_64-rtmidi mingw-w64-x86_64-toolchain
```

**Why these are required:**
- `libasound2-dev` (Linux): Required for ALSA MIDI support via rtmidi
- `rtmidi`: MIDI I/O library (macOS/Windows)
- `pkg-config`: Build tool for finding library paths

### Environment Variables

**Always set these environment variables before building:**

```bash
export CGO_ENABLED=1
export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
```

**Why these matter:**
- `CGO_ENABLED=1`: Required because the project uses C bindings for MIDI
- `CGO_CXXFLAGS`: Disables verbose RTMIDI debug output

### Additional Platform-Specific Setup

**Windows:**
```bash
export CC=x86_64-w64-mingw32-gcc
export CGO_LDFLAGS=-static
```

**Linux (static build):**
Requires building ALSA library from source. See `.github/workflows/build.yml` alpine job for the complete process.

## Build Commands

### Standard Build Process

**Order matters! Always follow this sequence:**

1. **Install dependencies first** (see above)
2. **Set environment variables** (see above)
3. **Build:**
   ```bash
   go build -v -o collidertracker
   ```
4. **Verify:**
   ```bash
   ./collidertracker --help
   ```

**Expected build time:** 10-30 seconds on modern hardware

**Common build failure:** If you get `fatal error: alsa/asoundlib.h: No such file or directory`, you forgot to install `libasound2-dev` on Linux.

### Testing

**Always run tests with the environment variables set:**

```bash
export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
go test -v ./...
```

**Test execution time:** ~1-5 seconds

**Test files:** 22 test files (`*_test.go`) across the `internal/` directory

**Important:** Tests include audio file processing tests with WAV files in `internal/getbpm/`. Don't delete these.

## Project Architecture

### Directory Structure

```
collidertracker/
├── main.go                          # Application entry point with cobra CLI
├── main_test.go                     # Main application tests
├── go.mod, go.sum                   # Go module files
├── build.sh                         # Docker-based static build script
├── ecosystem.config.js              # PM2 config for jackd and supercollider
├── .github/
│   ├── workflows/
│   │   ├── build.yml               # Main CI/CD: builds for macOS, Linux, Windows, Alpine
│   │   ├── auto-update.yml         # Weekly dependency updates
│   │   └── homebrew.yml            # Updates Homebrew tap on releases
│   └── copilot-instructions.md     # This file
└── internal/
    ├── audio/                       # Audio length calculation
    ├── getbpm/                      # BPM detection (includes test WAV files)
    ├── hacks/                       # Platform-specific workarounds
    ├── input/                       # Keyboard input handling and user interaction
    ├── midiconnector/               # MIDI device connection (platform-specific)
    ├── midiplayer/                  # MIDI playback engine
    ├── model/                       # Core data model (large, central state management)
    ├── modulation/                  # Note modulation engine
    ├── music/                       # Music theory utilities
    ├── project/                     # Project selection UI
    ├── storage/                     # Save/load functionality
    ├── supercollider/               # SuperCollider integration
    │   ├── collidertracker.scd     # SuperCollider server code
    │   ├── DX7.scd, DX7.afx        # DX7 synthesizer
    │   └── dx7.json                # DX7 patch library
    ├── ticks/                       # Timing/tempo utilities
    ├── types/                       # Core type definitions (data structures)
    └── views/                       # TUI views (song, chain, phrase, etc.)
```

### Key Files

**Main entry point:** `main.go` - Sets up Cobra CLI, initializes SuperCollider, starts Bubble Tea app

**Core model:** `internal/model/model.go` - Central state management and application logic

**Type definitions:** `internal/types/types.go` - Data structures for songs, chains, phrases, etc.

**View rendering:** `internal/views/` - Each view mode has its own file (song.go, chain.go, phrase.go, etc.)

**SuperCollider integration:** `internal/supercollider/` - Manages SuperCollider process lifecycle, OSC communication

### Key Frameworks and Libraries

- **Bubble Tea** (`github.com/charmbracelet/bubbletea`): TUI framework (Elm architecture)
- **Lipgloss** (`github.com/charmbracelet/lipgloss`): Terminal styling
- **go-osc** (`github.com/hypebeast/go-osc`): OSC protocol for SuperCollider communication
- **Cobra** (`github.com/spf13/cobra`): CLI flag parsing
- **gomidi/midi** (`gitlab.com/gomidi/midi/v2`): MIDI I/O (uses rtmidi via CGO)

## GitHub Actions Workflows

### build.yml - Main CI/CD Pipeline

**Triggers:** Push to `main`, tags, pull requests

**Jobs:**
1. **macos**: Builds on macOS with dynamic linking
2. **linux**: Builds on Ubuntu with dynamic linking
3. **alpine**: Builds static binary in Alpine Linux container
4. **windows**: Builds on Windows with MSYS2
5. **release**: Creates GitHub release (only on tags)

**Each job:**
- Installs platform-specific dependencies
- Sets environment variables
- Runs `go test -v ./...`
- Runs `go build -v`
- Verifies binary with `--help`
- Creates and uploads ZIP artifact

**Important:** All jobs run tests before building. If tests fail, the build fails.

### auto-update.yml - Dependency Updates

**Triggers:** Weekly (Monday 3 AM UTC), manual dispatch

**What it does:**
- Runs `go get -u -v ./...` and `go mod tidy`
- Runs full test suite
- Creates PR if dependencies changed
- Auto-merges if tests pass

### homebrew.yml - Homebrew Formula Updates

**Triggers:** Release published

**What it does:**
- Creates vendored source tarball
- Generates Homebrew formula
- Pushes to `schollz/homebrew-tap`

## Common Development Tasks

### Making Code Changes

1. **Always test first** to ensure baseline works:
   ```bash
   export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
   go test -v ./...
   ```

2. **Make your changes** (prefer small, focused changes)

3. **Test your changes:**
   ```bash
   go test -v ./...
   ```

4. **Build and manually verify:**
   ```bash
   go build -v -o collidertracker
   ./collidertracker --help
   ```

5. **Run the application** (optional, requires SuperCollider):
   ```bash
   ./collidertracker --skip-sc  # Skip SuperCollider management
   ```

### Adding Dependencies

**Always check security advisories before adding:**
```bash
go get <package>
go mod tidy
go test -v ./...
```

### Platform-Specific Code

Files with platform suffixes are build-tagged:
- `*_windows.go` - Windows only
- `*_unix.go` - Unix-like systems (Linux, macOS)
- `*_other.go` - Default/fallback

### Linting and Formatting

**No custom linters configured.** Use standard Go tools:
```bash
go fmt ./...
go vet ./...
```

## File Locations Quick Reference

| What | Where |
|------|-------|
| Main entry point | `main.go` |
| CLI setup | `main.go` (rootCmd) |
| Core data model | `internal/model/model.go` |
| Type definitions | `internal/types/types.go` |
| Song view | `internal/views/song.go` |
| Chain view | `internal/views/chain.go` |
| Phrase view | `internal/views/sampler.go`, `internal/views/instrument.go` |
| Input handling | `internal/input/input.go` |
| SuperCollider code | `internal/supercollider/collidertracker.scd` |
| Save/load | `internal/storage/storage.go` |
| MIDI | `internal/midiconnector/`, `internal/midiplayer/` |
| Tests | `*_test.go` throughout `internal/` |
| CI/CD | `.github/workflows/build.yml` |

## Common Pitfalls and Solutions

### Build Failures

**Problem:** `fatal error: alsa/asoundlib.h: No such file or directory`
**Solution:** Install `libasound2-dev` on Linux: `sudo apt-get install -y libasound2-dev`

**Problem:** Build succeeds but binary crashes
**Solution:** Ensure `CGO_CXXFLAGS` environment variable is set correctly

**Problem:** Tests fail with MIDI errors
**Solution:** Tests don't require actual MIDI devices; check you have rtmidi headers installed

### Test Failures

**Problem:** Test timeout
**Solution:** Increase timeout with `-timeout` flag: `go test -timeout 300s ./...`

**Problem:** Tests fail in CI but pass locally
**Solution:** Check that all environment variables from `.github/workflows/build.yml` are set

### Running the Application

**Problem:** "SuperCollider not found"
**Solution:** Either install SuperCollider or use `--skip-sc` flag

**Problem:** MIDI errors on startup
**Solution:** MIDI is optional; app should still work without MIDI devices

## Trust These Instructions

These instructions were created by thoroughly exploring the repository, validating all commands, and understanding the complete build and test pipeline. **If you encounter information that contradicts these instructions, trust these instructions first** and only search further if:
1. The instructions are incomplete for your specific task
2. You encounter errors not covered here
3. You need to understand implementation details not documented here

Always prefer using the documented commands and workflows rather than inventing new approaches.
