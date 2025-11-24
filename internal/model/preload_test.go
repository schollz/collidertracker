package model

import (
	"testing"

	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestPreloadUpcomingSamples(t *testing.T) {
	m := NewModel(57120, "test-preload.json", false)

	// Set track 4 as sampler track (default is true/sampler)
	assert.True(t, m.TrackTypes[4], "Track 4 should be sampler by default")

	// Add test files to sampler phrases
	m.SamplerPhrasesFiles = []string{
		"/tmp/test1.wav",
		"/tmp/test2.wav",
		"/tmp/test3.wav",
	}

	// Set up phrase 0 with some rows that reference these files
	// Row 0: file 0, DT=4 (playable)
	m.SamplerPhrasesData[0][0][types.ColFilename] = 0
	m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4

	// Row 1: file 1, DT=0 (not playable - should be skipped)
	m.SamplerPhrasesData[0][1][types.ColFilename] = 1
	m.SamplerPhrasesData[0][1][types.ColDeltaTime] = 0

	// Row 2: file 1, DT=4 (playable)
	m.SamplerPhrasesData[0][2][types.ColFilename] = 1
	m.SamplerPhrasesData[0][2][types.ColDeltaTime] = 4

	// Row 5: file 2, DT=4 (playable)
	m.SamplerPhrasesData[0][5][types.ColFilename] = 2
	m.SamplerPhrasesData[0][5][types.ColDeltaTime] = 4

	// Set playback state for track 4
	m.SongPlaybackPhrase[4] = 0
	m.SongPlaybackRowInPhrase[4] = 0

	// Call PreloadUpcomingSamples - should not panic and should handle the data correctly
	// Note: This will try to send OSC messages which may fail without a server, but should not panic
	m.PreloadUpcomingSamples(4, 8)

	// Test edge cases
	// Invalid track - should return safely
	m.PreloadUpcomingSamples(-1, 8)
	m.PreloadUpcomingSamples(8, 8)

	// Instrument track - should return safely (no files to preload)
	m.TrackTypes[0] = false // Set track 0 to instrument
	m.PreloadUpcomingSamples(0, 8)

	// Test with nil phrases files (edge case)
	oldFiles := m.SamplerPhrasesFiles
	m.SamplerPhrasesFiles = nil
	m.PreloadUpcomingSamples(4, 8) // Should handle gracefully
	m.SamplerPhrasesFiles = oldFiles
}

func TestGetPhrasesFilesForTrack(t *testing.T) {
	m := NewModel(0, "", false)

	// Test sampler track (default is sampler, TrackTypes[i] = true)
	files := m.GetPhrasesFilesForTrack(4)
	assert.NotNil(t, files, "Sampler track should return files")
	assert.Equal(t, &m.SamplerPhrasesFiles, files, "Should return SamplerPhrasesFiles")

	// Test instrument track
	m.TrackTypes[2] = false // Set track 2 to instrument (false = Instrument)
	files = m.GetPhrasesFilesForTrack(2)
	assert.Nil(t, files, "Instrument track should return nil (no files)")

	// Test invalid tracks
	files = m.GetPhrasesFilesForTrack(-1)
	assert.NotNil(t, files, "Invalid track defaults to sampler")

	files = m.GetPhrasesFilesForTrack(8)
	assert.NotNil(t, files, "Invalid track defaults to sampler")
}

func TestSendOSCPreloadMessage(t *testing.T) {
	m := NewModel(57120, "test-osc-preload.json", false)

	// Test sending preload message (should not panic even if OSC server isn't running)
	m.SendOSCPreloadMessage("/tmp/test.wav")

	// Test with relative path (should convert to absolute)
	m.SendOSCPreloadMessage("test.wav")

	// Test with nil OSC client (should return early)
	m.oscClient = nil
	m.SendOSCPreloadMessage("/tmp/test.wav") // Should not panic
}
