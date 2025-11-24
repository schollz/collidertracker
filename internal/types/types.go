package types

import (
	"math"
)

type ViewMode int

const (
	SongView ViewMode = iota
	ChainView
	PhraseView
	FileView
	SettingsView
	FileMetadataView
	RetriggerView
	TimestrechView
	ModulateView
	MixerView
	ArpeggioView
	MidiView
	SoundMakerView
	DuckingView
	WaveformView
)

type PhraseViewType int

const (
	SamplerPhraseView PhraseViewType = iota
	InstrumentPhraseView
)

type CellType int

const (
	HexCell CellType = iota
	FilenameCell
)

type ClipboardMode int

const (
	CellMode ClipboardMode = iota
	RowMode
)

type SOColumnMode int

const (
	SOModeSound SOColumnMode = iota // SO column shows SoundMaker
	SOModeMIDI                      // SO column shows MIDI
)

type PhraseColumn int

const (
	ColNote               PhraseColumn = iota // Column 0: Note value (hex)
	ColPitch                                  // Column 1: Pitch value (hex, default 80/0x80 = 0.0 pitch)
	ColDeltaTime                              // Column 2: Delta time (hex) - controls playback: >0=play, <=0=skip
	ColGate                                   // Column 3: Gate value (hex, default 80/0x50, sticky)
	ColRetrigger                              // Column 4: Retrigger setting index (hex, 00-FE)
	ColTimestretch                            // Column 5: Timestretch setting index (hex, 00-FE)
	ColModulate                               // Column 6: Modulate setting index (hex, 00-FE)
	ColEffectReverse                          // Column 7: Я (reverse) 0-F (0-15, probability where F=100%, 0=0%)
	ColPan                                    // Column 8: PA (pan) (hex, default 128/0x80, 00-FE maps -1.0 to 1.0)
	ColLowPassFilter                          // Column 9: LP (low pass filter) (hex, default FE/20kHz, 00-FE maps 20Hz to 20kHz exponentially)
	ColHighPassFilter                         // Column 10: HP (high pass filter) (hex, default -1/null, 00-FE maps 20Hz to 20kHz exponentially)
	ColEffectComb                             // Column 11: CO (00-FE)
	ColEffectReverb                           // Column 12: VE (00-FE)
	ColEffectDucking                          // Column 13: DU (00-FE, sticky)
	ColFilename                               // Column 14: Filename index
	ColChord                                  // Column 15: Chord (Instrument view only: "-", "M", "m", "d")
	ColChordAddition                          // Column 16: Chord Addition (Instrument view only: "-", "7", "9", "4")
	ColChordTransposition                     // Column 17: Chord Transposition (Instrument view only: "-", "0"-"F")
	ColArpeggio                               // Column 18: Arpeggio (Instrument view only: 00-FE)
	ColMidi                                   // Column 19: MIDI (Instrument view only: 00-FE, sticky)
	ColSoundMaker                             // Column 20: SoundMaker (Instrument view only: 00-FE, sticky)
	ColAttack                                 // Column 21: Attack (Instrument view only: 00-FE, 0.02-30s exponential, default -1, sticky)
	ColDecay                                  // Column 22: Decay (Instrument view only: 00-FE, 0.0-30.0s linear, default -1, sticky)
	ColSustain                                // Column 23: Sustain (Instrument view only: 00-FE, 0.0-1.0 linear, default -1, sticky)
	ColRelease                                // Column 24: Release (Instrument view only: 00-FE, 0.02-30s exponential, default -1, sticky)
	ColVelocity                               // Column 25: Velocity (VE) (00-7F, 0-127)
	// MIDI CC columns (Instrument view only, visible when SO/MI column is in MI mode)
	ColMidiCC0 // Column 26: MIDI CC 0 (00-7F, 0-127)
	ColMidiCC1 // Column 27: MIDI CC 1 (00-7F, 0-127)
	ColMidiCC2 // Column 28: MIDI CC 2 (00-7F, 0-127)
	ColMidiCC3 // Column 29: MIDI CC 3 (00-7F, 0-127)
	ColMidiCC4 // Column 30: MIDI CC 4 (00-7F, 0-127)
	ColMidiCC5 // Column 31: MIDI CC 5 (00-7F, 0-127)
	ColMidiCC6 // Column 32: MIDI CC 6 (00-7F, 0-127)
	ColMidiCC7 // Column 33: MIDI CC 7 (00-7F, 0-127)
	ColMidiCC8 // Column 34: MIDI CC 8 (00-7F, 0-127)
	ColCount   // Total number of columns
)

// ChordType represents different chord types for instrument tracks
type ChordType int

const (
	ChordNone      ChordType = iota // "-" (default)
	ChordMajor                      // "M"
	ChordMinor                      // "m"
	ChordDominant                   // "d"
	ChordTypeCount                  // Total number of chord types
)

// ChordAddition represents different chord additions for instrument tracks
type ChordAddition int

const (
	ChordAddNone       ChordAddition = iota // "-" (default)
	ChordAdd7                               // "7"
	ChordAdd9                               // "9"
	ChordAdd4                               // "4"
	ChordAdditionCount                      // Total number of chord additions
)

// ChordTransposition represents different chord transpositions for instrument tracks (hex values)
type ChordTransposition int

const (
	ChordTransNone          ChordTransposition = iota // "-" (default, same as 0)
	ChordTrans0                                       // "0"
	ChordTrans1                                       // "1"
	ChordTrans2                                       // "2"
	ChordTrans3                                       // "3"
	ChordTrans4                                       // "4"
	ChordTrans5                                       // "5"
	ChordTrans6                                       // "6"
	ChordTrans7                                       // "7"
	ChordTrans8                                       // "8"
	ChordTrans9                                       // "9"
	ChordTransA                                       // "A"
	ChordTransB                                       // "B"
	ChordTransC                                       // "C"
	ChordTransD                                       // "D"
	ChordTransE                                       // "E"
	ChordTransF                                       // "F"
	ChordTranspositionCount                           // Total number of chord transpositions
)

// ChordTypeToString converts a ChordType enum to its display string
func ChordTypeToString(chordType ChordType) string {
	switch chordType {
	case ChordNone:
		return "-"
	case ChordMajor:
		return "M"
	case ChordMinor:
		return "m"
	case ChordDominant:
		return "d"
	default:
		return "-"
	}
}

// ChordAdditionToString converts a ChordAddition enum to its display string
func ChordAdditionToString(chordAdd ChordAddition) string {
	switch chordAdd {
	case ChordAddNone:
		return "-"
	case ChordAdd7:
		return "7"
	case ChordAdd9:
		return "9"
	case ChordAdd4:
		return "4"
	default:
		return "-"
	}
}

// UI Column positions for Instrument Phrase View - to prevent hardcoding issues
type InstrumentUIColumn int

const (
	InstrumentColSL    InstrumentUIColumn = 0  // SL - Slice (display only)
	InstrumentColDT    InstrumentUIColumn = 1  // DT - Delta Time
	InstrumentColNOT   InstrumentUIColumn = 2  // NOT - Note
	InstrumentColMO    InstrumentUIColumn = 3  // MO - Modulation
	InstrumentColC     InstrumentUIColumn = 4  // C - Chord
	InstrumentColA     InstrumentUIColumn = 5  // A - Chord Addition
	InstrumentColT     InstrumentUIColumn = 6  // T - Chord Transposition
	InstrumentColVE    InstrumentUIColumn = 7  // VE - Velocity
	InstrumentColGT    InstrumentUIColumn = 8  // GT - Gate
	InstrumentColATK   InstrumentUIColumn = 9  // A - Attack
	InstrumentColDECAY InstrumentUIColumn = 10 // D - Decay
	InstrumentColSUS   InstrumentUIColumn = 11 // S - Sustain
	InstrumentColREL   InstrumentUIColumn = 12 // R - Release
	InstrumentColRE    InstrumentUIColumn = 13 // RE - Reverb
	InstrumentColCO    InstrumentUIColumn = 14 // CO - Comb
	InstrumentColPA    InstrumentUIColumn = 15 // PA - Pan
	InstrumentColLP    InstrumentUIColumn = 16 // LP - LowPass
	InstrumentColHP    InstrumentUIColumn = 17 // HP - HighPass
	InstrumentColAR    InstrumentUIColumn = 18 // AR - Arpeggio
	InstrumentColSOMI  InstrumentUIColumn = 19 // SO/MI - SoundMaker/MIDI (toggleable)
	InstrumentColDU    InstrumentUIColumn = 20 // DU - Ducking
)

// UI Column positions for Sampler Phrase View - to prevent hardcoding issues
type SamplerUIColumn int

const (
	SamplerColSL  SamplerUIColumn = 0  // SL - Slice (display only)
	SamplerColDT  SamplerUIColumn = 1  // DT - Delta Time
	SamplerColNN  SamplerUIColumn = 2  // NN - Note
	SamplerColMO  SamplerUIColumn = 3  // MO - Modulate
	SamplerColVE  SamplerUIColumn = 4  // VE - Velocity
	SamplerColPI  SamplerUIColumn = 5  // PI - Pitch
	SamplerColGT  SamplerUIColumn = 6  // GT - Gate
	SamplerColRT  SamplerUIColumn = 7  // RT - Retrigger
	SamplerColTS  SamplerUIColumn = 8  // TS - Timestretch
	SamplerColREV SamplerUIColumn = 9  // Я - Reverse
	SamplerColPA  SamplerUIColumn = 10 // PA - Pan
	SamplerColLP  SamplerUIColumn = 11 // LP - Low Pass Filter
	SamplerColHP  SamplerUIColumn = 12 // HP - High Pass Filter
	SamplerColCO  SamplerUIColumn = 13 // CO - Comb
	SamplerColRE  SamplerUIColumn = 14 // RE - Reverb
	SamplerColDU  SamplerUIColumn = 15 // DU - Ducking
	SamplerColFI  SamplerUIColumn = 16 // FI - Filename
)

// UI Column positions for Arpeggio View - to prevent hardcoding issues
type ArpeggioUIColumn int

const (
	ArpeggioColDI  ArpeggioUIColumn = 0 // DI - Direction
	ArpeggioColCO  ArpeggioUIColumn = 1 // CO - Count
	ArpeggioColDIV ArpeggioUIColumn = 2 // Divisor
)

// ChordTranspositionToString converts a ChordTransposition enum to its display string
func ChordTranspositionToString(chordTrans ChordTransposition) string {
	switch chordTrans {
	case ChordTransNone:
		return "-"
	case ChordTrans0:
		return "0"
	case ChordTrans1:
		return "1"
	case ChordTrans2:
		return "2"
	case ChordTrans3:
		return "3"
	case ChordTrans4:
		return "4"
	case ChordTrans5:
		return "5"
	case ChordTrans6:
		return "6"
	case ChordTrans7:
		return "7"
	case ChordTrans8:
		return "8"
	case ChordTrans9:
		return "9"
	case ChordTransA:
		return "A"
	case ChordTransB:
		return "B"
	case ChordTransC:
		return "C"
	case ChordTransD:
		return "D"
	case ChordTransE:
		return "E"
	case ChordTransF:
		return "F"
	default:
		return "-"
	}
}

func GetChordNotes(root int, ctype ChordType, add ChordAddition, transpose ChordTransposition) []int {
	notes := []int{root}

	if ctype == ChordNone {
		return notes
	}

	switch ctype {
	case ChordMajor:
		notes = append(notes, root+4, root+7)
	case ChordMinor:
		notes = append(notes, root+3, root+7)
	case ChordDominant:
		notes = append(notes, root+4, root+7) // dominant is like major triad, add7 handled below
	}

	switch add {
	case ChordAdd7:
		if ctype == ChordMinor {
			notes = append(notes, root+10) // minor 7th
		} else {
			notes = append(notes, root+11) // major 7th
		}
	case ChordAdd9:
		notes = append(notes, root+14) // 9th = collidertrackerd + octave
	case ChordAdd4:
		notes = append(notes, root+5) // 4th
	}

	if transpose > ChordTrans0 {
		for i := ChordTrans0; i < transpose; i++ {
			first := notes[0]
			notes = notes[1:]
			notes = append(notes, first+12)
		}
	}

	return notes
}

type FileMetadata struct {
	BPM          float32   `json:"bpm"`          // Source BPM for the file
	Slices       int       `json:"slices"`       // Number of slices in the file
	Playthrough  int       `json:"playthrough"`  // 0=Sliced, 1=Oneshot, 2=Slice Bounce, 3=Slice Stop
	SyncToBPM    int       `json:"synctobpm"`    // 0=No, 1=Yes (default)
	SliceType    int       `json:"slicetype"`    // 0=Even (default), 1=Onsets
	Onsets       []float64 `json:"onsets"`       // Onset times in seconds (populated when SliceType=1)
	WaveformFile string    `json:"waveformfile"` // Path to 16-bit mono .wav file for waveform visualization (generated by audiomorph)
}

type RetriggerSettings struct {
	Times              int     `json:"times"`              // Number of retriggers (0-256)
	Start              float32 `json:"start"`              // Starting rate (0-256, 0.05 increments) /beat
	End                float32 `json:"end"`                // Final rate (0-256, 0.05 increments, default 0) /beat
	Beats              int     `json:"beats"`              // Beats value (0-256)
	VolumeDB           float32 `json:"volumeDB"`           // Volume change (-16 to +16 dB, default 0) - NOTE: This is for retrigger-specific volume, not global
	PitchChange        float32 `json:"pitchChange"`        // Pitch change (-24 to +24, default 0)
	FinalPitchToStart  int     `json:"finalPitchToStart"`  // Final pitch to start: 0=No, 1=Yes (default 0)
	FinalVolumeToStart int     `json:"finalVolumeToStart"` // Final volume to start: 0=No, 1=Yes (default 0)
	Every              int     `json:"every"`              // Every N steps (1-64, default 1) - retrigger activates when step_count % Every == 0
	Probability        int     `json:"probability"`        // Probability percentage (0-100, default 100) - chance of activation after Every check
}

type TimestrechSettings struct {
	Start       float32 `json:"start"`       // Start value (0-256, 0.05 increments)
	End         float32 `json:"end"`         // End value (0-256, 0.05 increments, default 0)
	Beats       int     `json:"beats"`       // Beats value (0-256)
	Every       int     `json:"every"`       // Every N steps (1-64, default 1) - timestretch activates when step_count % Every == 0
	Probability int     `json:"probability"` // Probability percentage (0-100, default 100) - chance of activation after Every check
}

type ModulateSettings struct {
	Seed        int    `json:"seed"`        // Random seed: -1 for "none" (no randomization), 0 for "random" (time seeding), 1-128 for fixed seed
	IRandom     int    `json:"irandom"`     // Random range: 0-128 (0 means no randomization)
	Sub         int    `json:"sub"`         // Subtract value: 0-120
	Add         int    `json:"add"`         // Add value: 0-120
	Increment   int    `json:"increment"`   // Increment value: 0-128 (added to note when increment counter > -1)
	Wrap        int    `json:"wrap"`        // Wrap value: 0-128 (0 = none, wraps increment counter when exceeded)
	ScaleRoot   int    `json:"scaleRoot"`   // Scale root note: 0-11 (C, C#, D, D#, E, F, F#, G, G#, A, A#, B)
	Scale       string `json:"scale"`       // Scale selection: "all", "major", "minor", etc.
	Probability int    `json:"probability"` // Probability percentage: 0-100 (100 = always apply modulation)
}

type DuckingSettings struct {
	Type    int     `json:"type"`    // Type: 0=none, 1=ducking, 2=ducked
	Bus     int     `json:"bus"`     // Bus: 0-7
	Attack  float32 `json:"attack"`  // Attack time: 0.0-2.0 seconds
	Release float32 `json:"release"` // Release time: 0.0-2.0 seconds
	Depth   float32 `json:"depth"`   // Depth: 0.0-1.0
	Thresh  float32 `json:"thresh"`  // Threshold: 0.0-1.0, default 0.02
}

// ArpeggioDirection represents different arpeggio directions
type ArpeggioDirection int

const (
	ArpeggioDirectionNone ArpeggioDirection = iota // 0: "--"
	ArpeggioDirectionUp                            // 1: "u-"
	ArpeggioDirectionDown                          // 2: "d-"
)

// SoundMakerRow represents different rows in the SoundMaker settings view
type SoundMakerRow int

const (
	SoundMakerRowName   SoundMakerRow = iota // 0: Name row
	SoundMakerRowParamA                      // 1: Parameter A / Preset row
	SoundMakerRowParamB                      // 2: Parameter B row
	SoundMakerRowParamC                      // 3: Parameter C row
	SoundMakerRowParamD                      // 4: Parameter D row
)

// GlobalSettingsRow represents different rows in the Global settings column
type GlobalSettingsRow int

const (
	GlobalSettingsRowBPM            GlobalSettingsRow = iota // 0: BPM
	GlobalSettingsRowPPQ                                     // 1: PPQ
	GlobalSettingsRowPregainDB                               // 2: PregainDB
	GlobalSettingsRowPostgainDB                              // 3: PostgainDB
	GlobalSettingsRowBiasDB                                  // 4: BiasDB
	GlobalSettingsRowSaturationDB                            // 5: SaturationDB
	GlobalSettingsRowDriveDB                                 // 6: DriveDB
	GlobalSettingsRowTapePercent                             // 7: TapePercent
	GlobalSettingsRowShimmerPercent                          // 8: ShimmerPercent
)

// InputSettingsRow represents different rows in the Input settings column
type InputSettingsRow int

const (
	InputSettingsRowInputLevelDB      InputSettingsRow = iota // 0: InputLevelDB
	InputSettingsRowReverbSendPercent                         // 1: ReverbSendPercent
)

// BrailleDotRow represents different rows in a 2x4 Braille cell
type BrailleDotRow int

const (
	BrailleDotRow0 BrailleDotRow = iota // 0: Top row
	BrailleDotRow1                      // 1: Second row
	BrailleDotRow2                      // 2: Third row
	BrailleDotRow3                      // 3: Bottom row
)

// FileMetadataRow represents different rows in the file metadata view
type FileMetadataRow int

const (
	FileMetadataRowBPM         FileMetadataRow = iota // 0: BPM
	FileMetadataRowSlices                             // 1: Slices
	FileMetadataRowSliceType                          // 2: Slice Type
	FileMetadataRowPlaythrough                        // 3: Playthrough
	FileMetadataRowSyncToBPM                          // 4: Sync to BPM
)

// MidiSettingsRow represents different rows in the MIDI settings view
type MidiSettingsRow int

const (
	MidiSettingsRowDevice  MidiSettingsRow = iota // 0: MIDI Device
	MidiSettingsRowChannel                        // 1: MIDI Channel
)

// RetriggerSettingsRow represents different rows in the retrigger settings view
type RetriggerSettingsRow int

const (
	RetriggerSettingsRowTimes              RetriggerSettingsRow = iota // 0: Times
	RetriggerSettingsRowStartingRate                                   // 1: Starting Rate
	RetriggerSettingsRowFinalRate                                      // 2: Final Rate
	RetriggerSettingsRowBeats                                          // 3: Beats
	RetriggerSettingsRowVolume                                         // 4: Volume
	RetriggerSettingsRowPitch                                          // 5: Pitch
	RetriggerSettingsRowFinalPitchToStart                              // 6: FinalPitchToStart
	RetriggerSettingsRowFinalVolumeToStart                             // 7: FinalVolumeToStart
	RetriggerSettingsRowEvery                                          // 8: Every
	RetriggerSettingsRowProbability                                    // 9: Probability
)

// TimestrechSettingsRow represents different rows in the timestrech settings view
type TimestrechSettingsRow int

const (
	TimestrechSettingsRowStart       TimestrechSettingsRow = iota // 0: Start
	TimestrechSettingsRowEnd                                      // 1: End
	TimestrechSettingsRowBeats                                    // 2: Beats
	TimestrechSettingsRowEvery                                    // 3: Every
	TimestrechSettingsRowProbability                              // 4: Probability
)

// ModulateSettingsRow represents different rows in the modulate settings view
type ModulateSettingsRow int

const (
	ModulateSettingsRowSeed        ModulateSettingsRow = iota // 0: Seed
	ModulateSettingsRowIRandom                                // 1: IRandom
	ModulateSettingsRowSub                                    // 2: Sub
	ModulateSettingsRowAdd                                    // 3: Add
	ModulateSettingsRowIncrement                              // 4: Increment
	ModulateSettingsRowWrap                                   // 5: Wrap
	ModulateSettingsRowScaleRoot                              // 6: ScaleRoot
	ModulateSettingsRowScale                                  // 7: Scale
	ModulateSettingsRowProbability                            // 8: Probability
)

// DuckingSettingsRow represents different rows in the ducking settings view
type DuckingSettingsRow int

const (
	DuckingSettingsRowType    DuckingSettingsRow = iota // 0: Type
	DuckingSettingsRowBus                               // 1: Bus
	DuckingSettingsRowDepth                             // 2: Depth
	DuckingSettingsRowAttack                            // 3: Attack
	DuckingSettingsRowRelease                           // 4: Release
	DuckingSettingsRowThresh                            // 5: Thresh
)

type ArpeggioRow struct {
	Direction int `json:"direction"` // Direction: 0="--", 1="u-", 2="d-"
	Count     int `json:"count"`     // Count: -1="--", 0-254 for hex values 00-FE
	Divisor   int `json:"divisor"`   // Divisor: -1="--", 1-254 for hex values 01-FE
}

type ArpeggioSettings struct {
	Rows [16]ArpeggioRow `json:"rows"` // 16 rows (00-0F), each with its own DI and CO
}

type MidiSettings struct {
	Device  string `json:"device"`  // MIDI Device name
	Channel string `json:"channel"` // MIDI Channel (1-16 or "all")
}

type SoundMakerSettings struct {
	Name       string             `json:"name"`       // SoundMaker name ("PolyPerc", "Infinite Pad", "DX7", etc.)
	Parameters map[string]float32 `json:"parameters"` // Key-value pairs for parameters (e.g. "preset": 5, "A": 128)
	PatchName  string             `json:"patchName"`  // Patch name (used for DX7 when setting by name)
}

type ClipboardData struct {
	// Cell data
	Value    int
	CellType CellType
	// Row data
	RowData     []int
	RowFilename string
	SourceView  ViewMode
	// Arpeggio row data
	ArpeggioRowData struct {
		Direction []int
		Count     []int
		Divisor   []int
	}
	// Highlighting
	HighlightRow    int
	HighlightCol    int
	HighlightPhrase int
	HighlightView   ViewMode
	// Common
	Mode    ClipboardMode
	HasData bool
	// Flag to indicate this value is a fresh deep copy that shouldn't be deep copied again
	IsFreshDeepCopy bool
}

type SaveData struct {
	ViewMode      ViewMode     `json:"viewMode"`
	CurrentRow    int          `json:"currentRow"`
	CurrentCol    int          `json:"currentCol"`
	ScrollOffset  int          `json:"scrollOffset"`
	CurrentPhrase int          `json:"currentPhrase"`
	FileSelectRow int          `json:"fileSelectRow"`
	FileSelectCol int          `json:"fileSelectCol"`
	ChainsData    [][]int      `json:"chainsData"`
	PhrasesData   [255][][]int `json:"phrasesData"`
	// New separate data pools for Instruments and Samplers
	InstrumentChainsData       [][]int                 `json:"instrumentChainsData"`
	InstrumentPhrasesData      [255][][]int            `json:"instrumentPhrasesData"`
	SamplerChainsData          [][]int                 `json:"samplerChainsData"`
	SamplerPhrasesData         [255][][]int            `json:"samplerPhrasesData"`
	SamplerPhrasesFiles        []string                `json:"samplerPhrasesFiles"`
	LastEditRow                int                     `json:"lastEditRow"`
	PhrasesFiles               []string                `json:"phrasesFiles"`
	CurrentDir                 string                  `json:"currentDir"`
	BPM                        float32                 `json:"bpm"`
	PPQ                        int                     `json:"ppq"`
	PregainDB                  float32                 `json:"pregainDB"`
	PostgainDB                 float32                 `json:"postgainDB"`
	BiasDB                     float32                 `json:"biasDB"`
	SaturationDB               float32                 `json:"saturationDB"`
	DriveDB                    float32                 `json:"driveDB"`
	InputLevelDB               float32                 `json:"inputLevelDB"`
	ReverbSendPercent          float32                 `json:"reverbSendPercent"`
	TapePercent                float32                 `json:"tapePercent"`
	ShimmerPercent             float32                 `json:"shimmerPercent"`
	FileMetadata               map[string]FileMetadata `json:"fileMetadata"`
	LastChainRow               int                     `json:"lastChainRow"`
	LastPhraseRow              int                     `json:"lastPhraseRow"`
	LastPhraseCol              int                     `json:"lastPhraseCol"`
	RecordingEnabled           bool                    `json:"recordingEnabled"`
	RetriggerSettings          [255]RetriggerSettings  `json:"retriggerSettings"`
	TimestrechSettings         [255]TimestrechSettings `json:"timestrechSettings"`
	ModulateSettings           [255]ModulateSettings   `json:"modulateSettings"`           // Legacy field for backward compatibility
	InstrumentModulateSettings [255]ModulateSettings   `json:"instrumentModulateSettings"` // New separate pools
	SamplerModulateSettings    [255]ModulateSettings   `json:"samplerModulateSettings"`    // New separate pools
	DuckingSettings            [255]DuckingSettings    `json:"duckingSettings"`
	DuckingEditingIndex        int                     `json:"duckingEditingIndex"`
	ArpeggioSettings           [255]ArpeggioSettings   `json:"arpeggioSettings"`
	MidiSettings               [255]MidiSettings       `json:"midiSettings"`
	SoundMakerSettings         [255]SoundMakerSettings `json:"soundMakerSettings"`
	SongData                   [8][16]int              `json:"songData"`
	LastSongRow                int                     `json:"lastSongRow"`
	LastSongTrack              int                     `json:"lastSongTrack"`
	CurrentChain               int                     `json:"currentChain"`
	CurrentTrack               int                     `json:"currentTrack"`
	TrackSetLevels             [9]float32              `json:"trackSetLevels"`
	TrackTypes                 [9]bool                 `json:"trackTypes"`
	CurrentMixerTrack          int                     `json:"currentMixerTrack"`
	SOColumnMode               SOColumnMode            `json:"soColumnMode"`
	MidiCCNumbers              [9]int                  `json:"midiCCNumbers"`
}

const SaveFile = "tracker-save.json"
const WaveformHeight = 5

// ADSR mapping functions for Instrument view

// AttackToSeconds converts Attack hex value (00-FE) to seconds using exponential mapping
// 00 maps to 0.02s, FE maps to 30s
func AttackToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.02 // Default minimum value
	}
	// Exponential mapping: 00 -> 0.02s, FE -> 30s
	minSeconds := float32(0.02)
	maxSeconds := float32(30.0)
	ratio := float32(hexValue) / 254.0
	return minSeconds * float32(math.Pow(float64(maxSeconds/minSeconds), float64(ratio)))
}

// DecayToSeconds converts Decay hex value (00-FE) to seconds using linear mapping
// 00 maps to 0.0s, FE maps to 30.0s
func DecayToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.0 // Default minimum value
	}
	// Linear mapping: 00 -> 0.0s, FE -> 30.0s
	return (float32(hexValue) / 254.0) * 30.0
}

// SustainToLevel converts Sustain hex value (00-FE) to level using linear mapping
// 00 maps to 0.0, FE maps to 1.0
func SustainToLevel(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.0 // Default minimum value
	}
	// Linear mapping: 00 -> 0.0, FE -> 1.0
	return float32(hexValue) / 254.0
}

// ReleaseToSeconds converts Release hex value (00-FE) to seconds using exponential mapping
// 00 maps to 0.02s, FE maps to 30s (same as Attack)
func ReleaseToSeconds(hexValue int) float32 {
	if hexValue < 0 || hexValue > 254 {
		return 0.02 // Default minimum value
	}
	// Exponential mapping: 00 -> 0.02s, FE -> 30s
	minSeconds := float32(0.02)
	maxSeconds := float32(30.0)
	ratio := float32(hexValue) / 254.0
	return minSeconds * float32(math.Pow(float64(maxSeconds/minSeconds), float64(ratio)))
}

// VirtualDefaultConfig holds virtual default value for columns that display "--" but behave as a specific value
type VirtualDefaultConfig struct {
	DefaultValue int
}

// GetVirtualDefault returns the virtual default config for a column, or nil if no virtual default
func GetVirtualDefault(col PhraseColumn) *VirtualDefaultConfig {
	switch col {
	case ColPitch:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = center pitch
	case ColGate:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = default gate
	case ColPan:
		return &VirtualDefaultConfig{DefaultValue: 0x80} // 0x80 = center pan
	case ColLowPassFilter:
		return &VirtualDefaultConfig{DefaultValue: 0xFE} // 0xFE = 20kHz (no filtering)
	case ColVelocity:
		return &VirtualDefaultConfig{DefaultValue: 0x40} // 0x40 = 64 = default velocity
	default:
		return nil
	}
}

// Instrument Parameter Framework
type InstrumentParameterType int

const (
	ParameterTypeHex   InstrumentParameterType = iota // 0x00-0xFE hex values (default)
	ParameterTypeInt                                  // Integer values with custom range
	ParameterTypeFloat                                // Float values with custom range
)

// ParameterFormatter is a function type for custom parameter value formatting
type ParameterFormatter func(value float32) string

type InstrumentParameterDef struct {
	Key              string                  `json:"key"`           // Key for OSC sending (e.g. "preset", "cutoff")
	DisplayName      string                  `json:"displayName"`   // Name shown in UI (e.g. "Preset", "Cutoff")
	Type             InstrumentParameterType `json:"type"`          // Data type
	MinValue         float32                 `json:"minValue"`      // Minimum value
	MaxValue         float32                 `json:"maxValue"`      // Maximum value
	DefaultValue     float32                 `json:"defaultValue"`  // Default value (-1 for "--")
	Default          float32                 `json:"default"`       // Default value for new instances
	Column           int                     `json:"column"`        // Which column to display in (0 or 1)
	Order            int                     `json:"order"`         // Order within the column
	CoarseStep       float32                 `json:"coarseStep"`    // Step size for coarse control (0 = use default)
	FineStep         float32                 `json:"fineStep"`      // Step size for fine control (0 = use default)
	DisplayFormat    string                  `json:"displayFormat"` // Display format string (e.g. "%.1f Hz", empty = use default)
	DisplayFormatter ParameterFormatter      `json:"-"`             // Custom formatter function (not serialized to JSON)
}

type InstrumentDefinition struct {
	Name        string                   `json:"name"`        // Instrument name (e.g. "DX7", "PolyPerc")
	Description string                   `json:"description"` // Short description of the instrument
	Parameters  []InstrumentParameterDef `json:"parameters"`  // Parameter definitions
}

// Global registry of all instrument definitions
var InstrumentRegistry = map[string]InstrumentDefinition{
	"DX7": {
		Name:        "DX7",
		Description: "Classic FM synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "preset", DisplayName: "Preset", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 0,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"PolyPerc": {
		Name:        "PolyPerc",
		Description: "Polyphonic percussion synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "resonance", DisplayName: "Resonance", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 3.0, DefaultValue: 1.5, Column: 0, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 0,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"MiBraids": {
		Name:        "MiBraids",
		Description: "MiBraids is a digital macro oscillator that offers an atlas of waveform generation techniques.",
		Parameters: []InstrumentParameterDef{
			{
				Key: "timbre", DisplayName: "Timbre", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "color", DisplayName: "Color", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 1,
			},
			{
				Key: "model", DisplayName: "Model", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 47, DefaultValue: -1, Column: 0, Order: 2,
			},
			{
				Key: "resamp", DisplayName: "Resamp", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 2, DefaultValue: -1, Column: 1, Order: 0,
			},
			{
				Key: "decim", DisplayName: "Decim", Type: ParameterTypeInt,
				MinValue: 1, MaxValue: 32, DefaultValue: -1, Column: 1, Order: 1,
			},
			{
				Key: "bits", DisplayName: "Bits", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 6, DefaultValue: -1, Column: 1, Order: 2,
			},
			{
				Key: "ws", DisplayName: "WS", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 1, Order: 3,
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 4,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"MiPlaits": {
		Name:        "MiPlaits",
		Description: "A macro oscillator offering a multitude of synthesis methods.",
		Parameters: []InstrumentParameterDef{
			{
				Key: "engine", DisplayName: "Engine", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 15, DefaultValue: -1, Column: 0, Order: 0,
			},
			{
				Key: "harm", DisplayName: "Harm", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 1,
			},
			{
				Key: "timbre", DisplayName: "Timbre", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 2,
			},
			{
				Key: "morph", DisplayName: "Morph", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1000, DefaultValue: -1, Column: 0, Order: 3,
			},
			{
				Key: "fm_mod", DisplayName: "FM Mod", Type: ParameterTypeFloat,
				MinValue: -1.0, MaxValue: 1.0, DefaultValue: 0.0, Column: 1, Order: 0,
			},
			{
				Key: "timb_mod", DisplayName: "Timb Mod", Type: ParameterTypeFloat,
				MinValue: -1.0, MaxValue: 1.0, DefaultValue: 0.0, Column: 1, Order: 1,
			},
			{
				Key: "morph_mod", DisplayName: "Morph Mod", Type: ParameterTypeFloat,
				MinValue: -1.0, MaxValue: 1.0, DefaultValue: 0.0, Column: 1, Order: 2,
			},
			{
				Key: "decay", DisplayName: "Decay", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1.0, DefaultValue: 0.0, Column: 1, Order: 3,
			},
			{
				Key: "lpg_colour", DisplayName: "LPG Color", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 1.0, DefaultValue: 0.0, Column: 1, Order: 4,
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 5,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"SuperSaw": {
		Name:        "SuperSaw",
		Description: "stereo supersaw",
		Parameters: []InstrumentParameterDef{
			{
				Key: "vibrRate", DisplayName: "Vib Rate", Type: ParameterTypeFloat,
				MinValue: 0.1, MaxValue: 100.0, DefaultValue: 6.0, Default: 6.0, Column: 0, Order: 0,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f Hz",
			},
			{
				Key: "vibrDepth", DisplayName: "Vib Depth", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.3, Default: 0.3, Column: 0, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "drive", DisplayName: "Drive", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 10.0, DefaultValue: 1.5, Default: 1.5, Column: 0, Order: 2,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "detune", DisplayName: "Detune", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 4.0, DefaultValue: 0.2, Default: 0.2, Column: 0, Order: 3,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "spread", DisplayName: "Spread", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.6, Default: 0.6, Column: 1, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "lpenv", DisplayName: "LP Env", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 9.0, DefaultValue: 7.0, Default: 7.0, Column: 1, Order: 1,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "lpa", DisplayName: "LP Attack", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 10.0, DefaultValue: 1.0, Default: 1.0, Column: 1, Order: 2,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 3,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"TB303": {
		Name:        "TB303",
		Description: "Classic acid bassline synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "resonance", DisplayName: "Resonance", Type: ParameterTypeFloat,
				MinValue: 0.1, MaxValue: 3.0, DefaultValue: 1.0, Default: 1.0, Column: 0, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "glide", DisplayName: "Glide", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 0, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "envAmt", DisplayName: "Env Amount", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 10.0, DefaultValue: 1.0, Default: 1.0, Column: 0, Order: 2,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "envRelease", DisplayName: "Env Release", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 2.0, DefaultValue: 0.5, Default: 0.5, Column: 0, Order: 3,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "mixWave", DisplayName: "Pulse/Saw", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 1, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "drive", DisplayName: "Drive", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 10.0, DefaultValue: 1.0, Default: 1.0, Column: 1, Order: 1,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 2,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"Open303": {
		Name:        "Open303",
		Description: "Open-source 303 bassline synthesizer",
		Parameters: []InstrumentParameterDef{
			{
				Key: "waveform", DisplayName: "Waveform", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 0, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f sawtooth/square",
			},
			{
				Key: "tuning", DisplayName: "Tuning", Type: ParameterTypeFloat,
				MinValue: 110, MaxValue: 880, DefaultValue: 440, Default: 440, Column: 0, Order: 1,
				CoarseStep: 10, FineStep: 1, DisplayFormat: "%.0f Hz",
			},
			{
				Key: "resonance", DisplayName: "Resonance", Type: ParameterTypeFloat,
				MinValue: 1, MaxValue: 300, DefaultValue: 50, Default: 50, Column: 0, Order: 2,
				CoarseStep: 10, FineStep: 1, DisplayFormat: "%.0f%%",
			},
			{
				Key: "envMod", DisplayName: "Env Mod", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 400, DefaultValue: 50, Default: 50, Column: 0, Order: 3,
				CoarseStep: 10, FineStep: 1, DisplayFormat: "%.0f%%",
			},
			{
				Key: "normalDecay", DisplayName: "Decay", Type: ParameterTypeFloat,
				MinValue: 1, MaxValue: 3000, DefaultValue: 1000, Default: 1000, Column: 0, Order: 4,
				CoarseStep: 100, FineStep: 1, DisplayFormat: "%.0f ms",
			},
			{
				Key: "accent", DisplayName: "Accent", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 200, DefaultValue: 50, Default: 50, Column: 0, Order: 5,
				CoarseStep: 10, FineStep: 1, DisplayFormat: "%.0f%%",
			},
			{
				Key: "basePregain", DisplayName: "Pregain", Type: ParameterTypeFloat,
				MinValue: -60, MaxValue: 48, DefaultValue: 0, Default: 0, Column: 0, Order: 6,
				CoarseStep: 6, FineStep: 0.1, DisplayFormat: "%.1f dB",
			},
			{
				Key: "baseDrive", DisplayName: "Drive", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 100, DefaultValue: 0, Default: 0, Column: 0, Order: 7,
				CoarseStep: 5, FineStep: 1, DisplayFormat: "%.0f x",
			},
			{
				Key: "basePostgain", DisplayName: "Postgain", Type: ParameterTypeFloat,
				MinValue: -60, MaxValue: 24, DefaultValue: -12, Default: -12, Column: 0, Order: 8,
				CoarseStep: 1, FineStep: 0.1, DisplayFormat: "%.1f dB",
			},
			{
				Key: "ampDecay", DisplayName: "Amp Decay", Type: ParameterTypeFloat,
				MinValue: 30, MaxValue: 3000, DefaultValue: 1230, Default: 1230, Column: 1, Order: 0,
				CoarseStep: 100, FineStep: 10, DisplayFormat: "%.0f ms",
			},
			{
				Key: "ampRelease", DisplayName: "Amp Release", Type: ParameterTypeFloat,
				MinValue: 1, MaxValue: 2000, DefaultValue: 1, Default: 1, Column: 1, Order: 1,
				CoarseStep: 100, FineStep: 1, DisplayFormat: "%.0f ms",
			},
			{
				Key: "feedbackHPF", DisplayName: "Feedback HPF", Type: ParameterTypeFloat,
				MinValue: 20, MaxValue: 2000, DefaultValue: 150, Default: 150, Column: 1, Order: 2,
				CoarseStep: 100, FineStep: 10, DisplayFormat: "%.0f Hz",
			},
			{
				Key: "normalAttack", DisplayName: "Normal Attack", Type: ParameterTypeFloat,
				MinValue: 0.3, MaxValue: 30, DefaultValue: 3, Default: 3, Column: 1, Order: 3,
				CoarseStep: 1, FineStep: 0.1, DisplayFormat: "%.1f ms",
			},
			{
				Key: "accentAttack", DisplayName: "Accent Attack", Type: ParameterTypeFloat,
				MinValue: 0.3, MaxValue: 30, DefaultValue: 3, Default: 3, Column: 1, Order: 4,
				CoarseStep: 1, FineStep: 0.1, DisplayFormat: "%.1f ms",
			},
			{
				Key: "accentDecay", DisplayName: "Accent Decay", Type: ParameterTypeFloat,
				MinValue: 30, MaxValue: 3000, DefaultValue: 200, Default: 200, Column: 1, Order: 5,
				CoarseStep: 100, FineStep: 10, DisplayFormat: "%.0f ms",
			},
			{
				Key: "slideTime", DisplayName: "Slide Time", Type: ParameterTypeFloat,
				MinValue: 0, MaxValue: 300, DefaultValue: 60, Default: 60, Column: 1, Order: 6,
				CoarseStep: 10, FineStep: 1, DisplayFormat: "%.0f ms",
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 7,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"MollyThePoly": {
		Name:        "MollyThePoly",
		Description: "Classic polysynth with Juno-6 voice structure",
		Parameters: []InstrumentParameterDef{
			{
				Key: "oscWaveShape", DisplayName: "Osc Shape", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 2, DefaultValue: 1, Default: 1, Column: 0, Order: 0,
				DisplayFormatter: FormatOscWaveShape,
			},
			{
				Key: "pwMod", DisplayName: "PWM Amount", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.4, Default: 0.4, Column: 0, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "pwModSource", DisplayName: "PWM Source", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 2, DefaultValue: 0, Default: 0, Column: 0, Order: 2,
				DisplayFormatter: FormatPwmSource,
			},
			{
				Key: "mainOscLevel", DisplayName: "Main Osc", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 1.0, Default: 1.0, Column: 0, Order: 3,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "subOscLevel", DisplayName: "Sub Osc", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.2, Default: 0.2, Column: 0, Order: 4,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "subOscDetune", DisplayName: "Detune", Type: ParameterTypeFloat,
				MinValue: -5.0, MaxValue: 5.0, DefaultValue: 0.0, Default: 0.0, Column: 0, Order: 5,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f ST",
			},
			{
				Key: "noiseLevel", DisplayName: "Noise", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 0, Order: 6,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "lpFilterResonance", DisplayName: "Resonance", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.4, Default: 0.4, Column: 0, Order: 7,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "lpFilterCutoffModEnv", DisplayName: "Filter Env", Type: ParameterTypeFloat,
				MinValue: -1.0, MaxValue: 1.0, DefaultValue: 0.6, Default: 0.6, Column: 0, Order: 8,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "lpFilterCutoffModLfo", DisplayName: "Filter LFO", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "lpFilterTracking", DisplayName: "Key Track", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 2.0, DefaultValue: 1.0, Default: 1.0, Column: 1, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f:1",
			},
			{
				Key: "lfoFreq", DisplayName: "LFO Rate", Type: ParameterTypeFloat,
				MinValue: 0.1, MaxValue: 20.0, DefaultValue: 5.0, Default: 5.0, Column: 1, Order: 2,
				CoarseStep: 1.0, FineStep: 0.1, DisplayFormat: "%.1f Hz",
			},
			{
				Key: "lfoWaveShape", DisplayName: "LFO Shape", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 4, DefaultValue: 0, Default: 0, Column: 1, Order: 3,
				DisplayFormatter: FormatLfoWaveShape,
			},
			{
				Key: "ampMod", DisplayName: "Amp Mod", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 4,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "timbre", DisplayName: "Timbre", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 5,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "ringModMix", DisplayName: "Ring Mod", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 6,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "chorusMix", DisplayName: "Chorus", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 1.0, Default: 1.0, Column: 1, Order: 7,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "monophonic", DisplayName: "Monophonic", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 8,
				DisplayFormatter: FormatYesNo,
			},
		},
	},
	"Juno60": {
		Name:        "Juno60",
		Description: "Roland Juno-60 analog polysynth with chorus",
		Parameters: []InstrumentParameterDef{
			{
				Key: "juno60OscSaw", DisplayName: "Sawtooth", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.8, Default: 0.8, Column: 0, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60OscPulse", DisplayName: "Pulse", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 0, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60OscSub", DisplayName: "Sub Osc", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.3, Default: 0.3, Column: 0, Order: 2,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60OscNoise", DisplayName: "Noise", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 0, Order: 3,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60OscPwm", DisplayName: "PWM", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 0, Order: 4,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60Resonance", DisplayName: "Resonance", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.4, Default: 0.4, Column: 0, Order: 5,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60EnvMod", DisplayName: "Env Mod", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.6, Default: 0.6, Column: 0, Order: 6,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60VcfKey", DisplayName: "VCF Key", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 0, Order: 7,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60VcfDir", DisplayName: "VCF Dir", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 1, Default: 1, Column: 0, Order: 8,
				DisplayFormatter: FormatYesNo,
			},
			{
				Key: "juno60LfoRate", DisplayName: "LFO Rate", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 1, Order: 0,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60LfoDelay", DisplayName: "LFO Delay", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 1,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60DcoLfo", DisplayName: "DCO LFO", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 2,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60VcfLfo", DisplayName: "VCF LFO", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.0, Default: 0.0, Column: 1, Order: 3,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
			{
				Key: "juno60DcoRange", DisplayName: "DCO Range", Type: ParameterTypeFloat,
				MinValue: 0.5, MaxValue: 2.0, DefaultValue: 1.0, Default: 1.0, Column: 1, Order: 4,
				CoarseStep: 0.5, FineStep: 0.1, DisplayFormat: "%.1f",
			},
			{
				Key: "juno60LfoAuto", DisplayName: "LFO Auto", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 1, DefaultValue: 0, Default: 0, Column: 1, Order: 5,
				DisplayFormatter: FormatYesNo,
			},
			{
				Key: "juno60JunoChorus", DisplayName: "Chorus", Type: ParameterTypeInt,
				MinValue: 0, MaxValue: 6, DefaultValue: 3, Default: 3, Column: 1, Order: 6,
				DisplayFormatter: FormatJunoChorus,
			},
			{
				Key: "juno60ChorusDryWet", DisplayName: "Chorus Mix", Type: ParameterTypeFloat,
				MinValue: 0.0, MaxValue: 1.0, DefaultValue: 0.5, Default: 0.5, Column: 1, Order: 7,
				CoarseStep: 0.1, FineStep: 0.01, DisplayFormat: "%.2f",
			},
		},
	},
}

// Helper functions for the instrument framework

// FormatYesNo formats a 0/1 value as "No"/"Yes"
func FormatYesNo(value float32) string {
	if value == 0 {
		return "No"
	}
	return "Yes"
}

// FormatOscWaveShape formats oscillator waveform selection
func FormatOscWaveShape(value float32) string {
	switch int(value) {
	case 0:
		return "VarSaw"
	case 1:
		return "Saw"
	case 2:
		return "Pulse"
	default:
		return "Unknown"
	}
}

// FormatLfoWaveShape formats LFO waveform selection
func FormatLfoWaveShape(value float32) string {
	switch int(value) {
	case 0:
		return "Sine"
	case 1:
		return "Triangle"
	case 2:
		return "Saw"
	case 3:
		return "Pulse"
	case 4:
		return "Random"
	default:
		return "Unknown"
	}
}

// FormatPwmSource formats PWM modulation source selection
func FormatPwmSource(value float32) string {
	switch int(value) {
	case 0:
		return "LFO"
	case 1:
		return "Env"
	case 2:
		return "Manual"
	default:
		return "Unknown"
	}
}

// FormatJunoChorus formats Juno-60 chorus mode selection
func FormatJunoChorus(value float32) string {
	switch int(value) {
	case 0:
		return "Off"
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "I+II"
	case 4:
		return "I(junologue)"
	case 5:
		return "I-II(junologue)"
	case 6:
		return "II(junologue)"
	default:
		return "Unknown"
	}
}

// GetInstrumentDefinition returns the definition for a given instrument name
func GetInstrumentDefinition(name string) (InstrumentDefinition, bool) {
	def, exists := InstrumentRegistry[name]
	return def, exists
}

// GetAvailableSoundMakers returns a list of all available SoundMaker names from the InstrumentRegistry
// The list is sorted alphabetically for consistent ordering
func GetAvailableSoundMakers() []string {
	soundMakers := make([]string, 0, len(InstrumentRegistry))
	for name := range InstrumentRegistry {
		soundMakers = append(soundMakers, name)
	}

	// Sort alphabetically for consistent ordering
	for i := 0; i < len(soundMakers)-1; i++ {
		for j := i + 1; j < len(soundMakers); j++ {
			if soundMakers[i] > soundMakers[j] {
				soundMakers[i], soundMakers[j] = soundMakers[j], soundMakers[i]
			}
		}
	}

	return soundMakers
}

// GetAvailableSoundMakersWithNone returns a list of all available SoundMaker names with "None" as the first option
func GetAvailableSoundMakersWithNone() []string {
	soundMakers := []string{"None"}
	soundMakers = append(soundMakers, GetAvailableSoundMakers()...)
	return soundMakers
}

// GetInstrumentParameterByKey returns a specific parameter definition by key
func (def InstrumentDefinition) GetParameterByKey(key string) (InstrumentParameterDef, bool) {
	for _, param := range def.Parameters {
		if param.Key == key {
			return param, true
		}
	}
	return InstrumentParameterDef{}, false
}

// GetParametersSortedByColumn returns parameters sorted by column and order
func (def InstrumentDefinition) GetParametersSortedByColumn() (col0 []InstrumentParameterDef, col1 []InstrumentParameterDef) {
	for _, param := range def.Parameters {
		if param.Column == 0 {
			col0 = append(col0, param)
		} else {
			col1 = append(col1, param)
		}
	}

	// Sort by order within each column
	for i := 0; i < len(col0)-1; i++ {
		for j := i + 1; j < len(col0); j++ {
			if col0[i].Order > col0[j].Order {
				col0[i], col0[j] = col0[j], col0[i]
			}
		}
	}

	for i := 0; i < len(col1)-1; i++ {
		for j := i + 1; j < len(col1); j++ {
			if col1[i].Order > col1[j].Order {
				col1[i], col1[j] = col1[j], col1[i]
			}
		}
	}

	return col0, col1
}

// MiBraids model names for display
var MiBraidsModelNames = []string{
	"CSAW", "MORPH", "SAW_SQUARE", "SINE_TRIANGLE", "BUZZ", "SQUARE_SUB", "SAW_SUB", "SQUARE_SYNC",
	"SAW_SYNC", "TRIPLE_SAW", "TRIPLE_SQUARE", "TRIPLE_TRIANGLE", "TRIPLE_SINE", "TRIPLE_RING_MOD",
	"SAW_SWARM", "SAW_COMB", "TOY", "DIGITAL_FILTER_LP", "DIGITAL_FILTER_PK", "DIGITAL_FILTER_BP",
	"DIGITAL_FILTER_HP", "VOSIM", "VOWEL", "VOWEL_FOF", "HARMONICS", "FM", "FEEDBACK_FM",
	"CHAOTIC_FEEDBACK_FM", "PLUCKED", "BOWED", "BLOWN", "FLUTED", "STRUCK_BELL", "STRUCK_DRUM",
	"KICK", "CYMBAL", "SNARE", "WAVETABLES", "WAVE_MAP", "WAVE_LINE", "WAVE_PARAPHONIC",
	"FILTERED_NOISE", "TWIN_PEAKS_NOISE", "CLOCKED_NOISE", "GRANULAR_CLOUD", "PARTICLE_NOISE",
	"DIGITAL_MODULATION", "QUESTION_MARK",
}

// GetMiBraidsModelName returns the name for a given model index
func GetMiBraidsModelName(index int) string {
	if index >= 0 && index < len(MiBraidsModelNames) {
		return MiBraidsModelNames[index]
	}
	return "UNKNOWN"
}

// MiPlaits engine names for display
var MiPlaitsEngineNames = []string{
	"virtual_analog_engine", "waveshaping_engine", "fm_engine", "grain_engine",
	"additive_engine", "wavetable_engine", "chord_engine", "speech_engine",
	"swarm_engine", "noise_engine", "particle_engine", "string_engine",
	"modal_engine", "bass_drum_engine", "snare_drum_engine", "hi_hat_engine",
}

// GetMiPlaitsEngineName returns the name for a given engine index
func GetMiPlaitsEngineName(index int) string {
	if index >= 0 && index < len(MiPlaitsEngineNames) {
		return MiPlaitsEngineNames[index]
	}
	return "UNKNOWN"
}

// Helper functions for SoundMakerSettings with the new parameter framework

// GetParameterValue gets a parameter value with fallback to default
func (settings *SoundMakerSettings) GetParameterValue(key string) float32 {
	if settings.Parameters == nil {
		return -1
	}

	if value, exists := settings.Parameters[key]; exists {
		return value
	}

	// Return default value from instrument definition if available
	if def, exists := GetInstrumentDefinition(settings.Name); exists {
		if param, found := def.GetParameterByKey(key); found {
			// Use Default field if set
			if param.Default != 0 {
				return param.Default
			}
			// Use DefaultValue if set (not -1)
			if param.DefaultValue != -1 {
				return param.DefaultValue
			}
			// Otherwise use the minimum value in the acceptable range
			return param.MinValue
		}
	}

	return -1
}

// SetParameterValue sets a parameter value
func (settings *SoundMakerSettings) SetParameterValue(key string, value float32) {
	if settings.Parameters == nil {
		settings.Parameters = make(map[string]float32)
	}
	settings.Parameters[key] = value
}

// InitializeParameters ensures all parameters exist with default values
func (settings *SoundMakerSettings) InitializeParameters() {
	if def, exists := GetInstrumentDefinition(settings.Name); exists {
		if settings.Parameters == nil {
			settings.Parameters = make(map[string]float32)
		}

		for _, param := range def.Parameters {
			if _, exists := settings.Parameters[param.Key]; !exists {
				// Use Default field if set
				if param.Default != 0 {
					settings.Parameters[param.Key] = param.Default
				} else if param.DefaultValue != -1 {
					// Use DefaultValue if set (not -1)
					settings.Parameters[param.Key] = param.DefaultValue
				} else {
					// Otherwise use the minimum value in the acceptable range
					settings.Parameters[param.Key] = param.MinValue
				}
			}
		}
	}
}
