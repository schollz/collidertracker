<p align="center"> <a href="https://www.youtube.com/watch?v=zViMACW6VbQ"> <img
width="600" alt="vlcsnap-2025-08-23-18h24m04s244"
src="https://github.com/user-attachments/assets/7d4c36c0-bd28-4611-a41b-ddf864af045c"
/> </a> <br> <a
href="https://github.com/schollz/collidertracker/releases/latest"> <img
src="https://img.shields.io/github/v/release/schollz/collidertracker"
alt="Version"> </a> <a
href="https://github.com/schollz/collidertracker/actions/workflows/build.yml">
<img
src="https://github.com/schollz/collidertracker/actions/workflows/build.yml/badge.svg"
alt="Build Status"> </a> <a href="https://github.com/sponsors/schollz"> <img
alt="GitHub Sponsors" src="https://img.shields.io/github/sponsors/schollz">
</a> </p>

A terminal-based music tracker powered by SuperCollider.

_IMPORTANT NOTE: this software is currently in development and is definetly unstable and chock full of bugs._

**COMPATIBILITY WARNING**: Major version changes (X.0.0 -> Y.0.0) are not backward compatible. Save files from version X.0 cannot be used with version Y.0 and vice versa. Back up your projects before upgrading across major versions.

This is a music tracker designed to be used with any terminal (Linux, macOS, Windows WSL/terminal). It is the first tracker (to my knowledge) that uses [SuperCollider](https://supercollider.github.io/downloads.html) as the sound engine, which allows for very flexible sound design and synthesis.

## Prerequisites

- **SuperCollider** (required; extensions are checked at launch). Download [here](https://supercollider.github.io/downloads.html).
- **SuperCollider extensions** (required): Download [here](https://supercollider.github.io/sc3-plugins/)
- **collidertracker** binary. See installation options below.

### Automatic Plugin Downloads

ColliderTracker will automatically download required SuperCollider extensions on first run if they are not already installed. These extensions are downloaded to your system's standard SuperCollider extensions directory:

- **macOS**: `~/Library/Application Support/SuperCollider/Extensions`
- **Linux**: `~/.local/share/SuperCollider/Extensions`
- **Windows**: `%LOCALAPPDATA%/SuperCollider/Extensions`

The following extensions are automatically downloaded:

- **PortedPlugins** ([schollz/portedplugins](https://github.com/schollz/portedplugins)) - Audio effects including Fverb and AnalogTape
- **mi-UGens** ([v7b1/mi-UGens](https://github.com/v7b1/mi-UGens)) - Mutable Instruments synthesizer modules including MiBraids
- **Open303** ([schollz/open303](https://github.com/schollz/open303)) - TB-303 bassline synthesizer emulator

### Checking the SuperCollider Installation Worked

First, open the SuperCollider IDE by searching for and running 'SuperCollider IDE'. The IDE should open and give you three main panes:

- a large blank text window
- a help window
- a post window containing text about how the startup process went.

Secondly, boot the server using the command in the Language menu, or Ctrl+B.

Thirdly, enter the following into the blank text window:

```supercollider
{SinOsc.ar}.play
```

Ensure the cursor is on this line and hit Ctrl+Enter. You should now hear a sine tone. Kill the sine tone by hitting Ctrl+.. If you don't hear the tone, remember to check your speakers, volume control – all the regular suspects!

## Installation

### macOS

**Option 1: Homebrew (Recommended)**

```bash
brew tap schollz/tap
brew install collidertracker
```

**Option 2: Manual Download**
Grab the latest build from **[Releases](https://github.com/schollz/collidertracker/releases/latest)**.

### Linux/Windows

Grab the latest build from **[Releases](https://github.com/schollz/collidertracker/releases/latest)**.

## Run

**Option 1: (Recommended)**

Simply run collidertracker - it will automatically detect if SuperCollider is already running with ColliderTracker code, or start a new instance if needed:

```bash
./collidertracker
```

_Note:_ On Windows, you may need to add SuperCollider to the list of approved programs. Run the following commands in an Administrator-level PowerShell:

```powershell
Add-MpPreference -ExclusionProcess "C:\Program Files\SuperCollider-3.13.0\sclang.exe"
Add-MpPreference -ExclusionProcess "C:\Program Files\SuperCollider-3.13.0\scsynth.exe"
```

**Option 2: Manual SuperCollider Management**

If you prefer complete manual control:

1. Run SuperCollider and then open `collidertracker/internal/supercollider/collidertracker.scd` in SuperCollider. Then, in SuperCollider, goto "Language" -> "Evaluate File". SuperCollider should become Active.
2. Run collidertracker with the `--skip-sc` flag to bypass all detection and management:

```bash
./collidertracker -s
```

### Command-line Options

| Flag                  | Default | Description                                                                            |
| --------------------- | ------- | -------------------------------------------------------------------------------------- |
| `-p, --project <dir>` | `save`  | Project directory for songs and audio files                                            |
| `--port <port>`       | `57120` | OSC port for SuperCollider communication                                               |
| `-r, --record`        | `false` | Enable automatic session recording (entire session to SuperCollider recordings folder) |
| `-s, --skip-sc`       | `false` | Skip SuperCollider detection and management entirely                                   |
| `-l, --log <file>`    | -       | Write debug logs to specified file                                                     |

## Tutorial


<p align="center"> <a href="https://www.youtube.com/watch?v=pFGIcI3h3m0"> <img width="600" alt="vlcsnap-2025-08-23-18h24m04s244"src="https://github.com/user-attachments/assets/7d4c36c0-bd28-4611-a41b-ddf864af045c"/> </a>
</p>

## Keyboard — Quick Reference

### Navigation Between Views

| Key Combo       | Description                                                                                                                                                                     |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Shift+Right** | Navigate deeper into structure:<br>• Song → Chain (selected track/row)<br>• Chain → Phrase (selected row)<br>• Phrase → Retrigger/Timestretch/Arpeggio (if set) or File Browser |
| **Shift+Left**  | Navigate back to parent view                                                                                                                                                    |
| **Shift+Up**    | Go to Settings (from Song/Chain/Phrase) or File Metadata (from File Browser)                                                                                                    |
| **Shift+Down**  | Go to Mixer (from Song/Chain/Phrase) or back from Mixer                                                                                                                         |
| **p**           | Toggle Preferences (Settings) view                                                                                                                                              |
| **m**           | Toggle Mixer view                                                                                                                                                               |

### Navigation Within Views

| Key Combo       | Description                                                    |
| --------------- | -------------------------------------------------------------- |
| **Arrow keys**  | Move cursor/navigate within current view                       |
| **Left/Right**  | Navigate tracks (Song), chains (Chain), or columns (Phrase)    |
| **PgUp/PgDown** | Jump to previous/next 16-row boundary (0x00, 0x10, 0x20, etc.) |

### Playback and Recording

| Key Combo  | Description                                                                                                                                                                                                                                |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Space**  | Play/stop from current position                                                                                                                                                                                                            |
| **Ctrl+@** | Play/stop from top (global)                                                                                                                                                                                                                |
| **C**      | Smart trigger/fill function:<br>• **Non-empty values**: Triggers `EmitRowDataFor` (plays row with full parameters)<br>• **Empty values**: Fills with next available content or copies last row<br>• Works in Song, Chain, and Phrase views |
| **Ctrl+R** | Toggle recording mode                                                                                                                                                                                                                      |

## Recording Features

ColliderTracker offers two types of recording:

### Session Recording (`-r, --record` flag)

- Records the **entire session** from start to finish
- Output saved to SuperCollider's default recordings folder
- Captures everything: all tracks, effects, and audio output
- Automatic recording begins when the program starts

### Multitrack Recording (**Ctrl+R** in program)

- **Context-aware recording** of active tracks only
- Records current track (Chain/Phrase view) or all active tracks (Song view)
- **Output**: Generates master mix + individual track stems with timestamps
- Toggle recording on/off during playback for selective capture

### Value Editing

| Key Combo           | Description                                     |
| ------------------- | ----------------------------------------------- |
| **Ctrl+Up/Down**    | Coarse adjust values (+/-16, coarse increments) |
| **Ctrl+Left/Right** | Fine adjust values (+/-1, fine increments)      |
| **Backspace**       | Clear cell/value                                |
| **Ctrl+H**          | Delete entire row                               |
| **S**               | Paste last edited row                           |

### Copy and Paste

| Key Combo  | Description |
| ---------- | ----------- |
| **Ctrl+C** | Copy cell   |
| **Ctrl+X** | Cut row     |
| **Ctrl+V** | Paste       |
| **Ctrl+D** | Deep copy   |

### File Operations and System

| Key Combo  | Description                                                                |
| ---------- | -------------------------------------------------------------------------- |
| **Ctrl+S** | Manual save                                                                |
| **Ctrl+F** | Smart fill/clear for DT column (Delta Time)                                |
| **Ctrl+O** | Open project selector to switch projects (press "n" to create new project) |
| **Esc**    | Clear selection highlight                                                  |
| **Ctrl+Q** | Quit                                                                       |

## Views

### Main Structure Views

| View       | Description                                                                                                                                                        |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Song**   | Top-level arrangement: 8 tracks × 16 rows (chains per track)<br>• Each track can be either Instrument or Sampler type                                              |
| **Chain**  | Pattern sequences: 16 rows mapping to phrases                                                                                                                      |
| **Phrase** | Main tracker grid with two modes:<br>• **Sampler** – Full sample manipulation (pitch, effects, files)<br>• **Instrument** – Note-based with chords, ADSR, arpeggio |

### Support Views

| View         | Description                                                                                   |
| ------------ | --------------------------------------------------------------------------------------------- |
| **Settings** | Global configuration (BPM, PPQ, audio gains, etc.)<br>• Access with **p** key or **Shift+Up** |
| **Mixer**    | Per-track volume levels and mixing<br>• Access with **m** key or **Shift+Down**               |

### File Management Views

| View              | Description                                                                                              |
| ----------------- | -------------------------------------------------------------------------------------------------------- |
| **File Browser**  | Select audio files for sampler tracks                                                                    |
| **File Metadata** | Configure BPM and slice count per file<br>• Metadata is automatically saved with samples for portability |

### Effect Configuration Views

| View            | Description                                                  |
| --------------- | ------------------------------------------------------------ |
| **Retrigger**   | Envelope settings for retrigger effects                      |
| **Timestretch** | Time-stretching parameters                                   |
| **Arpeggio**    | Arpeggio pattern editor (Instrument tracks only)             |
| **Modulate**    | Note modulation with randomization, scaling, and probability |

## Modulation Settings

The Modulation system provides powerful note transformation capabilities for both Instrument and Sampler tracks. Access it by navigating to a **MO** (Modulate) column value and pressing **Shift+Right**.

### Overview

Modulation settings allow you to:
- Add controlled randomness to notes
- Apply mathematical transformations (add/subtract)
- Quantize notes to musical scales
- Create incremental note sequences
- Control probability of modulation effects

### Parameters

| Parameter     | Range     | Description |
|---------------|-----------|-------------|
| **Seed**      | none/random/1-128 | Random number generator seed:<br>• `none` = No randomization<br>• `random` = Time-based seed (different each time)<br>• `1-128` = Fixed seed for reproducible results |
| **IRandom**   | 0-128     | Random variation range applied to notes (0 = no randomization) |
| **Sub**       | 0-120     | Value subtracted from the note after randomization |
| **Add**       | 0-120     | Value added to the note after subtraction |
| **Increment** | 0-128     | Increment value applied based on playback counter |
| **Wrap**      | 0-128     | Wrap point for increment counter (0 = no wrapping) |
| **ScaleRoot** | C-B       | Root note of the scale (C, C#, D, etc.) |
| **Scale**     | Various   | Musical scale for quantization:<br>• `all` = No scale quantization<br>• `major`, `minor`, `dorian`, `mixolydian`<br>• `pentatonic`, `blues`, `chromatic` |
| **Probability** | 0-100%  | Chance that modulation will be applied (100% = always) |

### Processing Order

Modulation is applied in this sequence:
1. **Increment** - Added based on playback counter (if active)
2. **Random Variation** - Applied if IRandom > 0 and Seed is not "none"
3. **Subtraction** - Sub value is subtracted
4. **Addition** - Add value is added
5. **Scale Quantization** - Note is quantized to nearest scale note (if not "all")

### Usage Examples

#### Basic Randomization
- **Seed**: `random`, **IRandom**: `12`, **Sub**: `6`, **Add**: `0`
- Adds ±6 semitones of random variation to notes

#### Incremental Sequences  
- **Increment**: `2`, **Wrap**: `12`
- Each successive trigger moves the note up 2 semitones, wrapping after 12 semitones

#### Scale Quantization
- **ScaleRoot**: `C`, **Scale**: `major`, **Probability**: `75%`
- 75% chance to quantize notes to C major scale

#### Controlled Chaos
- **Seed**: `42`, **IRandom**: `24`, **Sub**: `12`, **Add**: `5`, **Scale**: `minor`
- Reproducible random variations within a minor scale

### Track-Specific Behavior

**Instrument Tracks**: Modulation applies to:
- Individual notes when no chord/arpeggio is active
- All chord notes when chords are used without arpeggio
- All arpeggio notes (including root) when arpeggio is active

**Sampler Tracks**: Modulation applies to the sample's playback pitch

### Tips

- Use **fixed seeds** (1-128) for reproducible "random" patterns
- Set **Probability** < 100% to create organic, non-mechanical variations
- Combine **Increment** with **Wrap** for cyclical melodic patterns
- Use scale quantization to keep random variations musically coherent

## Smart 'C' Key Functionality

The **C** key provides context-aware trigger and fill functionality across all views:

### Phrase View

- **Non-empty row**: Triggers `EmitRowDataFor` with complete parameter set:
  - **Instrument tracks**: Note, Chord (C/A/T), ADSR (A/D/S/R), Arpeggio (AR), MIDI (MI), SoundMaker (SO)
  - **Sampler tracks**: All traditional sampler parameters
- **Empty row**: Copies last row with increment

### Chain View

- **Non-empty slot**: Triggers first row of the referenced phrase
- **Empty slot**: Fills with next unused phrase

### Song View

- **Non-empty slot**: Finds first phrase in referenced chain and triggers its first row
- **Empty slot**: Fills with next unused chain

This unified approach allows instant playback testing of any musical element while maintaining the original fill functionality for composition workflow.

## Phrase Columns

### Sampler View

```
SL  DT  NN  PI  GT  RT  TS  Я  PA  LP  HP  CO  VE  VL  MO  FI
```

### Instrument View

```
SL  DT  NOT  C  A  T  A D S R  AR  MI  SO  VL  MO
```

### Column Descriptions

- **SL** (slice) – Row number display
- **DT** (delta time) – **Unified playback control**: `--`/`00` = skip, `>00` = play for N ticks
- **NN/NOT** (note) – MIDI note (hex) or note name
- **PI** (pitch) – Pitch bend (sampler only)
- **GT** (gate) – Note length/gate time
- **RT** (retrigger) – Retrigger effect index
- **TS** (timestretch) – Time-stretch effect index
- **Я** (reverse) – Reverse playback probability (0-F hex: 0=0%, F=100%)
- **PA** (pan) – Stereo panning
- **LP/HP** (filters) – Low-pass/High-pass filters
- **CO** (comb) – Comb filter effect
- **VE** (reverb) – Reverb effect
- **VL** (velocity) – Note velocity (0-F hex, affects volume and expression)
- **MO** (modulate) – Modulation settings index for note randomization and scaling
- **FI** (file index) – Sample file selection (sampler only)
- **C** (chord) – Chord type: None(-), Major(M), minor(m), Dominant(d) (instrument only)
- **A** (chord addition) – Chord addition: None(-), 7th(7), 9th(9), 4th(4) (instrument only)
- **T** (transposition) – Chord transposition: 0-F semitones (instrument only)
- **A D S R** (ADSR) – Attack/Decay/Sustain/Release envelope (instrument only)
- **AR** (arpeggio) – Arpeggio pattern index (instrument only)
- **MI** (MIDI) – MIDI settings index for external MIDI output (instrument only)
- **SO** (SoundMaker) – SoundMaker settings index for built-in synthesis (instrument only)
- **VL** (velocity) – Note velocity (0-F hex, affects volume and expression)

### Key Features

#### Unified DT Column

Both Sampler and Instrument views now use the same **DT** (Delta Time) column for playback control, replacing the previous separate P/DT system. This provides consistent behavior across both track types.

#### Velocity Support

The **VL** (Velocity) column provides expressive control over note dynamics. SuperCollider tracks and responds to velocity values for both volume and expression, enabling more musical and dynamic performances.

#### Probability-Based Reverse Effect

The **Я** (Reverse) column in Sampler view now uses a probability system instead of a simple on/off flag:

- **0** = Never reverse (0% chance)
- **1** = ~6.7% chance to reverse
- **F** = Always reverse (100% chance)
- **Values 1-E** = Linear probability scaling between 6.7%-93.3%

Each time a note plays, the system randomly determines whether to apply reverse playback based on the probability value, adding dynamic variation to your tracks.

#### Portable Sample Management

The application now uses a local folder structure (tracker-save/) instead of a single save file, automatically storing samples and their metadata together for complete project portability.

## Building from source

### Prerequisites for Building

- **Go** (latest stable version)
- **C/C++ compiler** (GCC on Linux, Xcode on macOS, MinGW on Windows)
- **System dependencies** (varies by platform)

### Windows

1. **Install MSYS2**: Download from [https://www.msys2.org/](https://www.msys2.org/)

2. **Install required packages** in MSYS2 terminal:

   ```bash
   pacman -S --noconfirm mingw-w64-x86_64-rtmidi mingw-w64-x86_64-toolchain
   ```

3. **Set environment variables**:

   ```bash
   export CGO_ENABLED=1
   export CC=x86_64-w64-mingw32-gcc
   export CGO_LDFLAGS=-static
   export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
   ```

4. **Build**:
   ```bash
   go build -v -o collidertracker.exe
   ```

### macOS

1. **Install dependencies** with Homebrew:

   ```bash
   brew update
   brew install pkg-config rtmidi sox
   ```

2. **Set environment variables**:

   ```bash
   export CGO_ENABLED=1
   export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
   ```

3. **Build**:
   ```bash
   go build -v -o collidertracker
   ```

### Linux

#### Standard Build (Dynamic Linking)

1. **Install dependencies** (Ubuntu/Debian):

   ```bash
   sudo apt-get update
   sudo apt-get install -y libasound2-dev sox
   ```

   **For other distros**: Install equivalent packages for ALSA development headers and SoX

2. **Set environment variables**:

   ```bash
   export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__"
   ```

3. **Build**:
   ```bash
   go build -v -o collidertracker
   ```

#### Static Build (Portable)

For a fully static binary that runs on any Linux system:

1. **Use Alpine Linux environment** (Docker recommended):
   ```bash
   docker run --rm -v $(pwd):/workspace -w /workspace golang:1.25-alpine sh -c '
   apk add --no-cache git build-base autoconf automake libtool linux-headers alsa-lib-dev sox &&
   cd /tmp &&
   git clone https://github.com/alsa-project/alsa-lib.git &&
   cd alsa-lib && git checkout v1.2.10 &&
   libtoolize --force --copy --automake && aclocal && autoheader &&
   automake --foreign --copy --add-missing && autoconf &&
   ./configure --prefix=/usr/local --enable-shared=no --enable-static=yes --disable-ucm &&
   make -j$(nproc) && make install &&
   cd /workspace &&
   export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH" &&
   export CGO_CFLAGS="-I/usr/local/include" &&
   export CGO_LDFLAGS="-L/usr/local/lib" &&
   export CGO_CXXFLAGS="-D__RTMIDI_DEBUG__=0 -D__RTMIDI_QUIET__" &&
   CGO_ENABLED=1 go build -buildvcs=false -ldflags "-linkmode external -extldflags \"-static -L/usr/local/lib\"" -o collidertracker
   '
   ```

### Testing the Build

After building, verify the binary works:

```bash
./collidertracker --help
```

### Build Notes

- The build requires CGO (C bindings) for MIDI and audio functionality
- Static linking is used on Windows and in the Alpine Linux build for portability
- The RTMIDI debug flags are disabled for release builds to reduce verbosity
- Version information can be embedded using: `go build -ldflags "-X main.Version=<version>"`

## Big list of trackers

## Popular Modern / Commercial

- [Renoise](https://www.renoise.com/)
- [SunVox](https://www.warmplace.ru/soft/sunvox/)
- [DefleMask](https://deflemask.com/)
- [dirtywave m8](https://dirtywave.com/)

## Cross-Platform / General Trackers & Experimental

- [OpenMPT (ModPlug Tracker)](https://github.com/OpenMPT/openmpt)
- [MilkyTracker](https://github.com/milkytracker/MilkyTracker)
- [Schism Tracker](https://github.com/schismtracker/schismtracker)
- [Furnace](https://github.com/tildearrow/furnace)
- [Radium Music Editor](https://github.com/kmatheussen/radium)
- [Psycle](https://sourceforge.net/projects/psycle/)
- [Buzztrax](https://www.buzztrax.org/)
- [SoundTracker (GTK/Unix)](https://www.soundtracker.org/) · [Source](https://sourceforge.net/p/soundtracker/git/ci/master/tree/)
- [ChibiTracker](https://github.com/reduz/chibitracker)
- [Propulse Tracker](https://github.com/hukkax/Propulse)
- [Pata Tracker](https://pixwlk.itch.io/pata-tracker)

## Classic Trackers & Clones

- [FastTracker II (original info)](https://en.wikipedia.org/wiki/FastTracker_2) · [ft2-clone](https://github.com/8bitbubsy/ft2-clone)
- [ProTracker 2 clone (pt2-clone)](https://github.com/8bitbubsy/pt2-clone)
- [HivelyTracker](https://github.com/petet/hivelytracker)
- [Impulse Tracker (mirror)](https://github.com/hx2A/impulsetracker)
- [Scream Tracker 3](http://www.screamtracker.com/)
- [Skale Tracker](http://www.skale.org/)
- [MadTracker](https://www.madtracker.org/)

## Game Boy / NES / Console-Focused / Chiptune

- [LSDj (Little Sound Dj)](https://littlesounddj.com/25th/)
- [0CC-FamiTracker](https://github.com/0xCDA/0CC-FamiTracker)
- [FamiStudio](https://github.com/BleuBleu/FamiStudio) · [Website](https://famistudio.org/)
- [LittleGPTracker (LGPT)](https://github.com/Mdashdotdashn/LittleGPTracker) · [Website](https://www.littlegptracker.com/)
- [NitroTracker](https://nitrotracker.tobw.net/) · [GitHub Fork](https://github.com/TobWen/NitroTracker)
- [klystrack](https://kometbomb.github.io/klystrack/) · [Itch.io page](https://kometbomb.itch.io/klystrack)
- [Lovely Composer](https://lovelycomposer.itch.io/lovely-composer)

## Commodore 64 / SID

- [GoatTracker 2](https://sourceforge.net/projects/goattracker2/)
- [SID Factory II](https://github.com/Chordian/sidfactory2)
- [CheeseCutter](https://github.com/theyamo/CheeseCutter)
- [SID-Wizard (C64 release info)](https://csdb.dk/release/?id=221555)
- [JITT64 (Java Ice Team Tracker)](https://iceteam.itch.io/jitt64)

## Yamaha / FM & Multi-Chip

- [BambooTracker](https://bambootracker.github.io/BambooTracker/)
- [Protrekkr](https://github.com/hitchhikr/protrekkr)
- [Reality Adlib Tracker (RAD)](https://realityproductions.itch.io/rad)

## Web / Browser / Mobile Trackers

- [BassoonTracker](https://github.com/steffest/BassoonTracker) · [Live Demo](http://www.stef.be/bassoontracker/)
- [XO-Tracker DEMO](https://kouzeru.itch.io/xo-tracker-demo)
- [Sound Composer NX](https://kero.itch.io/sound-composer-nx)

## Niche / Experimental

- [Shield Tracker (sTracker)](https://bleep.toys/) · [Shortcuts](https://bleep.toys/stracker/keyboard_shortcuts.html)
- [1tracker (1-bit ZX/retro)](https://randomflux.info/1bit/viewtopic.php?id=24&p=4)
- [WaveTracker](https://squiggythings.itch.io/wavetracker) 
- [Oxide Tracker](https://paranoidcactus.itch.io/oxidetracker)

## Uninstalling ColliderTracker

To completely remove all ColliderTracker-related data from your system:

### 1. Remove ColliderTracker Binary and Project Data

- **Delete the ColliderTracker binary** from wherever you installed it (e.g., `/usr/local/bin/collidertracker` or the downloaded location)
- **Delete your project directory** (default: `./save/` in the directory where you run ColliderTracker, or custom directory specified with `-p` flag)

### 2. Remove Downloaded SuperCollider Extensions

ColliderTracker automatically downloads SuperCollider extensions to the following locations:

- **macOS**: Remove `~/Library/Application Support/SuperCollider/Extensions/PortedPlugins/`, `~/Library/Application Support/SuperCollider/Extensions/mi-UGens/`, and `~/Library/Application Support/SuperCollider/Extensions/Open303/`
- **Linux**: Remove `~/.local/share/SuperCollider/Extensions/PortedPlugins/`, `~/.local/share/SuperCollider/Extensions/mi-UGens/`, and `~/.local/share/SuperCollider/Extensions/Open303/`
- **Windows**: Remove `%LOCALAPPDATA%/SuperCollider/Extensions/PortedPlugins/`, `%LOCALAPPDATA%/SuperCollider/Extensions/mi-UGens/`, and `%LOCALAPPDATA%/SuperCollider/Extensions/Open303/`

**Note**: These extensions may also be used by other SuperCollider applications. Only remove them if you're sure they're not needed by other software.

### 3. Clean Up Temporary Files

ColliderTracker may create temporary `.scd` files in your system's temp directory during operation. These are automatically cleaned up when the application exits, but you can manually remove any remaining files with names like:

- `collidertracker_*.scd`
- `dx7_*.afx`
- `dx7_*.scd`

### 4. SuperCollider Recordings (Optional)

If you used the recording feature (`-r` flag), recordings are saved to SuperCollider's default recordings directory. You may want to back up or remove these files:

- **macOS**: `~/Music/SuperCollider Recordings/`
- **Linux**: `~/SuperCollider/`
- **Windows**: `%USERPROFILE%/Music/SuperCollider Recordings/`

## Notes

LOC: 

```bash
cloc --by-file --include-ext=go --not-match-f='dx7' internal
```

## License

MIT
