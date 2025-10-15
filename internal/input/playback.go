package input

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func TogglePlayback(m *model.Model) tea.Cmd {
	// If currently playing and trying to start playback from a different context, stop first
	if m.IsPlaying {
		shouldStop := false

		// Check if switching between different playback contexts
		if m.ViewMode == types.ChainView && m.PlaybackMode != types.ChainView {
			shouldStop = true
		} else if m.ViewMode == types.PhraseView && m.PlaybackMode != types.PhraseView {
			shouldStop = true
		} else if m.ViewMode == types.ChainView && m.PlaybackMode == types.ChainView && m.PlaybackChain != m.CurrentChain {
			// Different chain
			shouldStop = true
		} else if m.ViewMode == types.PhraseView && m.PlaybackMode == types.PhraseView && m.PlaybackPhrase != m.CurrentPhrase {
			// Different phrase
			shouldStop = true
		}

		if shouldStop {
			log.Printf("Stopping playback before starting new playback in different context")
			m.IsPlaying = false
			if m.RecordingActive {
				stopRecording(m)
			}
			if m.CurrentlyPlayingFile != "" {
				m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
				m.CurrentlyPlayingFile = ""
			}
			m.SendStopOSC()
			// Reset playback state
			for t := 0; t < 8; t++ {
				m.SongPlaybackActive[t] = false
				m.SongPlaybackQueued[t] = 0
				m.SongPlaybackQueuedRow[t] = -1
			}
		}
	}

	var config PlaybackConfig

	if m.ViewMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: true,         // Start from current selected row/track
			Chain:         -1,           // Not used for song mode
			Phrase:        -1,           // Not used for song mode
			Row:           m.CurrentRow, // Song row
		}
	} else if m.ViewMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: true,           // Start from current chain row position
			Chain:         m.CurrentChain, // Use the chain we're currently viewing
			Phrase:        0,              // Will be determined from chain
			Row:           m.CurrentRow,   // Use current chain row
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: true, // Start from current selected row
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           m.CurrentRow,
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromTop(m *model.Model) tea.Cmd {
	var config PlaybackConfig

	if m.ViewMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: false, // Always start from song row 0
			Chain:         -1,    // Not used for song mode
			Phrase:        -1,    // Not used for song mode
			Row:           0,     // Start from song row 0
		}
	} else if m.ViewMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: false,          // Always start from top/first non-empty
			Chain:         m.CurrentChain, // Use current chain being viewed
			Phrase:        0,              // Will be determined from chain
			Row:           -1,             // Will be determined
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: false, // Start from first non-empty row in phrase
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           -1, // Will be determined
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromTopGlobal(m *model.Model) tea.Cmd {
	// Determine playback mode based on the current view
	var playbackMode types.ViewMode
	if m.ViewMode == types.SongView || m.ViewMode == types.ChainView || m.ViewMode == types.PhraseView {
		playbackMode = m.ViewMode
	} else {
		// Use PreviousView if it's Song, Chain or Phrase, otherwise default to Phrase
		if m.PreviousView == types.SongView || m.PreviousView == types.ChainView || m.PreviousView == types.PhraseView {
			playbackMode = m.PreviousView
		} else {
			// Default to phrase view if no clear editing history
			playbackMode = types.PhraseView
		}
	}

	var config PlaybackConfig

	if playbackMode == types.SongView {
		config = PlaybackConfig{
			Mode:          types.SongView,
			UseCurrentRow: false, // Always start from song row 0
			Chain:         -1,    // Not used for song mode
			Phrase:        -1,    // Not used for song mode
			Row:           0,     // Start from song row 0
		}
	} else if playbackMode == types.ChainView {
		config = PlaybackConfig{
			Mode:          types.ChainView,
			UseCurrentRow: false, // Always start from top/first non-empty
			Chain:         m.CurrentChain,
			Phrase:        0,  // Will be determined from chain
			Row:           -1, // Will be determined
		}
	} else {
		config = PlaybackConfig{
			Mode:          types.PhraseView,
			UseCurrentRow: false, // Start from first non-empty row in phrase
			Chain:         -1,
			Phrase:        m.CurrentPhrase,
			Row:           -1, // Will be determined
		}
	}

	return togglePlaybackWithConfig(m, config)
}

func TogglePlaybackFromLastSongRow(m *model.Model) tea.Cmd {
	// If playback was started from Chain or Phrase view, stop it first
	if m.IsPlaying && (m.PlaybackMode == types.ChainView || m.PlaybackMode == types.PhraseView) {
		log.Printf("Ctrl+Space: Stopping playback that was started from %v view", m.PlaybackMode)
		m.IsPlaying = false
		if m.RecordingActive {
			stopRecording(m)
		}
		if m.CurrentlyPlayingFile != "" {
			m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
			m.CurrentlyPlayingFile = ""
		}
		m.SendStopOSC()
		// Reset playback state
		for t := 0; t < 8; t++ {
			m.SongPlaybackActive[t] = false
			m.SongPlaybackQueued[t] = 0
			m.SongPlaybackQueuedRow[t] = -1
		}
		// Now continue with normal Ctrl+Space behavior (which will start new playback)
	}

	// Always play ALL tracks from the last Song view row, regardless of current view
	config := PlaybackConfig{
		Mode:          types.SongView,
		UseCurrentRow: false,
		Chain:         -1,            // Not used for song mode
		Phrase:        -1,            // Not used for song mode
		Row:           m.LastSongRow, // Start from last selected song row
	}

	return togglePlaybackWithConfigFromCtrlSpace(m, config)
}

// CancelQueuedAction cancels any queued playback action (start, stop, or jump) for the current track
func CancelQueuedAction(m *model.Model) {
	if m.ViewMode != types.SongView {
		return
	}

	track := m.CurrentCol
	if track < 0 || track >= 8 {
		return
	}

	// Check if there's a queued action for this track
	if m.SongPlaybackQueued[track] != 0 {
		log.Printf("ESC: Cancelling queued action for track %d (queued=%d, queuedRow=%d)",
			track, m.SongPlaybackQueued[track], m.SongPlaybackQueuedRow[track])
		m.SongPlaybackQueued[track] = 0
		m.SongPlaybackQueuedRow[track] = -1
	}
}

// ToggleSingleTrackPlayback handles Space key in Song View - affects only current track
func ToggleSingleTrackPlayback(m *model.Model) tea.Cmd {
	if m.ViewMode != types.SongView {
		// Not in song view, use regular playback
		return TogglePlayback(m)
	}

	track := m.CurrentCol
	if track < 0 || track >= 8 {
		log.Printf("Invalid track %d for single track playback", track)
		return nil
	}

	// If playback was started from Chain or Phrase view, stop it first
	if m.IsPlaying && (m.PlaybackMode == types.ChainView || m.PlaybackMode == types.PhraseView) {
		log.Printf("Stopping playback that was started from %v view", m.PlaybackMode)
		m.IsPlaying = false
		if m.RecordingActive {
			stopRecording(m)
		}
		if m.CurrentlyPlayingFile != "" {
			m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
			m.CurrentlyPlayingFile = ""
		}
		m.SendStopOSC()
		// Reset playback state
		for t := 0; t < 8; t++ {
			m.SongPlaybackActive[t] = false
			m.SongPlaybackQueued[t] = 0
			m.SongPlaybackQueuedRow[t] = -1
		}
		// Now continue with normal Space behavior (which will start new playback)
	}

	// Check if current track is playing
	isCurrentTrackPlaying := m.IsPlaying && m.PlaybackMode == types.SongView && m.SongPlaybackActive[track]

	// Check if any other tracks are playing
	hasOtherTracksPlaying := false
	if m.IsPlaying && m.PlaybackMode == types.SongView {
		for t := 0; t < 8; t++ {
			if t != track && m.SongPlaybackActive[t] {
				hasOtherTracksPlaying = true
				break
			}
		}
	}

	songRow := m.CurrentRow
	if songRow < 0 || songRow >= 16 {
		log.Printf("Invalid song row %d for single track playback", songRow)
		return nil
	}

	if isCurrentTrackPlaying {
		// Current track is playing
		// Check if cursor is on the currently playing cell
		isOnPlayingCell := (m.SongPlaybackRow[track] == songRow)

		if !isOnPlayingCell {
			// Track is playing but cursor is on a different cell - JUMP functionality
			// Queue stop for current cell and start for selected cell
			chainID := m.SongData[track][songRow]
			if chainID == -1 {
				log.Printf("Cannot jump: no chain at track %d, row %02X", track, songRow)
				return nil
			}

			// Verify target chain has phrases
			chainsData := m.GetChainsDataForTrack(track)
			hasValidPhrase := false
			for chainRow := 0; chainRow < 16; chainRow++ {
				if (*chainsData)[chainID][chainRow] != -1 {
					hasValidPhrase = true
					break
				}
			}

			if !hasValidPhrase {
				log.Printf("Cannot jump: chain %02X has no phrases for track %d", chainID, track)
				return nil
			}

			// Queue stop at current cell boundary and jump to target row
			m.SongPlaybackQueued[track] = -1
			m.SongPlaybackQueuedRow[track] = songRow // Store jump target
			log.Printf("JUMP: Queued track %d to jump from row %02X to row %02X at cell boundary", track, m.SongPlaybackRow[track], songRow)
		} else if hasOtherTracksPlaying {
			// On currently playing cell with other tracks playing - queue stop at end of current cell
			m.SongPlaybackQueued[track] = -1
			m.SongPlaybackQueuedRow[track] = -1 // Clear jump target (normal stop)
			log.Printf("Queued track %d to stop at cell boundary", track)
		} else {
			// On currently playing cell with no other tracks playing - stop immediately
			m.SongPlaybackActive[track] = false
			m.SongPlaybackQueued[track] = 0
			m.SongPlaybackQueuedRow[track] = -1
			// If this was the last playing track, stop playback entirely
			m.IsPlaying = false
			if m.RecordingActive {
				stopRecording(m)
			}
			if m.CurrentlyPlayingFile != "" {
				m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
				m.CurrentlyPlayingFile = ""
			}
			m.SendStopOSC()
			log.Printf("Stopped track %d immediately (no other tracks playing)", track)
		}
	} else {
		// Current track is not playing
		chainID := m.SongData[track][songRow]
		if chainID == -1 {
			log.Printf("No chain at track %d, row %d", track, songRow)
			return nil
		}

		// Check if chain has valid phrase data
		chainsData := m.GetChainsDataForTrack(track)
		firstPhraseID := -1
		firstChainRow := -1
		for chainRow := 0; chainRow < 16; chainRow++ {
			if (*chainsData)[chainID][chainRow] != -1 {
				firstPhraseID = (*chainsData)[chainID][chainRow]
				firstChainRow = chainRow
				break
			}
		}

		if firstPhraseID == -1 {
			log.Printf("Chain %d has no phrases for track %d", chainID, track)
			return nil
		}

		if hasOtherTracksPlaying {
			// Other tracks are playing - queue start at next cell boundary
			m.SongPlaybackQueued[track] = 1
			m.SongPlaybackQueuedRow[track] = songRow // Store the row to start from
			log.Printf("QUEUE_DEBUG: Queued track %d to start at next cell boundary from row %02X (other tracks playing)", track, songRow)
		} else {
			// No other tracks playing - start immediately
			// Initialize playback if not already running
			if !m.IsPlaying {
				m.IsPlaying = true
				m.PlaybackMode = types.SongView
				m.PlaybackPhrase = -1
				m.PlaybackRow = -1
				m.PlaybackChain = -1
				m.PlaybackChainRow = -1

				// Initialize increment counters for this track
				for phrase := 0; phrase < 255; phrase++ {
					for row := 0; row < 255; row++ {
						m.IncrementCounters[track][phrase][row] = -1
					}
				}
			}

			m.SongPlaybackActive[track] = true
			m.SongPlaybackQueued[track] = 0
			m.SongPlaybackRow[track] = songRow
			m.SongPlaybackChain[track] = chainID
			m.SongPlaybackChainRow[track] = firstChainRow
			m.SongPlaybackPhrase[track] = firstPhraseID
			m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)

			// Initialize ticks for this track
			m.LoadTicksLeftForTrack(track)

			// Emit initial row for this track
			EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
			log.Printf("Started track %d immediately at row %02X, chain %02X, phrase %02X with %d ticks",
				track, songRow, chainID, firstPhraseID, m.SongPlaybackTicksLeft[track])

			return Tick(m)
		}
	}

	return nil
}

func Tick(m *model.Model) tea.Cmd {
	us := rowDurationMicroseconds(m)
	return tea.Tick(time.Duration(us*1000), func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AdvancePlayback(m *model.Model) {
	oldRow := m.PlaybackRow

	// Increment tick counter for blinking indicators
	m.TickCount++

	if m.PlaybackMode == types.SongView {
		// Song playback mode with per-track tick counting
		log.Printf("Song playback advancing - checking %d tracks", 8)
		activeTrackCount := 0
		anyTrackAtCellBoundary := false // Track if any track reached a cell boundary this tick

		for track := 0; track < 8; track++ {
			if !m.SongPlaybackActive[track] {
				continue
			}
			activeTrackCount++
			log.Printf("DEBUG_SONG: Processing active track %d, ticksLeft=%d", track, m.SongPlaybackTicksLeft[track])

			// Decrement ticks for this track if > 0
			if m.SongPlaybackTicksLeft[track] > 0 {
				m.SongPlaybackTicksLeft[track]--
				log.Printf("Song track %d: %d ticks remaining", track, m.SongPlaybackTicksLeft[track])
			}

			// Only advance when ticks reach 0
			if m.SongPlaybackTicksLeft[track] > 0 {
				continue
			}

			// Mark that at least one track reached a cell boundary
			log.Printf("CELL_BOUNDARY: Song track %d: ticks exhausted, advancing (checking if song row changes)", track)

			// Remember the song row before advancing
			oldSongRow := m.SongPlaybackRow[track]

			// Now advance to next playable row for this track
			success, chainLooped := advanceToNextPlayableRowForTrack(m, track)
			if !success {
				// Track finished, deactivate
				m.SongPlaybackActive[track] = false
				m.SongPlaybackQueued[track] = 0 // Clear any queued action
				log.Printf("Song track %d deactivated (end of sequence)", track)
				continue
			}

			// Check if we advanced to a new song row (new chain)
			newSongRow := m.SongPlaybackRow[track]

			// Detect cell boundary: either song row changed OR chain looped back to beginning
			if newSongRow != oldSongRow || chainLooped {
				// Track advanced to a new song row OR chain looped back - this is a song-level cell boundary
				anyTrackAtCellBoundary = true
				if newSongRow != oldSongRow {
					log.Printf("SONG_CELL_BOUNDARY: Song track %d advanced from song row %02X to %02X (anyTrackAtCellBoundary=true)", track, oldSongRow, newSongRow)
				} else {
					log.Printf("SONG_CELL_BOUNDARY: Song track %d chain looped back to beginning at song row %02X (anyTrackAtCellBoundary=true)", track, oldSongRow)
				}

				// Check for queued stop action at SONG cell boundary (after finishing current chain)
				if m.SongPlaybackQueued[track] == -1 {
					jumpTargetRow := m.SongPlaybackQueuedRow[track]
					// Check if this is a jump (target row is set and different from current)
					if jumpTargetRow >= 0 && jumpTargetRow < 16 && jumpTargetRow != newSongRow {
						// This is a jump - queue start at target row instead of stopping
						m.SongPlaybackActive[track] = false
						m.SongPlaybackQueued[track] = 1 // Queue start
						// jumpTargetRow is already set in SongPlaybackQueuedRow
						log.Printf("JUMP_EXEC: Song track %d stopped at row %02X, queued to jump to row %02X at next cell boundary", track, newSongRow, jumpTargetRow)
					} else {
						// Regular queued stop - deactivate track after finishing the chain
						m.SongPlaybackActive[track] = false
						m.SongPlaybackQueued[track] = 0
						m.SongPlaybackQueuedRow[track] = -1
						log.Printf("Song track %d stopped (queued stop executed after chain finished)", track)
					}
					continue
				}
			} else {
				log.Printf("Song track %d advanced within chain (song row %02X unchanged)", track, oldSongRow)
			}

			// Load new ticks for the advanced row
			m.LoadTicksLeftForTrack(track)

			// Emit the newly advanced row immediately (at start of its DT period)
			phraseNum := m.SongPlaybackPhrase[track]
			currentRow := m.SongPlaybackRowInPhrase[track]
			if phraseNum >= 0 && phraseNum < 255 && currentRow >= 0 && currentRow < 255 {
				EmitRowDataFor(m, phraseNum, currentRow, track)
				log.Printf("Song track %d emitted phrase %02X row %d with %d ticks", track, phraseNum, currentRow, m.SongPlaybackTicksLeft[track])
			}
		}
		log.Printf("Song playback: processed %d active tracks", activeTrackCount)

		// Process queued start actions ONLY at cell boundaries (when at least one track advanced)
		log.Printf("QUEUE_CHECK: anyTrackAtCellBoundary=%v, checking queued starts", anyTrackAtCellBoundary)
		if anyTrackAtCellBoundary {
			for track := 0; track < 8; track++ {
				if m.SongPlaybackQueued[track] == 1 && !m.SongPlaybackActive[track] {
					// Queued to start - activate track
					songRow := m.SongPlaybackQueuedRow[track]
					// Validate song row bounds (0-15). This should not occur in normal operation
					// as the row is set from CurrentRow when queuing, but we check defensively.
					if songRow < 0 || songRow >= 16 {
						log.Printf("ERROR: Invalid queued song row %d for track %d (valid range: 0-15) - clearing queue", songRow, track)
						m.SongPlaybackQueued[track] = 0
						continue
					}

					chainID := m.SongData[track][songRow]
					if chainID == -1 {
						m.SongPlaybackQueued[track] = 0
						log.Printf("Cannot start track %d: no chain at row %02X (empty cell)", track, songRow)
						continue
					}

					// Find first phrase in chain
					chainsData := m.GetChainsDataForTrack(track)
					firstPhraseID := -1
					firstChainRow := -1
					for chainRow := 0; chainRow < 16; chainRow++ {
						if (*chainsData)[chainID][chainRow] != -1 {
							firstPhraseID = (*chainsData)[chainID][chainRow]
							firstChainRow = chainRow
							break
						}
					}

					if firstPhraseID == -1 {
						m.SongPlaybackQueued[track] = 0
						log.Printf("Cannot start track %d: chain %d has no phrases", track, chainID)
						continue
					}

					// Activate the track
					m.SongPlaybackActive[track] = true
					m.SongPlaybackQueued[track] = 0
					m.SongPlaybackRow[track] = songRow
					m.SongPlaybackChain[track] = chainID
					m.SongPlaybackChainRow[track] = firstChainRow
					m.SongPlaybackPhrase[track] = firstPhraseID
					m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)

					// Initialize ticks for this track
					m.LoadTicksLeftForTrack(track)

					// Emit initial row for this track
					EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
					log.Printf("QUEUE_EXEC: Song track %d started (queued start executed) at row %02X, chain %02X, phrase %02X with %d ticks",
						track, songRow, chainID, firstPhraseID, m.SongPlaybackTicksLeft[track])
				}
			}
		} // End of anyTrackAtCellBoundary check

		// Check if all tracks are now inactive - stop playback entirely
		allTracksInactive := true
		for track := 0; track < 8; track++ {
			if m.SongPlaybackActive[track] {
				allTracksInactive = false
				break
			}
		}
		if allTracksInactive {
			m.IsPlaying = false
			if m.RecordingActive {
				stopRecording(m)
			}
			if m.CurrentlyPlayingFile != "" {
				m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
				m.CurrentlyPlayingFile = ""
			}
			m.SendStopOSC()
			log.Printf("All tracks inactive - stopped playback")
		}
	} else if m.PlaybackMode == types.ChainView {
		// Chain playback mode - advance through phrases in sequence
		// Find next row with playback enabled (unified DT-based playback)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)

		// Validate PlaybackPhrase is within bounds before accessing array
		if m.PlaybackPhrase >= 0 && m.PlaybackPhrase < 255 {
			for i := m.PlaybackRow + 1; i < 255; i++ {
				// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
				dtValue := (*phrasesData)[m.PlaybackPhrase][i][types.ColDeltaTime]
				if IsRowPlayable(dtValue) {
					m.PlaybackRow = i
					DebugLogRowEmission(m)
					log.Printf("Chain playback advanced from row %d to %d", oldRow, m.PlaybackRow)
					return
				}
			}
		}

		// End of phrase reached, move to next phrase slot in the same chain
		chainsData := GetChainsDataForTrack(m, m.CurrentTrack)
		for i := m.PlaybackChainRow + 1; i < 16; i++ {
			phraseID := (*chainsData)[m.PlaybackChain][i]
			if phraseID != -1 && phraseID >= 0 && phraseID < 255 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = phraseID
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback moved to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}

		// End of chain reached, loop back to first phrase slot in the same chain
		for i := 0; i < 16; i++ {
			phraseID := (*chainsData)[m.PlaybackChain][i]
			if phraseID != -1 && phraseID >= 0 && phraseID < 255 {
				m.PlaybackChainRow = i
				m.PlaybackPhrase = phraseID
				m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)

				// Reset inheritance values when changing phrases would be handled in main

				DebugLogRowEmission(m)
				log.Printf("Chain playback looped back to chain row %d, phrase %d, row %d", m.PlaybackChainRow, m.PlaybackPhrase, m.PlaybackRow)
				return
			}
		}

		// No valid phrases found in this chain - stop playback
		log.Printf("Chain playback stopped - no valid phrases found in chain %d", m.PlaybackChain)
		return
	} else {
		// Phrase-only playback mode
		// Find next row with playback enabled (unified DT-based playback)
		phrasesData := GetPhrasesDataForTrack(m, m.CurrentTrack)
		for i := m.PlaybackRow + 1; i < 255; i++ {
			// Unified DT-based playback: DT > 0 means playable for both instruments and samplers
			dtValue := (*phrasesData)[m.PlaybackPhrase][i][types.ColDeltaTime]
			if IsRowPlayable(dtValue) {
				m.PlaybackRow = i
				DebugLogRowEmission(m)
				log.Printf("Phrase playback advanced from row %d to %d", oldRow, m.PlaybackRow)
				return
			}
		}

		// Loop back to beginning of phrase
		m.PlaybackRow = FindFirstNonEmptyRowInPhrase(m, m.PlaybackPhrase)
		DebugLogRowEmission(m)
		log.Printf("Phrase playback looped from row %d back to %d", oldRow, m.PlaybackRow)
	}
}

// advanceToNextPlayableRowForTrack advances a track to its next playable row
// Returns (success, chainLooped) where:
// - success: true if track advanced to a valid row, false if track should stop
// - chainLooped: true if chain completed and looped back to beginning (even on same song row)
func advanceToNextPlayableRowForTrack(m *model.Model, track int) (bool, bool) {
	if track < 0 || track >= 8 {
		return false, false
	}

	// Try to advance within current phrase first
	phraseNum := m.SongPlaybackPhrase[track]
	if phraseNum >= 0 && phraseNum < 255 {
		phrasesData := GetPhrasesDataForTrack(m, track)
		for i := m.SongPlaybackRowInPhrase[track] + 1; i < 255; i++ {
			dtValue := (*phrasesData)[phraseNum][i][types.ColDeltaTime]
			if dtValue >= 1 {
				m.SongPlaybackRowInPhrase[track] = i
				log.Printf("Song track %d advanced within phrase to row %d", track, i)
				return true, false
			}
		}
	}

	// End of phrase reached, try to advance within current chain
	currentChain := m.SongPlaybackChain[track]
	chainsData := m.GetChainsDataForTrack(track)
	for chainRow := m.SongPlaybackChainRow[track] + 1; chainRow < 16; chainRow++ {
		phraseID := (*chainsData)[currentChain][chainRow]
		if phraseID != -1 {
			// Found next phrase in chain, find its first playable row
			m.SongPlaybackChainRow[track] = chainRow
			m.SongPlaybackPhrase[track] = phraseID
			if findFirstPlayableRowInPhraseForTrack(m, phraseID, track) {
				log.Printf("Song track %d advanced to chain row %d, phrase %02X", track, chainRow, phraseID)
				return true, false
			}
		}
	}

	// End of chain reached, find next valid song row
	// This means the chain has completed - we'll mark this as a loop-back
	startSearchRow := m.SongPlaybackRow[track] + 1
	for searchOffset := 0; searchOffset < 16; searchOffset++ {
		searchRow := (startSearchRow + searchOffset) % 16
		chainID := m.SongData[track][searchRow]

		if chainID != -1 {
			// Check if this chain has any phrases with playable rows
			for chainRow := 0; chainRow < 16; chainRow++ {
				phraseID := (*chainsData)[chainID][chainRow]
				if phraseID != -1 {
					// Found a phrase, check if it has playable rows
					if findFirstPlayableRowInPhraseForTrack(m, phraseID, track) {
						// Valid chain found
						m.SongPlaybackRow[track] = searchRow
						m.SongPlaybackChain[track] = chainID
						m.SongPlaybackChainRow[track] = chainRow
						m.SongPlaybackPhrase[track] = phraseID
						log.Printf("Song track %d advanced to song row %02X, chain %02X", track, searchRow, chainID)
						// Return chainLooped=true since we completed the previous chain
						return true, true
					}
				}
			}
		}
	}

	// No valid sequences found, track should stop
	return false, false
}

// findFirstPlayableRowInPhraseForTrack finds the first playable row in a phrase for a track
// Sets the track's SongPlaybackRowInPhrase and returns true if found
func findFirstPlayableRowInPhraseForTrack(m *model.Model, phraseNum, track int) bool {
	if phraseNum < 0 || phraseNum >= 255 || track < 0 || track >= 8 {
		return false
	}

	phrasesData := GetPhrasesDataForTrack(m, track)
	for row := 0; row < 255; row++ {
		dtValue := (*phrasesData)[phraseNum][row][types.ColDeltaTime]
		if dtValue >= 1 {
			m.SongPlaybackRowInPhrase[track] = row
			return true
		}
	}
	return false
}
