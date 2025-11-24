package views

import (
	"fmt"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func GetArpeggioStatusMessage(m *model.Model) string {
	settings := m.ArpeggioSettings[m.ArpeggioEditingIndex]
	currentRow := &settings.Rows[m.CurrentRow]

	var columnStatus string
	switch m.CurrentCol {
	case 0: // DI (Direction) column
		directionText := "--"
		switch types.ArpeggioDirection(currentRow.Direction) {
		case types.ArpeggioDirectionUp:
			directionText = "u-"
		case types.ArpeggioDirectionDown:
			directionText = "d-"
		}
		columnStatus = fmt.Sprintf("Direction %s", directionText)
	case 1: // CO (Count) column
		countText := "--"
		if currentRow.Count != -1 {
			countText = fmt.Sprintf("%02d", currentRow.Count)
		}
		columnStatus = fmt.Sprintf("Count %s", countText)
	case 2: // Divisor (/) column
		divisorText := "--"
		if currentRow.Divisor != -1 {
			divisorText = fmt.Sprintf("%02d", currentRow.Divisor)
		}
		columnStatus = fmt.Sprintf("Divisor /%s", divisorText)
	}

	return columnStatus
}

func RenderArpeggioView(m *model.Model) string {
	statusMsg := GetArpeggioStatusMessage(m)
	return renderViewWithCommonPattern(m, "Arpeggio Settings", fmt.Sprintf("Arpeggio %02X", m.ArpeggioEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Render header for the arpeggio table
		headerRow := fmt.Sprintf("     %-4s %-4s %-4s", styles.Label.Render("DI"), styles.Label.Render("CO"), styles.Label.Render("/"))
		content.WriteString(headerRow)
		content.WriteString("\n")

		// Get current arpeggio settings
		settings := m.ArpeggioSettings[m.ArpeggioEditingIndex]

		// Render 16 rows (00 to 0F), each with its own DI and CO values
		for row := 0; row < 16; row++ {
			// Row label
			rowLabel := fmt.Sprintf("%02X", row)

			// Get DI and CO values for this specific row
			arpeggioRow := settings.Rows[row]

			// Direction (DI) text for this row
			var diText string
			switch types.ArpeggioDirection(arpeggioRow.Direction) {
			case types.ArpeggioDirectionNone:
				diText = "--"
			case types.ArpeggioDirectionUp:
				diText = "u-"
			case types.ArpeggioDirectionDown:
				diText = "d-"
			default:
				diText = "--"
			}

			// Count (CO) text for this row
			coText := "--"
			if arpeggioRow.Count != -1 {
				coText = fmt.Sprintf("%02X", arpeggioRow.Count)
			}

			// Divisor (/) text for this row
			divisorText := "--"
			if arpeggioRow.Divisor != -1 {
				divisorText = fmt.Sprintf("%02X", arpeggioRow.Divisor)
			}

			// Direction (DI) column - selectable if this row and column are selected
			var diCell string
			if m.CurrentRow == row && m.CurrentCol == 0 { // Column 0 for DI
				diCell = styles.Selected.Render(diText)
			} else {
				diCell = styles.Normal.Render(diText)
			}

			// Count (CO) column - selectable if this row and column are selected
			var coCell string
			if m.CurrentRow == row && m.CurrentCol == 1 { // Column 1 for CO
				coCell = styles.Selected.Render(coText)
			} else {
				coCell = styles.Normal.Render(coText)
			}

			// Divisor (/) column - selectable if this row and column are selected
			var divisorCell string
			if m.CurrentRow == row && m.CurrentCol == 2 { // Column 2 for Divisor
				divisorCell = styles.Selected.Render(divisorText)
			} else {
				divisorCell = styles.Normal.Render(divisorText)
			}

			rowData := fmt.Sprintf("  %-4s %-4s %-4s %-4s", styles.Label.Render(rowLabel), diCell, coCell, divisorCell)
			content.WriteString(rowData)
			content.WriteString("\n")
		}

		return content.String()
	}, fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey()), statusMsg, 18) // 16 rows + 1 header + 1 spacing
}
