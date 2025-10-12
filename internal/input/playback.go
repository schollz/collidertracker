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

// ToggleTrackPlayback handles track-specific playback toggling in song mode (spacebar)
func ToggleTrackPlayback(m *model.Model, track, row int) tea.Cmd {
	// Validate inputs
	if track < 0 || track >= 8 || row < 0 || row >= 16 {
		log.Printf("Invalid track/row for ToggleTrackPlayback: track=%d, row=%d", track, row)
		return nil
	}

	// Check if this is the currently playing cell for this track
	isCurrentlyPlayingCell := m.SongPlaybackActive[track] && m.SongPlaybackRow[track] == row

	// If track is not active
	if !m.SongPlaybackActive[track] {
		// Case 2: Track is not playing, start playing immediately
		log.Printf("Track %d not active, starting playback from row %02X", track, row)
		
		// Initialize playback state if not already playing anything
		if !m.IsPlaying {
			m.IsPlaying = true
			m.PlaybackMode = types.SongView
			// Reset single-track playback variables
			m.PlaybackPhrase = -1
			m.PlaybackRow = -1
			m.PlaybackChain = -1
			m.PlaybackChainRow = -1
			log.Printf("Starting song playback mode")
		}

		// Start this specific track
		chainID := m.SongData[track][row]
		if chainID == -1 {
			log.Printf("Track %d row %02X has no chain, cannot start", track, row)
			return nil
		}

		// Find first phrase in chain
		firstPhraseID := -1
		firstChainRow := -1
		chainsData := m.GetChainsDataForTrack(track)
		for chainRow := 0; chainRow < 16; chainRow++ {
			if (*chainsData)[chainID][chainRow] != -1 {
				firstPhraseID = (*chainsData)[chainID][chainRow]
				firstChainRow = chainRow
				break
			}
		}

		if firstPhraseID == -1 {
			log.Printf("Track %d chain %02X has no phrases, cannot start", track, chainID)
			return nil
		}

		// Activate track
		m.SongPlaybackActive[track] = true
		m.SongPlaybackRow[track] = row
		m.SongPlaybackChain[track] = chainID
		m.SongPlaybackChainRow[track] = firstChainRow
		m.SongPlaybackPhrase[track] = firstPhraseID
		m.SongPlaybackRowInPhrase[track] = FindFirstNonEmptyRowInPhraseForTrack(m, firstPhraseID, track)
		
		// Clear any queued operations for this track
		m.SongPlaybackQueuedRow[track] = -1
		m.SongPlaybackQueuedStop[track] = false

		// Initialize ticks for this track
		m.LoadTicksLeftForTrack(track)

		// Emit initial row for this track
		EmitRowDataFor(m, firstPhraseID, m.SongPlaybackRowInPhrase[track], track)
		log.Printf("Track %d started at song row %02X, chain %02X, phrase %02X", track, row, chainID, firstPhraseID)

		// Start ticking if this is the first active track
		return Tick(m)
	}

	// Track is active
	if isCurrentlyPlayingCell {
		// Case 1: Pressing spacebar on the currently playing cell - queue stop
		log.Printf("Track %d: queuing stop after current cell (row %02X) finishes", track, row)
		m.SongPlaybackQueuedStop[track] = true
		m.SongPlaybackQueuedRow[track] = -1 // Clear any queued row
		return nil
	} else {
		// Case 3: Track is playing but different cell - queue this cell to play next
		log.Printf("Track %d: queuing row %02X to play after current cell finishes", track, row)
		m.SongPlaybackQueuedRow[track] = row
		m.SongPlaybackQueuedStop[track] = false // Clear any queued stop
		return nil
	}
}


func Tick(m *model.Model) tea.Cmd {
	us := rowDurationMicroseconds(m)
	return tea.Tick(time.Duration(us*1000), func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func AdvancePlayback(m *model.Model) {
	oldRow := m.PlaybackRow

	if m.PlaybackMode == types.SongView {
		// Song playback mode with per-track tick counting
		log.Printf("Song playback advancing - checking %d tracks", 8)
		activeTrackCount := 0

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

			log.Printf("Song track %d: ticks exhausted, advancing", track)

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
		
		// Check if all tracks are now inactive
		allInactive := true
		for track := 0; track < 8; track++ {
			if m.SongPlaybackActive[track] {
				allInactive = false
				break
			}
		}
		
		// If all tracks are inactive, stop playback
		if allInactive {
			log.Printf("All tracks inactive, stopping playback")
			m.IsPlaying = false
			// Clear any queued operations
			for track := 0; track < 8; track++ {
				m.SongPlaybackQueuedRow[track] = -1
				m.SongPlaybackQueuedStop[track] = false
			}
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
	
	// Check if there's a queued operation for this track
	if m.SongPlaybackQueuedStop[track] {
		// Stop this track as queued
		log.Printf("Track %d: executing queued stop", track)
		m.SongPlaybackQueuedStop[track] = false
		m.SongPlaybackQueuedRow[track] = -1
		return false // Signal to deactivate track
	}
	
	if m.SongPlaybackQueuedRow[track] != -1 {
		// Start playing the queued row
		queuedRow := m.SongPlaybackQueuedRow[track]
		log.Printf("Track %d: executing queued row %02X", track, queuedRow)
		m.SongPlaybackQueuedRow[track] = -1
		
		chainID := m.SongData[track][queuedRow]
		if chainID != -1 {
			// Find first phrase in this chain
			for chainRow := 0; chainRow < 16; chainRow++ {
				phraseID := (*chainsData)[chainID][chainRow]
				if phraseID != -1 {
					if findFirstPlayableRowInPhraseForTrack(m, phraseID, track) {
						m.SongPlaybackRow[track] = queuedRow
						m.SongPlaybackChain[track] = chainID
						m.SongPlaybackChainRow[track] = chainRow
						m.SongPlaybackPhrase[track] = phraseID
						log.Printf("Track %d: started queued row %02X, chain %02X", track, queuedRow, chainID)
						return true
					}
				}
			}
		}
		// Queued row was invalid, fall through to normal advancement
		log.Printf("Track %d: queued row %02X was invalid, continuing normal advancement", track, queuedRow)
	}
	
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
