package views

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

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
func renderViewWithCommonPattern(m *model.Model, leftHeader, rightHeader string, renderContent func(styles *ViewStyles) string, helpText string, statusMsg string, contentLines int) string {
	styles := getCommonStyles()

	// Content builder - same pattern as working views
	var content strings.Builder

	// Render header (includes waveform) - same as working views
	// Only render if leftHeader or rightHeader is provided, otherwise the view handles it
	if leftHeader != "" || rightHeader != "" {
		content.WriteString(RenderHeader(m, leftHeader, rightHeader))
	}

	// Render view-specific content
	content.WriteString(renderContent(styles))

	// Render footer with navigation help text and status message
	content.WriteString(RenderFooter(m, contentLines, helpText, statusMsg))

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

// RenderNavigationLines renders the standard 3-line navigation display at the top of the status section
// Format:
//   O          ← Shift+Up target (Options)
// S-C-P        ← Current view indicator (highlighted) + Shift+Right target + help text
//   M          ← Shift+Down target (Mixer)
func RenderNavigationLines(m *model.Model, helpText string) string {
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	
	var line1, line2, line3 string
	
	// Determine what appears above and below the current view
	// The logic is:
	// - O (Options/Settings) appears above Song, Chain, Phrase, and sub-views (but not above itself or Mixer)
	// - M (Mixer) appears below Song, Chain, Phrase, sub-views, and Options (but not below itself)
	// - D (File Metadata) appears above File browser
	var topLabel string
	var bottomLabel string
	var highlightPosition int // Position of the highlighted character (0-based)
	var highlightTopLabel bool // Whether to highlight the top label
	var highlightBottomLabel bool // Whether to highlight the bottom label
	
	switch m.ViewMode {
	case types.SongView:
		// Song view: O above S, M below S
		topLabel = "O"
		bottomLabel = "M"
		highlightPosition = 0 // S is at position 0
		
	case types.ChainView:
		// Chain view: O above C, M below C
		topLabel = "O"
		bottomLabel = "M"
		highlightPosition = 2 // C is at position 2 (S-C)
		
	case types.PhraseView:
		// Phrase view: O above P, M below P
		topLabel = "O"
		bottomLabel = "M"
		highlightPosition = 4 // P is at position 4 (S-C-P)
		
	case types.SettingsView:
		// Settings (Options) view: O above, S-C-P in middle, M below
		// Determine position based on PreviousView
		switch m.PreviousView {
		case types.SongView:
			highlightPosition = 0 // S is at position 0
		case types.ChainView:
			highlightPosition = 2 // C is at position 2 (S-C)
		case types.PhraseView:
			highlightPosition = 4 // P is at position 4 (S-C-P)
		default:
			highlightPosition = 0 // Default to S position
		}
		topLabel = "O"
		bottomLabel = "M"
		highlightTopLabel = true // Highlight O in Settings view
		
	case types.MixerView:
		// Mixer view: O above, S-C-P in middle, M below
		// Determine position based on PreviousView
		switch m.PreviousView {
		case types.SongView:
			highlightPosition = 0 // S is at position 0
		case types.ChainView:
			highlightPosition = 2 // C is at position 2 (S-C)
		case types.PhraseView:
			highlightPosition = 4 // P is at position 4 (S-C-P)
		default:
			highlightPosition = 0 // Default to S position
		}
		topLabel = "O"
		bottomLabel = "M"
		highlightBottomLabel = true // Highlight M in Mixer view
		
	case types.FileView:
		// File browser: D above F
		topLabel = "D"
		bottomLabel = ""
		highlightPosition = 6 // F is at position 6 (S-C-P-F)
		
	case types.FileMetadataView:
		// File Metadata view: nothing above/below
		topLabel = ""
		bottomLabel = ""
		highlightPosition = 8 // D is at position 8 (S-C-P-F-D)
		
	case types.RetriggerView, types.TimestrechView, types.ModulateView, 
		types.ArpeggioView, types.MidiView, types.SoundMakerView, types.DuckingView:
		// Sub-views from Phrase: O above sub-view, M below
		topLabel = "O"
		bottomLabel = "M"
		highlightPosition = 6 // Sub-view character is at position 6 (S-C-P-X)
		
	default:
		topLabel = ""
		bottomLabel = ""
		highlightPosition = 0
	}
	
	// Build the 3 lines with proper alignment
	// Line 1: Top label aligned with the highlighted character
	if topLabel != "" {
		if highlightTopLabel {
			line1 = strings.Repeat(" ", highlightPosition) + highlightStyle.Render(topLabel)
		} else {
			line1 = strings.Repeat(" ", highlightPosition) + dimStyle.Render(topLabel)
		}
	} else {
		line1 = ""
	}
	
	// Line 2: Current view path (S-C-P or variations)
	line2 = buildNavigationChain(m, highlightStyle, dimStyle, helpText)
	
	// Line 3: Bottom label aligned with the highlighted character
	if bottomLabel != "" {
		if highlightBottomLabel {
			line3 = strings.Repeat(" ", highlightPosition) + highlightStyle.Render(bottomLabel)
		} else {
			line3 = strings.Repeat(" ", highlightPosition) + dimStyle.Render(bottomLabel)
		}
	} else {
		line3 = ""
	}
	
	return line1 + "\n" + line2 + "\n" + line3
}

// buildNavigationChain builds the main navigation chain line (e.g., "S-C-P")
func buildNavigationChain(m *model.Model, highlightStyle, dimStyle lipgloss.Style, helpText string) string {
	var chain string
	
	// Build the navigation chain based on current view
	switch m.ViewMode {
	case types.SongView:
		// S is highlighted, -C-P is shown dimmed
		chain = highlightStyle.Render("S") + dimStyle.Render("-C-P")
		
	case types.ChainView:
		// S-C-P with C highlighted
		chain = dimStyle.Render("S-") + highlightStyle.Render("C") + dimStyle.Render("-P")
		
	case types.PhraseView:
		// S-C-P-X format with P highlighted, X varies by column
		fourthChar := determinePhraseFourthChar(m)
		chain = dimStyle.Render("S-C-") + highlightStyle.Render("P") + dimStyle.Render(fourthChar)
		
	case types.FileView:
		// S-C-P-F with F highlighted (file browser from phrase view)
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("F")
		
	case types.RetriggerView:
		// S-C-P-R with R highlighted
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("R")
		
	case types.TimestrechView:
		// S-C-P-T with T highlighted
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("T")
		
	case types.ModulateView:
		// S-C-P-O with O highlighted (mOdulate)
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("O")
		
	case types.ArpeggioView:
		// S-C-P-A with A highlighted
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("A")
		
	case types.MidiView:
		// S-C-P-I with I highlighted (mIdI)
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("I")
		
	case types.SoundMakerView:
		// S-C-P-S with S highlighted
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("S")
		
	case types.DuckingView:
		// S-C-P-D with D highlighted
		chain = dimStyle.Render("S-C-P-") + highlightStyle.Render("D")
		
	case types.SettingsView:
		// Settings view: Show S-C-P all dimmed (no highlighting in chain)
		// Only the O label is highlighted
		chain = dimStyle.Render("S-C-P")
		
	case types.MixerView:
		// Mixer view: Show S-C-P all dimmed (no highlighting in chain)
		// Only the M label is highlighted
		chain = dimStyle.Render("S-C-P")
		
	case types.FileMetadataView:
		// S-C-P-F-D with D highlighted (File Metadata from File browser)
		chain = dimStyle.Render("S-C-P-F-") + highlightStyle.Render("D")
		
	default:
		chain = highlightStyle.Render("?")
	}
	
	// Add help text starting at column 15 if provided
	if helpText != "" {
		// Calculate padding needed to reach column 15
		// Account for ANSI codes by measuring the rendered width
		chainWidth := lipgloss.Width(chain)
		paddingNeeded := 14 - chainWidth
		if paddingNeeded < 1 {
			paddingNeeded = 1
		}
		return chain + strings.Repeat(" ", paddingNeeded) + helpText
	}
	
	return chain
}

// determinePhraseFourthChar determines the 4th character in the navigation based on the current column in phrase view
func determinePhraseFourthChar(m *model.Model) string {
	phraseViewType := m.GetPhraseViewType()
	
	if phraseViewType == types.InstrumentPhraseView {
		// Instrument phrase view columns
		switch m.CurrentCol {
		case int(types.InstrumentColAR):
			return "-A" // Arpeggio
		case int(types.InstrumentColSOMI):
			// Check the mode to determine if it's MIDI or SoundMaker
			if m.SOColumnMode == types.SOModeMIDI {
				return "-I" // mIdI
			}
			return "-S" // SoundMaker
		case int(types.InstrumentColDU):
			return "-D" // Ducking
		default:
			return "-F" // File browser (default)
		}
	} else {
		// Sampler phrase view columns
		switch m.CurrentCol {
		case int(types.SamplerColRT):
			return "-R" // Retrigger
		case int(types.SamplerColTS):
			return "-T" // Timestretch
		case int(types.SamplerColMO):
			return "-O" // mOdulate
		case int(types.SamplerColDU):
			return "-D" // Ducking
		default:
			return "-F" // File browser (default)
		}
	}
}

// RenderFooter handles the common pattern of filling remaining space and adding navigation + status
func RenderFooter(m *model.Model, contentLines int, helpText string, statusMsg string) string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	var content strings.Builder

	// Calculate how many lines the navigation and status will take
	navLines := 3 // Navigation always takes 3 lines
	statusLines := 0
	if statusMsg != "" {
		statusLines = 1
	}
	footerLines := navLines + statusLines
	
	// Fill remaining space if terminal is larger
	// Account for container padding (4) and footer lines
	maxContentLines := m.TermHeight - 4 - footerLines
	if m.TermHeight > 0 && contentLines < maxContentLines {
		for i := contentLines; i < maxContentLines; i++ {
			content.WriteString("\n")
		}
	}

	// Render navigation lines with help text
	content.WriteString(RenderNavigationLines(m, helpText))
	
	// Add status message if provided
	if statusMsg != "" {
		content.WriteString("\n")
		content.WriteString(statusStyle.Render(statusMsg))
	}

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
