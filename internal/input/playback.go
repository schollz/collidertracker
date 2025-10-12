package input

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func TogglePlayback(m *model.Model) tea.Cmd {
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

	if isCurrentTrackPlaying {
		// Current track is playing
		if hasOtherTracksPlaying {
			// Other tracks are playing - queue stop at end of current cell
			m.SongPlaybackQueued[track] = -1
			log.Printf("Queued track %d to stop at cell boundary", track)
		} else {
			// No other tracks playing - stop immediately
			m.SongPlaybackActive[track] = false
			m.SongPlaybackQueued[track] = 0
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
		songRow := m.CurrentRow
		if songRow < 0 || songRow >= 16 {
			log.Printf("Invalid song row %d for single track playback", songRow)
			return nil
		}

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
			anyTrackAtCellBoundary = true
			log.Printf("CELL_BOUNDARY: Song track %d: ticks exhausted, advancing (anyTrackAtCellBoundary=true)", track)

			// Check for queued stop action at cell boundary
			if m.SongPlaybackQueued[track] == -1 {
				// Queued to stop - deactivate track
				m.SongPlaybackActive[track] = false
				m.SongPlaybackQueued[track] = 0
				log.Printf("Song track %d stopped (queued stop executed)", track)
				continue
			}

			// Now advance to next playable row for this track
			if !advanceToNextPlayableRowForTrack(m, track) {
				// Track finished, deactivate
				m.SongPlaybackActive[track] = false
				log.Printf("Song track %d deactivated (end of sequence)", track)
				continue
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
// Returns true if successful, false if track should be stopped
func advanceToNextPlayableRowForTrack(m *model.Model, track int) bool {
	if track < 0 || track >= 8 {
		return false
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
				return true
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
				return true
			}
		}
	}

	// End of chain reached, find next valid song row
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
						return true
					}
				}
			}
		}
	}

	// No valid sequences found, track should stop
	return false
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
