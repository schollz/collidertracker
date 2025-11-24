package views

import (
	"fmt"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/ticks"
	"github.com/schollz/collidertracker/internal/types"
)

func RenderChainView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "", "", func(styles *ViewStyles) string {
		var content strings.Builder

		// Render header with chain name on the right (like Phrase View)
		columnHeader := "      PH"
		chainsData := m.GetCurrentChainsData()
		phrasesData := m.GetCurrentPhrasesData()
		totalTicks := ticks.CalculateChainTicks(chainsData, phrasesData, m.CurrentChain)
		chainHeader := fmt.Sprintf("Chain %02X (%d ticks)", m.CurrentChain, totalTicks)
		content.WriteString(RenderHeader(m, columnHeader, chainHeader))

		// Render 16 rows of the current chain
		visibleRows := 16            // Always show all 16 rows of a chain
		chainIndex := m.CurrentChain // We need to track which chain we're viewing

		for row := 0; row < visibleRows; row++ {
			// Row indicator with playback arrow
			rowIndicator := fmt.Sprintf(" %02X ", row)

			// Check if this chain row is currently playing (track-specific)
			playingOnThisRow := false
			if m.IsPlaying {
				if m.PlaybackMode == types.SongView {
					// In song playback, check if this track's chain row is playing
					if m.SongPlaybackActive[m.CurrentTrack] &&
						m.SongPlaybackChain[m.CurrentTrack] == chainIndex &&
						m.SongPlaybackChainRow[m.CurrentTrack] == row {
						playingOnThisRow = true
					}
				} else if m.PlaybackMode == types.ChainView {
					// In chain playback, check if this is the current chain row
					if m.PlaybackChain == chainIndex && m.PlaybackChainRow == row {
						playingOnThisRow = true
					}
				}
			}

			if playingOnThisRow {
				rowIndicator = styles.Playback.Render(fmt.Sprintf("▶%02X ", row))
			}

			content.WriteString(rowIndicator)

			// Get phrase ID for this chain row
			chainsData := m.GetCurrentChainsData()
			phraseID := (*chainsData)[chainIndex][row]
			var phraseCell string

			// Format the phrase ID
			if phraseID == -1 {
				phraseCell = "--"
			} else {
				phraseCell = fmt.Sprintf("%02X", phraseID)
			}

			// Determine cell styling
			isSelected := (m.CurrentRow == row)

			if isSelected {
				// Selected cell
				phraseCell = styles.Selected.Render(phraseCell)
			} else if m.Clipboard.HasData && m.Clipboard.HighlightView == types.ChainView &&
				m.Clipboard.HighlightRow == row {
				// Copied cell
				phraseCell = styles.Copied.Render(phraseCell)
			} else if phraseID == -1 {
				// Empty phrase - dimmed
				phraseCell = styles.Label.Render(phraseCell)
			} else {
				// Normal style
				phraseCell = styles.Normal.Render(phraseCell)
			}

			content.WriteString("  " + phraseCell)
			content.WriteString("\n")
		}

		return content.String()
	}, fmt.Sprintf("arrows: edit | %s+arrows: edit phrase", input.GetModifierKey()), GetChainStatusMessage(m), 17) // 16 rows + 1 for header
}
