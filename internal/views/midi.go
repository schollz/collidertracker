package views

import (
	"fmt"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func GetMidiStatusMessage(m *model.Model) string {
	settings := m.MidiSettings[m.MidiEditingIndex]

	var columnStatus string
	switch types.MidiSettingsRow(m.CurrentRow) {
	case types.MidiSettingsRowDevice: // MIDI Device row
		columnStatus = fmt.Sprintf("MIDI Device: %s", settings.Device)
	case types.MidiSettingsRowChannel: // MIDI Channel row
		columnStatus = fmt.Sprintf("MIDI Channel: %s", settings.Channel)
	default:
		// Device selection rows
		if m.CurrentRow >= 2 && m.CurrentRow-2+m.ScrollOffset < len(m.AvailableMidiDevices) {
			deviceIndex := m.CurrentRow - 2 + m.ScrollOffset
			columnStatus = fmt.Sprintf("Select Device: %s", m.AvailableMidiDevices[deviceIndex])
		} else {
			columnStatus = "Available MIDI Devices"
		}
	}

	return columnStatus
}

func RenderMidiView(m *model.Model) string {
	statusMsg := GetMidiStatusMessage(m)
	return renderViewWithCommonPattern(m, "MIDI Settings", fmt.Sprintf("MIDI %02X", m.MidiEditingIndex), func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current MIDI settings
		settings := m.MidiSettings[m.MidiEditingIndex]

		// Settings rows with common rendering pattern
		settingsRows := []struct {
			label string
			value string
			row   int
		}{
			{"Device:", settings.Device, int(types.MidiSettingsRowDevice)},
			{"Channel:", settings.Channel, int(types.MidiSettingsRowChannel)},
		}

		for _, setting := range settingsRows {
			var valueCell string
			if m.CurrentRow == setting.row {
				valueCell = styles.Selected.Render(setting.value)
			} else {
				valueCell = styles.Normal.Render(setting.value)
			}
			row := fmt.Sprintf("  %-8s %s", styles.Label.Render(setting.label), valueCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		// Separator line
		content.WriteString("\n")
		content.WriteString(styles.Label.Render("Available MIDI Devices:"))
		content.WriteString("\n\n")

		// Available MIDI devices list (scrollable)
		visibleRows := m.GetVisibleRows() - 6 // Reserve space for header, settings, and labels
		deviceStartRow := 2                   // Devices start at row 2 (after Device and Channel settings)

		for i := 0; i < visibleRows && i+m.ScrollOffset < len(m.AvailableMidiDevices); i++ {
			dataIndex := i + m.ScrollOffset
			deviceRow := deviceStartRow + i

			// Arrow for current selection
			arrow := " "
			if m.CurrentRow == deviceRow {
				arrow = "▶"
			}

			// Device name with appropriate styling
			deviceName := m.AvailableMidiDevices[dataIndex]
			var deviceCell string
			if m.CurrentRow == deviceRow {
				deviceCell = styles.Selected.Render(deviceName)
			} else {
				deviceCell = styles.Normal.Render(deviceName)
			}

			row := fmt.Sprintf("%s %s", arrow, deviceCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		return content.String()
	}, fmt.Sprintf("arrows: navigate | space: select | %s+arrows: adjust", input.GetModifierKey()), statusMsg, m.GetVisibleRows()) // Use dynamic visible rows
}
