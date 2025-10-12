package input

import (
	"testing"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestToggleSingleTrackPlayback_JumpCells(t *testing.T) {
	// Test the "jump" functionality: pressing Space on a playing track but different cell.
	
	t.Run("Jump from one cell to another when track is playing", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data with chains at rows 2 and 5
		m.SongData[0][2] = 0 // Track 0, Row 2, Chain 0
		m.SongData[0][5] = 1 // Track 0, Row 5, Chain 1
		
		// Track 0 uses sampler chains and phrases by default
		// Set up sampler chains with valid phrases
		m.SamplerChainsData[0][0] = 0  // Chain 0 has phrase 0
		m.SamplerChainsData[1][0] = 1  // Chain 1 has phrase 1
		
		// Set up phrases with valid DT values (playable rows)
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4  // Phrase 0, row 0 is playable
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4  // Phrase 1, row 0 is playable
		
		// Start playback at row 2, track 0
		m.CurrentCol = 0
		m.CurrentRow = 2
		m.IsPlaying = false
		
		// First press: Start playback at row 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.IsPlaying, "Playback should start")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be active")
		assert.Equal(t, 2, m.SongPlaybackRow[0], "Track 0 should be playing row 2")
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "No queued action initially")
		
		// Move cursor to row 5 (different cell)
		m.CurrentRow = 5
		
		// Second press: Jump to row 5
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.IsPlaying, "Playback should still be running")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
		assert.Equal(t, 2, m.SongPlaybackRow[0], "Track 0 should still be playing row 2 (until cell boundary)")
		assert.Equal(t, -1, m.SongPlaybackQueued[0], "Track 0 should be queued to stop")
		assert.Equal(t, 5, m.SongPlaybackQueuedRow[0], "Jump target should be row 5")
	})
	
	t.Run("Normal stop when pressing Space on currently playing cell", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data
		m.SongData[0][2] = 0 // Track 0, Row 2, Chain 0
		m.SamplerChainsData[0][0] = 0  // Chain 0 has phrase 0
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		
		// Start playback
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be playing")
		
		// Press Space again on same cell (should stop immediately, no other tracks)
		ToggleSingleTrackPlayback(m)
		
		assert.False(t, m.IsPlaying, "Playback should stop")
		assert.False(t, m.SongPlaybackActive[0], "Track 0 should be inactive")
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "No queued action")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "Jump target should be cleared")
	})
	
	t.Run("Queue stop when pressing Space on currently playing cell with other tracks", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data for two tracks
		m.SongData[0][2] = 0 // Track 0, Row 2
		m.SongData[1][3] = 1 // Track 1, Row 3
		m.SamplerChainsData[0][0] = 0
		m.SamplerChainsData[1][0] = 1
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4
		
		// Start both tracks
		m.IsPlaying = true
		m.PlaybackMode = types.SongView
		m.SongPlaybackActive[0] = true
		m.SongPlaybackActive[1] = true
		m.SongPlaybackRow[0] = 2
		m.SongPlaybackRow[1] = 3
		
		// Press Space on track 0's playing cell
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.IsPlaying, "Playback should continue (other track playing)")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
		assert.Equal(t, -1, m.SongPlaybackQueued[0], "Track 0 should be queued to stop")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "No jump target (normal stop)")
		assert.True(t, m.SongPlaybackActive[1], "Track 1 should still be active")
	})
	
	t.Run("Cannot jump to empty cell", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data with chain at row 2 only
		m.SongData[0][2] = 0 // Track 0, Row 2, Chain 0
		// m.SongData[0][5] is empty (contains default -1)
		m.SamplerChainsData[0][0] = 0
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		
		// Start playback
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be playing")
		
		// Try to jump to empty row 5
		m.CurrentRow = 5
		ToggleSingleTrackPlayback(m)
		
		// Should do nothing (stay playing)
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be playing")
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "No queued action (jump failed)")
		assert.Equal(t, 2, m.SongPlaybackRow[0], "Still playing row 2")
	})
}

func TestCancelQueuedAction(t *testing.T) {
	// Test cancelling queued actions with ESC key
	
	t.Run("Cancel queued start action", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data with chains at rows 2 and 5
		m.SongData[0][2] = 0 // Track 0, Row 2, Chain 0
		m.SongData[0][5] = 1 // Track 0, Row 5, Chain 1
		
		// Set up chains with valid phrases
		m.SamplerChainsData[0][0] = 0
		m.SamplerChainsData[1][0] = 1
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4
		
		// Track 1 is playing to create the condition where track 0 will queue
		m.IsPlaying = true
		m.PlaybackMode = types.SongView
		m.SongPlaybackActive[1] = true
		m.SongPlaybackRow[1] = 3
		
		// Start playback on track 0 (will queue because track 1 is playing)
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.Equal(t, 1, m.SongPlaybackQueued[0], "Track 0 should have queued start")
		assert.Equal(t, 2, m.SongPlaybackQueuedRow[0], "Queued row should be 2")
		assert.False(t, m.SongPlaybackActive[0], "Track 0 should not be active yet")
		
		// Cancel the queued action with ESC
		CancelQueuedAction(m)
		
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Queued action should be cancelled")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "Queued row should be cleared")
		assert.False(t, m.SongPlaybackActive[0], "Track 0 should still not be active")
	})
	
	t.Run("Cancel queued stop action", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data
		m.SongData[0][2] = 0
		m.SongData[1][3] = 1
		
		// Set up chains with valid phrases
		m.SamplerChainsData[0][0] = 0
		m.SamplerChainsData[1][0] = 1
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4
		
		// Start playback on track 0
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		// Manually start track 1 (simulating it already playing)
		m.IsPlaying = true
		m.PlaybackMode = types.SongView
		m.SongPlaybackActive[1] = true
		m.SongPlaybackRow[1] = 3
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be playing")
		assert.True(t, m.SongPlaybackActive[1], "Track 1 should be playing")
		
		// Queue stop for track 0 (press space on playing cell with other tracks active)
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.Equal(t, -1, m.SongPlaybackQueued[0], "Track 0 should have queued stop")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "No jump target")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
		
		// Cancel the queued stop with ESC
		CancelQueuedAction(m)
		
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Queued stop should be cancelled")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "Queued row should be cleared")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
	})
	
	t.Run("Cancel queued jump action", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data with chains at rows 2 and 5
		m.SongData[0][2] = 0
		m.SongData[0][5] = 1
		
		// Set up chains with valid phrases
		m.SamplerChainsData[0][0] = 0
		m.SamplerChainsData[1][0] = 1
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4
		
		// Start playback at row 2
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be playing")
		assert.Equal(t, 2, m.SongPlaybackRow[0], "Should be playing row 2")
		
		// Queue jump to row 5
		m.CurrentRow = 5
		ToggleSingleTrackPlayback(m)
		
		assert.Equal(t, -1, m.SongPlaybackQueued[0], "Track 0 should be queued for jump")
		assert.Equal(t, 5, m.SongPlaybackQueuedRow[0], "Jump target should be row 5")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
		
		// Cancel the queued jump with ESC
		CancelQueuedAction(m)
		
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Queued jump should be cancelled")
		assert.Equal(t, -1, m.SongPlaybackQueuedRow[0], "Jump target should be cleared")
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be active")
	})
	
	t.Run("ESC does nothing when no queued action", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data
		m.SongData[0][2] = 0
		m.SamplerChainsData[0][0] = 0
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		
		// Start playback
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should be playing")
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "No queued action")
		
		// Try to cancel (should do nothing)
		CancelQueuedAction(m)
		
		assert.True(t, m.SongPlaybackActive[0], "Track 0 should still be playing")
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Still no queued action")
	})
	
	t.Run("ESC only affects current track", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.SongView
		
		// Set up song data for two tracks
		m.SongData[0][2] = 0
		m.SongData[1][3] = 1
		
		// Set up chains with valid phrases
		m.SamplerChainsData[0][0] = 0
		m.SamplerChainsData[1][0] = 1
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 4
		m.SamplerPhrasesData[1][0][types.ColDeltaTime] = 4
		
		// Manually set track 1 as playing (to force queuing on track 0)
		m.IsPlaying = true
		m.PlaybackMode = types.SongView
		m.SongPlaybackActive[1] = true
		m.SongPlaybackRow[1] = 3
		
		// Queue start on track 0
		m.CurrentCol = 0
		m.CurrentRow = 2
		ToggleSingleTrackPlayback(m)
		
		assert.Equal(t, 1, m.SongPlaybackQueued[0], "Track 0 should have queued start")
		
		// Manually queue stop on track 1 (simulating it being queued)
		m.SongPlaybackQueued[1] = -1
		m.SongPlaybackQueuedRow[1] = -1
		
		// Cancel action on current track (track 0) by setting current col to 0
		m.CurrentCol = 0
		CancelQueuedAction(m)
		
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Track 0 queue should be cancelled")
		assert.Equal(t, -1, m.SongPlaybackQueued[1], "Track 1 queue should be unchanged")
		
		// Now cancel on track 1
		m.CurrentCol = 1
		CancelQueuedAction(m)
		
		assert.Equal(t, 0, m.SongPlaybackQueued[0], "Track 0 queue should remain cancelled")
		assert.Equal(t, 0, m.SongPlaybackQueued[1], "Track 1 queue should now be cancelled")
	})
	
	t.Run("ESC does nothing in non-Song view", func(t *testing.T) {
		m := model.NewModel(0, "test.json", false)
		m.ViewMode = types.PhraseView
		
		// Artificially set a queued action (shouldn't happen in phrase view, but test defensive code)
		m.SongPlaybackQueued[0] = 1
		m.SongPlaybackQueuedRow[0] = 2
		
		// Try to cancel (should do nothing in non-Song view)
		CancelQueuedAction(m)
		
		// Should remain unchanged
		assert.Equal(t, 1, m.SongPlaybackQueued[0], "Queue should be unchanged in non-Song view")
		assert.Equal(t, 2, m.SongPlaybackQueuedRow[0], "Queued row should be unchanged in non-Song view")
	})
}
