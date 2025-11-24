package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/music"
	"github.com/schollz/collidertracker/internal/ticks"
	"github.com/schollz/collidertracker/internal/types"
)

func RenderInstrumentPhraseView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0"))          // Lighter background, dark text
	selectedDefaultStyle := lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("0")) // Darker background for default pool values, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)                                // White text for normal cells
	normalDefaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)                          // Dimmed text for default pool values when not selected
	sliceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sliceDownbeatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))                          // Lighter gray for downbeats
	playbackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))                              // Green
	copiedStyle := lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")) // Yellow background

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header for Instrument view (row, playback, note, modulation, and chord columns)
	// SO/MI column displays dynamically based on mode
	// ADSR columns display as CC# in MI mode
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")).Bold(true)

	somiHeader := headerStyle.Render("SO")
	adsrHeader := headerStyle.Render("A D S R  ")
	effectHeader := headerStyle.Render(" RE  CO  PA  LP  HP")

	if m.SOColumnMode == types.SOModeMIDI {
		somiHeader = headerStyle.Render("MI")
		// Change ADSR and effect columns to show CC numbers from MidiCCNumbers array
		// Format each CC number with highlighting if on header row

		cc0 := fmt.Sprintf("%02X", m.MidiCCNumbers[0])
		cc1 := fmt.Sprintf("%02X", m.MidiCCNumbers[1])
		cc2 := fmt.Sprintf("%02X", m.MidiCCNumbers[2])
		cc3 := fmt.Sprintf("%02X", m.MidiCCNumbers[3])
		cc4 := fmt.Sprintf("%02X", m.MidiCCNumbers[4])
		cc5 := fmt.Sprintf("%02X", m.MidiCCNumbers[5])
		cc6 := fmt.Sprintf("%02X", m.MidiCCNumbers[6])
		cc7 := fmt.Sprintf("%02X", m.MidiCCNumbers[7])
		cc8 := fmt.Sprintf("%02X", m.MidiCCNumbers[8])

		// Apply header style to all CC numbers first (white)
		cc0 = headerStyle.Render(cc0)
		cc1 = headerStyle.Render(cc1)
		cc2 = headerStyle.Render(cc2)
		cc3 = headerStyle.Render(cc3)
		cc4 = headerStyle.Render(cc4)
		cc5 = headerStyle.Render(cc5)
		cc6 = headerStyle.Render(cc6)
		cc7 = headerStyle.Render(cc7)
		cc8 = headerStyle.Render(cc8)

		// Then highlight the selected column header if on header row
		if m.CurrentRow == -1 {
			switch m.CurrentCol {
			case int(types.InstrumentColATK):
				cc0 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[0]))
			case int(types.InstrumentColDECAY):
				cc1 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[1]))
			case int(types.InstrumentColSUS):
				cc2 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[2]))
			case int(types.InstrumentColREL):
				cc3 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[3]))
			case int(types.InstrumentColRE):
				cc4 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[4]))
			case int(types.InstrumentColCO):
				cc5 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[5]))
			case int(types.InstrumentColPA):
				cc6 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[6]))
			case int(types.InstrumentColLP):
				cc7 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[7]))
			case int(types.InstrumentColHP):
				cc8 = highlightStyle.Render(fmt.Sprintf("%02X", m.MidiCCNumbers[8]))
			}
		}

		// the spacing is important here to keep alignment
		adsrHeader = cc0 + cc1 + cc2 + cc3 + "  "
		effectHeader = cc4 + "  " + cc5 + "  " + cc6 + "  " + cc7 + "  " + cc8
	}

	// Highlight SO/MI column header if we're on header row (-1) and on that column
	if m.CurrentRow == -1 && m.CurrentCol == int(types.InstrumentColSOMI) {
		somiHeader = highlightStyle.Render("SO")
		if m.SOColumnMode == types.SOModeMIDI {
			somiHeader = highlightStyle.Render("MI")
		}
	}

	columnHeader := headerStyle.Render("  SL  DT  NOT  MO  CAT  VE  GT ") + adsrHeader + effectHeader + headerStyle.Render("  AR  ") + somiHeader + headerStyle.Render("  DU")
	phrasesData := m.GetCurrentPhrasesData()
	totalTicks := ticks.CalculatePhraseTicks(phrasesData, m.CurrentPhrase)
	phraseHeader := headerStyle.Render(fmt.Sprintf("Instrument %02X (%d ticks)", m.CurrentPhrase, totalTicks))
	content.WriteString(RenderHeader(m, columnHeader, phraseHeader))

	// Data rows
	visibleRows := m.GetVisibleRows()
	for i := 0; i < visibleRows && i+m.ScrollOffset < 255; i++ {
		dataIndex := i + m.ScrollOffset

		// Arrow for current row or playback
		arrow := " "
		if m.IsPlaying {
			if m.PlaybackMode == types.SongView {
				// Song playback mode - check the current track context
				if m.SongPlaybackActive[m.CurrentTrack] &&
					m.SongPlaybackPhrase[m.CurrentTrack] == m.CurrentPhrase &&
					m.SongPlaybackRowInPhrase[m.CurrentTrack] == dataIndex {
					arrow = playbackStyle.Render("▶")
				}
			} else {
				// Chain/Phrase playback mode - use existing logic
				if m.PlaybackPhrase == m.CurrentPhrase && m.PlaybackRow == dataIndex {
					arrow = playbackStyle.Render("▶")
				}
			}
		} else if m.CurrentRow == dataIndex {
			// Not playing - show cursor arrow
			arrow = "▶"
		}

		// Slice number (hex)
		sliceHex := fmt.Sprintf("%02X", dataIndex)
		var sliceCell string
		if dataIndex%4 == 0 {
			sliceCell = sliceDownbeatStyle.Render(sliceHex) // Lighter for downbeats
		} else {
			sliceCell = sliceStyle.Render(sliceHex)
		}

		// Delta Time (DT) column - unified playback control for both Sampler and Instrument views
		phrasesData := m.GetCurrentPhrasesData()
		dtValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDeltaTime]
		dtText := input.GetEffectiveDTValue(dtValue)

		var dtCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColDT) { // Column 1 is the DT column
			dtCell = selectedStyle.Render(dtText)
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColDT)) {
				dtCell = copiedStyle.Render(dtText)
			} else {
				dtCell = normalStyle.Render(dtText)
			}
		} else {
			dtCell = normalStyle.Render(dtText)
		}

		// Note (NOT) - use ColNote but display as note name
		// For Instrument view, we're using the Note column to store MIDI note values (0-127)
		noteValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColNote]
		noteText := "---"
		if noteValue != -1 {
			noteText = music.MidiToNoteName(noteValue)
		}

		var noteCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColNOT) { // Column 2 is the NOT column
			noteCell = selectedStyle.Render(fmt.Sprintf("%3s", noteText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColNOT)) {
				noteCell = copiedStyle.Render(fmt.Sprintf("%3s", noteText))
			} else {
				noteCell = normalStyle.Render(fmt.Sprintf("%3s", noteText))
			}
		} else {
			noteCell = normalStyle.Render(fmt.Sprintf("%3s", noteText))
		}

		// Modulation (MO) - display modulation index
		modulateValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColModulate]
		modulateText := "--"
		if modulateValue != -1 {
			modulateText = fmt.Sprintf("%02X", modulateValue)
		}
		// Check if this modulate setting is still at default values (instrument uses InstrumentModulateSettings)
		moIsDefault := modulateValue != -1 && m.IsModulateSettingDefault(m.InstrumentModulateSettings[modulateValue])

		var modulateCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColMO) { // Column 3 is the MO column
			if moIsDefault {
				modulateCell = selectedDefaultStyle.Render(fmt.Sprintf("%2s", modulateText))
			} else {
				modulateCell = selectedStyle.Render(fmt.Sprintf("%2s", modulateText))
			}
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColMO)) {
				modulateCell = copiedStyle.Render(fmt.Sprintf("%2s", modulateText))
			} else {
				if moIsDefault {
					modulateCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", modulateText))
				} else {
					modulateCell = normalStyle.Render(fmt.Sprintf("%2s", modulateText))
				}
			}
		} else {
			if moIsDefault {
				modulateCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", modulateText))
			} else {
				modulateCell = normalStyle.Render(fmt.Sprintf("%2s", modulateText))
			}
		}

		// Chord (C) - display chord type
		chordValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChord]
		chordText := types.ChordTypeToString(types.ChordType(chordValue))

		var chordCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColC) { // Column 4 is the C column
			chordCell = selectedStyle.Render(fmt.Sprintf("%1s", chordText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColC)) {
				chordCell = copiedStyle.Render(fmt.Sprintf("%1s", chordText))
			} else {
				chordCell = normalStyle.Render(fmt.Sprintf("%1s", chordText))
			}
		} else {
			chordCell = normalStyle.Render(fmt.Sprintf("%1s", chordText))
		}

		// Chord Addition (A) - display chord addition
		chordAddValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChordAddition]
		chordAddText := types.ChordAdditionToString(types.ChordAddition(chordAddValue))

		var chordAddCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColA) { // Column 5 is the A column
			chordAddCell = selectedStyle.Render(fmt.Sprintf("%1s", chordAddText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColA)) {
				chordAddCell = copiedStyle.Render(fmt.Sprintf("%1s", chordAddText))
			} else {
				chordAddCell = normalStyle.Render(fmt.Sprintf("%1s", chordAddText))
			}
		} else {
			chordAddCell = normalStyle.Render(fmt.Sprintf("%1s", chordAddText))
		}

		// Chord Transposition (T) - display transposition value
		chordTransValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColChordTransposition]
		chordTransText := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))

		var chordTransCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColT) { // Column 6 is the T column
			chordTransCell = selectedStyle.Render(fmt.Sprintf("%1s", chordTransText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColT)) {
				chordTransCell = copiedStyle.Render(fmt.Sprintf("%1s", chordTransText))
			} else {
				chordTransCell = normalStyle.Render(fmt.Sprintf("%1s", chordTransText))
			}
		} else {
			chordTransCell = normalStyle.Render(fmt.Sprintf("%1s", chordTransText))
		}

		// Velocity (VE) - display velocity value (00-7F)
		velocityValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColVelocity]
		velocityText := "--"
		if velocityValue != -1 {
			velocityText = fmt.Sprintf("%02X", velocityValue)
		}

		var velocityCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColVE) { // Column 7 is the VE column
			velocityCell = selectedStyle.Render(fmt.Sprintf("%2s", velocityText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColVE)) {
				velocityCell = copiedStyle.Render(fmt.Sprintf("%2s", velocityText))
			} else {
				velocityCell = normalStyle.Render(fmt.Sprintf("%2s", velocityText))
			}
		} else {
			velocityCell = normalStyle.Render(fmt.Sprintf("%2s", velocityText))
		}

		// Gate (GT) - display gate value
		gateValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColGate]
		gateText := "--"
		if gateValue != -1 {
			gateText = fmt.Sprintf("%02X", gateValue)
		}

		var gateCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColGT) { // Column 8 is the GT column
			gateCell = selectedStyle.Render(fmt.Sprintf("%2s", gateText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColGT)) {
				gateCell = copiedStyle.Render(fmt.Sprintf("%2s", gateText))
			} else {
				gateCell = normalStyle.Render(fmt.Sprintf("%2s", gateText))
			}
		} else {
			gateCell = normalStyle.Render(fmt.Sprintf("%2s", gateText))
		}

		// Attack (A) or MIDI CC 0 - display based on mode
		var attackValue int
		if m.SOColumnMode == types.SOModeMIDI {
			attackValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC0]
		} else {
			attackValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColAttack]
		}
		attackText := "--"
		if attackValue != -1 {
			attackText = fmt.Sprintf("%02X", attackValue)
		}

		var attackCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColATK) { // Column 9 is the A column
			attackCell = selectedStyle.Render(fmt.Sprintf("%2s", attackText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColATK)) {
				attackCell = copiedStyle.Render(fmt.Sprintf("%2s", attackText))
			} else {
				attackCell = normalStyle.Render(fmt.Sprintf("%2s", attackText))
			}
		} else {
			attackCell = normalStyle.Render(fmt.Sprintf("%2s", attackText))
		}

		// Decay (D) or MIDI CC 1 - display based on mode
		var decayValue int
		if m.SOColumnMode == types.SOModeMIDI {
			decayValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC1]
		} else {
			decayValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColDecay]
		}
		decayText := "--"
		if decayValue != -1 {
			decayText = fmt.Sprintf("%02X", decayValue)
		}

		var decayCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColDECAY) { // Column 10 is the D column
			decayCell = selectedStyle.Render(fmt.Sprintf("%2s", decayText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColDECAY)) {
				decayCell = copiedStyle.Render(fmt.Sprintf("%2s", decayText))
			} else {
				decayCell = normalStyle.Render(fmt.Sprintf("%2s", decayText))
			}
		} else {
			decayCell = normalStyle.Render(fmt.Sprintf("%2s", decayText))
		}

		// Sustain (S) or MIDI CC 2 - display based on mode
		var sustainValue int
		if m.SOColumnMode == types.SOModeMIDI {
			sustainValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC2]
		} else {
			sustainValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColSustain]
		}
		sustainText := "--"
		if sustainValue != -1 {
			sustainText = fmt.Sprintf("%02X", sustainValue)
		}

		var sustainCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColSUS) { // Column 11 is the S column
			sustainCell = selectedStyle.Render(fmt.Sprintf("%2s", sustainText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColSUS)) {
				sustainCell = copiedStyle.Render(fmt.Sprintf("%2s", sustainText))
			} else {
				sustainCell = normalStyle.Render(fmt.Sprintf("%2s", sustainText))
			}
		} else {
			sustainCell = normalStyle.Render(fmt.Sprintf("%2s", sustainText))
		}

		// Release (R) or MIDI CC 3 - display based on mode
		var releaseValue int
		if m.SOColumnMode == types.SOModeMIDI {
			releaseValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC3]
		} else {
			releaseValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColRelease]
		}
		releaseText := "--"
		if releaseValue != -1 {
			releaseText = fmt.Sprintf("%02X", releaseValue)
		}

		var releaseCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColREL) { // Column 12 is the R column
			releaseCell = selectedStyle.Render(fmt.Sprintf("%2s", releaseText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColREL)) {
				releaseCell = copiedStyle.Render(fmt.Sprintf("%2s", releaseText))
			} else {
				releaseCell = normalStyle.Render(fmt.Sprintf("%2s", releaseText))
			}
		} else {
			releaseCell = normalStyle.Render(fmt.Sprintf("%2s", releaseText))
		}

		// Reverb (RE) or MIDI CC 4 - display based on mode
		var reverbValue int
		if m.SOColumnMode == types.SOModeMIDI {
			reverbValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC4]
		} else {
			reverbValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectReverb]
		}
		reverbText := "--"
		if reverbValue != -1 {
			reverbText = fmt.Sprintf("%02X", reverbValue)
		}

		var reverbCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColRE) {
			reverbCell = selectedStyle.Render(fmt.Sprintf("%2s", reverbText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColRE)) {
				reverbCell = copiedStyle.Render(fmt.Sprintf("%2s", reverbText))
			} else {
				reverbCell = normalStyle.Render(fmt.Sprintf("%2s", reverbText))
			}
		} else {
			reverbCell = normalStyle.Render(fmt.Sprintf("%2s", reverbText))
		}

		// Comb (CO) or MIDI CC 5 - display based on mode
		var combValue int
		if m.SOColumnMode == types.SOModeMIDI {
			combValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC5]
		} else {
			combValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectComb]
		}
		combText := "--"
		if combValue != -1 {
			combText = fmt.Sprintf("%02X", combValue)
		}

		var combCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColCO) {
			combCell = selectedStyle.Render(fmt.Sprintf("%2s", combText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColCO)) {
				combCell = copiedStyle.Render(fmt.Sprintf("%2s", combText))
			} else {
				combCell = normalStyle.Render(fmt.Sprintf("%2s", combText))
			}
		} else {
			combCell = normalStyle.Render(fmt.Sprintf("%2s", combText))
		}

		// Pan (PA) or MIDI CC 6 - display based on mode
		var panValue int
		if m.SOColumnMode == types.SOModeMIDI {
			panValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC6]
		} else {
			panValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColPan]
		}
		panText := "--"
		if panValue != -1 {
			panText = fmt.Sprintf("%02X", panValue)
		}

		var panCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColPA) {
			panCell = selectedStyle.Render(fmt.Sprintf("%2s", panText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColPA)) {
				panCell = copiedStyle.Render(fmt.Sprintf("%2s", panText))
			} else {
				panCell = normalStyle.Render(fmt.Sprintf("%2s", panText))
			}
		} else {
			panCell = normalStyle.Render(fmt.Sprintf("%2s", panText))
		}

		// LowPass (LP) or MIDI CC 7 - display based on mode
		var lpValue int
		if m.SOColumnMode == types.SOModeMIDI {
			lpValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC7]
		} else {
			lpValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColLowPassFilter]
		}
		lpText := "--"
		if lpValue != -1 {
			lpText = fmt.Sprintf("%02X", lpValue)
		}

		var lpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColLP) {
			lpCell = selectedStyle.Render(fmt.Sprintf("%2s", lpText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColLP)) {
				lpCell = copiedStyle.Render(fmt.Sprintf("%2s", lpText))
			} else {
				lpCell = normalStyle.Render(fmt.Sprintf("%2s", lpText))
			}
		} else {
			lpCell = normalStyle.Render(fmt.Sprintf("%2s", lpText))
		}

		// HighPass (HP) or MIDI CC 8 - display based on mode
		var hpValue int
		if m.SOColumnMode == types.SOModeMIDI {
			hpValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidiCC8]
		} else {
			hpValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColHighPassFilter]
		}
		hpText := "--"
		if hpValue != -1 {
			hpText = fmt.Sprintf("%02X", hpValue)
		}

		var hpCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColHP) {
			hpCell = selectedStyle.Render(fmt.Sprintf("%2s", hpText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColHP)) {
				hpCell = copiedStyle.Render(fmt.Sprintf("%2s", hpText))
			} else {
				hpCell = normalStyle.Render(fmt.Sprintf("%2s", hpText))
			}
		} else {
			hpCell = normalStyle.Render(fmt.Sprintf("%2s", hpText))
		}

		// Arpeggio (AR) - display arpeggio index
		arpeggioValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColArpeggio]
		arpeggioText := "--"
		if arpeggioValue != -1 {
			arpeggioText = fmt.Sprintf("%02X", arpeggioValue)
		}

		var arpeggioCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColAR) { // Column 18 is the AR column
			arpeggioCell = selectedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColAR)) {
				arpeggioCell = copiedStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			} else {
				arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
			}
		} else {
			arpeggioCell = normalStyle.Render(fmt.Sprintf("%2s", arpeggioText))
		}

		// SO/MI (toggleable) - display SoundMaker or MIDI index based on mode
		var somiValue int
		var somiText string
		if m.SOColumnMode == types.SOModeMIDI {
			somiValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColMidi]
		} else {
			somiValue = (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColSoundMaker]
		}
		somiText = "--"
		if somiValue != -1 {
			somiText = fmt.Sprintf("%02X", somiValue)
		}
		// Check if this SoundMaker setting is still at default values (only check for SO mode, not MIDI)
		soIsDefault := m.SOColumnMode == types.SOModeSound && somiValue != -1 && m.IsSoundMakerSettingDefault(somiValue)

		var somiCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColSOMI) { // Column 19 is the SO/MI column
			if soIsDefault {
				somiCell = selectedDefaultStyle.Render(fmt.Sprintf("%2s", somiText))
			} else {
				somiCell = selectedStyle.Render(fmt.Sprintf("%2s", somiText))
			}
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColSOMI)) {
				somiCell = copiedStyle.Render(fmt.Sprintf("%2s", somiText))
			} else {
				if soIsDefault {
					somiCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", somiText))
				} else {
					somiCell = normalStyle.Render(fmt.Sprintf("%2s", somiText))
				}
			}
		} else {
			if soIsDefault {
				somiCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", somiText))
			} else {
				somiCell = normalStyle.Render(fmt.Sprintf("%2s", somiText))
			}
		}

		// Ducking (DU) - display ducking index
		duckingValue := (*phrasesData)[m.CurrentPhrase][dataIndex][types.ColEffectDucking]
		duckingText := "--"
		if duckingValue != -1 {
			duckingText = fmt.Sprintf("%02X", duckingValue)
		}
		// Check if this ducking setting is still at default values
		duckingIsDefault := duckingValue != -1 && m.IsDuckingSettingDefault(duckingValue)

		var duckingCell string
		if m.CurrentRow == dataIndex && m.CurrentCol == int(types.InstrumentColDU) { // Column 21 is the DU column
			if duckingIsDefault {
				duckingCell = selectedDefaultStyle.Render(fmt.Sprintf("%2s", duckingText))
			} else {
				duckingCell = selectedStyle.Render(fmt.Sprintf("%2s", duckingText))
			}
		} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.PhraseView && m.Clipboard.HighlightPhrase == m.CurrentPhrase && m.Clipboard.HighlightRow == dataIndex {
			if m.Clipboard.Mode == types.RowMode || (m.Clipboard.Mode == types.CellMode && m.Clipboard.HighlightCol == int(types.InstrumentColDU)) {
				duckingCell = copiedStyle.Render(fmt.Sprintf("%2s", duckingText))
			} else {
				if duckingIsDefault {
					duckingCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", duckingText))
				} else {
					duckingCell = normalStyle.Render(fmt.Sprintf("%2s", duckingText))
				}
			}
		} else {
			if duckingIsDefault {
				duckingCell = normalDefaultStyle.Render(fmt.Sprintf("%2s", duckingText))
			} else {
				duckingCell = normalStyle.Render(fmt.Sprintf("%2s", duckingText))
			}
		}

		row := fmt.Sprintf("%s %-3s  %s  %s  %s  %s%s%s  %s  %s %s%s%s%s  %s  %s  %s  %s  %s  %s  %s  %s", arrow, sliceCell, dtCell, noteCell, modulateCell, chordCell, chordAddCell, chordTransCell, velocityCell, gateCell, attackCell, decayCell, sustainCell, releaseCell, reverbCell, combCell, panCell, lpCell, hpCell, arpeggioCell, somiCell, duckingCell)
		content.WriteString(row)
		content.WriteString("\n")
	}

	// Footer with status
	helpText := GetPhraseHelpText(m)
	statusMsg := GetInstrumentPhraseStatusMessage(m)
	content.WriteString(RenderFooter(m, visibleRows+1, helpText, statusMsg)) // +1 for header

	// Apply container padding to entire content
	return containerStyle.Render(content.String())
}

func GetInstrumentPhraseStatusMessage(m *model.Model) string {
	var statusMsg string

	// Handle header row (row -1) for SO/MI column mode switching and CC number editing
	if m.CurrentRow == -1 {
		if m.CurrentCol == int(types.InstrumentColSOMI) {
			if m.SOColumnMode == types.SOModeMIDI {
				statusMsg = "Column Mode: MI (MIDI) | Ctrl+Down/Left: Switch to SO"
			} else {
				statusMsg = "Column Mode: SO (SoundMaker) | Ctrl+Up/Right: Switch to MI"
			}
		} else if m.SOColumnMode == types.SOModeMIDI {
			// Check if on a CC column
			ccIndex := -1
			ccName := ""
			switch m.CurrentCol {
			case int(types.InstrumentColATK):
				ccIndex = 0
				ccName = "Attack/CC0"
			case int(types.InstrumentColDECAY):
				ccIndex = 1
				ccName = "Decay/CC1"
			case int(types.InstrumentColSUS):
				ccIndex = 2
				ccName = "Sustain/CC2"
			case int(types.InstrumentColREL):
				ccIndex = 3
				ccName = "Release/CC3"
			case int(types.InstrumentColRE):
				ccIndex = 4
				ccName = "Reverb/CC4"
			case int(types.InstrumentColCO):
				ccIndex = 5
				ccName = "Comb/CC5"
			case int(types.InstrumentColPA):
				ccIndex = 6
				ccName = "Pan/CC6"
			case int(types.InstrumentColLP):
				ccIndex = 7
				ccName = "LowPass/CC7"
			case int(types.InstrumentColHP):
				ccIndex = 8
				ccName = "HighPass/CC8"
			}
			if ccIndex != -1 {
				statusMsg = fmt.Sprintf("%s CC#: %d | Ctrl+Up: +16, Ctrl+Right: +1, Ctrl+Down: -16, Ctrl+Left: -1", ccName, m.MidiCCNumbers[ccIndex])
			} else {
				statusMsg = "Header row"
			}
		} else {
			statusMsg = "Header row"
		}
		return statusMsg
	}

	// Use centralized column mapping to determine current column
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	phrasesData := m.GetCurrentPhrasesData()

	if columnMapping != nil && (columnMapping.DataColumnIndex == int(types.ColNote) ||
		columnMapping.DataColumnIndex == int(types.ColChord) ||
		columnMapping.DataColumnIndex == int(types.ColChordAddition) ||
		columnMapping.DataColumnIndex == int(types.ColChordTransposition)) { // NOT, C, A, or T columns
		// Get current row data
		noteValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColNote]
		chordValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChord]
		chordAddValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChordAddition]
		chordTransValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColChordTransposition]

		if noteValue >= 0 && noteValue <= 127 {
			noteName := music.MidiToNoteName(noteValue)

			// Check if chord is defined (not null/"-")
			if chordValue > int(types.ChordNone) {
				// Extract note name and octave from MIDI note
				noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
				rootNote := noteNames[noteValue%12]
				octave := (noteValue / 12) - 1

				// Build chord name
				var chordName string
				switch types.ChordType(chordValue) {
				case types.ChordMajor:
					chordName = rootNote + "maj"
				case types.ChordMinor:
					chordName = rootNote + "min"
				case types.ChordDominant:
					chordName = rootNote // Dominant chords have no suffix
				default:
					chordName = rootNote
				}

				// Add chord addition if defined
				if chordAddValue > int(types.ChordAddNone) {
					switch types.ChordAddition(chordAddValue) {
					case types.ChordAdd7:
						chordName += "7"
					case types.ChordAdd9:
						chordName += "9"
					case types.ChordAdd4:
						chordName += "4"
					}
				}

				// Add transposition if defined and not 0
				if chordTransValue > int(types.ChordTrans0) {
					transpositionStr := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))
					statusMsg = fmt.Sprintf("Chord: %s (octave %d, transpose %s)", chordName, octave, transpositionStr)
				} else {
					statusMsg = fmt.Sprintf("Chord: %s (octave %d)", chordName, octave)
				}
			} else {
				// Chord is null, show simple note info with transposition if defined and not 0
				if chordTransValue > int(types.ChordTrans0) {
					transpositionStr := types.ChordTranspositionToString(types.ChordTransposition(chordTransValue))
					statusMsg = fmt.Sprintf("Note: %s (transpose %s)", noteName, transpositionStr)
				} else {
					statusMsg = fmt.Sprintf("Note: %s", noteName)
				}
			}
		} else {
			statusMsg = "No note selected"
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColModulate) { // MO column
		// Show Modulation info
		modulateValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColModulate]
		if modulateValue >= 0 && modulateValue < 255 {
			statusMsg = fmt.Sprintf("Modulate: %02X", modulateValue)
		} else {
			statusMsg = "No modulate selected"
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColDeltaTime) { // DT column
		// Show DT playback info
		playbackValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDeltaTime]
		if playbackValue > 0 {
			statusMsg = fmt.Sprintf("DT: %02X (play %d ticks)", playbackValue, playbackValue)
		} else {
			statusMsg = "DT: not played"
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColGate) { // GT column
		// Show Gate info with percentage (80 = 100% gate)
		gateValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColGate]
		if gateValue == -1 {
			// Check for effective (sticky) Gate value
			effectiveGateValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColGate), m.CurrentTrack)
			if effectiveGateValue == -1 {
				statusMsg = "Gate: -- (80/100%, sticky)"
			} else {
				gatePercent := float32(effectiveGateValue) / 128.0 * 100.0
				statusMsg = fmt.Sprintf("Gate: -- (%02X/%.0f%%, sticky)", effectiveGateValue, gatePercent)
			}
		} else {
			gatePercent := float32(gateValue) / 128.0 * 100.0
			statusMsg = fmt.Sprintf("Gate: %02X (%.0f%%, sticky)", gateValue, gatePercent)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColAttack) { // A column
		// Show Attack info
		attackValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColAttack]
		if attackValue == -1 {
			// Check for effective (sticky) Attack value
			effectiveAttackValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColAttack), m.CurrentTrack)
			if effectiveAttackValue == -1 {
				statusMsg = "Attack: -- (sticky)"
			} else {
				attackSeconds := types.AttackToSeconds(effectiveAttackValue)
				statusMsg = fmt.Sprintf("Attack: -- (%.3fs, sticky)", attackSeconds)
			}
		} else {
			attackSeconds := types.AttackToSeconds(attackValue)
			statusMsg = fmt.Sprintf("Attack: %02X (%.3fs, sticky)", attackValue, attackSeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColDecay) { // D column
		// Show Decay info
		decayValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColDecay]
		if decayValue == -1 {
			// Check for effective (sticky) Decay value
			effectiveDecayValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColDecay), m.CurrentTrack)
			if effectiveDecayValue == -1 {
				statusMsg = "Decay: -- (sticky)"
			} else {
				decaySeconds := types.DecayToSeconds(effectiveDecayValue)
				statusMsg = fmt.Sprintf("Decay: -- (%.3fs, sticky)", decaySeconds)
			}
		} else {
			decaySeconds := types.DecayToSeconds(decayValue)
			statusMsg = fmt.Sprintf("Decay: %02X (%.3fs, sticky)", decayValue, decaySeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColSustain) { // S column
		// Show Sustain info
		sustainValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSustain]
		if sustainValue == -1 {
			// Check for effective (sticky) Sustain value
			effectiveSustainValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColSustain), m.CurrentTrack)
			if effectiveSustainValue == -1 {
				statusMsg = "Sustain: -- (sticky)"
			} else {
				sustainLevel := types.SustainToLevel(effectiveSustainValue)
				statusMsg = fmt.Sprintf("Sustain: -- (%.3f, sticky)", sustainLevel)
			}
		} else {
			sustainLevel := types.SustainToLevel(sustainValue)
			statusMsg = fmt.Sprintf("Sustain: %02X (%.3f, sticky)", sustainValue, sustainLevel)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColRelease) { // R column
		// Show Release info
		releaseValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColRelease]
		if releaseValue == -1 {
			// Check for effective (sticky) Release value
			effectiveReleaseValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColRelease), m.CurrentTrack)
			if effectiveReleaseValue == -1 {
				statusMsg = "Release: -- (sticky)"
			} else {
				releaseSeconds := types.ReleaseToSeconds(effectiveReleaseValue)
				statusMsg = fmt.Sprintf("Release: -- (%.3fs, sticky)", releaseSeconds)
			}
		} else {
			releaseSeconds := types.ReleaseToSeconds(releaseValue)
			statusMsg = fmt.Sprintf("Release: %02X (%.3fs, sticky)", releaseValue, releaseSeconds)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColEffectReverb) { // RE column
		// Show Reverb info with sticky behavior
		reverbValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColEffectReverb]
		if reverbValue == -1 {
			// Check for effective (sticky) Reverb value
			effectiveReverbValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColEffectReverb), m.CurrentTrack)
			if effectiveReverbValue == -1 {
				statusMsg = "Reverb: -- (sticky)"
			} else {
				reverbFloat := float32(effectiveReverbValue) / 254.0
				statusMsg = fmt.Sprintf("Reverb: -- (%.2f, sticky)", reverbFloat)
			}
		} else {
			reverbFloat := float32(reverbValue) / 254.0
			statusMsg = fmt.Sprintf("Reverb: %02X (%.2f, sticky)", reverbValue, reverbFloat)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColEffectComb) { // CO column
		// Show Comb info with sticky behavior
		combValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColEffectComb]
		if combValue == -1 {
			// Check for effective (sticky) Comb value
			effectiveCombValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColEffectComb), m.CurrentTrack)
			if effectiveCombValue == -1 {
				statusMsg = "Comb: -- (sticky)"
			} else {
				combFloat := float32(effectiveCombValue) / 254.0
				statusMsg = fmt.Sprintf("Comb: -- (%.2f, sticky)", combFloat)
			}
		} else {
			combFloat := float32(combValue) / 254.0
			statusMsg = fmt.Sprintf("Comb: %02X (%.2f, sticky)", combValue, combFloat)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColPan) { // PA column
		// Show Pan info with sticky behavior
		panValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColPan]
		if panValue == -1 {
			// Check for effective (sticky) Pan value - default is 80 (center/0.0)
			effectivePanValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColPan), m.CurrentTrack)
			if effectivePanValue == -1 {
				statusMsg = "Pan: -- (0.0, sticky)"
			} else {
				panFloat := (float32(effectivePanValue) - 127.0) / 127.0 // Map 0-254 to -1.0 to 1.0
				statusMsg = fmt.Sprintf("Pan: -- (%.2f, sticky)", panFloat)
			}
		} else {
			panFloat := (float32(panValue) - 127.0) / 127.0 // Map 0-254 to -1.0 to 1.0
			statusMsg = fmt.Sprintf("Pan: %02X (%.2f, sticky)", panValue, panFloat)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColLowPassFilter) { // LP column
		// Show Low Pass Filter info with sticky behavior
		lpValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColLowPassFilter]
		if lpValue == -1 {
			// Check for effective (sticky) Low Pass value - default is 20kHz
			effectiveLpValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColLowPassFilter), m.CurrentTrack)
			if effectiveLpValue == -1 {
				statusMsg = "Low Pass: -- (20kHz, sticky)"
			} else {
				// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
				logMin := float32(1.301) // log10(20)
				logMax := float32(4.301) // log10(20000)
				logFreq := logMin + (float32(effectiveLpValue)/254.0)*(logMax-logMin)
				freq := float32(math.Pow(10, float64(logFreq)))
				if freq >= 1000 {
					statusMsg = fmt.Sprintf("Low Pass: -- (%.1fkHz, sticky)", freq/1000)
				} else {
					statusMsg = fmt.Sprintf("Low Pass: -- (%.0fHz, sticky)", freq)
				}
			}
		} else {
			// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
			logMin := float32(1.301) // log10(20)
			logMax := float32(4.301) // log10(20000)
			logFreq := logMin + (float32(lpValue)/254.0)*(logMax-logMin)
			freq := float32(math.Pow(10, float64(logFreq)))
			if freq >= 1000 {
				statusMsg = fmt.Sprintf("Low Pass: %02X (%.1fkHz, sticky)", lpValue, freq/1000)
			} else {
				statusMsg = fmt.Sprintf("Low Pass: %02X (%.0fHz, sticky)", lpValue, freq)
			}
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColHighPassFilter) { // HP column
		// Show High Pass Filter info with sticky behavior
		hpValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColHighPassFilter]
		if hpValue == -1 {
			// Check for effective (sticky) High Pass value - default is 20Hz
			effectiveHpValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColHighPassFilter), m.CurrentTrack)
			if effectiveHpValue == -1 {
				statusMsg = "High Pass: -- (20Hz, sticky)"
			} else {
				// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
				logMin := float32(1.301) // log10(20)
				logMax := float32(4.301) // log10(20000)
				logFreq := logMin + (float32(effectiveHpValue)/254.0)*(logMax-logMin)
				freq := float32(math.Pow(10, float64(logFreq)))
				if freq >= 1000 {
					statusMsg = fmt.Sprintf("High Pass: -- (%.1fkHz, sticky)", freq/1000)
				} else {
					statusMsg = fmt.Sprintf("High Pass: -- (%.0fHz, sticky)", freq)
				}
			}
		} else {
			// Exponential mapping: 00 -> 20Hz, FE -> 20kHz
			logMin := float32(1.301) // log10(20)
			logMax := float32(4.301) // log10(20000)
			logFreq := logMin + (float32(hpValue)/254.0)*(logMax-logMin)
			freq := float32(math.Pow(10, float64(logFreq)))
			if freq >= 1000 {
				statusMsg = fmt.Sprintf("High Pass: %02X (%.1fkHz, sticky)", hpValue, freq/1000)
			} else {
				statusMsg = fmt.Sprintf("High Pass: %02X (%.0fHz, sticky)", hpValue, freq)
			}
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColArpeggio) { // AR column
		// Show Arpeggio info
		arpeggioValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColArpeggio]
		if arpeggioValue == -1 {
			statusMsg = "Arpeggio: -- (not assigned)"
		} else {
			statusMsg = fmt.Sprintf("Arpeggio: %02X", arpeggioValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColMidi) { // MI column
		// Show MIDI info with sticky behavior
		midiValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColMidi]
		if midiValue == -1 {
			// Check for effective (sticky) MIDI value
			effectiveMidiValue := input.GetEffectiveMidiValueForTrack(m, m.CurrentPhrase, m.CurrentRow, m.CurrentTrack)
			if effectiveMidiValue == -1 {
				statusMsg = "MIDI: -- (sticky)"
			} else {
				statusMsg = fmt.Sprintf("MIDI: -- (%02X sticky)", effectiveMidiValue)
			}
		} else {
			statusMsg = fmt.Sprintf("MIDI: %02X (sticky)", midiValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColSoundMaker) { // SO column
		// Show SoundMaker info with sticky behavior
		soundMakerValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColSoundMaker]
		if soundMakerValue == -1 {
			// Check for effective (sticky) SoundMaker value
			effectiveSoundMakerValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColSoundMaker), m.CurrentTrack)
			if effectiveSoundMakerValue == -1 {
				statusMsg = "SoundMaker: -- (sticky)"
			} else {
				statusMsg = fmt.Sprintf("SoundMaker: -- (%02X sticky)", effectiveSoundMakerValue)
			}
		} else {
			statusMsg = fmt.Sprintf("SoundMaker: %02X (sticky)", soundMakerValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColVelocity) { // VE column
		// Show Velocity info with sticky behavior
		velocityValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColVelocity]
		if velocityValue == -1 {
			// Check for effective (sticky) Velocity value
			effectiveVelocityValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColVelocity), m.CurrentTrack)
			if effectiveVelocityValue == -1 {
				statusMsg = "Velocity: -- (64, sticky)"
			} else {
				statusMsg = fmt.Sprintf("Velocity: -- (%02X/%d, sticky)", effectiveVelocityValue, effectiveVelocityValue)
			}
		} else {
			statusMsg = fmt.Sprintf("Velocity: %02X (%d, sticky)", velocityValue, velocityValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex == int(types.ColEffectDucking) { // DU column
		// Show Ducking info with sticky behavior (reuse logic from sampler view)
		duckingValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.ColEffectDucking]
		if duckingValue == -1 {
			// Check for effective (sticky) Ducking value
			effectiveDuckingValue := input.GetEffectiveValueForTrack(m, m.CurrentPhrase, m.CurrentRow, int(types.ColEffectDucking), m.CurrentTrack)
			if effectiveDuckingValue == -1 {
				statusMsg = "Ducking: -- (sticky)"
			} else {
				statusMsg = fmt.Sprintf("Ducking: -- (%02X, sticky)", effectiveDuckingValue)
			}
		} else {
			statusMsg = fmt.Sprintf("Ducking: %02X (sticky)", duckingValue)
		}
	} else if columnMapping != nil && columnMapping.DataColumnIndex >= int(types.ColMidiCC0) && columnMapping.DataColumnIndex <= int(types.ColMidiCC8) {
		// Show MIDI CC info with controller number and decimal value
		ccIndex := columnMapping.DataColumnIndex - int(types.ColMidiCC0)
		ccValue := (*phrasesData)[m.CurrentPhrase][m.CurrentRow][types.PhraseColumn(columnMapping.DataColumnIndex)]
		ccNumber := m.MidiCCNumbers[ccIndex]

		if ccValue == -1 {
			statusMsg = fmt.Sprintf("CC%d: -- (Controller=%d)", ccIndex, ccNumber)
		} else {
			statusMsg = fmt.Sprintf("CC%d: %02X (Controller=%d, Value=%d)", ccIndex, ccValue, ccNumber, ccValue)
		}
	} else {
		// On other columns - show basic info
		statusMsg = fmt.Sprintf("Instrument Phrase %02X Row %02X", m.CurrentPhrase, m.CurrentRow)
	}

	if m.IsPlaying {
		if m.PlaybackMode == types.ChainView {
			statusMsg += fmt.Sprintf(" | Chain playing (C:%02X P:%02X) (SPACE to stop)", m.PlaybackChain, m.PlaybackPhrase)
		} else {
			statusMsg += " | Phrase playing (SPACE to stop)"
		}
	} else {
		statusMsg += " | Stopped (SPACE to play)"
	}

	// Add context-sensitive column mode info based on current column
	if m.CurrentCol == int(types.InstrumentColSOMI) {
		if m.SOColumnMode == types.SOModeMIDI {
			statusMsg += " | MIDI mode | Ctrl+Down/Left: Switch to SO"
		} else {
			statusMsg += " | SoundMaker mode | Ctrl+Up/Right: Switch to MI"
		}
	} else if m.CurrentCol == int(types.InstrumentColAR) {
		statusMsg += " | Arpeggio mode"
	} else if m.CurrentCol == int(types.InstrumentColDU) {
		statusMsg += " | Ducking mode"
	}
	return statusMsg
}
