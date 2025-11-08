package views

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/schollz/gowaveform"

	"github.com/schollz/collidertracker/internal/getbpm"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

// RenderWaveformView renders the waveform editing view for the current track's file
func RenderWaveformView(m *model.Model) string {
	styles := getCommonStyles()
	
	// Build the view
	var content strings.Builder
	
	// Get file for current track
	file := m.WaveformFile
	if file == "" {
		// Return error message if no file
		content.WriteString(RenderHeader(m, "Waveform View", ""))
		content.WriteString("\n")
		content.WriteString(styles.Label.Render("No audio file for current track"))
		content.WriteString("\n\n")
		content.WriteString(styles.Label.Render("Press 'w' to return"))
		content.WriteString("\n")
		return styles.Container.Render(content.String())
	}
	
	filename := filepath.Base(file)
	header := fmt.Sprintf("Waveform: %s", filename)
	
	// Get file metadata for onsets
	metadata, hasMetadata := m.FileMetadata[file]
	if !hasMetadata {
		// Initialize default metadata
		metadata = types.FileMetadata{
			BPM:         120.0,
			Slices:      16,
			Playthrough: 0,
			SyncToBPM:   1,
			SliceType:   0,
			Onsets:      []float64{},
		}
	}
	
	// Render header with waveform indicator
	content.WriteString(RenderHeader(m, header, ""))
	content.WriteString("\n")
	
	// Calculate available space for waveform
	// Header takes 2 lines, footer takes 5 lines, we want some padding
	headerLines := 2
	footerLines := 5
	paddingLines := 2
	contentHeight := m.TermHeight - headerLines - footerLines - paddingLines
	if contentHeight < 10 {
		contentHeight = 10
	}
	
	// Reserve space for controls (3 lines) and info (2 lines)
	waveformHeight := contentHeight - 5
	if waveformHeight < 5 {
		waveformHeight = 5
	}
	
	waveWidth := m.TermWidth - 4 // account for container padding
	if waveWidth < 40 {
		waveWidth = 40
	}
	
	// Get audio duration
	duration, _, _, err := getbpm.Length(file)
	if err != nil {
		content.WriteString(styles.Label.Render(fmt.Sprintf("Error loading file: %v", err)))
		content.WriteString("\n")
		return styles.Container.Render(content.String())
	}
	
	// Render the waveform with markers
	// Use the waveform file if available (converted for visualization), otherwise use original
	waveformFile := file
	if hasMetadata && metadata.WaveformFile != "" {
		waveformFile = metadata.WaveformFile
	}
	
	waveformStr, err := renderWaveformWithMarkers(waveformFile, waveWidth, waveformHeight, 
		m.WaveformStart, m.WaveformEnd, metadata.Onsets, m.WaveformSelectedSlice)
	if err != nil {
		content.WriteString(styles.Label.Render(fmt.Sprintf("Error rendering waveform: %v", err)))
		content.WriteString("\n")
	} else {
		content.WriteString(waveformStr)
		content.WriteString("\n")
	}
	
	// Display information
	viewDuration := m.WaveformEnd - m.WaveformStart
	content.WriteString(styles.Label.Render(
		fmt.Sprintf("Duration: %.2fs | Viewing: %.2fs - %.2fs (%.2fs) | Slices: %d",
			duration, m.WaveformStart, m.WaveformEnd, viewDuration, len(metadata.Onsets))))
	if m.WaveformSelectedSlice >= 0 && m.WaveformSelectedSlice < len(metadata.Onsets) {
		content.WriteString(styles.Selected.Render(fmt.Sprintf(" | Selected: %.3fs", metadata.Onsets[m.WaveformSelectedSlice])))
	}
	content.WriteString("\n")
	
	// Display controls
	content.WriteString(styles.Label.Render("Controls: m/Space (add slice) | Tab (select) | d/Backspace (delete) | Esc (unselect)"))
	content.WriteString("\n")
	content.WriteString(styles.Label.Render("          ← → (jog) | Shift+← → (fast jog) | ↑ ↓ (zoom) | w (exit)"))
	content.WriteString("\n")
	
	return styles.Container.Render(content.String())
}

// renderWaveformWithMarkers renders a waveform with slice markers overlaid
func renderWaveformWithMarkers(filepath string, width, height int, start, end float64, 
	markers []float64, selectedMarker int) (string, error) {
	
	// Load waveform
	wf, err := gowaveform.LoadWaveform(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to load waveform: %w", err)
	}
	
	// Generate view with current width
	view, err := wf.GenerateView(gowaveform.WaveformOptions{
		Start: start,
		End:   end,
		Width: width,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate view: %w", err)
	}
	
	if view == nil || len(view.Data) == 0 {
		return "No waveform data", nil
	}
	
	// Use 8 vertical segments per character for higher resolution
	const segmentsPerChar = 8
	virtualHeight := height * segmentsPerChar
	
	// Create a higher resolution grid (8 segments per character height)
	grid := make([][]bool, virtualHeight)
	for i := range grid {
		grid[i] = make([]bool, width)
	}
	
	// Find the maximum absolute value for normalization
	var maxAbs int16
	for _, val := range view.Data {
		if val < 0 {
			if -val > maxAbs {
				maxAbs = -val
			}
		} else {
			if val > maxAbs {
				maxAbs = val
			}
		}
	}
	
	if maxAbs == 0 {
		maxAbs = 1 // Prevent division by zero
	}
	
	// Plot each min/max pair
	for i := 0; i < len(view.Data)/2 && i < width; i++ {
		minVal := view.Data[i*2]
		maxVal := view.Data[i*2+1]
		
		// Normalize to virtual height
		center := virtualHeight / 2
		
		minY := center - int(float64(minVal)/float64(maxAbs)*float64(center))
		maxY := center - int(float64(maxVal)/float64(maxAbs)*float64(center))
		
		// Clamp values
		if minY < 0 {
			minY = 0
		}
		if minY >= virtualHeight {
			minY = virtualHeight - 1
		}
		if maxY < 0 {
			maxY = 0
		}
		if maxY >= virtualHeight {
			maxY = virtualHeight - 1
		}
		
		// Ensure minY <= maxY (since we're working in screen coordinates)
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		
		// Fill the column
		for y := minY; y <= maxY; y++ {
			grid[y][i] = true
		}
	}
	
	// Calculate marker positions in pixels
	markerPositions := make(map[int]bool) // x positions of all markers
	selectedMarkerPos := -1                // x position of selected marker
	duration := end - start
	
	for i, markerTime := range markers {
		if markerTime >= start && markerTime <= end {
			// Calculate x position
			xPos := int(float64(width-1) * (markerTime - start) / duration)
			if xPos >= 0 && xPos < width {
				markerPositions[xPos] = true
				if i == selectedMarker {
					selectedMarkerPos = xPos
				}
			}
		}
	}
	
	// Convert high-resolution grid to block characters
	var sb strings.Builder
	centerY := height / 2
	
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorYellow = "\033[33m" // Unselected markers
		colorCyan   = "\033[36m" // Selected marker
	)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Determine if we're in upper or lower half
			var char string
			if y < centerY {
				// Upper half: use lower blocks inverted (hanging from top of cell)
				char = getUpperHalfChar(grid, x, y, segmentsPerChar)
			} else {
				// Lower half: use upper blocks (extending from bottom of cell)
				char = getLowerHalfChar(grid, x, y, segmentsPerChar)
			}
			
			// Apply color if this is a marker position
			if x == selectedMarkerPos {
				sb.WriteString(colorCyan + char + colorReset)
			} else if markerPositions[x] {
				sb.WriteString(colorYellow + char + colorReset)
			} else {
				sb.WriteString(char)
			}
		}
		sb.WriteString("\n")
	}
	
	// Add timestamp ruler
	sb.WriteString(generateTimestampRuler(width, start, end))
	
	return sb.String(), nil
}

// getUpperHalfChar returns block character for upper half of waveform
// Uses upper blocks (measuring down from top of character cell)
func getUpperHalfChar(grid [][]bool, x, y, segmentsPerChar int) string {
	baseY := y * segmentsPerChar
	
	// Find the lowest filled segment (deepest extent into this cell from top)
	lowestFilled := -1
	for i := segmentsPerChar - 1; i >= 0; i-- {
		segY := baseY + i
		if segY < len(grid) && grid[segY][x] {
			lowestFilled = i
			break
		}
	}
	
	// If nothing filled, return empty
	if lowestFilled == -1 {
		return " "
	}
	
	// Use upper blocks that hang from the top
	// lowestFilled ranges from 0 (top) to 7 (bottom of cell)
	// Upper blocks fill from top, so we map based on extent
	extent := lowestFilled + 1 // +1 because index 0 means 1 segment filled
	
	switch extent {
	case 1:
		return "▔" // U+2594 Upper one eighth
	case 2:
		return "🮂" // U+1FB02 Upper one quarter
	case 3:
		return "🮃" // U+1FB03 Upper three eighths
	case 4:
		return "▀" // U+2580 Upper half
	case 5:
		return "🮄" // U+1FB04 Upper five eighths
	case 6:
		return "🮅" // U+1FB05 Upper three quarters
	case 7:
		return "🮆" // U+1FB06 Upper seven eighths
	default: // 8
		return "█" // U+2588 - full cell
	}
}

// getLowerHalfChar returns block character for lower half of waveform
// Uses lower blocks (measuring up from bottom of character cell)
func getLowerHalfChar(grid [][]bool, x, y, segmentsPerChar int) string {
	baseY := y * segmentsPerChar
	
	// Find the highest filled segment (highest extent into this cell from bottom)
	highestFilled := -1
	for i := 0; i < segmentsPerChar; i++ {
		segY := baseY + i
		if segY < len(grid) && grid[segY][x] {
			highestFilled = i
			break
		}
	}
	
	// If nothing filled, return empty
	if highestFilled == -1 {
		return " "
	}
	
	// Use lower blocks that extend from the bottom
	// highestFilled ranges from 0 (top of cell) to 7 (bottom of cell)
	// Lower blocks fill from bottom, so we need to invert
	// If segment 0 (top) is filled, we need a full or near-full block
	// If segment 7 (bottom) is filled, we need just a small bottom block
	extent := segmentsPerChar - highestFilled
	
	switch extent {
	case 1:
		return "▁" // U+2581 - one eighth from bottom
	case 2:
		return "▂" // U+2582
	case 3:
		return "▃" // U+2583
	case 4:
		return "▄" // U+2584
	case 5:
		return "▅" // U+2585
	case 6:
		return "▆" // U+2586
	case 7:
		return "▇" // U+2587
	default: // 8
		return "█" // U+2588 - full cell
	}
}

// generateTimestampRuler creates a timestamp ruler below the waveform
func generateTimestampRuler(width int, start, end float64) string {
	duration := end - start
	
	// Determine the precision based on the duration
	var precision int
	var interval float64
	
	if duration < 0.1 {
		precision = 4
		interval = 0.01
	} else if duration < 1.0 {
		precision = 3
		interval = 0.05
	} else if duration < 10.0 {
		precision = 2
		interval = 0.5
	} else if duration < 60.0 {
		precision = 1
		interval = 2.0
	} else {
		precision = 0
		interval = 10.0
	}
	
	// Calculate number of timestamps to show (aim for ~8-12 timestamps)
	numTimestamps := int(duration / interval)
	if numTimestamps < 5 {
		numTimestamps = 5
		interval = duration / float64(numTimestamps)
	} else if numTimestamps > 15 {
		numTimestamps = 12
		interval = duration / float64(numTimestamps)
	}
	
	var sb strings.Builder
	
	// Create tick marks line
	tickLine := make([]rune, width)
	for i := range tickLine {
		tickLine[i] = ' '
	}
	
	// Create timestamp labels
	timestamps := make(map[int]string)
	
	for i := 0; i <= numTimestamps; i++ {
		time := start + float64(i)*interval
		if time > end {
			time = end
		}
		
		// Calculate position
		pos := int(float64(width-1) * (time - start) / duration)
		if pos >= 0 && pos < width {
			tickLine[pos] = '|'
			
			// Format timestamp based on precision
			var label string
			if precision == 0 {
				label = fmt.Sprintf("%.0f", time)
			} else {
				label = fmt.Sprintf("%.*f", precision, time)
			}
			timestamps[pos] = label
		}
	}
	
	// Write tick line
	sb.WriteString(string(tickLine))
	sb.WriteString("\n")
	
	// Write timestamp labels
	labelLine := make([]rune, width)
	for i := range labelLine {
		labelLine[i] = ' '
	}
	
	for pos, label := range timestamps {
		// Center the label on the tick mark
		startPos := pos - len(label)/2
		if startPos < 0 {
			startPos = 0
		}
		if startPos+len(label) > width {
			startPos = width - len(label)
		}
		
		// Write label
		for i, ch := range label {
			if startPos+i >= 0 && startPos+i < width {
				labelLine[startPos+i] = ch
			}
		}
	}
	
	sb.WriteString(string(labelLine))
	sb.WriteString("\n")
	
	return sb.String()
}
