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

	// Render footer
	content.WriteString(RenderFooter(m, contentLines, statusMsg))

	// Apply container padding to entire content - same as working views
	return styles.Container.Render(content.String())
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

// RenderFooter handles the common pattern of filling remaining space and adding status
func RenderFooter(m *model.Model, contentLines int, statusMsg string) string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	var content strings.Builder

	// Fill remaining space if terminal is larger
	if m.TermHeight > 0 && contentLines < m.TermHeight-4 { // -4 for container padding
		for i := contentLines; i < m.TermHeight-4; i++ {
			content.WriteString("\n")
		}
	}

	// Status message
	content.WriteString(statusStyle.Render(statusMsg))

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
