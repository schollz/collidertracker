package input

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/schollz/collidertracker/internal/getbpm"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/modulation"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/types"
)

// GetPhrasesDataForTrack returns the appropriate phrases data based on track type
func GetPhrasesDataForTrack(m *model.Model, track int) *[255][][]int {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentPhrasesData
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerPhrasesData
}

// GetChainsDataForTrack returns the appropriate chains data based on track type
func GetChainsDataForTrack(m *model.Model, track int) *[][]int {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentChainsData
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerChainsData
}

// GetModulateSettingsForTrack returns the appropriate modulate settings based on track type
func GetModulateSettingsForTrack(m *model.Model, track int) *[255]types.ModulateSettings {
	if track >= 0 && track < 8 && !m.TrackTypes[track] {
		// TrackTypes[track] = false means Instrument
		return &m.InstrumentModulateSettings
	}
	// TrackTypes[track] = true means Sampler (or invalid track defaults to Sampler)
	return &m.SamplerModulateSettings
}

// ValueModifier represents a function that modifies a value with bounds checking
type ValueModifier struct {
	GetValue         func() interface{}
	SetValue         func(interface{})
	ApplyDelta       func(current interface{}, delta float32) interface{}
	ValidateAndClamp func(value interface{}) interface{}
	FormatValue      func(value interface{}) string
	LogPrefix        string
}

// modifyValueWithBounds provides common logic for value modification with bounds checking
func modifyValueWithBounds(modifier ValueModifier, delta float32) {
	current := modifier.GetValue()
	newValue := modifier.ApplyDelta(current, delta)
	clampedValue := modifier.ValidateAndClamp(newValue)

	// Log the change
	oldFormatted := modifier.FormatValue(current)
	newFormatted := modifier.FormatValue(clampedValue)
	log.Printf("Modified %s: %s -> %s (delta: %.2f)", modifier.LogPrefix, oldFormatted, newFormatted, delta)

	modifier.SetValue(clampedValue)
}

// createFloatModifier creates a ValueModifier for float32 values
func createFloatModifier(getValue func() float32, setValue func(float32), min, max float32, logPrefix string) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(float32)) },
		ApplyDelta: func(current interface{}, delta float32) interface{} {
			return current.(float32) + delta
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(float32)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%.2f", value.(float32))
		},
		LogPrefix: logPrefix,
	}
}

// createIntModifier creates a ValueModifier for int values
func createIntModifier(getValue func() int, setValue func(int), min, max int, logPrefix string) ValueModifier {
	return ValueModifier{
		GetValue: func() interface{} { return getValue() },
		SetValue: func(v interface{}) { setValue(v.(int)) },
		ApplyDelta: func(current interface{}, delta float32) interface{} {
			// Handle fine control (Ctrl+Left/Right): 0.05 or -0.05 should become +/-1
			var intDelta int
			if delta == 0.05 || delta == -0.05 {
				intDelta = int(delta / 0.05) // Fine control: convert to +/-1
			} else {
				intDelta = int(delta) // Coarse control or other
			}
			return current.(int) + intDelta
		},
		ValidateAndClamp: func(value interface{}) interface{} {
			v := value.(int)
			if v < min {
				return min
			} else if v > max {
				return max
			}
			return v
		},
		FormatValue: func(value interface{}) string {
			return fmt.Sprintf("%d", value.(int))
		},
		LogPrefix: logPrefix,
	}
}

// deltaHandler handles different delta increments (coarse vs fine control)
func deltaHandler(baseDelta float32, coarseMultiplier float32) float32 {
	if baseDelta == 1.0 || baseDelta == -1.0 {
		return baseDelta * coarseMultiplier // Coarse control (Ctrl+Up/Down)
	} else if baseDelta == 0.05 || baseDelta == -0.05 {
		return baseDelta / 0.05 // Fine control (Ctrl+Left/Right): converts to +/-1
	}
	return baseDelta // Fallback
}

func ModifyValue(m *model.Model, delta int) {
	if m.ViewMode == types.ChainView {
		// Chain view now only has phrase editing
		chainsData := m.GetCurrentChainsData()
		currentValue := (*chainsData)[m.CurrentChain][m.CurrentRow]

		var newValue int
		if currentValue == -1 {
			// First edit on an empty cell: initialize to 00 and DO NOT apply delta
			newValue = 0
		} else {
			newValue = currentValue + delta
		}

		if newValue < 0 {
			newValue = 0
		} else if newValue > 254 {
			newValue = 254
		}
		(*chainsData)[m.CurrentChain][m.CurrentRow] = newValue

		log.Printf("Modified chain %02X row %02X phrase: %d -> %d (delta: %d)", m.CurrentChain, m.CurrentRow, currentValue, newValue, delta)
		storage.AutoSave(m)
		return
	}

	// Phrase view: modify cell under cursor
	// Use centralized column mapping system
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	if columnMapping == nil || !columnMapping.IsEditable {
		return // Invalid or non-editable column
	}

	colIndex := columnMapping.DataColumnIndex
	phrasesData := m.GetCurrentPhrasesData()
	currentValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex]

	if colIndex == int(types.ColDeltaTime) {
		// DT column: clamp 0..254 (hex range)
		if currentValue == -1 {
			currentValue = 0
		}
		newValue := currentValue + delta
		if newValue < 0 {
			newValue = 0
		} else if newValue > 254 {
			newValue = 254
		}
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

	} else if colIndex == int(types.ColEffectReverse) {
		// Я column: probability 0..15 (0x0-0xF, where 15 = 100% chance, 0 = 0% chance)
		if currentValue == -1 {
			currentValue = 0
		}
		newValue := currentValue + delta
		if newValue < 0 {
			newValue = 0 // Clamp to min
		} else if newValue > 15 {
			newValue = 15 // Clamp to max
		}
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

	} else {
		// Handle different behavior for Instrument vs Sampler views
		phraseViewType := m.GetPhraseViewType()

		if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColNote) {
			// Instrument view note column: MIDI notes (0-127) with special increment behavior
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to the last note above (or middle C (60) if no notes above)
				newValue = FindFirstNonEmptyNoteAbove(phrasesData, m.CurrentPhrase, m.CurrentRow)
			} else {
				// Apply special increment logic for instrument notes
				// Coarse (Ctrl+Up/Down) should increment by 12 (octaves)
				// Fine (Ctrl+Left/Right) should increment by 1 (semitones)
				if delta == 16 || delta == -16 {
					// This is coarse increment - convert to octave increment (12 semitones)
					octaveDelta := (delta / 16) * 12
					newValue = currentValue + octaveDelta
				} else {
					// This is fine increment (+/-1)
					newValue = currentValue + delta
				}
			}

			// Clamp to MIDI range (0-127)
			if newValue < 0 {
				newValue = 0
			} else if newValue > 127 {
				newValue = 127
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue

			// Auto-set DT only when changing from no note (-1) to a note AND DT is currently -1
			// Use the first non "--" DT value above current row, or default to 1 if none found
			if currentValue == -1 && newValue != -1 && (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] == -1 {
				dtValue := FindFirstNonEmptyDTAbove(phrasesData, m.CurrentPhrase, m.CurrentRow)
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = dtValue
			}
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChord) {
			// Instrument view chord column: Cycle through chord types, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordNone if unset
				currentValue = int(types.ChordNone)
			}

			// Determine direction based on delta (Ctrl+Up/Right = forward, Ctrl+Down/Left = backward)
			if delta > 0 {
				// Forward: "-" -> "M" -> "m" -> "d" (stop at "d")
				newValue = currentValue + 1
				if newValue >= int(types.ChordTypeCount) {
					newValue = int(types.ChordTypeCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "d" -> "m" -> "M" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordAddition) {
			// Instrument view chord addition column: Cycle through addition types, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordAddNone if unset
				currentValue = int(types.ChordAddNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward: "-" -> "7" -> "9" -> "4" (stop at "4")
				newValue = currentValue + 1
				if newValue >= int(types.ChordAdditionCount) {
					newValue = int(types.ChordAdditionCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "4" -> "9" -> "7" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColChordTransposition) {
			// Instrument view chord transposition column: Cycle through transposition values, stop at ends
			var newValue int
			if currentValue == -1 {
				// Initialize to ChordTransNone if unset
				currentValue = int(types.ChordTransNone)
			}

			// Determine direction based on delta
			if delta > 0 {
				// Forward: "-" -> "0" -> "1" -> ... -> "F" (stop at "F")
				newValue = currentValue + 1
				if newValue >= int(types.ChordTranspositionCount) {
					newValue = int(types.ChordTranspositionCount) - 1 // Stop at last valid value
				}
			} else {
				// Backward: "F" -> "E" -> ... -> "0" -> "-" (stop at "-")
				newValue = currentValue - 1
				if newValue < 0 {
					newValue = 0 // Stop at first valid value
				}
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if phraseViewType == types.InstrumentPhraseView && colIndex == int(types.ColMidi) {
			// Instrument view MIDI column: hex values 00-FE
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to 00 and DO NOT apply delta
				newValue = 0
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 254 {
				newValue = 254
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if colIndex == int(types.ColVelocity) {
			// VE (Velocity) column: special handling to limit to 0x7F (127)
			virtualDefault := types.GetVirtualDefault(types.PhraseColumn(colIndex))
			var newValue int
			if currentValue == -1 {
				if virtualDefault != nil {
					// Virtual default column: start from virtual default value and apply delta
					newValue = virtualDefault.DefaultValue + delta
				} else {
					// Regular column: initialize to 00 and DO NOT apply delta
					newValue = 0
				}
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 127 { // Limit VE to 0x7F (127)
				newValue = 127
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else if colIndex >= int(types.ColMidiCC0) && colIndex <= int(types.ColMidiCC8) {
			// MIDI CC columns: special handling to limit to 0x7F (127)
			var newValue int
			if currentValue == -1 {
				// First edit on an empty cell: initialize to 00 and DO NOT apply delta
				newValue = 0
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 127 { // Limit MIDI CC to 0x7F (127)
				newValue = 127
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		} else {
			// All other hex-ish columns (NN, DT, GT, RT, TS, CO, FI index) - check for virtual defaults
			virtualDefault := types.GetVirtualDefault(types.PhraseColumn(colIndex))
			var newValue int
			if currentValue == -1 {
				if virtualDefault != nil {
					// Virtual default column: start from virtual default value and apply delta
					newValue = virtualDefault.DefaultValue + delta
				} else {
					// Regular column: initialize to 00 and DO NOT apply delta
					newValue = 0
				}
			} else {
				newValue = currentValue + delta
			}

			if newValue < 0 {
				newValue = 0
			} else if newValue > 254 {
				newValue = 254
			}
			(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex] = newValue
		}

		// Auto-enable playback on first note entry - use DT for both views
		if colIndex == int(types.ColNote) {
			// Only auto-set DT when changing from no note (-1) to a note (not -1) AND DT is currently -1
			if currentValue == -1 && (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColNote] != -1 &&
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] == -1 {
				// Auto-set DT using the first non "--" DT value above current row, or default to 01 if none found
				dtValue := FindFirstNonEmptyDTAbove(phrasesData, m.CurrentPhrase, m.CurrentRow)
				(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] = dtValue
				log.Printf("Auto-set DT=%02X for phrase %d row %d due to note change from -1 to note", dtValue, m.CurrentPhrase, m.CurrentRow)
			}
		}
	}

	m.LastEditRow = m.CurrentRow
	log.Printf("Modified phrase %d row %d, col %d: %d -> %d (delta: %d)",
		m.CurrentPhrase, m.CurrentRow, colIndex, currentValue,
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][colIndex], delta)

	// If this row is currently playing, send an update OSC message
	if m.IsRowCurrentlyPlaying(m.CurrentPhrase, m.CurrentRow, m.CurrentTrack) {
		log.Printf("Row is currently playing, sending update OSC message")
		EmitRowDataFor(m, m.CurrentPhrase, m.CurrentRow, m.CurrentTrack, true) // true indicates this is an update
	}

	storage.AutoSave(m)
}

func DebugLogRowEmission(m *model.Model) {
	// Delegate to the single canonical emitter so "space" playback and "c" manual emit behave identically.
	if m.PlaybackPhrase < 0 || m.PlaybackPhrase >= 255 || m.PlaybackRow < 0 || m.PlaybackRow >= 255 {
		log.Printf("ROW_EMIT: Invalid playback position - Phrase: %d, Row: %d", m.PlaybackPhrase, m.PlaybackRow)
		return
	}
	// Use current track context for chain and phrase playback modes
	EmitRowDataFor(m, m.PlaybackPhrase, m.PlaybackRow, m.CurrentTrack)
}

func FindFirstNonEmptyRowInPhrase(m *model.Model, phraseNum int) int {
	// Use current track context to determine which data pool to search
	return FindFirstNonEmptyRowInPhraseForTrack(m, phraseNum, m.CurrentTrack)
}

func FindFirstNonEmptyRowInPhraseForTrack(m *model.Model, phraseNum int, track int) int {
	if phraseNum >= 0 && phraseNum < 255 {
		phrasesData := GetPhrasesDataForTrack(m, track)
		log.Printf("DEBUG: FindFirstNonEmptyRowInPhraseForTrack - phrase=%d, track=%d", phraseNum, track)
		for i := 0; i < 255; i++ {
			// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
			dtValue := (*phrasesData)[phraseNum][i][types.ColDeltaTime]
			if IsRowPlayable(dtValue) {
				log.Printf("DEBUG: Found playback row %d with DT=%d", i, dtValue)
				return i
			}
		}
		log.Printf("DEBUG: No playback rows found in phrase %d track %d", phraseNum, track)
	}
	return 0 // Fallback to row 0 if no playback rows found
}

func FindFirstNonEmptyChain(m *model.Model) int {
	chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
	for i := 0; i < 255; i++ {
		// Check if any phrase is assigned in this chain
		for row := 0; row < 16; row++ {
			if (*chainsData)[i][row] != -1 {
				return i
			}
		}
	}
	return 0 // Fallback to chain 0 if no data found
}

func CopyLastRowWithIncrement(m *model.Model) {
	if m.ViewMode != types.PhraseView {
		return // Only works in phrase view
	}

	log.Printf("copyLastRowWithIncrement called - currentRow: %d", m.CurrentRow)

	// Check if we're on a header row (invalid row index)
	if m.CurrentRow < 0 || m.CurrentRow >= 256 {
		log.Printf("Cannot copy on header row (row %d)", m.CurrentRow)
		return
	}

	// Get the appropriate phrases data based on view type
	phrasesData := m.GetCurrentPhrasesData()
	phraseViewType := m.GetPhraseViewType()

	// Check if current row is empty (note, deltatime, filename are -1, playback is 0)
	if (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColNote] != -1 ||
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime] != -1 ||
		(*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColFilename] != -1 {
		log.Printf("Current row %d is not empty, skipping copy", m.CurrentRow)
		return
	}

	// Find the first non-null note above the current row
	var sourceNote int = -1
	for r := m.CurrentRow - 1; r >= 0; r-- {
		if (*phrasesData)[m.CurrentPhrase][r][types.ColNote] != -1 {
			sourceNote = (*phrasesData)[m.CurrentPhrase][r][types.ColNote]
			log.Printf("Found non-null note %d at row %d", sourceNote, r)
			break
		}
	}

	// If no non-null note found above, use default starting note
	if sourceNote == -1 {
		if phraseViewType == types.InstrumentPhraseView {
			// For Instrument view, start with middle C (60)
			sourceNote = 59 // Will be incremented to 60
		} else {
			// For Sampler view, start with 0
			sourceNote = -1 // Will be incremented to 0
		}
		log.Printf("No non-null note found above current row, using default start")
	}

	// Set DT to 1 for both view types
	(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColDeltaTime)] = 1

	// Increment the note and set it
	var newNote int
	if phraseViewType == types.InstrumentPhraseView {
		// For Instrument view: increment MIDI notes (0-127), with chromatic progression
		newNote = sourceNote + 1
		if newNote > 127 { // Wrap around MIDI range
			newNote = 0
		}
	} else {
		// For Sampler view: increment sample/note numbers (0-254)
		newNote = sourceNote + 1
		if newNote > 254 { // Wrap around if needed
			newNote = 0
		}
	}
	(*phrasesData)[m.CurrentPhrase][m.CurrentRow][int(types.ColNote)] = newNote

	log.Printf("Set note on row %d: %d->%d, P=1", m.CurrentRow, sourceNote, newNote)

	// Update the last edit row to current row
	m.LastEditRow = m.CurrentRow
}

// IsRowEmpty checks if the current row in phrase view is empty
func IsRowEmpty(m *model.Model) bool {
	if m.ViewMode != types.PhraseView {
		return true
	}

	// Check if we're on a header row (invalid row index)
	if m.CurrentRow < 0 || m.CurrentRow >= 256 {
		return true
	}

	// Use track-aware data access
	phrasesData := m.GetCurrentPhrasesData()
	rowData := (*phrasesData)[m.CurrentPhrase][m.CurrentRow]

	// A row is considered empty if all key data fields are at their default values
	// For both instrument and sampler tracks, check Note and DeltaTime
	isEmpty := rowData[types.ColNote] == -1 && rowData[types.ColDeltaTime] == -1

	// For sampler tracks, also check filename
	if m.GetPhraseViewType() == types.SamplerPhraseView {
		isEmpty = isEmpty && rowData[types.ColFilename] == -1
	}

	return isEmpty
}

// GetEffectiveValue searches backwards from the current row to find the first non-null value for a given column
func GetEffectiveValue(m *model.Model, phrase, row, colIndex int) int {
	return GetEffectiveValueForTrack(m, phrase, row, colIndex, m.CurrentTrack)
}

// GetEffectiveValueForTrack searches backwards from the current row to find the first non-null value for a given column
func GetEffectiveValueForTrack(m *model.Model, phrase, row, colIndex, trackId int) int {
	// Use track-specific data pool
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	// Search backwards from the given row to find first non-null value
	for r := row; r >= 0; r-- {
		value := (*phrasesData)[phrase][r][colIndex]
		if value != -1 {
			return value
		}
	}
	return -1 // No non-null value found
}

// GetEffectiveMidiValue gets the effective MIDI value for a row (sticky behavior)
func GetEffectiveMidiValue(m *model.Model, phrase, row int) int {
	return GetEffectiveMidiValueForTrack(m, phrase, row, m.CurrentTrack)
}

// GetEffectiveMidiValueForTrack gets the effective MIDI value for a row for a specific track
func GetEffectiveMidiValueForTrack(m *model.Model, phrase, row, trackId int) int {
	return GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
}

// GetEffectiveFilename gets the effective filename for a row
func GetEffectiveFilename(m *model.Model, phrase, row int) string {
	return GetEffectiveFilenameForTrack(m, phrase, row, m.CurrentTrack)
}

// GetEffectiveFilenameForTrack gets the effective filename for a row for a specific track
func GetEffectiveFilenameForTrack(m *model.Model, phrase, row, trackId int) string {
	effectiveFileIndex := GetEffectiveValueForTrack(m, phrase, row, int(types.ColFilename), trackId)

	// Use track-specific files array based on track type
	var phrasesFiles *[]string
	if trackId >= 0 && trackId < 8 && !m.TrackTypes[trackId] {
		// TrackTypes[trackId] = false means Instrument - don't use files
		return "none"
	} else {
		// TrackTypes[trackId] = true means Sampler - use SamplerPhrasesFiles
		phrasesFiles = &m.SamplerPhrasesFiles
	}

	if effectiveFileIndex >= 0 && effectiveFileIndex < len(*phrasesFiles) && (*phrasesFiles)[effectiveFileIndex] != "" {
		return (*phrasesFiles)[effectiveFileIndex]
	}
	return "none"
}

func EmitRowData(m *model.Model) {
	if m.ViewMode != types.PhraseView {
		return
	}
	EmitRowDataFor(m, m.CurrentPhrase, m.CurrentRow, m.CurrentTrack)
}

func EmitLastSelectedPhraseRowData(m *model.Model) {
	EmitRowDataFor(m, m.CurrentPhrase, m.LastPhraseRow, m.CurrentTrack)
}

// EmitRowDataFor logs row data (rich debug) and emits OSC if applicable.
// This is the single canonical emitter used by both manual "c" triggers and playback ("space").
// If isUpdate is true, the OSC message will include "update",1 to indicate this is an update to a playing row.
func EmitRowDataFor(m *model.Model, phrase, row, trackId int, isUpdate ...bool) {
	shouldUpdate := len(isUpdate) > 0 && isUpdate[0]
	log.Printf("DEBUG_EMIT: EmitRowDataFor called with phrase=%d, row=%d, trackId=%d, isUpdate=%v", phrase, row, trackId, shouldUpdate)

	// Defensive null check to prevent crashes
	if m == nil {
		log.Printf("ERROR: EmitRowDataFor called with nil model")
		return
	}

	// Validate input parameters
	if phrase < 0 || phrase >= 255 || row < 0 || row >= 255 || trackId < 0 || trackId >= 8 {
		log.Printf("ERROR: EmitRowDataFor called with invalid parameters - phrase=%d, row=%d, trackId=%d", phrase, row, trackId)
		return
	}

	// Use track-aware data access for correct playback
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	if phrasesData == nil {
		log.Printf("ERROR: GetPhrasesDataForTrack returned nil for trackId=%d", trackId)
		return
	}
	rowData := (*phrasesData)[phrase][row]

	// Raw values - DT used for playback control in both views
	rawNote := rowData[types.ColNote]
	rawPitch := rowData[types.ColPitch]
	rawDeltaTime := rowData[types.ColDeltaTime]
	rawGate := rowData[types.ColGate]
	rawRetrigger := rowData[types.ColRetrigger]
	rawTimestretch := rowData[types.ColTimestretch]
	rawModulate := rowData[types.ColModulate]
	rawFilenameIndex := rowData[types.ColFilename]

	// Effect columns (may be -1) - get both raw and effective values
	rawEffectReverse := -1
	rawPan := -1
	rawLowPassFilter := -1
	rawHighPassFilter := -1
	rawEffectComb := -1
	rawEffectReverb := -1
	if int(types.ColCount) > int(types.ColEffectReverb) { // guard in case of older saves
		rawEffectReverse = rowData[types.ColEffectReverse]
		rawPan = rowData[types.ColPan]
		rawLowPassFilter = rowData[types.ColLowPassFilter]
		rawHighPassFilter = rowData[types.ColHighPassFilter]
		rawEffectComb = rowData[types.ColEffectComb]
		rawEffectReverb = rowData[types.ColEffectReverb]
	}

	// Get effective values for sticky columns (PA, LP, HP, CO, VE)
	effectivePan := GetEffectiveValueForTrack(m, phrase, row, int(types.ColPan), trackId)
	effectiveLowPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColLowPassFilter), trackId)
	effectiveHighPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColHighPassFilter), trackId)
	effectiveComb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectComb), trackId)
	effectiveReverb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectReverb), trackId)

	// Effective/inherited values
	effectiveNote := GetEffectiveValueForTrack(m, phrase, row, int(types.ColNote), trackId)
	rawNoteModulated := rawNote

	// For sampler tracks, apply modulation as before (current behavior)
	if !isInstrumentTrack(m, trackId) {
		// Apply modulation if there's a modulate setting active on this row
		if rawModulate != -1 && rawModulate >= 0 && rawModulate < 255 && effectiveNote != -1 {
			modulateSettings := (*GetModulateSettingsForTrack(m, trackId))[rawModulate]
			originalNote := effectiveNote

			// Apply increment before other modulation operations if counter > -1
			incrementCounter := m.IncrementCounters[trackId][phrase][row]
			originalNote = modulation.ApplyIncrement(originalNote, incrementCounter, modulateSettings.Increment, modulateSettings.Wrap)

			// Get track-specific RNG for modulation
			var trackRng *rand.Rand
			if trackId >= 0 && trackId < 8 {
				trackRng = m.ModulateRngs[trackId]
			} else {
				// Fallback to creating a temporary RNG for invalid track IDs
				trackRng = rand.New(rand.NewSource(time.Now().UnixNano()))
			}

			modulatedNote := modulation.ApplyModulation(originalNote, modulation.ModulateSettings{
				Seed:        modulateSettings.Seed,
				IRandom:     modulateSettings.IRandom,
				Sub:         modulateSettings.Sub,
				Add:         modulateSettings.Add,
				Increment:   modulateSettings.Increment,
				Wrap:        modulateSettings.Wrap,
				ScaleRoot:   modulateSettings.ScaleRoot,
				Scale:       modulateSettings.Scale,
				Probability: modulateSettings.Probability,
			}, trackRng)

			// Apply the same logic for raw note
			rawNoteWithIncrement := modulation.ApplyIncrement(rawNote, incrementCounter, modulateSettings.Increment, modulateSettings.Wrap)
			rawNoteModulated = modulation.ApplyModulation(rawNoteWithIncrement, modulation.ModulateSettings{
				Seed:        modulateSettings.Seed,
				IRandom:     modulateSettings.IRandom,
				Sub:         modulateSettings.Sub,
				Add:         modulateSettings.Add,
				Increment:   modulateSettings.Increment,
				Wrap:        modulateSettings.Wrap,
				ScaleRoot:   modulateSettings.ScaleRoot,
				Scale:       modulateSettings.Scale,
				Probability: modulateSettings.Probability,
			}, trackRng)
			log.Printf("Applied modulation %02X: NN %02X -> %02X (Seed=%d, IRandom=%d, Sub=%d, Add=%d, Increment=%d, Scale=%s)",
				rawModulate, effectiveNote, modulatedNote, modulateSettings.Seed, modulateSettings.IRandom, modulateSettings.Sub, modulateSettings.Add, modulateSettings.Increment, modulateSettings.Scale)
			effectiveNote = modulatedNote
		}
	}

	effectiveDeltaTime := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDeltaTime), trackId)
	effectiveFilenameIndex := GetEffectiveValueForTrack(m, phrase, row, int(types.ColFilename), trackId)
	effectiveFilename := GetEffectiveFilenameForTrack(m, phrase, row, trackId)
	effectiveDucking := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectDucking), trackId)

	// Helper to format hex-ish cells
	formatHex := func(value int) string {
		if value == -1 {
			return "--"
		}
		return fmt.Sprintf("%02X", value)
	}

	// RICH DEBUG LOG (adapted from previous DebugLogRowEmission)
	log.Printf("=== ROW EMISSION ===")
	if m.IsPlaying {
		log.Printf("Context: PlaybackMode=%v Chain:%02X Phrase:%02X Row:%02X IsPlaying=%v BPM=%.2f PPQ=%d",
			m.PlaybackMode, m.PlaybackChain, m.PlaybackPhrase, m.PlaybackRow, m.IsPlaying, m.BPM, m.PPQ)
	} else {
		log.Printf("Context: Manual emit  Phrase:%02X Row:%02X BPM=%.2f PPQ=%d", phrase, row, m.BPM, m.PPQ)
	}
	log.Printf("DeltaTime (playback control): %d", rawDeltaTime)
	// Show different debug info based on track type
	if trackId >= 0 && trackId < 8 && !m.TrackTypes[trackId] {
		// Instrument track - show all instrument parameters
		rawChord := rowData[types.ColChord]
		rawChordAdd := rowData[types.ColChordAddition]
		rawChordTrans := rowData[types.ColChordTransposition]
		rawAttack := rowData[types.ColAttack]
		rawDecay := rowData[types.ColDecay]
		rawSustain := rowData[types.ColSustain]
		rawRelease := rowData[types.ColRelease]
		rawArpeggio := rowData[types.ColArpeggio]
		rawMidi := rowData[types.ColMidi]
		rawSoundMaker := rowData[types.ColSoundMaker]

		log.Printf("Raw:  NN=%s GT=%s C=%d A=%d T=%d A=%s D=%s S=%s R=%s AR=%s MI=%s SO=%s",
			formatHex(rawNote),
			formatHex(rawGate),
			rawChord,
			rawChordAdd,
			rawChordTrans,
			formatHex(rawAttack),
			formatHex(rawDecay),
			formatHex(rawSustain),
			formatHex(rawRelease),
			formatHex(rawArpeggio),
			formatHex(rawMidi),
			formatHex(rawSoundMaker),
		)

		// Show effective values for sticky parameters
		effAttack := GetEffectiveValueForTrack(m, phrase, row, int(types.ColAttack), trackId)
		effDecay := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDecay), trackId)
		effSustain := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSustain), trackId)
		effRelease := GetEffectiveValueForTrack(m, phrase, row, int(types.ColRelease), trackId)
		effArpeggio := rowData[types.ColArpeggio] // Arpeggio should NOT be sticky - use current row only
		effMidi := GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
		effSoundMaker := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSoundMaker), trackId)

		log.Printf("Eff:  NN=%s A=%s D=%s S=%s R=%s AR=%s MI=%s SO=%s",
			formatHex(effectiveNote),
			formatHex(effAttack),
			formatHex(effDecay),
			formatHex(effSustain),
			formatHex(effRelease),
			formatHex(effArpeggio),
			formatHex(effMidi),
			formatHex(effSoundMaker),
		)
	} else {
		// Sampler track - show traditional sampler information
		log.Printf("Raw:  NN=%s DT=%s GT=%s RT=%s TS=%s FI=%d Я=%s PA=%s LP=%s HP=%s CO=%s VE=%s",
			formatHex(rawNote),
			formatHex(rawDeltaTime),
			formatHex(rawGate),
			formatHex(rawRetrigger),
			formatHex(rawTimestretch),
			rawFilenameIndex,
			func() string {
				if rawEffectReverse == -1 {
					return "-"
				}
				return fmt.Sprintf("%X", rawEffectReverse)
			}(),
			formatHex(rawPan),
			formatHex(rawLowPassFilter),
			formatHex(rawHighPassFilter),
			formatHex(rawEffectComb),
			formatHex(rawEffectReverb),
		)
		log.Printf("Eff:  NN=%s DT=%s FI=%d FN=%s PA=%s LP=%s HP=%s CO=%s VE=%s",
			formatHex(effectiveNote),
			formatHex(effectiveDeltaTime),
			effectiveFilenameIndex,
			effectiveFilename,
			formatHex(effectivePan),
			formatHex(effectiveLowPassFilter),
			formatHex(effectiveHighPassFilter),
			formatHex(effectiveComb),
			formatHex(effectiveReverb),
		)
	}

	// Only emit if we have playback enabled and a concrete note
	// For samplers, also check that we have a filename
	needsFile := trackId >= 0 && trackId < 8 && m.TrackTypes[trackId] // Sampler tracks need files

	// Check if any CC values are set for instrument tracks
	hasCCValues := false
	if isInstrumentTrack(m, trackId) {
		for i := 0; i < 9; i++ {
			if rowData[types.ColMidiCC0+types.PhraseColumn(i)] != -1 {
				hasCCValues = true
				break
			}
		}
	}

	// Unified DT-based playback condition: DT > 0 means play for both instruments and samplers
	// For instrument tracks, allow emission if CC values are set even without a note
	if !IsRowPlayable(rawDeltaTime) || (rawNote == -1 && !hasCCValues) || (needsFile && effectiveFilename == "none") {
		if !IsRowPlayable(rawDeltaTime) {
			log.Printf("ROW_EMIT: skipped (DT<=0, value=%d)", rawDeltaTime)
		} else if rawNote == -1 && !hasCCValues {
			log.Printf("ROW_EMIT: skipped (NN==null and no CC values)")
		} else if needsFile && effectiveFilename == "none" {
			log.Printf("ROW_EMIT: skipped (sampler track needs filename)")
		}
		return
	}

	// shouldEmitRow() hook (kept from playback path; e.g., to skip DT==00, etc.)
	if shouldEmitRowForTrackAtPosition(m, phrase, row, trackId) == false {
		log.Printf("ROW_EMIT: skipped by shouldEmitRow()")
		return
	}

	// ONLY cancel any existing arpeggio on this track when a new note is actually going to start
	// This ensures arpeggios are cancelled only when a real note is triggered, not just during row processing
	if trackId >= 0 && trackId < 8 {
		log.Printf("DEBUG_EMIT: About to cancel any existing arpeggio for track %d (new note starting)", trackId)
		m.CancelArpeggioForTrack(int32(trackId))
		log.Printf("DEBUG_EMIT: Cancelled any existing arpeggio for track %d (new note starting)", trackId)
	}

	// Build slice/OSC params
	fileMetadata, exists := m.FileMetadata[effectiveFilename]
	sliceCount := 16
	bpmSource := float32(120.0)
	playthrough := 0 // Default: Sliced
	syncToBPM := 1   // Default: Yes
	if exists {
		sliceCount = fileMetadata.Slices
		bpmSource = fileMetadata.BPM
		playthrough = fileMetadata.Playthrough
		syncToBPM = fileMetadata.SyncToBPM
	}
	sliceNumber := rawNoteModulated % sliceCount

	// Get effective gate value (handles sticky behavior and virtual defaults)
	effectiveGate := GetEffectiveValueForTrack(m, phrase, row, int(types.ColGate), trackId)
	if effectiveGate == -1 {
		effectiveGate = 0x80 // Default Gate value (128)
	}

	baseDuration := 1.0 / float32(m.PPQ)
	gateMultiplier := float32(effectiveGate) / 96.0
	sliceDuration := baseDuration * gateMultiplier

	// Apply DT multiplier if DT value is non-zero
	if rawDeltaTime > 0 {
		sliceDuration *= float32(rawDeltaTime)
	}

	// Calculate delta time in seconds (time per row * DT)
	deltaTimeSeconds := calculateDeltaTimeSeconds(m, phrase, row, trackId)

	// Calculate velocity from velocity column (sticky behavior)
	rawVelocity := GetEffectiveValueForTrack(m, phrase, row, int(types.ColVelocity), trackId)
	velocity := 64 // Default velocity (0x40)
	if rawVelocity != -1 {
		velocity = rawVelocity // Keep as integer (0x00-0x7F = 0-127)
	}
	if velocity > 127 {
		velocity = 127
	}

	// Increment step counter for this position (for effect Every functionality)
	// Add defensive check to ensure model is not nil and arrays are properly initialized
	if m != nil && trackId >= 0 && trackId < 8 && phrase >= 0 && phrase < 255 && row >= 0 && row < 255 {
		m.EffectStepCounter[trackId][phrase][row]++
		log.Printf("DEBUG_EFFECTS: Incremented step counter for track=%d phrase=%d row=%d, count=%d", trackId, phrase, row, m.EffectStepCounter[trackId][phrase][row])

		// Handle increment counter logic
		if rawModulate != -1 && rawModulate >= 0 && rawModulate < 255 {
			modulateSettings := (*GetModulateSettingsForTrack(m, trackId))[rawModulate]
			if modulateSettings.Increment > 0 {
				// Add increment value to the counter
				m.IncrementCounters[trackId][phrase][row] += modulateSettings.Increment
				log.Printf("DEBUG_INCREMENT: Added increment %d to counter for track=%d phrase=%d row=%d, new counter=%d",
					modulateSettings.Increment, trackId, phrase, row, m.IncrementCounters[trackId][phrase][row])
			}
		}
	} else {
		log.Printf("WARNING: Invalid parameters for step counter - model=%v, trackId=%d, phrase=%d, row=%d", m != nil, trackId, phrase, row)
	}

	var oscParams model.SamplerOSCParams
	// Check if retrigger is set and should be active based on Every setting
	isRetriggerActive := false
	if rawRetrigger != -1 && rawRetrigger >= 0 && rawRetrigger < 255 {
		retriggerSettings := m.RetriggerSettings[rawRetrigger]

		// Validate Every field to prevent division by zero
		if retriggerSettings.Every <= 0 {
			log.Printf("WARNING: Invalid retrigger Every value %d for index %d, defaulting to 1", retriggerSettings.Every, rawRetrigger)
			retriggerSettings.Every = 1
			m.RetriggerSettings[rawRetrigger] = retriggerSettings // Update the model with corrected value
		}

		if m != nil && trackId >= 0 && trackId < 8 && phrase >= 0 && phrase < 255 && row >= 0 && row < 255 {
			stepCount := m.EffectStepCounter[trackId][phrase][row]
			everyActive := stepCount%retriggerSettings.Every == 0

			// Apply probability AFTER Every check
			if everyActive {
				// Validate Probability field
				if retriggerSettings.Probability < 0 || retriggerSettings.Probability > 100 {
					log.Printf("WARNING: Invalid retrigger Probability value %d for index %d, defaulting to 100", retriggerSettings.Probability, rawRetrigger)
					retriggerSettings.Probability = 100
					m.RetriggerSettings[rawRetrigger] = retriggerSettings
				}

				// Generate random number 1-100 and check against probability
				randomValue := (stepCount*31+trackId*17+phrase*13+row*7)%100 + 1
				isRetriggerActive = randomValue <= retriggerSettings.Probability
				log.Printf("DEBUG_RETRIGGER: track=%d phrase=%d row=%d, stepCount=%d, Every=%d, everyActive=%v, probability=%d%%, random=%d, finalActive=%v",
					trackId, phrase, row, stepCount, retriggerSettings.Every, everyActive, retriggerSettings.Probability, randomValue, isRetriggerActive)
			} else {
				log.Printf("DEBUG_RETRIGGER: track=%d phrase=%d row=%d, stepCount=%d, Every=%d, skipped by Every check", trackId, phrase, row, stepCount, retriggerSettings.Every)
			}
		} else {
			log.Printf("WARNING: Invalid parameters for retrigger check - model=%v, trackId=%d, phrase=%d, row=%d", m != nil, trackId, phrase, row)
		}

		if isRetriggerActive {
			oscParams = model.NewSamplerOSCParamsWithRetrigger(
				effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration,
				retriggerSettings.Times,
				float32(retriggerSettings.Beats),
				retriggerSettings.Start,
				retriggerSettings.End,
				retriggerSettings.PitchChange,
				retriggerSettings.VolumeDB,
				deltaTimeSeconds,
				velocity,
				retriggerSettings.FinalPitchToStart,
				retriggerSettings.FinalVolumeToStart,
			)
		} else {
			// Retrigger is set but not active this time, play normally without retrigger
			oscParams = model.NewSamplerOSCParams(effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration, deltaTimeSeconds, velocity)
		}
	} else {
		oscParams = model.NewSamplerOSCParams(effectiveFilename, trackId, sliceCount, sliceNumber, bpmSource, m.BPM, sliceDuration, deltaTimeSeconds, velocity)
	}

	// Apply onset-based slicing if enabled
	if exists && fileMetadata.SliceType == 1 && len(fileMetadata.Onsets) > 0 {
		// Calculate which onset to use with modulo wrapping
		onsetIndex := sliceNumber % len(fileMetadata.Onsets)
		
		// Get the onset time in seconds
		onsetTime := fileMetadata.Onsets[onsetIndex]
		
		// Get the audio file length to calculate normalized positions
		audioLength, _, _, err := getbpm.Length(effectiveFilename)
		if err != nil {
			log.Printf("WARNING: Failed to get audio length for %s: %v", effectiveFilename, err)
		} else {
			// Calculate start position (normalized 0.0-1.0)
			oscParams.SliceStart = float32(onsetTime / audioLength)
			
			// Calculate end position (next onset or end of file)
			if onsetIndex < len(fileMetadata.Onsets)-1 {
				nextOnsetTime := fileMetadata.Onsets[onsetIndex+1]
				oscParams.SliceEnd = float32(nextOnsetTime / audioLength)
			} else {
				// Last onset - play to end of file
				oscParams.SliceEnd = 1.0
			}
			
			log.Printf("Onset slicing: file=%s, slice=%d, onsetIndex=%d/%d, start=%.3f, end=%.3f", 
				effectiveFilename, sliceNumber, onsetIndex, len(fileMetadata.Onsets), oscParams.SliceStart, oscParams.SliceEnd)
		}
	}

	// Pitch conversion from hex to float: 128 (0x80) = 0.0, range 0-254 maps to -24 to +24
	if rawPitch != -1 {
		// Map 0-254 to -24 to +24, with 128 as center (0.0)
		oscParams.Pitch = ((float32(rawPitch) - 128.0) / 128.0) * 24.0
	} else {
		// Default pitch is 0.0 when cleared (-1)
		oscParams.Pitch = 0.0
	}

	// Timestretch - check if it should be active based on Every setting
	if rawTimestretch != -1 && rawTimestretch >= 0 && rawTimestretch < 255 {
		ts := m.TimestrechSettings[rawTimestretch]

		// Validate Every field to prevent division by zero
		if ts.Every <= 0 {
			log.Printf("WARNING: Invalid timestretch Every value %d for index %d, defaulting to 1", ts.Every, rawTimestretch)
			ts.Every = 1
			m.TimestrechSettings[rawTimestretch] = ts // Update the model with corrected value
		}

		isTimestrechActive := false
		if m != nil && trackId >= 0 && trackId < 8 && phrase >= 0 && phrase < 255 && row >= 0 && row < 255 {
			stepCount := m.EffectStepCounter[trackId][phrase][row]
			everyActive := stepCount%ts.Every == 0

			// Apply probability AFTER Every check
			if everyActive {
				// Validate Probability field
				if ts.Probability < 0 || ts.Probability > 100 {
					log.Printf("WARNING: Invalid timestretch Probability value %d for index %d, defaulting to 100", ts.Probability, rawTimestretch)
					ts.Probability = 100
					m.TimestrechSettings[rawTimestretch] = ts
				}

				// Generate random number 1-100 and check against probability (different seed than retrigger)
				randomValue := (stepCount*37+trackId*23+phrase*19+row*11)%100 + 1
				isTimestrechActive = randomValue <= ts.Probability
				log.Printf("DEBUG_TIMESTRETCH: track=%d phrase=%d row=%d, stepCount=%d, Every=%d, everyActive=%v, probability=%d%%, random=%d, finalActive=%v",
					trackId, phrase, row, stepCount, ts.Every, everyActive, ts.Probability, randomValue, isTimestrechActive)
			} else {
				log.Printf("DEBUG_TIMESTRETCH: track=%d phrase=%d row=%d, stepCount=%d, Every=%d, skipped by Every check", trackId, phrase, row, stepCount, ts.Every)
			}
		} else {
			log.Printf("WARNING: Invalid parameters for timestretch check - model=%v, trackId=%d, phrase=%d, row=%d", m != nil, trackId, phrase, row)
		}

		if isTimestrechActive {
			oscParams.TimestretchStart = float32(ts.Start)
			oscParams.TimestretchEnd = float32(ts.End)
			oscParams.TimestretchBeats = float32(ts.Beats)
		} else {
			// Timestretch is set but not active this time, use defaults (no timestretch)
			oscParams.TimestretchStart = 0.0
			oscParams.TimestretchEnd = 0.0
			oscParams.TimestretchBeats = 0.0
		}
	}

	oscParams.DuckingIndex = effectiveDucking

	// NEW effect params
	if rawEffectReverse != -1 {
		if rawEffectReverse == 0 {
			oscParams.EffectReverse = 0
		} else {
			// Probability-based reverse: 1-15 maps to ~6.67%-100% chance
			// Use true random for each playback
			probability := float64(rawEffectReverse) / 15.0 * 100.0 // Convert to percentage
			randomValue := float64(rand.Intn(100) + 1)              // 1-100
			if randomValue <= probability {
				oscParams.EffectReverse = 1
			} else {
				oscParams.EffectReverse = 0
			}
		}
	}
	// Pan: Use effective value, map 0-254 to -1.0 to 1.0, with 128 as center (0.0)
	if effectivePan != -1 {
		if effectivePan == 128 {
			oscParams.Pan = 0.0 // Explicit center
		} else {
			oscParams.Pan = (float32(effectivePan) - 127.0) / 127.0
		}
	} else {
		oscParams.Pan = 0.0 // Default to center when no effective value found
	}

	// Low Pass Filter: Use effective value, send 20000 Hz when -1, otherwise exponential mapping
	if effectiveLowPassFilter != -1 {
		// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
		logMin := float32(1.301) // log10(20)
		logMax := float32(4.301) // log10(20000)
		logFreq := logMin + (float32(effectiveLowPassFilter)/254.0)*(logMax-logMin)
		oscParams.LowPassFilter = float32(math.Pow(10, float64(logFreq)))
	} else {
		oscParams.LowPassFilter = 20000.0 // Send 20kHz when no effective value found
	}

	// High Pass Filter: Use effective value, send 20 Hz when -1, otherwise exponential mapping
	if effectiveHighPassFilter != -1 {
		// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
		logMin := float32(1.301) // log10(20)
		logMax := float32(4.301) // log10(20000)
		logFreq := logMin + (float32(effectiveHighPassFilter)/254.0)*(logMax-logMin)
		oscParams.HighPassFilter = float32(math.Pow(10, float64(logFreq)))
	} else {
		oscParams.HighPassFilter = 20.0 // Send 20Hz when no effective value found
	}

	// Comb: Use effective value
	if effectiveComb != -1 {
		oscParams.EffectComb = float32(effectiveComb) / 254.0
	}

	// Reverb: Use effective value
	if effectiveReverb != -1 {
		oscParams.EffectReverb = float32(effectiveReverb) / 254.0
	}

	// Set file metadata parameters
	oscParams.Playthrough = playthrough
	oscParams.SyncToBPM = syncToBPM

	// Set update flag if this is an update call
	if shouldUpdate {
		oscParams.Update = 1
	}

	log.Printf("[EmitRowDataFor] oscParams: %+v", oscParams)

	// Determine track type and emit appropriate message
	if isInstrumentTrack(m, trackId) {
		// For instrument tracks, extract all instrument-specific parameters
		rawVelocity := GetEffectiveValueForTrack(m, phrase, row, int(types.ColVelocity), trackId)
		velocity := float32(64) // Default velocity as float for instrument OSC (0x40)
		if rawVelocity != -1 {
			velocity = float32(rawVelocity) // Keep as integer value, just convert to float32 for OSC
		}
		if velocity > 127.0 {
			velocity = 127.0
		}

		// Extract chord parameters
		rawChord := rowData[types.ColChord]
		rawChordAdd := rowData[types.ColChordAddition]
		rawChordTrans := rowData[types.ColChordTransposition]

		// Extract ADSR parameters with effective values (sticky)
		rawAttack := GetEffectiveValueForTrack(m, phrase, row, int(types.ColAttack), trackId)
		rawDecay := GetEffectiveValueForTrack(m, phrase, row, int(types.ColDecay), trackId)
		rawSustain := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSustain), trackId)
		rawRelease := GetEffectiveValueForTrack(m, phrase, row, int(types.ColRelease), trackId)

		// Extract effect parameters with effective values (sticky)
		rawPan := GetEffectiveValueForTrack(m, phrase, row, int(types.ColPan), trackId)
		rawLowPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColLowPassFilter), trackId)
		rawHighPassFilter := GetEffectiveValueForTrack(m, phrase, row, int(types.ColHighPassFilter), trackId)
		rawEffectComb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectComb), trackId)
		rawEffectReverb := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectReverb), trackId)
		rawEffectDucking := GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectDucking), trackId)

		// Extract other parameters with effective values (sticky)
		rawArpeggio := rowData[types.ColArpeggio] // Arpeggio should NOT be sticky - use current row only
		rawMidi := GetEffectiveValueForTrack(m, phrase, row, int(types.ColMidi), trackId)
		rawSoundMaker := GetEffectiveValueForTrack(m, phrase, row, int(types.ColSoundMaker), trackId)

		// Extract Gate parameter with effective value (sticky)
		effectiveGate := GetEffectiveValueForTrack(m, phrase, row, int(types.ColGate), trackId)
		if effectiveGate == -1 {
			effectiveGate = 0x80 // Default Gate value (128)
		}

		// Calculate delta time in seconds (time per row * DT)
		deltaTimeSeconds := calculateDeltaTimeSeconds(m, phrase, row, trackId)

		// Convert ADSR values using the conversion functions from types package
		attack := float32(0.02) // Default
		if rawAttack != -1 {
			attack = types.AttackToSeconds(rawAttack)
		}

		decay := float32(0.0) // Default
		if rawDecay != -1 {
			decay = types.DecayToSeconds(rawDecay)
		}

		sustain := float32(1.0) // Default
		if rawSustain != -1 {
			sustain = types.SustainToLevel(rawSustain)
		}

		release := float32(0.02) // Default
		if rawRelease != -1 {
			release = types.ReleaseToSeconds(rawRelease)
		}

		// Convert effect parameters (similar to sampler conversion logic)
		pan := float32(0.0) // Default center pan
		if rawPan != -1 {
			pan = (float32(rawPan) - 127.0) / 127.0 // Map 0-254 to -1.0 to 1.0
		}

		lowPassFilter := float32(20000) // Default no filter (20kHz)
		if rawLowPassFilter != -1 {
			// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
			logMin := float32(1.301) // log10(20)
			logMax := float32(4.301) // log10(20000)
			logFreq := logMin + (float32(rawLowPassFilter)/254.0)*(logMax-logMin)
			lowPassFilter = float32(math.Pow(10, float64(logFreq)))
		}

		highPassFilter := float32(20) // Default no filter (20Hz)
		if rawHighPassFilter != -1 {
			// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
			logMin := float32(1.301) // log10(20)
			logMax := float32(4.301) // log10(20000)
			logFreq := logMin + (float32(rawHighPassFilter)/254.0)*(logMax-logMin)
			highPassFilter = float32(math.Pow(10, float64(logFreq)))
		}

		effectComb := float32(0) // Default
		if rawEffectComb != -1 {
			effectComb = float32(rawEffectComb) / 254.0
		}

		effectReverb := float32(0) // Default
		if rawEffectReverb != -1 {
			effectReverb = float32(rawEffectReverb) / 254.0
		}

		// Extract MIDI CC values (only used when in MIDI mode)
		var midiCC [9]int
		midiCC[0] = rowData[types.ColMidiCC0]
		midiCC[1] = rowData[types.ColMidiCC1]
		midiCC[2] = rowData[types.ColMidiCC2]
		midiCC[3] = rowData[types.ColMidiCC3]
		midiCC[4] = rowData[types.ColMidiCC4]
		midiCC[5] = rowData[types.ColMidiCC5]
		midiCC[6] = rowData[types.ColMidiCC6]
		midiCC[7] = rowData[types.ColMidiCC7]
		midiCC[8] = rowData[types.ColMidiCC8]

		instrumentParams := model.NewInstrumentOSCParams(
			int32(trackId),
			velocity,
			rawChord,
			rawChordAdd,
			rawChordTrans,
			effectiveGate,
			deltaTimeSeconds,
			attack,
			decay,
			sustain,
			release,
			pan,
			lowPassFilter,
			highPassFilter,
			effectComb,
			effectReverb,
			rawArpeggio,
			rawMidi,
			rawSoundMaker,
			rawEffectDucking,
			midiCC,
		)
		// Generate chord notes and apply modulation according to user specification
		midiNotes := types.GetChordNotes(rowData[types.ColNote], types.ChordType(rawChord), types.ChordAddition(rawChordAdd), types.ChordTransposition(rawChordTrans))
		instrumentParams.Notes = make([]float32, len(midiNotes))

		// Apply modulation to notes according to the new logic for instrument view:
		// - If there is an arpeggio: apply modulation to each note in the arpeggio INCLUDING the root note
		// - If it is not an arpeggio and not a chord: apply modulation to the main note
		// - If it is a chord but no arpeggio: apply modulation to all notes

		hasArpeggio := rawArpeggio != -1
		hasChord := rawChord != int(types.ChordNone)
		hasModulation := rawModulate != -1 && rawModulate >= 0 && rawModulate < 255

		if hasModulation {
			modulateSettings := (*GetModulateSettingsForTrack(m, trackId))[rawModulate]

			// Get track-specific RNG for modulation
			var trackRng *rand.Rand
			if trackId >= 0 && trackId < 8 {
				trackRng = m.ModulateRngs[trackId]
			} else {
				trackRng = rand.New(rand.NewSource(time.Now().UnixNano()))
			}

			incrementCounter := m.IncrementCounters[trackId][phrase][row]

			for i, note := range midiNotes {
				// Apply increment before other modulation operations if counter > -1
				noteWithIncrement := modulation.ApplyIncrement(note, incrementCounter, modulateSettings.Increment, modulateSettings.Wrap)

				modulatedNote := modulation.ApplyModulation(noteWithIncrement, modulation.ModulateSettings{
					Seed:        modulateSettings.Seed,
					IRandom:     modulateSettings.IRandom,
					Sub:         modulateSettings.Sub,
					Add:         modulateSettings.Add,
					Increment:   modulateSettings.Increment,
					Wrap:        modulateSettings.Wrap,
					ScaleRoot:   modulateSettings.ScaleRoot,
					Scale:       modulateSettings.Scale,
					Probability: modulateSettings.Probability,
				}, trackRng)
				instrumentParams.Notes[i] = float32(modulatedNote)
				log.Printf("Applied modulation to instrument note %d: %d -> %d (increment=%d, hasArpeggio=%v, hasChord=%v)", i, note, modulatedNote, modulateSettings.Increment, hasArpeggio, hasChord)
			}
		} else {
			// No modulation - copy notes as-is
			for i, note := range midiNotes {
				instrumentParams.Notes[i] = float32(note)
			}
		}
		// Set update flag for instrument params if this is an update
		if shouldUpdate {
			instrumentParams.Update = 1
		}
		m.SendOSCInstrumentMessageWithArpeggio(instrumentParams)
	} else {
		// For sampler tracks, emit full sampler message

		m.SendOSCSamplerMessage(oscParams)
	}
}

// isInstrumentTrack determines if the given track should use instrument OSC messages
// Uses the track type from mixer settings (false = Instrument, true = Sampler)
func isInstrumentTrack(m *model.Model, trackId int) bool {
	if trackId >= 0 && trackId < 8 {
		return !m.TrackTypes[trackId] // false = Instrument, true = Sampler
	}
	return false // Invalid track defaults to Sampler
}

// rowDurationMicroseconds returns the per-row duration in microseconds.
// DT is read *row-locally* and never inherited.
// New behavior:
// - DT == -1 (--) -> behaves like DT == 1 (1 tick duration)
// - DT == 0       -> skip row (duration irrelevant, handled by shouldEmitRow)
// - DT > 0        -> hold for DT number of ticks (baseUs * DT)
func rowDurationMicroseconds(m *model.Model) float64 {
	// Guard against invalid BPM/PPQ
	if m.BPM <= 0 || m.PPQ <= 0 {
		// Fallback to a sane default: 120 BPM, PPQ=2  => 250ms (250000us) per row
		return 250000.0
	}

	beatsPerSecond := float64(m.BPM) / 60.0
	ticksPerSecond := beatsPerSecond * float64(m.PPQ)
	baseUs := 1000000.0 / ticksPerSecond

	p := m.PlaybackPhrase
	r := m.PlaybackRow
	// Bounds checks (255 x 255 grid)
	if p < 0 || p >= 255 || r < 0 || r >= 255 {
		return baseUs
	}

	// Get the correct phrases data based on current track type
	phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
	dtRaw := (*phrasesData)[p][r][types.ColDeltaTime] // row-local DT
	if dtRaw == -1 {
		// -- behaves like 01 (1 tick)
		return baseUs
	} else if dtRaw == 0 {
		// 00 means skip row, but we still return base duration for timing consistency
		return baseUs
	} else {
		// DT > 0: hold for DT number of ticks
		return baseUs * float64(dtRaw)
	}
}

// calculateDeltaTimeSeconds calculates the DT value in seconds for a specific phrase/row
// This is the time per row (based on BPM/PPQ) multiplied by the DT value
func calculateDeltaTimeSeconds(m *model.Model, phrase, row, trackId int) float32 {
	// Guard against invalid BPM/PPQ
	if m.BPM <= 0 || m.PPQ <= 0 {
		// Fallback to a sane default: 120 BPM, PPQ=2  => 0.25s per row
		return 0.25
	}

	// Calculate base time per row (tick) in seconds
	beatsPerSecond := float64(m.BPM) / 60.0
	ticksPerSecond := beatsPerSecond * float64(m.PPQ)
	baseSecondsPerTick := 1.0 / ticksPerSecond

	// Bounds checks (255 x 255 grid)
	if phrase < 0 || phrase >= 255 || row < 0 || row >= 255 {
		return float32(baseSecondsPerTick)
	}

	// Get the correct phrases data based on specified track type
	phrasesData := GetPhrasesDataForTrack(m, trackId)
	dtRaw := (*phrasesData)[phrase][row][types.ColDeltaTime] // row-local DT

	if dtRaw == -1 {
		// -- behaves like 01 (1 tick)
		return float32(baseSecondsPerTick)
	} else if dtRaw == 0 {
		// 00 means skip row (0 ticks)
		return 0.0
	} else {
		// DT > 0: hold for DT number of ticks
		return float32(baseSecondsPerTick * float64(dtRaw))
	}
}

// shouldEmitRow enforces:
// - NN must be present (not -1)
// - P (playback flag) must be 1
// - DT == 00 => do NOT emit (for all track types)
// DT is read row-locally and never inherited.
func shouldEmitRow(m *model.Model) bool {
	return shouldEmitRowForTrack(m, m.CurrentTrack)
}

func shouldEmitRowForTrack(m *model.Model, trackId int) bool {
	p := m.PlaybackPhrase
	r := m.PlaybackRow
	if p < 0 || p >= 255 || r < 0 || r >= 255 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, trackId)
	nn := (*phrasesData)[p][r][types.ColNote]         // may still be inherited elsewhere in your code
	dtRaw := (*phrasesData)[p][r][types.ColDeltaTime] // unified playback control

	// Unified DT-based playback: DT > 0 means play, DT <= 0 means don't play
	if !IsRowPlayable(dtRaw) {
		return false
	}

	if nn == -1 {
		return false
	}

	return true
}

func shouldEmitRowForTrackAtPosition(m *model.Model, phrase, row, trackId int) bool {
	if phrase < 0 || phrase >= 255 || row < 0 || row >= 255 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, trackId)
	nn := (*phrasesData)[phrase][row][types.ColNote]
	dtRaw := (*phrasesData)[phrase][row][types.ColDeltaTime]

	// Unified DT-based playback: DT > 0 means play, DT <= 0 means don't play
	if !IsRowPlayable(dtRaw) {
		return false
	}

	// For instrument tracks, check if there are any CC values set
	// If so, allow emission even without a note
	if isInstrumentTrack(m, trackId) {
		// Check if any CC values are set
		ccCols := []types.PhraseColumn{
			types.ColMidiCC0, types.ColMidiCC1, types.ColMidiCC2,
			types.ColMidiCC3, types.ColMidiCC4, types.ColMidiCC5,
			types.ColMidiCC6, types.ColMidiCC7, types.ColMidiCC8,
		}
		for _, ccCol := range ccCols {
			if (*phrasesData)[phrase][row][ccCol] != -1 {
				return true // Allow emission if any CC value is set
			}
		}
	}

	if nn == -1 {
		return false
	}
	return true
}

func goToPhrase(m *model.Model, phrase int) {
	if phrase < 0 || phrase >= 255 {
		return
	}
	m.CurrentPhrase = phrase
	// Reset to beginning of phrase if it hasn't been visited/edited
	if m.LastEditRow == -1 || len(m.PhrasesData[phrase]) == 0 {
		m.CurrentRow = 0
		m.CurrentCol = 1
		m.ScrollOffset = 0
	} else {
		// keep last edit row if desired, but usually we just reset
		m.CurrentRow = 0
		m.CurrentCol = 1
		m.ScrollOffset = 0
	}
	storage.AutoSave(m)
}

// ModifySongValue modifies chain ID values in song view
func ModifySongValue(m *model.Model, delta int) {
	track := m.CurrentCol
	row := m.CurrentRow

	// Bounds check
	if track < 0 || track >= 8 || row < 0 || row >= 16 {
		return
	}

	currentValue := m.SongData[track][row]
	var newValue int

	if currentValue == -1 {
		// First edit on an empty cell: initialize to 00 and DO NOT apply delta
		newValue = 0
	} else {
		newValue = currentValue + delta
	}

	// Clamp to valid chain range (0-254, which is 00-FE in hex)
	if newValue < 0 {
		newValue = 0
	} else if newValue > 254 {
		newValue = 254
	}

	m.SongData[track][row] = newValue
	log.Printf("Modified song track %d row %d: %d -> %d (delta: %d)", track, row, currentValue, newValue, delta)

	storage.AutoSave(m)
}

// PlaybackConfig represents the configuration for starting playback
type PlaybackConfig struct {
	Mode          types.ViewMode
	UseCurrentRow bool // false means start from top/first non-empty
	Chain         int  // for chain playback
	Phrase        int  // for phrase playback
	Row           int  // starting row
}

// stopPlayback provides common logic for stopping playback
func stopPlayback(m *model.Model) {
	m.IsPlaying = false

	// Stop recording if active
	if m.RecordingActive {
		stopRecording(m)
	}

	// Clear file browser playback state when stopping tracker playback
	if m.CurrentlyPlayingFile != "" {
		m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
		m.CurrentlyPlayingFile = ""
		log.Printf("Stopped file browser playback when stopping tracker playback")
	}

	m.SendStopOSC()
	log.Printf("Playback stopped")
}

// startPlaybackWithConfig provides common logic for starting playback
func startPlaybackWithConfig(m *model.Model, config PlaybackConfig) tea.Cmd {
	m.IsPlaying = true
	m.PlaybackMode = config.Mode

	// Initialize increment counters to -1 for all tracks/phrases/rows
	for track := 0; track < 8; track++ {
		for phrase := 0; phrase < 255; phrase++ {
			for row := 0; row < 255; row++ {
				m.IncrementCounters[track][phrase][row] = -1
			}
		}
	}
	log.Printf("DEBUG_INCREMENT: Initialized all increment counters to -1 for playback start")

	if config.Mode == types.SongView {
		// Song playback mode - reset single-track playback variables and initialize all tracks with data
		m.PlaybackPhrase = -1
		m.PlaybackRow = -1
		m.PlaybackChain = -1
		m.PlaybackChainRow = -1

		startRow := 0
		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			startRow = config.Row
		}
		log.Printf("Song playback starting from row %02X", startRow)
		// Debug: show song data for first few rows
		for r := 0; r < 4 && r < 16; r++ {
			log.Printf("Song row %02X data: %v", r, [8]int{
				m.SongData[0][r], m.SongData[1][r], m.SongData[2][r], m.SongData[3][r],
				m.SongData[4][r], m.SongData[5][r], m.SongData[6][r], m.SongData[7][r],
			})
		}

		for track := 0; track < 8; track++ {
			chainID := m.SongData[track][startRow]
			log.Printf("Song track %d at row %02X: chainID = %d", track, startRow, chainID)
			if chainID == -1 {
				// No chain at this position
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d: no chain data, skipping", track)
				continue
			}

			// Check if chain has valid phrase data (find first phrase in chain)
			firstPhraseID := -1
			firstChainRow := -1
			chainsData := m.GetChainsDataForTrack(track)
			for chainRow := 0; chainRow < 16; chainRow++ {
				if (*chainsData)[chainID][chainRow] != -1 {
					firstPhraseID = (*chainsData)[chainID][chainRow]
					firstChainRow = chainRow
					log.Printf("Song track %d: found phrase %d in chain %d at row %d", track, firstPhraseID, chainID, chainRow)
					break
				}
			}

			if firstPhraseID != -1 {
				m.SongPlaybackActive[track] = true
				m.SongPlaybackRow[track] = startRow
				m.SongPlaybackChain[track] = chainID
				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)

				// Initialize ticks for this track
				m.LoadTicksLeftForTrack(track)

				// Emit initial row for this track
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d started at row %02X, chain %02X (chain row %d), phrase %02X with %d ticks", track, startRow, chainID, firstChainRow, firstPhraseID, m.SongPlaybackTicksLeft[track])
			} else {
				// Chain exists but has no phrases
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d skipped at row %02X (chain %02X has no phrases)", track, startRow, chainID)
			}
		}

		// Count how many tracks will be active
		activeTracks := 0
		for t := 0; t < 8; t++ {
			if m.SongPlaybackActive[t] {
				activeTracks++
			}
		}
		log.Printf("Song playback started from row %02X with %d active tracks", startRow, activeTracks)

		// If no tracks are active, make track 0 active with chain 0 as fallback
		if activeTracks == 0 {
			log.Printf("No active tracks found, activating track 0 with chain 0 as fallback")
			m.SongPlaybackActive[0] = true
			m.SongPlaybackRow[0] = startRow
			m.SongPlaybackChain[0] = 0
			m.SongPlaybackChainRow[0] = 0
			m.SongPlaybackPhrase[0] = 0 // Use phrase 0 as fallback
			m.SongPlaybackRowInPhrase[0] = FindFirstNonEmptyRowInPhraseForTrack(m, 0, 0)
			// Initialize ticks for fallback track 0
			m.LoadTicksLeftForTrack(0)
			EmitRowDataFor(m, 0, m.SongPlaybackRowInPhrase[0], 0)
			log.Printf("Song track 0 fallback started at phrase 0, row %d with %d ticks", m.SongPlaybackRowInPhrase[0], m.SongPlaybackTicksLeft[0])
		}
	} else if config.Mode == types.ChainView {
		// Chain playback mode - find appropriate starting phrase
		m.PlaybackChain = config.Chain
		m.PlaybackPhrase = -1

		chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			// Start from specified chain row
			if (*chainsData)[config.Chain][config.Row] != -1 {
				m.PlaybackChainRow = config.Row
				m.PlaybackPhrase = (*chainsData)[config.Chain][config.Row]
			}
		}

		// If no phrase found yet, find first non-empty phrase slot in this chain
		if m.PlaybackPhrase == -1 {
			for row := 0; row < 16; row++ {
				if (*chainsData)[config.Chain][row] != -1 {
					m.PlaybackPhrase = (*chainsData)[config.Chain][row]
					m.PlaybackChainRow = row
					break
				}
			}
		}

		if m.PlaybackPhrase == -1 {
			// This chain has no phrases, find a different chain
			m.PlaybackChain = FindFirstNonEmptyChain(m)
			log.Printf("Chain playback fallback: switching to chain %d", m.PlaybackChain)
			m.PlaybackChainRow = 0
			for row := 0; row < 16; row++ {
				if (*chainsData)[m.PlaybackChain][row] != -1 {
					m.PlaybackPhrase = (*chainsData)[m.PlaybackChain][row]
					m.PlaybackChainRow = row
					log.Printf("Chain playback fallback: found phrase %d at chain row %d", m.PlaybackPhrase, row)
					break
				}
			}
		}

		// If still no valid phrase found, just start with phrase 0 as fallback
		if m.PlaybackPhrase == -1 {
			log.Printf("Chain playback warning - no valid phrases found, using phrase 0 as fallback (Chain: %d, ChainRow: %d)", m.PlaybackChain, m.PlaybackChainRow)
			// Let's log the chain data for debugging
			log.Printf("Chain %d contents: %v", m.PlaybackChain, (*chainsData)[m.PlaybackChain])
			m.PlaybackPhrase = 0
			m.PlaybackChainRow = 0
		}

		if config.UseCurrentRow && config.Row >= 0 {
			m.PlaybackRow = config.Row
		} else {
			m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
		}

		DebugLogRowEmission(m)
		log.Printf("Chain playback started at chain %d, chain row %d, phrase %d, phrase row %d", m.PlaybackChain, m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
	} else {
		// Phrase playback mode
		m.PlaybackPhrase = config.Phrase
		m.PlaybackChain = -1

		trackType := "Sampler"
		if m.CurrentTrack >= 0 && m.CurrentTrack < 8 && !m.TrackTypes[m.CurrentTrack] {
			trackType = "Instrument"
		}
		log.Printf("DEBUG: Phrase playback starting - CurrentTrack=%d (%s), Phrase=%d", m.CurrentTrack, trackType, m.PlaybackPhrase)

		if config.UseCurrentRow && config.Row >= 0 {
			m.PlaybackRow = config.Row
		} else {
			m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
			log.Printf("DEBUG: FindFirstNonEmptyRowInPhrase returned row %d for track %d", m.PlaybackRow, m.CurrentTrack)
		}

		DebugLogRowEmission(m)
		log.Printf("Phrase playback started at phrase %d, row %d", m.PlaybackPhrase, m.PlaybackRow)
	}

	// Start recording if enabled
	if m.RecordingEnabled && !m.RecordingActive {
		// Determine context based on playback mode
		fromSongView := (config.Mode == types.SongView)
		fromCtrlSpace := false // This is not from Ctrl+Space, but from regular playback start
		startRecordingWithContext(m, fromSongView, fromCtrlSpace)
	}

	return Tick(m)
}

// startPlaybackWithConfigFromCtrlSpace is specialized for Ctrl+Space recording context
func startPlaybackWithConfigFromCtrlSpace(m *model.Model, config PlaybackConfig) tea.Cmd {
	// Initialize increment counters to -1 for all tracks/phrases/rows
	for track := 0; track < 8; track++ {
		for phrase := 0; phrase < 255; phrase++ {
			for row := 0; row < 255; row++ {
				m.IncrementCounters[track][phrase][row] = -1
			}
		}
	}
	log.Printf("DEBUG_INCREMENT: Initialized all increment counters to -1 for Ctrl+Space playback start")
	m.IsPlaying = true
	m.PlaybackMode = config.Mode

	if config.Mode == types.SongView {
		// Song playback mode - reset single-track playback variables and initialize all tracks with data
		m.PlaybackPhrase = -1
		m.PlaybackRow = -1
		m.PlaybackChain = -1
		m.PlaybackChainRow = -1

		startRow := 0
		if config.UseCurrentRow && config.Row >= 0 && config.Row < 16 {
			startRow = config.Row
		}
		log.Printf("Song playback starting from row %02X (Ctrl+Space)", startRow)

		for track := 0; track < 8; track++ {
			chainID := m.SongData[track][startRow]
			log.Printf("Song track %d at row %02X: chainID = %d", track, startRow, chainID)
			if chainID == -1 {
				// No chain at this position
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d: no chain data, skipping", track)
				continue
			}

			// Check if chain has valid phrase data (find first phrase in chain)
			firstPhraseID := -1
			firstChainRow := -1
			chainsData := m.GetChainsDataForTrack(track)
			for chainRow := 0; chainRow < 16; chainRow++ {
				if (*chainsData)[chainID][chainRow] != -1 {
					firstPhraseID = (*chainsData)[chainID][chainRow]
					firstChainRow = chainRow
					log.Printf("Song track %d: found phrase %d in chain %d at row %d", track, firstPhraseID, chainID, chainRow)
					break
				}
			}

			if firstPhraseID != -1 {
				m.SongPlaybackActive[track] = true
				m.SongPlaybackRow[track] = startRow
				m.SongPlaybackPhrase[track] = firstPhraseID
				m.SongPlaybackChain[track] = chainID
				m.SongPlaybackChainRow[track] = firstChainRow
				m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)
				m.LoadTicksLeftForTrack(track)

				// Emit the initial row immediately
				EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
				log.Printf("Song track %d initialized: phrase %d, row %d, ticks %d", track, firstPhraseID, m.SongPlaybackRowInPhrase[track], m.SongPlaybackTicksLeft[track])
			} else {
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d: chain %d has no phrases, skipping", track, chainID)
			}
		}

		log.Printf("Song playback initialized (Ctrl+Space)")
	} else {
		// Chain/Phrase playback modes - same logic as regular playback
		if config.Mode == types.ChainView {
			chainsData := m.GetCurrentChainsData()
			m.PlaybackChain = config.Chain
			m.PlaybackChainRow = 0

			if config.UseCurrentRow && config.Row >= 0 {
				m.PlaybackChainRow = config.Row
				m.PlaybackPhrase = (*chainsData)[config.Chain][config.Row]
			}

			// If no phrase found yet, find first non-empty phrase slot in this chain
			if m.PlaybackPhrase == -1 {
				for row := 0; row < 16; row++ {
					if (*chainsData)[config.Chain][row] != -1 {
						m.PlaybackPhrase = (*chainsData)[config.Chain][row]
						m.PlaybackChainRow = row
						break
					}
				}
			}

			if config.UseCurrentRow && config.Row >= 0 {
				m.PlaybackRow = config.Row
			} else {
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
			}

			log.Printf("Chain playback started (Ctrl+Space): chain %d, phrase %d, row %d", m.PlaybackChain, m.PlaybackPhrase, m.PlaybackRow)
		} else {
			// Phrase playback mode
			m.PlaybackPhrase = config.Phrase
			m.PlaybackChain = -1

			if config.UseCurrentRow && config.Row >= 0 {
				m.PlaybackRow = config.Row
			} else {
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
			}

			log.Printf("Phrase playback started (Ctrl+Space): phrase %d, row %d", m.PlaybackPhrase, m.PlaybackRow)
		}
	}

	// Start recording if enabled (with Ctrl+Space context)
	if m.RecordingEnabled && !m.RecordingActive {
		fromSongView := (config.Mode == types.SongView)
		fromCtrlSpace := true // This IS from Ctrl+Space
		startRecordingWithContext(m, fromSongView, fromCtrlSpace)
	}

	return Tick(m)
}

// togglePlaybackWithConfig provides common toggle logic
func togglePlaybackWithConfig(m *model.Model, config PlaybackConfig) tea.Cmd {
	if m.IsPlaying {
		stopPlayback(m)
		return nil
	}
	return startPlaybackWithConfig(m, config)
}

// togglePlaybackWithConfigFromCtrlSpace provides toggle logic for Ctrl+Space
func togglePlaybackWithConfigFromCtrlSpace(m *model.Model, config PlaybackConfig) tea.Cmd {
	if m.IsPlaying {
		stopPlayback(m)
		return nil
	}
	return startPlaybackWithConfigFromCtrlSpace(m, config)
}

// Deep copy functionality for Ctrl+B
func DeepCopyToClipboard(m *model.Model) {
	if m.ViewMode == types.SongView {
		DeepCopyChainToClipboard(m)
	} else if m.ViewMode == types.ChainView {
		DeepCopyPhraseToClipboard(m)
	} else if m.ViewMode == types.PhraseView {
		// Check if we're in specific columns that support individual deep copy
		columnMapping := m.GetColumnMapping(m.CurrentCol)
		if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColRetrigger) {
			DeepCopyRetriggerToClipboard(m)
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColTimestretch) {
			DeepCopyTimestrechToClipboard(m)
		} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColArpeggio) {
			DeepCopyArpeggioToClipboard(m)
		} else {
			DeepCopyCurrentPhraseToClipboard(m)
		}
	} else if m.ViewMode == types.RetriggerView {
		DeepCopyRetriggerToClipboard(m)
	} else if m.ViewMode == types.TimestrechView {
		DeepCopyTimestrechToClipboard(m)
	} else {
		log.Printf("Deep copy not supported in this view")
	}
}

func DeepCopyChainToClipboard(m *model.Model) {
	sourceChainID := m.SongData[m.CurrentCol][m.CurrentRow]
	if sourceChainID == -1 {
		log.Printf("Cannot deep copy: no chain at track %d row %02X (cell is empty)", m.CurrentCol, m.CurrentRow)
		return
	}

	// Find next unused chain
	destChainID := FindNextUnusedChain(m, sourceChainID)
	if destChainID == -1 {
		log.Printf("Cannot deep copy: no unused chains available")
		return
	}

	// Copy all 16 phrase slots from source chain to destination chain
	for row := 0; row < 16; row++ {
		m.ChainsData[destChainID][row] = m.ChainsData[sourceChainID][row]
	}

	// Put the new chain ID in clipboard
	clipboard := types.ClipboardData{
		Value:           destChainID,
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: -1,
		HighlightView:   types.SongView,
	}
	m.Clipboard = clipboard

	log.Printf("Deep copied chain %02X to chain %02X", sourceChainID, destChainID)
}

func DeepCopyPhraseToClipboard(m *model.Model) {
	// Bounds check for current row
	if m.CurrentRow < 0 || m.CurrentRow >= 16 {
		log.Printf("Cannot deep copy: invalid row %d", m.CurrentRow)
		return
	}

	// In Chain view, get the phrase from the current row
	chainsData := m.GetCurrentChainsData()
	sourcePhraseID := (*chainsData)[m.CurrentChain][m.CurrentRow]
	if sourcePhraseID == -1 {
		log.Printf("Cannot deep copy: no phrase at chain %02X row %02X (cell is empty)", m.CurrentChain, m.CurrentRow)
		return
	}

	// Additional bounds check for source phrase ID
	if sourcePhraseID < 0 || sourcePhraseID >= 255 {
		log.Printf("Cannot deep copy: invalid source phrase ID %d", sourcePhraseID)
		return
	}

	// Find next unused phrase in the same pool (Sampler/Instrument)
	destPhraseID := FindNextUnusedPhrase(m, sourcePhraseID)
	if destPhraseID == -1 {
		log.Printf("Cannot deep copy: no unused phrases available")
		return
	}

	// Additional bounds check for destination phrase ID
	if destPhraseID < 0 || destPhraseID >= 255 {
		log.Printf("Cannot deep copy: invalid destination phrase ID %d", destPhraseID)
		return
	}

	// Get the appropriate phrases data for the current pool
	phrasesData := m.GetCurrentPhrasesData()

	// Copy all 255 rows of phrase data from source to destination
	for row := 0; row < 255; row++ {
		for col := 0; col < int(types.ColCount); col++ {
			(*phrasesData)[destPhraseID][row][col] = (*phrasesData)[sourcePhraseID][row][col]
		}
	}

	// For sampler phrases, also copy any associated file references
	if m.GetPhraseViewType() == types.SamplerPhraseView {
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if phrasesFiles != nil {
			// Copy file references by duplicating entries in the files array
			for row := 0; row < 255; row++ {
				fileIndex := (*phrasesData)[sourcePhraseID][row][types.ColFilename]
				if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
					// Add the same file to the files array and update the destination phrase
					filename := (*phrasesFiles)[fileIndex]
					newFileIndex := len(*phrasesFiles)
					*phrasesFiles = append(*phrasesFiles, filename)
					(*phrasesData)[destPhraseID][row][types.ColFilename] = newFileIndex
				}
			}
		}
	}

	// Copy and remap arpeggio settings referenced in the phrase
	arpeggioMapping := make(map[int]int) // Map from source arpeggio index to destination arpeggio index
	for row := 0; row < 255; row++ {
		arpeggioIndex := (*phrasesData)[destPhraseID][row][types.ColArpeggio]
		if arpeggioIndex >= 0 && arpeggioIndex < 255 {
			// Check if we already have a mapping for this arpeggio index
			if newArpeggioIndex, exists := arpeggioMapping[arpeggioIndex]; exists {
				// Use existing mapping
				(*phrasesData)[destPhraseID][row][types.ColArpeggio] = newArpeggioIndex
			} else {
				// Find next unused arpeggio slot and copy the settings
				newArpeggioIndex := FindNextUnusedArpeggio(m, arpeggioIndex)
				if newArpeggioIndex != -1 {
					// Copy the arpeggio settings
					m.ArpeggioSettings[newArpeggioIndex] = m.ArpeggioSettings[arpeggioIndex]
					// Store mapping and update the phrase data
					arpeggioMapping[arpeggioIndex] = newArpeggioIndex
					(*phrasesData)[destPhraseID][row][types.ColArpeggio] = newArpeggioIndex
					log.Printf("Deep copied arpeggio settings %02X to %02X", arpeggioIndex, newArpeggioIndex)
				} else {
					// No unused arpeggio slots available, keep original reference
					log.Printf("Warning: No unused arpeggio slots available for copying arpeggio %02X", arpeggioIndex)
				}
			}
		}
	}

	// Put the new phrase ID in clipboard
	clipboard := types.ClipboardData{
		Value:           destPhraseID,
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: -1,
		HighlightView:   types.ChainView,
	}
	m.Clipboard = clipboard

	log.Printf("Deep copied phrase %02X to phrase %02X", sourcePhraseID, destPhraseID)
}

func DeepCopyCurrentPhraseToClipboard(m *model.Model) {
	sourcePhraseID := m.CurrentPhrase
	if sourcePhraseID < 0 || sourcePhraseID >= 255 {
		log.Printf("Cannot deep copy: invalid current phrase ID %d", sourcePhraseID)
		return
	}

	// Find next unused phrase in the same pool (Sampler/Instrument)
	destPhraseID := FindNextUnusedPhrase(m, sourcePhraseID)
	if destPhraseID == -1 {
		log.Printf("Cannot deep copy: no unused phrases available")
		return
	}

	// Get the appropriate phrases data for the current pool
	phrasesData := m.GetCurrentPhrasesData()

	// Copy all 255 rows of phrase data from source to destination
	for row := 0; row < 255; row++ {
		for col := 0; col < int(types.ColCount); col++ {
			(*phrasesData)[destPhraseID][row][col] = (*phrasesData)[sourcePhraseID][row][col]
		}
	}

	// For sampler phrases, also copy any associated file references
	if m.GetPhraseViewType() == types.SamplerPhraseView {
		phrasesFiles := m.GetCurrentPhrasesFiles()
		if phrasesFiles != nil {
			// Copy file references by duplicating entries in the files array
			for row := 0; row < 255; row++ {
				fileIndex := (*phrasesData)[sourcePhraseID][row][types.ColFilename]
				if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
					// Add the same file to the files array and update the destination phrase
					filename := (*phrasesFiles)[fileIndex]
					newFileIndex := len(*phrasesFiles)
					*phrasesFiles = append(*phrasesFiles, filename)
					(*phrasesData)[destPhraseID][row][types.ColFilename] = newFileIndex
				}
			}
		}
	}

	// Copy and remap arpeggio settings referenced in the phrase
	arpeggioMapping := make(map[int]int) // Map from source arpeggio index to destination arpeggio index
	for row := 0; row < 255; row++ {
		arpeggioIndex := (*phrasesData)[destPhraseID][row][types.ColArpeggio]
		if arpeggioIndex >= 0 && arpeggioIndex < 255 {
			// Check if we already have a mapping for this arpeggio index
			if newArpeggioIndex, exists := arpeggioMapping[arpeggioIndex]; exists {
				// Use existing mapping
				(*phrasesData)[destPhraseID][row][types.ColArpeggio] = newArpeggioIndex
			} else {
				// Find next unused arpeggio slot and copy the settings
				newArpeggioIndex := FindNextUnusedArpeggio(m, arpeggioIndex)
				if newArpeggioIndex != -1 {
					// Copy the arpeggio settings
					m.ArpeggioSettings[newArpeggioIndex] = m.ArpeggioSettings[arpeggioIndex]
					// Store mapping and update the phrase data
					arpeggioMapping[arpeggioIndex] = newArpeggioIndex
					(*phrasesData)[destPhraseID][row][types.ColArpeggio] = newArpeggioIndex
					log.Printf("Deep copied arpeggio settings %02X to %02X", arpeggioIndex, newArpeggioIndex)
				} else {
					// No unused arpeggio slots available, keep original reference
					log.Printf("Warning: No unused arpeggio slots available for copying arpeggio %02X", arpeggioIndex)
				}
			}
		}
	}

	// Put the new phrase ID in clipboard for pasting
	clipboard := types.ClipboardData{
		Value:           destPhraseID,
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: destPhraseID,
		HighlightView:   types.PhraseView,
	}
	m.Clipboard = clipboard

	log.Printf("Deep copied phrase %02X to phrase %02X", sourcePhraseID, destPhraseID)
}

func DeepCopyRetriggerToClipboard(m *model.Model) {
	var sourceRetriggerIndex int

	if m.ViewMode == types.RetriggerView {
		// In RetriggerView, use the currently editing retrigger index
		sourceRetriggerIndex = m.RetriggerEditingIndex
	} else {
		// In PhraseView, get the retrigger index from the current cell
		phrasesData := m.GetCurrentPhrasesData()
		sourceRetriggerIndex = (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColRetrigger]

		if sourceRetriggerIndex == -1 {
			log.Printf("Cannot deep copy retrigger: no retrigger set in current cell")
			return
		}
	}

	if sourceRetriggerIndex < 0 || sourceRetriggerIndex >= 255 {
		log.Printf("Cannot deep copy retrigger: invalid retrigger index %d", sourceRetriggerIndex)
		return
	}

	// Put the ORIGINAL retrigger index in clipboard, but mark it for deep copy on paste
	clipboard := types.ClipboardData{
		Value:           sourceRetriggerIndex, // Keep original reference
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: m.CurrentPhrase,
		HighlightView:   m.ViewMode,
		IsFreshDeepCopy: true, // Mark for deep copy on paste
	}
	m.Clipboard = clipboard

	log.Printf("Marked retrigger %02X for deep copy on paste", sourceRetriggerIndex)
}

func DeepCopyTimestrechToClipboard(m *model.Model) {
	var sourceTimestrechIndex int

	if m.ViewMode == types.TimestrechView {
		// In TimestrechView, use the currently editing timestrech index
		sourceTimestrechIndex = m.TimestrechEditingIndex
	} else {
		// In PhraseView, get the timestrech index from the current cell
		phrasesData := m.GetCurrentPhrasesData()
		sourceTimestrechIndex = (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColTimestretch]

		if sourceTimestrechIndex == -1 {
			log.Printf("Cannot deep copy timestrech: no timestrech set in current cell")
			return
		}
	}

	if sourceTimestrechIndex < 0 || sourceTimestrechIndex >= 255 {
		log.Printf("Cannot deep copy timestrech: invalid timestrech index %d", sourceTimestrechIndex)
		return
	}

	// Put the ORIGINAL timestrech index in clipboard, but mark it for deep copy on paste
	clipboard := types.ClipboardData{
		Value:           sourceTimestrechIndex, // Keep original reference
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: m.CurrentPhrase,
		HighlightView:   m.ViewMode,
		IsFreshDeepCopy: true, // Mark for deep copy on paste
	}
	m.Clipboard = clipboard

	log.Printf("Marked timestrech %02X for deep copy on paste", sourceTimestrechIndex)
}

func DeepCopyArpeggioToClipboard(m *model.Model) {
	// Get the arpeggio index from the current cell
	phrasesData := m.GetCurrentPhrasesData()
	sourceArpeggioIndex := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColArpeggio]

	if sourceArpeggioIndex == -1 {
		log.Printf("Cannot deep copy arpeggio: no arpeggio set in current cell")
		return
	}

	if sourceArpeggioIndex < 0 || sourceArpeggioIndex >= 255 {
		log.Printf("Cannot deep copy arpeggio: invalid arpeggio index %d", sourceArpeggioIndex)
		return
	}

	// Put the ORIGINAL arpeggio index in clipboard, but mark it for deep copy on paste
	clipboard := types.ClipboardData{
		Value:           sourceArpeggioIndex, // Keep original reference
		CellType:        types.HexCell,
		Mode:            types.CellMode,
		HasData:         true,
		HighlightRow:    m.CurrentRow,
		HighlightCol:    m.CurrentCol,
		HighlightPhrase: m.CurrentPhrase,
		HighlightView:   types.PhraseView,
		IsFreshDeepCopy: true, // Mark for deep copy on paste
	}
	m.Clipboard = clipboard

	log.Printf("Marked arpeggio %02X for deep copy on paste", sourceArpeggioIndex)
}

func FindNextUnusedChain(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		chainID := (startingFrom + offset) % 255
		if chainID >= 0 && chainID < 255 && IsChainUnused(m, chainID) {
			return chainID
		}
	}
	return -1 // No unused chain found
}

func FindNextUnusedPhrase(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		phraseID := (startingFrom + offset) % 255

		if phraseID >= 0 && phraseID < 255 && IsPhraseUnused(m, phraseID) {
			return phraseID
		}
	}
	return -1 // No unused phrase found
}

func IsChainUnused(m *model.Model, chainID int) bool {
	// Bounds check first
	if chainID < 0 || chainID >= 255 {
		return false
	}

	// Check if chain is referenced in song data
	for track := 0; track < 8; track++ {
		for row := 0; row < 16; row++ {
			if m.SongData[track][row] == chainID {
				return false
			}
		}
	}

	// Check if chain has any phrase data
	for row := 0; row < 16; row++ {
		if m.ChainsData[chainID][row] != -1 {
			return false
		}
	}

	return true
}

func IsPhraseUnused(m *model.Model, phraseID int) bool {
	// Bounds check first
	if phraseID < 0 || phraseID >= 255 {
		return false
	}

	// Check if phrase is referenced in any chain
	for chain := 0; chain < 255; chain++ { // ChainsData has 255 elements (0-254)
		for row := 0; row < 16; row++ {
			if m.ChainsData[chain][row] == phraseID {
				return false
			}
		}
	}

	// Check if phrase has any playback-enabled rows
	phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
	for row := 0; row < 255; row++ {
		// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
		dtValue := (*phrasesData)[phraseID][row][types.ColDeltaTime]
		if IsRowPlayable(dtValue) {
			return false
		}
	}

	return true
}

func FindNextUnusedArpeggio(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		arpeggioID := (startingFrom + offset) % 255
		if arpeggioID >= 0 && arpeggioID < 255 && IsArpeggioUnused(m, arpeggioID) {
			return arpeggioID
		}
	}
	return -1 // No unused arpeggio found
}

func IsArpeggioUnused(m *model.Model, arpeggioID int) bool {
	// Bounds check first
	if arpeggioID < 0 || arpeggioID >= 255 {
		return false
	}

	// Check if arpeggio is referenced in any phrase data
	for phrase := 0; phrase < 255; phrase++ {
		for row := 0; row < 255; row++ {
			// Check both sampler and instrument phrases
			if m.SamplerPhrasesData[phrase][row][types.ColArpeggio] == arpeggioID ||
				m.InstrumentPhrasesData[phrase][row][types.ColArpeggio] == arpeggioID {
				return false
			}
		}
	}

	// Check if arpeggio has any non-default settings
	settings := m.ArpeggioSettings[arpeggioID]
	for _, row := range settings.Rows {
		if row.Direction != 0 || row.Count != -1 || row.Divisor != -1 {
			return false
		}
	}

	return true
}

// ModifyMixerSetLevel adjusts the set level for the currently selected track in mixer view
func ModifyMixerSetLevel(m *model.Model, delta float32) {
	// Bounds check (support tracks 0-8, including Input track at index 8)
	if m.CurrentMixerTrack < 0 || m.CurrentMixerTrack >= 9 {
		return
	}

	oldValue := m.TrackSetLevels[m.CurrentMixerTrack]
	newValue := oldValue + delta

	// Clamp to valid range (-96.0 to +32.0 dB)
	if newValue < -96.0 {
		newValue = -96.0
	} else if newValue > 32.0 {
		newValue = 32.0
	}

	m.TrackSetLevels[m.CurrentMixerTrack] = newValue
	if m.CurrentMixerTrack == 8 {
		log.Printf("Modified mixer Input track set level: %.2f -> %.2f (delta: %.2f)", oldValue, newValue, delta)
	} else {
		log.Printf("Modified mixer track %d set level: %.2f -> %.2f (delta: %.2f)", m.CurrentMixerTrack+1, oldValue, newValue, delta)
	}

	// Send OSC message for track set level
	m.SendOSCTrackSetLevelMessage(m.CurrentMixerTrack)

	storage.AutoSave(m)
}

// ToggleTrackType toggles the track type for the specified track (used in Song view)
func ToggleTrackType(m *model.Model, track int) {
	// Bounds check
	if track < 0 || track >= 8 {
		return
	}

	// Toggle the track type
	oldType := m.TrackTypes[track]
	m.TrackTypes[track] = !oldType

	var oldTypeStr, newTypeStr string
	if oldType {
		oldTypeStr = "SA"
	} else {
		oldTypeStr = "IN"
	}
	if m.TrackTypes[track] {
		newTypeStr = "SA"
	} else {
		newTypeStr = "IN"
	}

	log.Printf("Toggled track %d type: %s -> %s", track, oldTypeStr, newTypeStr)
	storage.AutoSave(m)
}

// FillSequential fills from the last null cell to the current cell in increments of 1
func FillSequential(m *model.Model) {
	if m.ViewMode == types.SongView {
		FillSequentialSong(m)
	} else if m.ViewMode == types.ChainView {
		FillSequentialChain(m)
	} else if m.ViewMode == types.PhraseView {
		FillSequentialPhrase(m)
	}
}

// FillSequentialSong fills chain IDs in song view
func FillSequentialSong(m *model.Model) {
	track := m.CurrentCol
	currentRow := m.CurrentRow

	// Find the last non-null value going upward
	var startValue int = 0
	var startRow int = 0

	for row := currentRow - 1; row >= 0; row-- {
		if m.SongData[track][row] != -1 {
			startValue = m.SongData[track][row] + 1
			startRow = row + 1
			break
		}
	}

	// Fill from startRow to currentRow
	for row := startRow; row <= currentRow; row++ {
		value := startValue + (row - startRow)
		if value > 254 {
			value = value % 255 // Wrap around (0-254)
		}
		m.SongData[track][row] = value
	}

	log.Printf("Filled song track %d from row %d to %d, starting with %d", track, startRow, currentRow, startValue)
}

// FillSequentialChain fills phrase IDs in chain view
func FillSequentialChain(m *model.Model) {
	currentRow := m.CurrentRow
	chainsData := m.GetCurrentChainsData()

	// Find where to start filling - look for the last non-empty row above current position
	var startValue int = 0
	var startRow int = 0

	for row := currentRow - 1; row >= 0; row-- {
		if (*chainsData)[m.CurrentChain][row] != -1 {
			startValue = (*chainsData)[m.CurrentChain][row] + 1
			startRow = row + 1
			break
		}
	}

	// Fill from startRow to currentRow with sequential values
	for row := startRow; row <= currentRow; row++ {
		value := startValue + (row - startRow)
		if value > 254 {
			value = value % 255 // Wrap around (0-254)
		}
		(*chainsData)[m.CurrentChain][row] = value
	}

	log.Printf("Filled chain %02X from row %d to %d, starting with %d", m.CurrentChain, startRow, currentRow, startValue)
}

// FillSequentialPhrase fills values in phrase view for the current column
func FillSequentialPhrase(m *model.Model) {
	currentRow := m.CurrentRow

	// Use centralized column mapping system
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	if columnMapping == nil || !columnMapping.IsEditable {
		return // Invalid or non-editable column
	}

	colIndex := columnMapping.DataColumnIndex

	// Find the last non-null value going upward
	var startValue int = 0
	var startRow int = 0

	// Use current phrases data based on view type
	phrasesData := m.GetCurrentPhrasesData()

	for row := currentRow - 1; row >= 0; row-- {
		cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
		if cellValue != -1 {
			startValue = cellValue + 1
			startRow = row + 1
			break
		}
	}

	// Handle different column types
	if colIndex == int(types.ColNote) {
		// Special Ctrl+F logic for NN (Note) column
		// Find the fill range - start at the first "--" cell at or above current position
		fillStartRow := currentRow
		for row := currentRow; row >= 0; row-- {
			if (*phrasesData)[m.CurrentPhrase][row][types.ColNote] == -1 {
				fillStartRow = row
			} else {
				break
			}
		}

		// Get the last non-null note value before the fill range (this is the starting note)
		lastNote := -1
		for row := fillStartRow - 1; row >= 0; row-- {
			noteValue := (*phrasesData)[m.CurrentPhrase][row][types.ColNote]
			if noteValue != -1 {
				lastNote = noteValue
				break
			}
		}

		// If no previous note found, we can't fill - need a starting point
		if lastNote == -1 {
			log.Printf("No previous note found to use as starting point for NN fill")
			return
		}

		// Fill each cell in the range
		currentNote := lastNote
		for row := fillStartRow; row <= currentRow; row++ {
			// Get DT value for this row (use effective/sticky value)
			dtValue := (*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime]

			// If no DT in current row, search backwards for last non-null DT
			if dtValue == -1 {
				dtValue = 1 // Default to 1
				for searchRow := row - 1; searchRow >= 0; searchRow-- {
					searchDT := (*phrasesData)[m.CurrentPhrase][searchRow][types.ColDeltaTime]
					if searchDT != -1 {
						dtValue = searchDT
						break
					}
				}
				// Fill in the DT value that we're using
				(*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime] = dtValue
			}

			// Calculate new note value based on DT
			currentNote = currentNote + dtValue

			// Clamp to valid MIDI range (0-127)
			if currentNote > 127 {
				currentNote = 127
			} else if currentNote < 0 {
				currentNote = 0
			}

			// Set the note value
			(*phrasesData)[m.CurrentPhrase][row][types.ColNote] = currentNote
		}

		log.Printf("Filled NN column from row %d to %d, starting note=%d", fillStartRow, currentRow, lastNote)
	} else if colIndex == int(types.ColDeltaTime) {
		// Special Ctrl+F logic for DT column
		currentValue := (*phrasesData)[m.CurrentPhrase][currentRow][colIndex]

		if currentValue <= 0 {
			// Current cell is "--" (value <= 0): Find last non-"--" value and copy it down
			lastNonEmptyValue := 1 // Default to 1 if no previous value found
			fillStartRow := 0

			// Find the last non-"--" value going upward
			for row := currentRow - 1; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue > 0 {
					lastNonEmptyValue = cellValue
					fillStartRow = row + 1
					break
				}
			}

			// Fill from fillStartRow to currentRow with lastNonEmptyValue
			for row := fillStartRow; row <= currentRow; row++ {
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = lastNonEmptyValue
			}
		} else {
			// Current cell is "XX" (value > 0): Switch XX to "--" on current and all above until first non-XX
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != currentValue {
					break // Stop at first cell that doesn't match current value
				}
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = -1 // Set to "--"
			}
		}
	} else if colIndex == int(types.ColModulate) {
		// Special Ctrl+F logic for MO (Modulate) column - similar to DT toggle behavior
		currentValue := (*phrasesData)[m.CurrentPhrase][currentRow][colIndex]

		if currentValue == -1 {
			// Current cell is "--": Find last non-"--" value and copy it down
			lastNonEmptyValue := -1
			fillStartRow := 0

			// Find the last non-"--" value going upward
			for row := currentRow - 1; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != -1 {
					lastNonEmptyValue = cellValue
					fillStartRow = row + 1
					break
				}
			}

			// Only fill if we found a non-empty value
			if lastNonEmptyValue != -1 {
				// Fill from fillStartRow to currentRow with lastNonEmptyValue
				for row := fillStartRow; row <= currentRow; row++ {
					(*phrasesData)[m.CurrentPhrase][row][colIndex] = lastNonEmptyValue
				}
				log.Printf("Filled MO column from row %d to %d with value %02X", fillStartRow, currentRow, lastNonEmptyValue)
			}
		} else {
			// Current cell is not "--": Clear current and all consecutive cells above with the same value
			clearCount := 0
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != currentValue {
					break // Stop at first cell that doesn't match current value
				}
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = -1 // Set to "--"
				clearCount++
			}
			log.Printf("Cleared %d MO cells with value %02X", clearCount, currentValue)
		}
	} else if colIndex == int(types.ColRetrigger) || colIndex == int(types.ColTimestretch) {
		// RT and TS columns should keep the reference value constant
		// Find the last non-null value and fill all cells with that same value
		referenceValue := 0
		fillStartRow := 0
		for row := currentRow - 1; row >= 0; row-- {
			cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
			if cellValue != -1 {
				referenceValue = cellValue
				fillStartRow = row + 1
				break
			}
		}
		// Fill from the found starting row to current row with the same reference value
		for row := fillStartRow; row <= currentRow; row++ {
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = referenceValue
		}
	} else if colIndex == int(types.ColEffectReverse) {
		// Special Ctrl+F logic for Reverse column - similar to DT toggle behavior
		currentValue := (*phrasesData)[m.CurrentPhrase][currentRow][colIndex]

		if currentValue == -1 {
			// Current cell is empty: Find last non-empty value and repeat it, or fill incrementally
			lastNonEmptyValue := -1
			fillStartRow := 0

			// Find the last non-empty value going upward
			for row := currentRow - 1; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != -1 {
					lastNonEmptyValue = cellValue
					fillStartRow = row + 1
					break
				}
			}

			if lastNonEmptyValue != -1 {
				// Found a reference value - repeat it
				for row := fillStartRow; row <= currentRow; row++ {
					(*phrasesData)[m.CurrentPhrase][row][colIndex] = lastNonEmptyValue
				}
			} else {
				// No reference value found - fill incrementally (0-15, wrapping)
				for row := startRow; row <= currentRow; row++ {
					value := startValue + (row - startRow)
					if value > 15 {
						value = value % 16 // Wrap: 0-15, 0-15, ...
					}
					(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
				}
			}
		} else {
			// Current cell has value (0-15): Toggle/erase this value and matching values above
			for row := currentRow; row >= 0; row-- {
				cellValue := (*phrasesData)[m.CurrentPhrase][row][colIndex]
				if cellValue != currentValue {
					break // Stop at first cell that doesn't match current value
				}
				(*phrasesData)[m.CurrentPhrase][row][colIndex] = -1 // Set to empty
			}
		}
	} else if colIndex == int(types.ColPitch) {
		// Pitch column has default of 128, not -1
		if startValue == 0 && startRow == 0 {
			// Special case: if all above are empty, start with default pitch (128)
			startValue = 128
		}
		// Hex columns (0-254)
		maxValue := 254
		for row := startRow; row <= currentRow; row++ {
			value := startValue + (row - startRow)
			if value > maxValue {
				value = value % (maxValue + 1) // Wrap: 0-254, 0-254, ...
			}
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
		}
	} else {
		// Hex columns (0-254)
		maxValue := 254
		for row := startRow; row <= currentRow; row++ {
			value := startValue + (row - startRow)
			if value > maxValue {
				value = value % (maxValue + 1) // Wrap: 0-254, 0-254, ...
			}
			(*phrasesData)[m.CurrentPhrase][row][colIndex] = value
		}
	}

	// Auto-enable playback when filling note column (NN or NOT)
	if colIndex == int(types.ColNote) {
		for row := startRow; row <= currentRow; row++ {
			// Use DT column for both views (only if not already set)
			if (*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime] == -1 {
				(*phrasesData)[m.CurrentPhrase][row][types.ColDeltaTime] = 1
			}
		}
	}

	// Update last edit row
	m.LastEditRow = currentRow

	log.Printf("Filled phrase %02X column %d from row %d to %d, starting with %d", m.CurrentPhrase, colIndex, startRow, currentRow, startValue)
}

// Shared DT (Delta Time) utility functions for both Sampler and Instrument views
// DT controls playback: >0 = play for N ticks, 0 = skip, -1 = skip

// IsRowPlayable checks if a row should be played based on its DT value
func IsRowPlayable(dtValue int) bool {
	return dtValue > 0
}

// GetEffectiveDTValue gets the effective DT value for a row (for display/status)
func GetEffectiveDTValue(dtValue int) string {
	if dtValue == -1 {
		return "--"
	}
	return fmt.Sprintf("%02X", dtValue)
}

// SetDTForPlayback sets DT to 01 (default playback value) for a row
func SetDTForPlayback(phrasesData *[255][][]int, phrase, row int) {
	(*phrasesData)[phrase][row][types.ColDeltaTime] = 1
}

// FindFirstNonEmptyDTAbove finds the first non "--" DT value above the current row
// Returns the DT value if found, or 1 (default) if none found
func FindFirstNonEmptyDTAbove(phrasesData *[255][][]int, phrase, currentRow int) int {
	// Search upward from currentRow-1 to row 0
	for row := currentRow - 1; row >= 0; row-- {
		dtValue := (*phrasesData)[phrase][row][types.ColDeltaTime]
		if dtValue != -1 {
			// Found a non "--" DT value
			return dtValue
		}
	}
	// No non "--" DT found above, return default value of 1
	return 1
}

// FindFirstNonEmptyNoteAbove finds the first non "--" note value above the current row
// Returns the note value if found, or 60 (middle C / C-4) if none found
func FindFirstNonEmptyNoteAbove(phrasesData *[255][][]int, phrase, currentRow int) int {
	// Search upward from currentRow-1 to row 0
	for row := currentRow - 1; row >= 0; row-- {
		noteValue := (*phrasesData)[phrase][row][types.ColNote]
		if noteValue != -1 {
			// Found a non "--" note value
			return noteValue
		}
	}
	// No non "--" note found above, return default value of 60 (middle C / C-4)
	return 60
}

// GetDTStatusMessage returns a status message for DT column
func GetDTStatusMessage(dtValue int) string {
	if dtValue == -1 {
		return "Delta Time: -- (row not played)"
	} else if dtValue == 0 {
		return "Delta Time: 00 (row not played)"
	} else {
		return fmt.Sprintf("Delta Time: %02X (%d ticks, row played)", dtValue, dtValue)
	}
}

// FindNextUnusedRetrigger finds the next unused retrigger slot starting from a given index
func FindNextUnusedRetrigger(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		retriggerID := (startingFrom + offset) % 255
		if retriggerID >= 0 && retriggerID < 255 && IsRetriggerUnused(m, retriggerID) {
			return retriggerID
		}
	}
	return -1 // No unused retrigger found
}

// IsRetriggerUnused checks if a retrigger slot is unused (Times == 0)
func IsRetriggerUnused(m *model.Model, retriggerID int) bool {
	// Bounds check first
	if retriggerID < 0 || retriggerID >= 255 {
		return false
	}

	// Simple check: if Times is 0, the retrigger slot is unused
	return m.RetriggerSettings[retriggerID].Times == 0
}

// FindNextUnusedTimestrech finds the next unused timestrech slot starting from a given index
func FindNextUnusedTimestrech(m *model.Model, startingFrom int) int {
	// Bounds check input
	if startingFrom < 0 || startingFrom >= 255 {
		return -1
	}

	// Search from startingFrom+1 to 254, then wrap to 0 to startingFrom-1
	for offset := 1; offset < 255; offset++ {
		timestrechID := (startingFrom + offset) % 255
		if timestrechID >= 0 && timestrechID < 255 && IsTimestrechUnused(m, timestrechID) {
			return timestrechID
		}
	}
	return -1 // No unused timestrech found
}

// IsTimestrechUnused checks if a timestrech slot is unused (Start == 0, End == 0, and Beats == 0)
func IsTimestrechUnused(m *model.Model, timestrechID int) bool {
	// Bounds check first
	if timestrechID < 0 || timestrechID >= 255 {
		return false
	}

	// Check if timestrech has any meaningful settings - if Start, End, or Beats are non-zero, it's used
	settings := m.TimestrechSettings[timestrechID]
	return settings.Start == 0.0 && settings.End == 0.0 && settings.Beats == 0
}

// ResolveDuckingIndex returns the sticky/effective DU value for a row (or -1)
func ResolveDuckingIndex(m *model.Model, phrase, row, trackId int) int {
	return GetEffectiveValueForTrack(m, phrase, row, int(types.ColEffectDucking), trackId)
}

// ApplyDuckingToSamplerParams clamps and assigns the ducking index to the outgoing params
func ApplyDuckingToSamplerParams(m *model.Model, params *model.SamplerOSCParams, duckIdx int) {
	if duckIdx >= 0 && duckIdx < 255 {
		params.DuckingIndex = duckIdx
	} else {
		params.DuckingIndex = -1
	}
}
