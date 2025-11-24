package views

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/supercollider"
	"github.com/schollz/collidertracker/internal/types"
)

// stripAnsiCodes removes ANSI escape codes from a string to measure actual text width
func stripAnsiCodes(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

func GetSoundMakerStatusMessage(m *model.Model) string {
	settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

	var columnStatus string

	// Check if we're on the name row
	if m.CurrentRow == 0 {
		columnStatus = fmt.Sprintf("SoundMaker: %s", settings.Name)
	} else {
		// Check if we're on a parameter row
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			// Get parameters for the current column
			col0, col1 := def.GetParametersSortedByColumn()
			var params []types.InstrumentParameterDef
			if m.CurrentCol == 0 {
				params = col0
			} else {
				params = col1
			}

			// Parameter rows start at row 1
			paramIndex := m.CurrentRow - 1
			if paramIndex >= 0 && paramIndex < len(params) {
				param := params[paramIndex]
				value := settings.GetParameterValue(param.Key)

				// Special handling for DX7 preset display
				if param.Key == "preset" && settings.Name == "DX7" {
					patchName, err := supercollider.GetDX7PatchName(int(value))
					if err == nil {
						columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, patchName, value)
					} else {
						columnStatus = fmt.Sprintf("%s: %.0f", param.DisplayName, value)
					}
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special handling for MiBraids model display
					modelName := types.GetMiBraidsModelName(int(value))
					columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, modelName, value)
				} else if param.Key == "engine" && settings.Name == "MiPlaits" {
					// Special handling for MiPlaits engine display
					engineName := types.GetMiPlaitsEngineName(int(value))
					columnStatus = fmt.Sprintf("%s: %s (%.0f)", param.DisplayName, engineName, value)
				} else {
					// Standard parameter display
					// Use DisplayFormatter if available, otherwise use DisplayFormat or default formatting
					if param.DisplayFormatter != nil {
						formattedValue := param.DisplayFormatter(value)
						columnStatus = fmt.Sprintf("%s: %s", param.DisplayName, formattedValue)
					} else if param.DisplayFormat != "" {
						formattedValue := fmt.Sprintf(param.DisplayFormat, value)
						columnStatus = fmt.Sprintf("%s: %s", param.DisplayName, formattedValue)
					} else if param.Type == types.ParameterTypeHex {
						columnStatus = fmt.Sprintf("%s: %02X", param.DisplayName, int(value))
					} else if param.Type == types.ParameterTypeFloat {
						columnStatus = fmt.Sprintf("%s: %.2f", param.DisplayName, value)
					} else {
						columnStatus = fmt.Sprintf("%s: %.0f", param.DisplayName, value)
					}
				}
			} else {
				columnStatus = "Use Up/Down to navigate parameters"
			}
		} else {
			columnStatus = "Unknown SoundMaker"
		}
	}

	return columnStatus
}

func RenderSoundMakerView(m *model.Model) string {
	statusMsg := GetSoundMakerStatusMessage(m)
	return renderViewWithCommonPattern(m, "SoundMaker Settings", fmt.Sprintf("SoundMaker %02X", m.SoundMakerEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current SoundMaker settings
		settings := m.SoundMakerSettings[m.SoundMakerEditingIndex]

		// Initialize parameters if needed
		settings.InitializeParameters()

		// Always show Name row first
		var nameCell string
		if m.CurrentRow == 0 {
			nameCell = styles.Selected.Render(settings.Name)
		} else {
			nameCell = styles.Normal.Render(settings.Name)
		}
		content.WriteString(fmt.Sprintf("  %-12s %s\n", styles.Label.Render("Name:"), nameCell))

		// Show description if available
		if def, exists := types.GetInstrumentDefinition(settings.Name); exists && def.Description != "" {
			content.WriteString(fmt.Sprintf("  %-12s %s\n", styles.Label.Render("Description:"), styles.Normal.Render(def.Description)))
		}
		content.WriteString("\n")

		// Get instrument definition and render parameters in single column
		// Always reserve space for maximum parameters (7) to keep stable height
		content.WriteString("\n")

		if def, exists := types.GetInstrumentDefinition(settings.Name); exists {
			// Get parameters sorted by column
			col0, col1 := def.GetParametersSortedByColumn()

			// Helper function to render a parameter
			renderParam := func(param types.InstrumentParameterDef, paramIndex int, currentCol int) string {
				value := settings.GetParameterValue(param.Key)
				var valueStr string

				// Special formatting for DX7 preset
				if param.Key == "preset" && settings.Name == "DX7" {
					patchName, err := supercollider.GetDX7PatchName(int(value))
					if err == nil {
						valueStr = fmt.Sprintf("%s", patchName)
					} else {
						valueStr = fmt.Sprintf("%.0f", value)
					}
				} else if param.Key == "model" && settings.Name == "MiBraids" {
					// Special formatting for MiBraids model
					modelName := types.GetMiBraidsModelName(int(value))
					valueStr = fmt.Sprintf("%s", modelName)
				} else if param.Key == "engine" && settings.Name == "MiPlaits" {
					// Special formatting for MiPlaits engine
					engineName := types.GetMiPlaitsEngineName(int(value))
					valueStr = fmt.Sprintf("%s", engineName)
				} else {
					// Use DisplayFormatter if available, otherwise use DisplayFormat or default formatting
					if param.DisplayFormatter != nil {
						valueStr = param.DisplayFormatter(value)
					} else if param.DisplayFormat != "" {
						valueStr = fmt.Sprintf(param.DisplayFormat, value)
					} else if param.Type == types.ParameterTypeHex {
						valueStr = fmt.Sprintf("%02X", int(value))
					} else if param.Type == types.ParameterTypeFloat {
						// Display float parameters directly
						valueStr = fmt.Sprintf("%.2f", value)
					} else {
						valueStr = fmt.Sprintf("%.0f", value)
					}
				}

				var valueCell string
				// paramIndex is 1-based (row 1, 2, 3...), row 0 is the name
				if m.CurrentRow == paramIndex && m.CurrentCol == currentCol {
					valueCell = styles.Selected.Render(valueStr)
				} else {
					valueCell = styles.Normal.Render(valueStr)
				}

				return fmt.Sprintf("  %-10s %s", styles.Label.Render(param.DisplayName+":"), valueCell)
			}

			// Render parameters in two columns side by side
			maxRows := len(col0)
			if len(col1) > maxRows {
				maxRows = len(col1)
			}

			const leftColWidth = 40 // Fixed width for left column

			for i := 0; i < maxRows; i++ {
				var leftCol, rightCol string

				// Render left column (column 0)
				if i < len(col0) {
					leftCol = renderParam(col0[i], i+1, 0)
				} else {
					leftCol = ""
				}

				// Render right column (column 1)
				if i < len(col1) {
					rightCol = renderParam(col1[i], i+1, 1)
				} else {
					rightCol = ""
				}

				// Pad left column to fixed width for proper alignment
				leftColPadded := leftCol
				// Strip ANSI codes to measure actual text width
				leftColStripped := stripAnsiCodes(leftCol)
				if len(leftColStripped) < leftColWidth {
					leftColPadded = leftCol + strings.Repeat(" ", leftColWidth-len(leftColStripped))
				}

				content.WriteString(leftColPadded)
				if rightCol != "" {
					content.WriteString(rightCol)
				}
				content.WriteString("\n")
			}

			// Add empty rows to maintain consistent height (max parameters is 9)
			const maxParameters = 9
			for i := maxRows; i < maxParameters; i++ {
				content.WriteString("\n") // Empty row for consistent spacing
			}
		} else {
			// If no instrument definition found, add empty rows to maintain height
			const maxParameters = 9
			for i := 0; i < maxParameters; i++ {
				content.WriteString("\n")
			}
		}

		return content.String()
	}, fmt.Sprintf("arrows: navigate | space: select | %s+arrows: adjust", input.GetModifierKey()), statusMsg, 15) // Fixed height for stable view
}
