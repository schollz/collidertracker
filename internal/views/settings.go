package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

func RenderSettingsView(m *model.Model) string {
	return renderViewWithCommonPattern(m, "Preferences ", "", func(styles *ViewStyles) string {
		// Column widths
		const globalColWidth = 18
		const inputColWidth = 16

		// Column styles
		columnStyle := lipgloss.NewStyle().
			Width(globalColWidth).
			Align(lipgloss.Left)

		inputColumnStyle := lipgloss.NewStyle().
			Width(inputColWidth).
			Align(lipgloss.Left)

		// Column headers
		var globalHeader, inputHeader string
		if m.CurrentCol == 0 {
			globalHeader = styles.Selected.Render("Global")
		} else {
			globalHeader = styles.Label.Render("Global")
		}
		if m.CurrentCol == 1 {
			inputHeader = styles.Selected.Render("Input")
		} else {
			inputHeader = styles.Label.Render("Input")
		}

		// Create header row
		globalHeaderCell := columnStyle.Render(globalHeader)
		inputHeaderCell := inputColumnStyle.Render(inputHeader)
		headerRow := lipgloss.JoinHorizontal(lipgloss.Top, globalHeaderCell, inputHeaderCell)

		// Global settings (column 0)
		globalSettings := []struct {
			label string
			value string
			row   int
		}{
			{"BPM:", fmt.Sprintf("%.2f", m.BPM), 0},
			{"PPQ:", fmt.Sprintf("%d", m.PPQ), 1},
			{"Pre:", fmt.Sprintf("%.1f dB", m.PregainDB), 2},
			{"Post:", fmt.Sprintf("%.1f dB", m.PostgainDB), 3},
			{"Bias:", fmt.Sprintf("%.1f dB", m.BiasDB), 4},
			{"Sat:", fmt.Sprintf("%.1f dB", m.SaturationDB), 5},
			{"Drive:", fmt.Sprintf("%.1f dB", m.DriveDB), 6},
			{"Tape:", fmt.Sprintf("%.1f%%", m.TapePercent), 7},
			{"Shimmer:", fmt.Sprintf("%.1f%%", m.ShimmerPercent), 8},
		}

		// Input settings (column 1)
		inputSettings := []struct {
			label string
			value string
			row   int
		}{
			{"Input:", fmt.Sprintf("%.1f dB", m.InputLevelDB), 0},
			{"Reverb:", fmt.Sprintf("%.1f%%", m.ReverbSendPercent), 1},
		}

		// Build column content
		var globalRows []string
		var inputRows []string

		maxRows := len(globalSettings)
		if len(inputSettings) > maxRows {
			maxRows = len(inputSettings)
		}

		for i := 0; i < maxRows; i++ {
			// Global column row
			if i < len(globalSettings) {
				setting := globalSettings[i]
				var valueStyle lipgloss.Style
				if m.CurrentCol == 0 && m.CurrentRow == setting.row {
					valueStyle = styles.Selected
				} else {
					valueStyle = styles.Normal
				}
				row := fmt.Sprintf("%-6s %s", styles.Label.Render(setting.label), valueStyle.Render(setting.value))
				globalRows = append(globalRows, row)
			} else {
				globalRows = append(globalRows, "") // Empty row
			}

			// Input column row
			if i < len(inputSettings) {
				setting := inputSettings[i]
				var valueStyle lipgloss.Style
				if m.CurrentCol == 1 && m.CurrentRow == setting.row {
					valueStyle = styles.Selected
				} else {
					valueStyle = styles.Normal
				}
				row := fmt.Sprintf("%-7s %s", styles.Label.Render(setting.label), valueStyle.Render(setting.value))
				inputRows = append(inputRows, row)
			} else {
				inputRows = append(inputRows, "") // Empty row
			}
		}

		// Join rows in each column
		globalColumn := columnStyle.Render(strings.Join(globalRows, "\n"))
		inputColumn := inputColumnStyle.Render(strings.Join(inputRows, "\n"))

		// Join columns horizontally
		columnsRow := lipgloss.JoinHorizontal(lipgloss.Top, globalColumn, inputColumn)

		// Timing info
		beatsPerSecond := float64(m.BPM) / 60.0
		ticksPerSecond := beatsPerSecond * float64(m.PPQ)
		secondsPerTick := 1.0 / ticksPerSecond
		timingInfo := styles.Normal.Render(fmt.Sprintf("Timing: %.3f seconds per row", secondsPerTick))

		// Join everything vertically
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			"", // Empty line after title
			headerRow,
			"", // Empty line after headers
			columnsRow,
			"", // Empty line before timing
			timingInfo,
			"", // Final empty line
		)

		return content
	}, fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey()), "", 15)
}
