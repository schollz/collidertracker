package views

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

// Common styles used across all views
type ViewStyles struct {
	Selected      lipgloss.Style
	Normal        lipgloss.Style
	Label         lipgloss.Style
	Container     lipgloss.Style
	Playback      lipgloss.Style
	Copied        lipgloss.Style
	Chain         lipgloss.Style
	Slice         lipgloss.Style
	SliceDownbeat lipgloss.Style
	Dir           lipgloss.Style
	AssignedFile  lipgloss.Style
}

// getCommonStyles returns the standard style definitions used across views
func getCommonStyles() *ViewStyles {
	return &ViewStyles{
		Selected:      lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")),
		Normal:        lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Label:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Container:     lipgloss.NewStyle().Padding(1, 2),
		Playback:      lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Copied:        lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")),
		Chain:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Slice:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		SliceDownbeat: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Dir:           lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		AssignedFile:  lipgloss.NewStyle().Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0")),
	}
}

// renderViewWithCommonPattern provides a common structure for rendering views
func renderViewWithCommonPattern(m *model.Model, leftHeader, rightHeader string, renderContent func(styles *ViewStyles) string, statusMsg string, contentLines int) string {
	styles := getCommonStyles()

	// Content builder - same pattern as working views
	var content strings.Builder

	// Render header (includes waveform) - same as working views
	content.WriteString(RenderHeader(m, leftHeader, rightHeader))

	// Render view-specific content
	content.WriteString(renderContent(styles))

	// Render footer with three-line status
	content.WriteString(RenderFooterWithThreeLineStatus(m, contentLines, statusMsg))

	// Apply container padding to entire content - same as working views
	return styles.Container.Render(content.String())
}

// RenderFooterWithThreeLineStatus renders footer with three navigation status lines
func RenderFooterWithThreeLineStatus(m *model.Model, contentLines int, helpText string) string {
	var content strings.Builder

	// Fill remaining space if terminal is larger
	// Adjust for 3 status lines instead of 1
	if m.TermHeight > 0 && contentLines < m.TermHeight-6 { // -6 for container padding + 3 status lines
		for i := contentLines; i < m.TermHeight-6; i++ {
			content.WriteString("\n")
		}
	}

	// Split help text into three lines if needed
	// For now, put all help text on line 2
	content.WriteString(RenderThreeLineStatus(m, "", helpText, ""))

	return content.String()
}

func getRecordingIndicator(m *model.Model) string {
	if m.RecordingActive {
		// Closed red circle for active recording
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("●")
	} else if m.RecordingEnabled {
		// Open red circle for queued recording
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("○")
	}
	// No indicator when recording is disabled
	return ""
}

// RenderHeader renders the common waveform + header pattern used by all views
func RenderHeader(m *model.Model, leftContent, rightContent string) string {
	var content strings.Builder

	// Render waveform
	cellsHigh := (types.WaveformHeight+1)/2 - 1 // round up consistently
	waveWidth := m.TermWidth - 4                // account for container padding
	if waveWidth < 1 {
		waveWidth = 1
	}

	// Select waveform data based on current view and track context
	var waveformData []float64

	// Determine which track's waveform to display
	trackIndex := -1
	switch m.ViewMode {
	case types.SongView:
		// In Song View, use the track under the cursor
		trackIndex = m.CurrentCol
	case types.ChainView, types.PhraseView, types.RetriggerView, types.TimestrechView,
		types.ModulateView, types.ArpeggioView, types.MidiView, types.SoundMakerView,
		types.DuckingView, types.MixerView:
		// In Chain/Phrase/Settings views, use CurrentTrack
		trackIndex = m.CurrentTrack
	}

	// Get the appropriate waveform buffer
	if trackIndex >= 0 && trackIndex < 8 {
		waveformData = m.TrackWaveformBuf[trackIndex]
	} else {
		// Fall back to summed waveform for other views
		waveformData = m.WaveformBuf
	}

	// If no waveform data available, create a simple test pattern to show the waveform area
	if len(waveformData) == 0 {
		// Generate a simple sine wave for display when no OSC data is available
		testLength := waveWidth * 2 / 3
		if testLength < 10 {
			testLength = 10
		}
		waveformData = make([]float64, testLength)
		for i := range waveformData {
			waveformData[i] = 0.5 * math.Sin(2*math.Pi*float64(i)/float64(testLength)*3)
		}
	}

	content.WriteString(RenderWaveform(waveWidth, cellsHigh, waveformData))
	content.WriteString("\n")

	// Build header with recording indicator
	recordingIndicator := getRecordingIndicator(m)

	// Calculate available space for padding (account for container padding)
	availableWidth := m.TermWidth - 4 // Container padding (2 on each side)
	leftLen := lipgloss.Width(leftContent)
	rightLen := lipgloss.Width(rightContent)
	indicatorLen := 0
	if recordingIndicator != "" {
		indicatorLen = 2 // Space + circle
	}

	// Ensure we have enough space
	paddingSize := availableWidth - leftLen - rightLen - indicatorLen
	if paddingSize < 1 {
		paddingSize = 1
	}

	// Build full header
	fullHeader := leftContent
	if rightContent != "" {
		fullHeader += strings.Repeat(" ", paddingSize) + rightContent
	}
	if recordingIndicator != "" {
		fullHeader += " " + recordingIndicator
	}

	content.WriteString(fullHeader)
	content.WriteString("\n")

	return content.String()
}

// getShiftRightDestination returns the Shift+Right navigation label for Phrase view based on current column
func getShiftRightDestination(m *model.Model) string {
	if m.ViewMode != types.PhraseView {
		return ""
	}

	// Get column mapping to determine what data column we're on
	columnMapping := m.GetColumnMapping(m.CurrentCol)
	if columnMapping == nil {
		return "F" // Default to File view
	}

	// Check which data column we're on and return appropriate indicator
	switch columnMapping.DataColumnIndex {
	case int(types.ColRetrigger):
		return "R" // Retrigger view
	case int(types.ColTimestretch):
		return "T" // Timestretch view
	case int(types.ColModulate):
		return "O" // Modulate view (O for mOdulate)
	case int(types.ColArpeggio):
		return "A" // Arpeggio view
	case int(types.ColMidi):
		return "I" // MIDI view (I for mIdI)
	case int(types.ColSoundMaker):
		return "S" // SoundMaker view
	case int(types.ColEffectDucking):
		return "D" // Ducking view
	default:
		return "F" // File view (default for most columns)
	}
}

// getNavigationInfo returns the Shift+Up and Shift+Down navigation labels for a view
func getNavigationInfo(m *model.Model) (shiftUp, shiftDown string) {
	switch m.ViewMode {
	case types.SongView, types.ChainView, types.PhraseView:
		return "B", "M" // Settings (BPM) and Mixer
	case types.SettingsView:
		return "", "" // Shift+Down goes back to previous view, but we don't show it
	case types.MixerView:
		return "", "" // Shift+Up goes back to previous view, but we don't show it
	case types.RetriggerView, types.TimestrechView, types.ModulateView,
		types.ArpeggioView, types.MidiView, types.SoundMakerView, types.DuckingView:
		return "", "" // These are sub-views of Phrase, Shift+Left goes back
	case types.FileView:
		return "", "" // File browser, Shift+Left goes back to phrase
	case types.FileMetadataView:
		return "", "" // Metadata view, Shift+Down goes back to file view
	case types.WaveformView:
		return "", "" // Waveform view
	default:
		return "", ""
	}
}

// getCurrentViewIndicator returns the S-C-P indicator with current view highlighted
// and appends Shift+Right destination if in Phrase view
func getCurrentViewIndicator(m *model.Model) string {
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var s, c, p string

	// Determine which is highlighted
	switch m.ViewMode {
	case types.SongView:
		s = highlightStyle.Render("S")
		c = normalStyle.Render("C")
		p = normalStyle.Render("P")
	case types.ChainView:
		s = normalStyle.Render("S")
		c = highlightStyle.Render("C")
		p = normalStyle.Render("P")
	case types.PhraseView:
		s = normalStyle.Render("S")
		c = normalStyle.Render("C")
		p = highlightStyle.Render("P")
	default:
		// For other views, show unhighlighted
		s = normalStyle.Render("S")
		c = normalStyle.Render("C")
		p = normalStyle.Render("P")
	}

	scpIndicator := s + "-" + c + "-" + p

	// If in Phrase view, append Shift+Right destination
	if m.ViewMode == types.PhraseView {
		shiftRightDest := getShiftRightDestination(m)
		if shiftRightDest != "" {
			scpIndicator += "-" + normalStyle.Render(shiftRightDest)
		}
	}

	return scpIndicator
}

// calculateSpacingForHelpText calculates the spacing needed between indicator and help text
func calculateSpacingForHelpText(indicatorLength, minSpacing int) int {
	if indicatorLength < minSpacing {
		return minSpacing - indicatorLength
	}
	return 2
}

// RenderThreeLineStatus renders three navigation status lines
func RenderThreeLineStatus(m *model.Model, helpLine1, helpLine2, helpLine3 string) string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	var content strings.Builder

	shiftUp, shiftDown := getNavigationInfo(m)
	scpIndicator := getCurrentViewIndicator(m)

	// Minimum spacing between navigation indicator and help text
	const minSpacing = 10

	// Line 1: Shift-up navigation + help text
	line1 := "  " + shiftUp
	if helpLine1 != "" {
		spacing := calculateSpacingForHelpText(len(line1), minSpacing)
		line1 += strings.Repeat(" ", spacing) + helpLine1
	}
	content.WriteString(statusStyle.Render(line1))
	content.WriteString("\n")

	// Line 2: S-C-P indicator + help text
	line2 := scpIndicator
	if helpLine2 != "" {
		// Calculate raw length without ANSI codes
		// For phrase view with shift-right: "S-C-P-X" = 7 characters
		// For other views: "S-C-P" = 5 characters
		rawSCPLen := len("S-C-P")
		if m.ViewMode == types.PhraseView {
			rawSCPLen = len("S-C-P-X") // Account for the extra character
		}
		spacing := calculateSpacingForHelpText(rawSCPLen, minSpacing)
		line2 += statusStyle.Render(strings.Repeat(" ", spacing) + helpLine2)
	}
	content.WriteString(line2)
	content.WriteString("\n")

	// Line 3: Shift-down navigation + help text
	line3 := "  " + shiftDown
	if helpLine3 != "" {
		spacing := calculateSpacingForHelpText(len(line3), minSpacing)
		line3 += strings.Repeat(" ", spacing) + helpLine3
	}
	content.WriteString(statusStyle.Render(line3))

	return content.String()
}

func RenderPhraseView(m *model.Model) string {
	// Route to appropriate sub-view based on track context
	phraseViewType := m.GetPhraseViewType()
	if phraseViewType == types.InstrumentPhraseView {
		return RenderInstrumentPhraseView(m)
	}
	return RenderSamplerPhraseView(m)
}

func GetChainStatusMessage(m *model.Model) string {
	chainsData := m.GetCurrentChainsData()
	phraseID := (*chainsData)[m.CurrentChain][m.CurrentRow]

	var statusMsg string
	if phraseID == -1 {
		statusMsg = fmt.Sprintf("Chain %02X Row %02X: --", m.CurrentChain, m.CurrentRow)
	} else {
		statusMsg = fmt.Sprintf("Chain %02X Row %02X: Phrase %02X", m.CurrentChain, m.CurrentRow, phraseID)
	}

	statusMsg += fmt.Sprintf(" | Shift+Right: Enter phrase | %s+Arrow: Edit phrase", input.GetModifierKey())
	return statusMsg
}

func IsCurrentRowFile(m *model.Model, filename string) bool {
	// Check if this file is assigned to the current fileSelectRow
	phrasesData := m.GetCurrentPhrasesData()
	fileIndex := (*phrasesData)[m.CurrentPhrase][m.FileSelectRow][types.ColFilename]
	phrasesFiles := m.GetCurrentPhrasesFiles()
	if fileIndex >= 0 && fileIndex < len(*phrasesFiles) && (*phrasesFiles)[fileIndex] != "" {
		assignedFile := (*phrasesFiles)[fileIndex]
		fullPath := filepath.Join(m.CurrentDir, filename)
		return assignedFile == filename || assignedFile == fullPath
	}
	return false
}
