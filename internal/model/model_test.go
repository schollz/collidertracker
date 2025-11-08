package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/schollz/collidertracker/internal/types"
)

func TestNewModel(t *testing.T) {
	m := NewModel(57120, "test-save.json", false)

	// Test default initialization
	assert.Equal(t, 0, m.CurrentRow)
	assert.Equal(t, 0, m.CurrentCol) // Starts at first track in song view
	assert.Equal(t, 0, m.ScrollOffset)
	assert.Equal(t, types.SongView, m.ViewMode) // Starts with song view
	assert.False(t, m.IsPlaying)
	assert.Equal(t, float32(120), m.BPM)
	assert.Equal(t, 2, m.PPQ) // Default PPQ is 2
	assert.Equal(t, float32(0.0), m.PregainDB)
	assert.Equal(t, float32(0.0), m.PostgainDB)
	assert.Equal(t, float32(-6.0), m.BiasDB)
	assert.Equal(t, float32(-6.0), m.SaturationDB)
	assert.Equal(t, float32(-6.0), m.DriveDB)

	// Test data structure initialization
	assert.Len(t, m.PhrasesData, 255)
	assert.Len(t, m.InstrumentPhrasesData, 255)
	assert.Len(t, m.SamplerPhrasesData, 255)
	assert.Len(t, m.RetriggerSettings, 255)
	assert.Len(t, m.TimestrechSettings, 255)
	assert.Len(t, m.ArpeggioSettings, 255)
	assert.Len(t, m.MidiSettings, 255)
	assert.Len(t, m.SoundMakerSettings, 255)

	// Test song data structure (8 tracks × 16 rows)
	assert.Len(t, m.SongData, 8)
	for i := 0; i < 8; i++ {
		assert.Len(t, m.SongData[i], 16)
		// All should be initialized to -1 (empty)
		for j := 0; j < 16; j++ {
			assert.Equal(t, -1, m.SongData[i][j])
		}
	}

	// Test effect column initialization in InstrumentPhrasesData
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColEffectReverb])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColEffectComb])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColPan])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColLowPassFilter])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColHighPassFilter])
}

func TestModelDataStructures(t *testing.T) {
	m := NewModel(0, "", false)

	// Test accessing existing methods
	data := m.GetCurrentPhrasesData()
	assert.NotNil(t, data)

	chainsData := m.GetCurrentChainsData()
	assert.NotNil(t, chainsData)

	// Test phrase view type determination
	m.CurrentTrack = 2 // Instrument track
	phraseType := m.GetPhraseViewType()
	assert.Equal(t, types.SamplerPhraseView, phraseType) // Track type defaults to sampler

	m.CurrentTrack = 5 // Sampler track
	phraseType = m.GetPhraseViewType()
	assert.Equal(t, types.SamplerPhraseView, phraseType)
}

func TestModelDataManipulation(t *testing.T) {
	m := NewModel(0, "", false)

	// Test SetChainsData method
	m.SetChainsData(0, 0, 10)
	assert.Equal(t, 10, m.ChainsData[0][0])

	// Test SetPhrasesData method
	m.SetPhrasesData(5, 10, int(types.ColNote), 0x60)
	assert.Equal(t, 0x60, m.PhrasesData[5][10][types.ColNote])

	// Test bounds checking - should not panic
	m.SetChainsData(-1, 0, 10)
	m.SetChainsData(300, 0, 10)
	m.SetPhrasesData(-1, 0, 0, 10)
	m.SetPhrasesData(300, 0, 0, 10)
}

func TestModelWaveformOperations(t *testing.T) {
	m := NewModel(0, "", false)

	// Test PushWaveformSample method
	maxCols := 50
	m.PushWaveformSample(0.5, maxCols)
	m.PushWaveformSample(-0.3, maxCols)
	m.PushWaveformSample(0.8, maxCols)

	assert.Len(t, m.WaveformBuf, 3)
	assert.Equal(t, 0.5, m.WaveformBuf[0])
	assert.Equal(t, -0.3, m.WaveformBuf[1])
	assert.Equal(t, 0.8, m.WaveformBuf[2])

	// Test LastWaveform tracking
	m.LastWaveform = 0.25
	assert.Equal(t, 0.25, m.LastWaveform)
}

func TestModelOSCIntegration(t *testing.T) {
	m := NewModel(57120, "", false)

	// Test OSC message methods (they should not panic)
	m.SendOSCPregainMessage()
	m.SendOSCPostgainMessage()
	m.SendOSCBiasMessage()
	m.SendOSCSaturationMessage()
	m.SendOSCDriveMessage()

	// Test track level OSC messages
	for track := 0; track < 8; track++ {
		m.SendOSCTrackSetLevelMessage(track)
	}

	// Test playback OSC messages
	m.SendOSCPlaybackMessage("test.wav", true)
	m.SendOSCPlaybackMessage("test.wav", false)

	// Test stop OSC
	m.SendStopOSC()
}

func TestModelFileOperations(t *testing.T) {
	m := NewModel(0, "", false)

	// Test file appending for sampler tracks
	m.CurrentTrack = 5 // Sampler track
	fileIndex := m.AppendPhrasesFile("test.wav")
	assert.Equal(t, 0, fileIndex) // First file
	assert.Len(t, m.SamplerPhrasesFiles, 1)
	assert.Equal(t, "test.wav", m.SamplerPhrasesFiles[0])

	// Test adding another file
	fileIndex = m.AppendPhrasesFile("test2.wav")
	assert.Equal(t, 1, fileIndex) // Second file
	assert.Len(t, m.SamplerPhrasesFiles, 2)
}

func TestModelColumnMapping(t *testing.T) {
	m := NewModel(0, "", false)

	// Test column mapping for instrument view
	m.CurrentTrack = 1 // Instrument track
	mapping := m.GetColumnMapping(0)
	assert.NotNil(t, mapping)

	// Test column mapping for sampler view
	m.CurrentTrack = 6 // Sampler track
	mapping = m.GetColumnMapping(0)
	assert.NotNil(t, mapping)
}

func TestModelRecordingFilename(t *testing.T) {
	m := NewModel(0, "", false)

	filename := m.GenerateRecordingFilename()
	assert.NotEmpty(t, filename)
	assert.Contains(t, filename, ".wav")
}

func TestModelPlaybackState(t *testing.T) {
	m := NewModel(0, "", false)

	// Test initial playback state
	assert.False(t, m.IsPlaying)
	assert.Equal(t, 0, m.PlaybackRow)
	assert.Equal(t, 0, m.PlaybackChain)
	assert.Equal(t, 0, m.PlaybackChainRow)
	assert.Equal(t, 0, m.PlaybackPhrase)

	// Test basic playback state changes
	m.IsPlaying = true
	assert.True(t, m.IsPlaying)

	m.PlaybackRow = 5
	assert.Equal(t, 5, m.PlaybackRow)
}

func TestModelDataInitialization(t *testing.T) {
	m := NewModel(0, "", false)

	// Test that data is properly initialized with defaults
	// Check phrase data defaults for sampler - virtual default columns now start as -1
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColNote])
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColPitch]) // Virtual default: displays "--", behaves as 80
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColDeltaTime])
	assert.Equal(t, -1, m.SamplerPhrasesData[0][0][types.ColGate]) // Virtual default: displays "--", behaves as 80

	// Check phrase data defaults for instrument
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColNote])
	assert.Equal(t, -1, m.InstrumentPhrasesData[0][0][types.ColDeltaTime])
	assert.Equal(t, int(types.ChordNone), m.InstrumentPhrasesData[0][0][types.ColChord])
	assert.Equal(t, int(types.ChordAddNone), m.InstrumentPhrasesData[0][0][types.ColChordAddition])

	// Check chain data initialization
	assert.Equal(t, -1, m.InstrumentChainsData[0][0])
	assert.Equal(t, -1, m.SamplerChainsData[0][0])
}

func BenchmarkModelInitialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModel(57120, "bench-test.json", false)
	}
}

func BenchmarkWaveformBufferPush(b *testing.B) {
	m := NewModel(0, "", false)
	maxCols := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.PushWaveformSample(0.5, maxCols)
	}
}

func TestProcessArpeggio(t *testing.T) {
	// Create a model with test data
	model := NewModel(0, "", false)

	tests := []struct {
		name                string
		arpeggioIndex       int
		arpeggioSettings    types.ArpeggioSettings
		params              InstrumentOSCParams
		expectedNotes       []float32
		expectedDivisors    []float32
		expectedEmptyResult bool
	}{
		{
			name:          "No arpeggio (invalid index)",
			arpeggioIndex: -1,
			params: InstrumentOSCParams{
				ArpeggioIndex: -1,
				Notes:         []float32{60}, // C4
			},
			expectedEmptyResult: true,
		},
		{
			name:          "Single note - up 2",
			arpeggioIndex: 1,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 2, Divisor: 4}, // up, count 2, divisor 4
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 1,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordNone),
			},
			// Up 2: 72, 84 (no removal)
			expectedNotes:    []float32{72, 84},
			expectedDivisors: []float32{4, 4},
		},
		{
			name:          "Single note - down 2",
			arpeggioIndex: 2,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 2, Count: 2, Divisor: 2}, // down, count 2, divisor 2
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 2,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordNone),
			},
			// Down 2: 48, 36 (no removal)
			expectedNotes:    []float32{48, 36},
			expectedDivisors: []float32{2, 2},
		},
		{
			name:          "Major chord - up 3",
			arpeggioIndex: 3,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 3, Divisor: 1}, // up, count 3, divisor 1
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 3,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordMajor),
			},
			// C Major = [60, 64, 67]. Up 3: 64, 67, 72 (no removal)
			expectedNotes:    []float32{64, 67, 72},
			expectedDivisors: []float32{1, 1, 1},
		},
		{
			name:          "Major chord - up 6 (cross octave)",
			arpeggioIndex: 4,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 6, Divisor: 2}, // up, count 6, divisor 2
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 4,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordMajor),
			},
			// C Major = [60, 64, 67]. Up 6: 64, 67, 72(C5), 76(E5), 79(G5), 84(C6) (no removal)
			expectedNotes:    []float32{64, 67, 72, 76, 79, 84},
			expectedDivisors: []float32{2, 2, 2, 2, 2, 2},
		},
		{
			name:          "Major chord - down 3",
			arpeggioIndex: 5,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 2, Count: 3, Divisor: 3}, // down, count 3, divisor 3
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 5,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordMajor),
			},
			// C Major = [60, 64, 67]. Down 3: 67(prev octave), 64(prev octave), 60(prev octave)
			// Which is: 55, 52, 48 (no removal)
			expectedNotes:    []float32{55, 52, 48},
			expectedDivisors: []float32{3, 3, 3},
		},
		{
			name:          "Multiple rows - up then down",
			arpeggioIndex: 6,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 2, Divisor: 1}, // up 2
					{Direction: 2, Count: 2, Divisor: 2}, // down 2
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 6,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordMajor),
			},
			// C Major = [60, 64, 67]. Up 2: 64, 67. Then down 2 from 67: 64, 60 (no removal)
			expectedNotes:    []float32{64, 67, 64, 60},
			expectedDivisors: []float32{1, 1, 2, 2},
		},
		{
			name:          "Skip empty rows",
			arpeggioIndex: 7,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 0, Count: -1, Divisor: -1}, // empty row
					{Direction: 1, Count: 1, Divisor: 4},   // up 1
					{Direction: 0, Count: 2, Divisor: -1},  // invalid row
				},
			},
			params: InstrumentOSCParams{
				ArpeggioIndex: 7,
				Notes:         []float32{60}, // C4
				ChordType:     int(types.ChordNone),
			},
			// Only middle row executes: single note up 1 = 72 (no removal)
			expectedNotes:    []float32{72},
			expectedDivisors: []float32{4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the arpeggio settings
			if tt.arpeggioIndex >= 0 {
				model.ArpeggioSettings[tt.arpeggioIndex] = tt.arpeggioSettings
			}

			// Update params to have the full chord notes, just like in real usage
			// The test params currently have single notes but expect chord behavior
			// We need to populate Notes with the actual chord notes
			testParams := tt.params
			if len(testParams.Notes) == 1 && testParams.ChordType != int(types.ChordNone) {
				// Generate the chord notes that would be set by helpers.go
				baseNote := int(testParams.Notes[0])
				chordNotes := types.GetChordNotes(baseNote,
					types.ChordType(testParams.ChordType),
					types.ChordAddition(testParams.ChordAddition),
					types.ChordTransposition(testParams.ChordTransposition))

				testParams.Notes = make([]float32, len(chordNotes))
				for i, note := range chordNotes {
					testParams.Notes[i] = float32(note)
				}
			}

			// Call the function
			notes, divisors := model.ProcessArpeggio(testParams)

			// Check results
			if tt.expectedEmptyResult {
				if len(notes) != 0 || len(divisors) != 0 {
					t.Errorf("Expected empty result, got notes: %v, divisors: %v", notes, divisors)
				}
				return
			}

			if len(notes) != len(tt.expectedNotes) {
				t.Errorf("Expected %d notes, got %d: %v", len(tt.expectedNotes), len(notes), notes)
				return
			}

			if len(divisors) != len(tt.expectedDivisors) {
				t.Errorf("Expected %d divisors, got %d: %v", len(tt.expectedDivisors), len(divisors), divisors)
				return
			}

			for i, expectedNote := range tt.expectedNotes {
				if notes[i] != expectedNote {
					t.Errorf("Note %d: expected %f, got %f", i, expectedNote, notes[i])
				}
			}

			for i, expectedDivisor := range tt.expectedDivisors {
				if divisors[i] != expectedDivisor {
					t.Errorf("Divisor %d: expected %f, got %f", i, expectedDivisor, divisors[i])
				}
			}
		})
	}
}

func TestProcessArpeggioWithChordAdditions(t *testing.T) {
	model := NewModel(0, "", false)

	tests := []struct {
		name             string
		arpeggioIndex    int
		arpeggioSettings types.ArpeggioSettings
		baseNote         float32
		chordType        types.ChordType
		chordAddition    types.ChordAddition
		expectedNotes    []float32
	}{
		{
			name:          "Major 7 chord - up 4",
			arpeggioIndex: 10,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 4, Divisor: 1},
				},
			},
			baseNote:      60, // C4
			chordType:     types.ChordMajor,
			chordAddition: types.ChordAdd7,
			// C Major 7 = [60, 64, 67, 71]. Up 4: 64, 67, 71, 72 (no removal)
			expectedNotes: []float32{64, 67, 71, 72},
		},
		{
			name:          "Minor chord - up 3",
			arpeggioIndex: 11,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 3, Divisor: 1},
				},
			},
			baseNote:      60, // C4
			chordType:     types.ChordMinor,
			chordAddition: types.ChordAddNone,
			// C Minor = [60, 63, 67]. Up 3: 63, 67, 72 (no removal)
			expectedNotes: []float32{63, 67, 72},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.ArpeggioSettings[tt.arpeggioIndex] = tt.arpeggioSettings

			// Generate the full chord notes that would be set by helpers.go
			chordNotes := types.GetChordNotes(int(tt.baseNote), tt.chordType, tt.chordAddition, types.ChordTransNone)

			params := InstrumentOSCParams{
				ArpeggioIndex:      tt.arpeggioIndex,
				Notes:              make([]float32, len(chordNotes)),
				ChordType:          int(tt.chordType),
				ChordAddition:      int(tt.chordAddition),
				ChordTransposition: int(types.ChordTransNone),
			}

			// Populate with the actual chord notes
			for i, note := range chordNotes {
				params.Notes[i] = float32(note)
			}

			notes, _ := model.ProcessArpeggio(params)

			if len(notes) != len(tt.expectedNotes) {
				t.Errorf("Expected %d notes, got %d: %v", len(tt.expectedNotes), len(notes), notes)
				return
			}

			for i, expectedNote := range tt.expectedNotes {
				if notes[i] != expectedNote {
					t.Errorf("Note %d: expected %f, got %f", i, expectedNote, notes[i])
				}
			}
		})
	}
}

func TestGetNextChordNote(t *testing.T) {
	model := NewModel(0, "", false)

	tests := []struct {
		name        string
		currentNote float32
		baseChord   []float32
		isUp        bool
		expected    float32
	}{
		{
			name:        "Major chord - up from root",
			currentNote: 60,                    // C4
			baseChord:   []float32{60, 64, 67}, // C Major
			isUp:        true,
			expected:    64, // E4
		},
		{
			name:        "Major chord - up from 3rd (wrap to next octave)",
			currentNote: 67,                    // G4
			baseChord:   []float32{60, 64, 67}, // C Major
			isUp:        true,
			expected:    72, // C5
		},
		{
			name:        "Major chord - down from root (wrap to previous octave)",
			currentNote: 60,                    // C4
			baseChord:   []float32{60, 64, 67}, // C Major
			isUp:        false,
			expected:    55, // G3
		},
		{
			name:        "Major chord - down from 3rd",
			currentNote: 67,                    // G4
			baseChord:   []float32{60, 64, 67}, // C Major
			isUp:        false,
			expected:    64, // E4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.getNextChordNote(tt.currentNote, tt.baseChord, tt.isUp)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGetNextChordNoteWithRotatedChord(t *testing.T) {
	model := NewModel(0, "", false)

	tests := []struct {
		name        string
		currentNote float32
		baseChord   []float32
		isUp        bool
		expected    float32
	}{
		{
			name:        "Rotated chord (t=2) - up from G (should go to C+octave)",
			currentNote: 67,                    // G4 (first note in rotated chord)
			baseChord:   []float32{67, 72, 76}, // G, C+oct, E+oct (C major t=2)
			isUp:        true,
			expected:    72, // C5
		},
		{
			name:        "Rotated chord (t=2) - up from C+octave (should go to E+octave)",
			currentNote: 72,                    // C5 (second note in rotated chord)
			baseChord:   []float32{67, 72, 76}, // G, C+oct, E+oct (C major t=2)
			isUp:        true,
			expected:    76, // E5
		},
		{
			name:        "Rotated chord (t=2) - up from E+octave (should wrap to G+octave)",
			currentNote: 76,                    // E5 (third note in rotated chord)
			baseChord:   []float32{67, 72, 76}, // G, C+oct, E+oct (C major t=2)
			isUp:        true,
			expected:    79, // G5 (67+12)
		},
		{
			name:        "Rotated chord (t=2) - down from G (should wrap to E+octave-octave)",
			currentNote: 67,                    // G4 (first note in rotated chord)
			baseChord:   []float32{67, 72, 76}, // G, C+oct, E+oct (C major t=2)
			isUp:        false,
			expected:    64, // E4 (76-12)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.getNextChordNote(tt.currentNote, tt.baseChord, tt.isUp)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestProcessArpeggioWithRotatedChord(t *testing.T) {
	model := NewModel(0, "", false)

	tests := []struct {
		name             string
		arpeggioIndex    int
		arpeggioSettings types.ArpeggioSettings
		baseNote         int
		chordType        types.ChordType
		transpose        types.ChordTransposition
		expectedNotes    []float32
	}{
		{
			name:          "C Major chord with t=2 - up 3 (should go G->C+oct->E+oct)",
			arpeggioIndex: 20,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 3, Divisor: 1}, // up 3
				},
			},
			baseNote:  60, // C4
			chordType: types.ChordMajor,
			transpose: types.ChordTrans2, // t=2 gives us [G, C+oct, E+oct] = [67, 72, 76]
			// Starting from G(67), up 3 should give us: C+oct(72), E+oct(76), G+oct(79)
			expectedNotes: []float32{72, 76, 79},
		},
		{
			name:          "C Minor chord with t=1 - up 4",
			arpeggioIndex: 21,
			arpeggioSettings: types.ArpeggioSettings{
				Rows: [16]types.ArpeggioRow{
					{Direction: 1, Count: 4, Divisor: 1}, // up 4
				},
			},
			baseNote:  60, // C4
			chordType: types.ChordMinor,
			transpose: types.ChordTrans1, // t=1 gives us [Eb, G, C+oct] = [63, 67, 72]
			// Starting from Eb(63), up 4 should give us: G(67), C+oct(72), Eb+oct(75), G+oct(79)
			expectedNotes: []float32{67, 72, 75, 79},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.ArpeggioSettings[tt.arpeggioIndex] = tt.arpeggioSettings

			// Generate the rotated chord notes
			chordNotes := types.GetChordNotes(tt.baseNote, tt.chordType, types.ChordAddNone, tt.transpose)

			params := InstrumentOSCParams{
				ArpeggioIndex:      tt.arpeggioIndex,
				Notes:              make([]float32, len(chordNotes)),
				ChordType:          int(tt.chordType),
				ChordTransposition: int(tt.transpose),
			}

			// Convert to float32
			for i, note := range chordNotes {
				params.Notes[i] = float32(note)
			}

			notes, _ := model.ProcessArpeggio(params)

			if len(notes) != len(tt.expectedNotes) {
				t.Errorf("Expected %d notes, got %d: %v (from chord %v)",
					len(tt.expectedNotes), len(notes), notes, params.Notes)
				return
			}

			for i, expectedNote := range tt.expectedNotes {
				if notes[i] != expectedNote {
					t.Errorf("Note %d: expected %f, got %f (from chord %v)",
						i, expectedNote, notes[i], params.Notes)
				}
			}
		})
	}
}

func TestSendOSCInstrumentMessageWithArpeggioInitialNote(t *testing.T) {
	// This test verifies that when an arpeggio is active, only the root note is sent initially
	// We can't easily test the actual OSC sending without mocking, but we can test the logic
	model := NewModel(0, "", false)

	// Set up arpeggio settings for index 1
	model.ArpeggioSettings[1] = types.ArpeggioSettings{
		Rows: [16]types.ArpeggioRow{
			{Direction: 1, Count: 2, Divisor: 1}, // up 2
		},
	}

	// Test case 1: Chord with arpeggio should only send root note initially
	chordParams := InstrumentOSCParams{
		TrackId:       0,
		ArpeggioIndex: 1,                     // Valid arpeggio
		Notes:         []float32{60, 64, 67}, // C Major chord
		ChordType:     int(types.ChordMajor),
	}

	// This should process the arpeggio and send only the root note (60) initially
	// The actual testing of OSC message content would require mocking the OSC client
	// For now, we just verify that the function doesn't panic and processes correctly
	model.SendOSCInstrumentMessageWithArpeggio(chordParams)

	// Test case 2: No arpeggio should send full chord
	noArpeggioParams := InstrumentOSCParams{
		TrackId:       0,
		ArpeggioIndex: -1,                    // No arpeggio
		Notes:         []float32{60, 64, 67}, // C Major chord
		ChordType:     int(types.ChordMajor),
	}

	// This should send the full chord since no arpeggio is active
	model.SendOSCInstrumentMessageWithArpeggio(noArpeggioParams)
}

func TestGenerateEqualSlices(t *testing.T) {
	m := NewModel(0, "", false)

	// Test with a real audio file
	testFile := "../getbpm/Break120.wav"
	
	// Set up file metadata
	m.FileMetadata[testFile] = types.FileMetadata{
		BPM:         120.0,
		Slices:      16,
		SliceType:   0, // Equal mode
		Playthrough: 0,
		SyncToBPM:   1,
	}

	// Generate equal slices
	m.GenerateEqualSlices(testFile)

	// Verify that slices were generated
	metadata := m.FileMetadata[testFile]
	assert.Len(t, metadata.Onsets, 16, "Should generate 16 equal slices")

	// Verify that slices are evenly spaced
	// First slice should be at 0.0
	assert.InDelta(t, 0.0, metadata.Onsets[0], 0.001, "First slice should start at 0.0")

	// Check that slices are evenly spaced
	if len(metadata.Onsets) > 1 {
		sliceDuration := metadata.Onsets[1] - metadata.Onsets[0]
		for i := 1; i < len(metadata.Onsets)-1; i++ {
			actualDuration := metadata.Onsets[i+1] - metadata.Onsets[i]
			assert.InDelta(t, sliceDuration, actualDuration, 0.001, 
				"Slices should be evenly spaced (slice %d)", i)
		}
	}
}

func TestGenerateEqualSlicesWithDifferentCounts(t *testing.T) {
	m := NewModel(0, "", false)
	testFile := "../getbpm/Break120.wav"

	testCases := []struct {
		name       string
		sliceCount int
	}{
		{"4 slices", 4},
		{"8 slices", 8},
		{"16 slices", 16},
		{"32 slices", 32},
		{"64 slices", 64},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m.FileMetadata[testFile] = types.FileMetadata{
				BPM:         120.0,
				Slices:      tc.sliceCount,
				SliceType:   0, // Equal mode
				Playthrough: 0,
				SyncToBPM:   1,
			}

			m.GenerateEqualSlices(testFile)

			metadata := m.FileMetadata[testFile]
			assert.Len(t, metadata.Onsets, tc.sliceCount, 
				"Should generate %d equal slices", tc.sliceCount)
			assert.InDelta(t, 0.0, metadata.Onsets[0], 0.001, 
				"First slice should start at 0.0")
		})
	}
}

func TestGenerateEqualSlicesOnlyInEqualMode(t *testing.T) {
	m := NewModel(0, "", false)
	testFile := "../getbpm/Break120.wav"

	// Set up file metadata with Onset mode (SliceType=1)
	m.FileMetadata[testFile] = types.FileMetadata{
		BPM:         120.0,
		Slices:      16,
		SliceType:   1, // Onset mode - should not generate equal slices
		Playthrough: 0,
		SyncToBPM:   1,
	}

	// Try to generate equal slices (should not do anything)
	m.GenerateEqualSlices(testFile)

	// Verify that no slices were generated (function should exit early)
	metadata := m.FileMetadata[testFile]
	assert.Len(t, metadata.Onsets, 0, 
		"Should not generate slices when in Onset mode (SliceType=1)")
}
