package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

// getUnicodeBlock returns the appropriate Unicode block character for a fill ratio (0-1)
func getUnicodeBlock(fillRatio float64) string {
	if fillRatio <= 0 {
		return "  " // Empty
	} else if fillRatio <= 0.125 {
		return "▁▁" // 1/8 block
	} else if fillRatio <= 0.25 {
		return "▂▂" // 2/8 block
	} else if fillRatio <= 0.375 {
		return "▃▃" // 3/8 block
	} else if fillRatio <= 0.5 {
		return "▄▄" // 4/8 block
	} else if fillRatio <= 0.625 {
		return "▅▅" // 5/8 block
	} else if fillRatio <= 0.75 {
		return "▆▆" // 6/8 block
	} else if fillRatio <= 0.875 {
		return "▇▇" // 7/8 block
	} else {
		return "██" // Full block
	}
}

// createVerticalBar creates a simple vertical level meter bar with Unicode blocks
func createVerticalBar(currentLevel, setLevel float32, height int, isSelected bool) []string {
	// Convert dB range (-48 to +12) to bar height scale (60 dB range)
	currentPos := (float64(currentLevel) + 48.0) / 60.0 * float64(height)
	setPos := (float64(setLevel) + 48.0) / 60.0 * float64(height)

	// Clamp positions to valid range
	currentPos = math.Max(0, math.Min(float64(height), currentPos))
	setPos = math.Max(0, math.Min(float64(height), setPos))

	lines := make([]string, height)
	profile := termenv.ColorProfile()

	// Choose colors based on selection
	var fillColor, emptyColor colorful.Color
	if isSelected {
		fillColor, _ = colorful.Hex("#FFFFFF")  // White for selected
		emptyColor, _ = colorful.Hex("#808080") // Medium gray for empty parts of selected
	} else {
		fillColor, _ = colorful.Hex("#C0C0C0")  // Light gray for unselected
		emptyColor, _ = colorful.Hex("#404040") // Dark gray for empty parts of unselected
	}

	// Fill from bottom to top (invert the display)
	for row := 0; row < height; row++ {
		displayRow := float64(height - 1 - row) // Invert so 0dB is at top

		var barContent string
		var color colorful.Color

		// Determine what to show at this row
		if math.Abs(displayRow-setPos) < 0.5 && math.Abs(displayRow-currentPos) > 0.5 {
			// Set level marker (horizontal line)
			barContent = "━━"
			color = fillColor // Use same color as fill for consistency
		} else if displayRow < currentPos {
			// Full fill
			color = fillColor
			barContent = "██"
		} else if displayRow >= currentPos && displayRow < currentPos+1 {
			// Partial fill - use Unicode blocks for smooth edges
			partialFill := currentPos - math.Floor(currentPos) // Fractional part

			if partialFill > 0 {
				color = fillColor
				// Use appropriate Unicode block
				barContent = getUnicodeBlock(partialFill)
			} else {
				// Empty
				barContent = "▒▒"
				color = emptyColor
			}
		} else {
			// Empty space
			barContent = "▒▒"
			color = emptyColor
		}

		// Apply color using termenv
		colorHex := color.Hex()
		termColor := profile.Color(colorHex)
		styledContent := termenv.String(barContent).Foreground(termColor).String()

		lines[row] = styledContent
	}

	return lines
}

// dbToHex converts dB value (-48 to +12) to hex (00 to FE)
func dbToHex(db float32) int {
	// Clamp to valid range
	if db < -48.0 {
		db = -48.0
	}
	if db > 12.0 {
		db = 12.0
	}

	// Map -48 to +12 dB (60 dB range) to 0 to 254 (255 values)
	hex := int(((db + 48.0) * 254.0) / 60.0)
	if hex < 0 {
		hex = 0
	}
	if hex > 254 {
		hex = 254
	}
	return hex
}

// hexToDb converts hex value (0 to 254) back to dB (-48 to +12)
func hexToDb(hex int) float32 {
	if hex < 0 {
		hex = 0
	}
	if hex > 254 {
		hex = 254
	}

	// Map 0 to 254 back to -48 to +12 dB
	return ((float32(hex) * 60.0) / 254.0) - 48.0
}

// getMixerStatusMessage returns the status message for mixer view
func getMixerStatusMessage(m *model.Model) string {
	track := m.CurrentMixerTrack
	setLevel := m.TrackSetLevels[track]

	var trackLabel string
	if track == 8 {
		trackLabel = "Input"
	} else {
		trackLabel = fmt.Sprintf("Track %d", track+1)
	}

	statusMsg := fmt.Sprintf("%s: Set %.1fdB (Hex %02X)",
		trackLabel, setLevel, dbToHex(setLevel))
	statusMsg += fmt.Sprintf(" | Left/Right: Select │ %s+Arrow: Adjust │ Shift+Up: Back", input.GetModifierKey())

	return statusMsg
}

// RenderMixerView renders a modern, sleek mixer view with vertical level meters
func RenderMixerView(m *model.Model) string {
	// Column headers (matching song view format)
	columnHeader := "    " // 4 spaces for left padding like song view row numbers
	for track := 0; track < 8; track++ {
		columnHeader += fmt.Sprintf("  T%d", track+1)
	}
	// Add Input track (Track 9, index 8)
	columnHeader += "  In"

	var mixerHeader string
	if m.CurrentMixerTrack == 8 {
		mixerHeader = "Input"
	} else {
		mixerHeader = fmt.Sprintf("Track %d", m.CurrentMixerTrack+1)
	}

	return renderViewWithCommonPattern(m, columnHeader, mixerHeader, func(styles *ViewStyles) string {
		var content strings.Builder

		// Calculate bar height - smaller to fit with waveform
		barHeight := 8
		if m.TermHeight > 25 {
			barHeight = 10
		}

		// Create vertical bars for all tracks (including Input track at index 8)
		trackBars := make([][]string, 9)
		for track := 0; track < 9; track++ {
			isSelected := track == m.CurrentMixerTrack
			trackBars[track] = createVerticalBar(m.TrackVolumes[track], m.TrackSetLevels[track], barHeight, isSelected)
		}

		// Render the vertical bars row by row
		for row := 0; row < barHeight; row++ {
			content.WriteString("    ") // Left padding like song view
			for track := 0; track < 8; track++ {
				content.WriteString("  ") // 2 spaces before each track (like song view)
				content.WriteString(trackBars[track][row])
			}
			// Add Input track (Track 9, index 8) with slightly different spacing
			content.WriteString("  ") // 2 spaces before Input track
			content.WriteString(trackBars[8][row])
			content.WriteString("\n")
		}

		// Current level values row (hex codes)
		content.WriteString("    ")
		for track := 0; track < 8; track++ {
			content.WriteString("  ")
			currentLevel := m.TrackVolumes[track]
			levelHex := fmt.Sprintf("%02X", dbToHex(currentLevel))

			if track == m.CurrentMixerTrack {
				content.WriteString(styles.Selected.Render(levelHex))
			} else {
				content.WriteString(styles.Normal.Render(levelHex))
			}
		}
		// Add Input track (Track 9, index 8) current level
		content.WriteString("  ")
		inputCurrentLevel := m.TrackVolumes[8]
		inputLevelHex := fmt.Sprintf("%02X", dbToHex(inputCurrentLevel))
		if m.CurrentMixerTrack == 8 {
			content.WriteString(styles.Selected.Render(inputLevelHex))
		} else {
			content.WriteString(styles.Normal.Render(inputLevelHex))
		}
		content.WriteString("\n")

		// Set level values row (hex codes)
		content.WriteString("    ")
		for track := 0; track < 8; track++ {
			content.WriteString("  ")
			setLevel := m.TrackSetLevels[track]
			setHex := fmt.Sprintf("%02X", dbToHex(setLevel))

			if track == m.CurrentMixerTrack && m.CurrentMixerRow == 0 {
				content.WriteString(styles.Selected.Render(setHex))
			} else {
				content.WriteString(styles.Label.Render(setHex))
			}
		}
		// Add Input track (Track 9, index 8) set level
		content.WriteString("  ")
		inputSetLevel := m.TrackSetLevels[8]
		inputSetHex := fmt.Sprintf("%02X", dbToHex(inputSetLevel))
		if m.CurrentMixerTrack == 8 && m.CurrentMixerRow == 0 {
			content.WriteString(styles.Selected.Render(inputSetHex))
		} else {
			content.WriteString(styles.Label.Render(inputSetHex))
		}
		content.WriteString("\n\n")

		return content.String()
	}, getMixerStatusMessage(m), 12)
}
