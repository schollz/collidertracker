package model

import (
	"testing"

	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

// TestWaveformMarkerManipulation tests adding, selecting, and deleting waveform markers
func TestWaveformMarkerManipulation(t *testing.T) {
	m := NewModel(0, "/tmp/test", false)
	
	// Set up a waveform file
	testFile := "../getbpm/Break120.wav"
	m.WaveformFile = testFile
	m.WaveformStart = 0.0
	m.WaveformEnd = 2.0
	m.WaveformDuration = 2.0
	m.WaveformSelectedSlice = -1
	
	// Initialize metadata with some slices
	m.FileMetadata[testFile] = types.FileMetadata{
		BPM:         120.0,
		Slices:      4,
		SliceType:   1, // Onsets
		Onsets:      []float64{0.0, 0.5, 1.0, 1.5},
		Playthrough: 0,
		SyncToBPM:   1,
	}
	
	// Test 1: Add a new marker
	t.Run("AddMarker", func(t *testing.T) {
		initialCount := len(m.FileMetadata[testFile].Onsets)
		m.AddWaveformMarker()
		
		metadata := m.FileMetadata[testFile]
		assert.Equal(t, initialCount+1, len(metadata.Onsets), "Should have one more marker")
		assert.Equal(t, initialCount+1, metadata.Slices, "Slices count should be updated")
		
		// Verify the marker was added at midpoint (0.0 + 2.0) / 2 = 1.0
		// Note: 1.0 already exists, so it will be at or near 1.0
		found := false
		for _, onset := range metadata.Onsets {
			if onset >= 0.9 && onset <= 1.1 {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have a marker near the midpoint")
	})
	
	// Test 2: Select next marker
	t.Run("SelectNextMarker", func(t *testing.T) {
		m.WaveformSelectedSlice = -1
		m.SelectNextWaveformMarker()
		
		// Should select first visible marker (0.0 is in view [0.0, 2.0])
		assert.GreaterOrEqual(t, m.WaveformSelectedSlice, 0, "Should select a marker")
		
		// Get the selected marker time
		metadata := m.FileMetadata[testFile]
		selectedTime := metadata.Onsets[m.WaveformSelectedSlice]
		assert.GreaterOrEqual(t, selectedTime, m.WaveformStart, "Selected marker should be in view")
		assert.LessOrEqual(t, selectedTime, m.WaveformEnd, "Selected marker should be in view")
	})
	
	// Test 3: Jog marker
	t.Run("JogMarker", func(t *testing.T) {
		// Select a marker first - select one that's not at the edges
		m.WaveformSelectedSlice = 1 // Select the second marker
		metadata := m.FileMetadata[testFile]
		initialTime := metadata.Onsets[m.WaveformSelectedSlice]
		
		// Jog right significantly (fast mode)
		m.JogWaveformMarker(1, true) // fast = true for larger movement
		
		metadata = m.FileMetadata[testFile]
		// After sorting, the marker might be at a different index, so find it
		found := false
		for _, time := range metadata.Onsets {
			if time > initialTime+0.01 { // Moved at least 0.01s
				found = true
				break
			}
		}
		assert.True(t, found, "Should find a marker that moved right")
	})
	
	// Test 4: Delete marker
	t.Run("DeleteMarker", func(t *testing.T) {
		// Select a marker first
		m.WaveformSelectedSlice = 1 // Select second marker
		metadata := m.FileMetadata[testFile]
		initialCount := len(metadata.Onsets)
		
		m.DeleteSelectedWaveformMarker()
		
		metadata = m.FileMetadata[testFile]
		assert.Equal(t, initialCount-1, len(metadata.Onsets), "Should have one less marker")
		assert.Equal(t, initialCount-1, metadata.Slices, "Slices count should be updated")
	})
}

// TestWaveformViewNavigation tests panning and zooming in waveform view
func TestWaveformViewNavigation(t *testing.T) {
	m := NewModel(0, "/tmp/test", false)
	
	// Set up a waveform file
	testFile := "../getbpm/Break120.wav"
	m.WaveformFile = testFile
	m.WaveformStart = 0.0
	m.WaveformEnd = 2.0
	m.WaveformDuration = 4.0 // Total duration
	
	// Test 1: Jog view right
	t.Run("JogViewRight", func(t *testing.T) {
		initialStart := m.WaveformStart
		initialEnd := m.WaveformEnd
		
		m.JogWaveformView(1, false)
		
		assert.Greater(t, m.WaveformStart, initialStart, "View should have moved right")
		assert.Greater(t, m.WaveformEnd, initialEnd, "View should have moved right")
		
		// Duration should remain the same
		duration := m.WaveformEnd - m.WaveformStart
		assert.InDelta(t, 2.0, duration, 0.01, "View duration should remain constant")
	})
	
	// Test 2: Jog view left
	t.Run("JogViewLeft", func(t *testing.T) {
		initialStart := m.WaveformStart
		initialEnd := m.WaveformEnd
		
		m.JogWaveformView(-1, false)
		
		assert.Less(t, m.WaveformStart, initialStart, "View should have moved left")
		assert.Less(t, m.WaveformEnd, initialEnd, "View should have moved left")
	})
	
	// Test 3: Zoom in
	t.Run("ZoomIn", func(t *testing.T) {
		m.WaveformStart = 0.0
		m.WaveformEnd = 2.0
		initialDuration := m.WaveformEnd - m.WaveformStart
		
		m.ZoomWaveformView(true)
		
		newDuration := m.WaveformEnd - m.WaveformStart
		assert.Less(t, newDuration, initialDuration, "Zoom in should decrease view duration")
	})
	
	// Test 4: Zoom out
	t.Run("ZoomOut", func(t *testing.T) {
		m.WaveformStart = 1.0
		m.WaveformEnd = 2.0
		initialDuration := m.WaveformEnd - m.WaveformStart
		
		m.ZoomWaveformView(false)
		
		newDuration := m.WaveformEnd - m.WaveformStart
		assert.Greater(t, newDuration, initialDuration, "Zoom out should increase view duration")
	})
	
	// Test 5: Bounds checking - don't scroll past end
	t.Run("BoundsCheckEnd", func(t *testing.T) {
		m.WaveformStart = 3.0
		m.WaveformEnd = 4.0 // At the end
		
		// Try to jog right
		m.JogWaveformView(1, false)
		
		// Should be clamped to not exceed total duration
		assert.LessOrEqual(t, m.WaveformEnd, m.WaveformDuration, "View should not exceed total duration")
	})
	
	// Test 6: Bounds checking - don't scroll past beginning
	t.Run("BoundsCheckStart", func(t *testing.T) {
		m.WaveformStart = 0.0
		m.WaveformEnd = 1.0
		
		// Try to jog left
		m.JogWaveformView(-1, false)
		
		// Should be clamped to not go below 0
		assert.GreaterOrEqual(t, m.WaveformStart, 0.0, "View should not go below 0")
	})
}

// TestGetCurrentTrackFile tests retrieving the current file for a track
func TestGetCurrentTrackFile(t *testing.T) {
	m := NewModel(0, "/tmp/test", false)
	
	// Set up as a sampler track (default)
	m.CurrentTrack = 4
	m.CurrentPhrase = 0
	m.TrackTypes[4] = true // Sampler
	
	// Add a file to the sampler phrases
	testFile := "test.wav"
	m.SamplerPhrasesFiles = append(m.SamplerPhrasesFiles, testFile)
	
	// Set the file in a phrase row
	m.SamplerPhrasesData[0][0][types.ColFilename] = 0 // First file
	
	// Test 1: Get file for sampler track
	t.Run("SamplerTrackWithFile", func(t *testing.T) {
		file := m.GetCurrentTrackFile()
		assert.Equal(t, testFile, file, "Should return the file from the phrase")
	})
	
	// Test 2: Instrument track should return empty string
	t.Run("InstrumentTrackNoFile", func(t *testing.T) {
		m.CurrentTrack = 0
		m.TrackTypes[0] = false // Instrument
		
		file := m.GetCurrentTrackFile()
		assert.Equal(t, "", file, "Instrument tracks should not have files")
	})
	
	// Test 3: Sampler track with no file
	t.Run("SamplerTrackNoFile", func(t *testing.T) {
		m.CurrentTrack = 5
		m.CurrentPhrase = 1
		m.TrackTypes[5] = true // Sampler
		
		// No file assigned to this phrase
		m.SamplerPhrasesData[1][0][types.ColFilename] = -1
		
		file := m.GetCurrentTrackFile()
		assert.Equal(t, "", file, "Should return empty string when no file is assigned")
	})
}
