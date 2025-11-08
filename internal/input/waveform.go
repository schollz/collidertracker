package input

import (
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/schollz/collidertracker/internal/getbpm"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/types"
)

// handleW toggles the waveform view for sampler tracks
func handleW(m *model.Model) tea.Cmd {
	// If already in waveform view, return to previous view
	if m.ViewMode == types.WaveformView {
		m.ViewMode = m.WaveformPreviousView
		storage.AutoSave(m)
		return nil
	}
	
	// Only allow waveform view for sampler tracks
	if m.GetPhraseViewType() == types.InstrumentPhraseView {
		log.Printf("Waveform view only available for Sampler tracks")
		return nil
	}
	
	// Get the current file for this track
	file := m.GetCurrentTrackFile()
	if file == "" {
		log.Printf("No audio file for current track")
		return nil
	}
	
	// Make sure the file is absolute path
	if !filepath.IsAbs(file) {
		file = filepath.Join(m.SaveFolder, file)
	}
	
	// Get audio duration
	duration, _, _, err := getbpm.Length(file)
	if err != nil {
		log.Printf("Error getting audio duration: %v", err)
		return nil
	}
	
	// Initialize waveform view state
	m.WaveformPreviousView = m.ViewMode
	m.WaveformFile = file
	m.WaveformStart = 0.0
	m.WaveformEnd = duration
	m.WaveformDuration = duration // Cache duration
	m.WaveformSelectedSlice = -1
	
	// Switch to waveform view
	m.ViewMode = types.WaveformView
	storage.AutoSave(m)
	
	return nil
}

// HandleWaveformInput handles input for waveform view
func HandleWaveformInput(m *model.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "w", "q":
		// Exit waveform view
		m.ViewMode = m.WaveformPreviousView
		storage.AutoSave(m)
		return nil
		
	case "m", " ":
		// Add marker at midpoint
		m.AddWaveformMarker()
		storage.AutoSave(m)
		return nil
		
	case "tab":
		// Select next marker
		m.SelectNextWaveformMarker()
		return nil
		
	case "d", "backspace":
		// Delete selected marker
		m.DeleteSelectedWaveformMarker()
		storage.AutoSave(m)
		return nil
		
	case "left":
		// Jog marker or view left
		if m.WaveformSelectedSlice >= 0 {
			m.JogWaveformMarker(-1, false)
			storage.AutoSave(m)
		} else {
			m.JogWaveformView(-1, false)
		}
		return nil
		
	case "right":
		// Jog marker or view right
		if m.WaveformSelectedSlice >= 0 {
			m.JogWaveformMarker(1, false)
			storage.AutoSave(m)
		} else {
			m.JogWaveformView(1, false)
		}
		return nil
		
	case "shift+left":
		// Fast jog view left (always view, not marker)
		m.JogWaveformView(-1, true)
		return nil
		
	case "shift+right":
		// Fast jog view right (always view, not marker)
		m.JogWaveformView(1, true)
		return nil
		
	case "up":
		// Zoom in
		m.ZoomWaveformView(true)
		return nil
		
	case "down":
		// Zoom out
		m.ZoomWaveformView(false)
		return nil
	}
	
	return nil
}
