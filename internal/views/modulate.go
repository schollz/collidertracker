package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

func RenderModulateView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")) // Lighter background, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header (includes waveform)
	header := "Modulate Settings"
	modulateHeader := fmt.Sprintf("Modulate %02X", m.ModulateEditingIndex)
	content.WriteString(RenderHeader(m, header, modulateHeader))
	content.WriteString("\n")

	// Get current modulate settings
	settings := (*m.GetCurrentModulateSettings())[m.ModulateEditingIndex]

	// Seed setting
	seedLabel := "Seed:"
	seedValue := "none"
	if settings.Seed == 0 {
		seedValue = "random"
	} else if settings.Seed > 0 {
		seedValue = fmt.Sprintf("%d", settings.Seed)
	}
	var seedCell string
	if m.CurrentRow == 0 {
		seedCell = selectedStyle.Render(seedValue)
	} else {
		seedCell = normalStyle.Render(seedValue)
	}
	seedRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(seedLabel), seedCell)
	content.WriteString(seedRow)
	content.WriteString("\n")

	// IRandom setting
	irandomLabel := "IRandom:"
	irandomValue := fmt.Sprintf("%d", settings.IRandom)
	var irandomCell string
	if m.CurrentRow == 1 {
		irandomCell = selectedStyle.Render(irandomValue)
	} else {
		irandomCell = normalStyle.Render(irandomValue)
	}
	irandomRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(irandomLabel), irandomCell)
	content.WriteString(irandomRow)
	content.WriteString("\n")

	// Sub setting
	subLabel := "Sub:"
	subValue := fmt.Sprintf("%d", settings.Sub)
	var subCell string
	if m.CurrentRow == 2 {
		subCell = selectedStyle.Render(subValue)
	} else {
		subCell = normalStyle.Render(subValue)
	}
	subRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(subLabel), subCell)
	content.WriteString(subRow)
	content.WriteString("\n")

	// Add setting
	addLabel := "Add:"
	addValue := fmt.Sprintf("%d", settings.Add)
	var addCell string
	if m.CurrentRow == 3 {
		addCell = selectedStyle.Render(addValue)
	} else {
		addCell = normalStyle.Render(addValue)
	}
	addRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(addLabel), addCell)
	content.WriteString(addRow)
	content.WriteString("\n")

	// Increment setting
	incrementLabel := "Increment:"
	incrementValue := fmt.Sprintf("%d", settings.Increment)
	var incrementCell string
	if m.CurrentRow == 4 {
		incrementCell = selectedStyle.Render(incrementValue)
	} else {
		incrementCell = normalStyle.Render(incrementValue)
	}
	incrementRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(incrementLabel), incrementCell)
	content.WriteString(incrementRow)
	content.WriteString("\n")

	// Wrap setting
	wrapLabel := "Wrap:"
	wrapValue := "none"
	if settings.Wrap > 0 {
		wrapValue = fmt.Sprintf("%d", settings.Wrap)
	}
	var wrapCell string
	if m.CurrentRow == 5 {
		wrapCell = selectedStyle.Render(wrapValue)
	} else {
		wrapCell = normalStyle.Render(wrapValue)
	}
	wrapRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(wrapLabel), wrapCell)
	content.WriteString(wrapRow)
	content.WriteString("\n")

	// ScaleRoot setting
	scaleRootLabel := "ScaleRoot:"
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	scaleRootValue := "C" // Default
	if settings.ScaleRoot >= 0 && settings.ScaleRoot < len(noteNames) {
		scaleRootValue = noteNames[settings.ScaleRoot]
	}
	var scaleRootCell string
	if m.CurrentRow == 6 {
		scaleRootCell = selectedStyle.Render(scaleRootValue)
	} else {
		scaleRootCell = normalStyle.Render(scaleRootValue)
	}
	scaleRootRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(scaleRootLabel), scaleRootCell)
	content.WriteString(scaleRootRow)
	content.WriteString("\n")

	// Scale setting
	scaleLabel := "Scale:"
	scaleValue := settings.Scale
	if scaleValue == "" {
		scaleValue = "all"
	}
	var scaleCell string
	if m.CurrentRow == 7 {
		scaleCell = selectedStyle.Render(scaleValue)
	} else {
		scaleCell = normalStyle.Render(scaleValue)
	}
	scaleRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(scaleLabel), scaleCell)
	content.WriteString(scaleRow)
	content.WriteString("\n")

	// Probability setting
	probabilityLabel := "Probability:"
	probabilityValue := fmt.Sprintf("%d%%", settings.Probability)
	var probabilityCell string
	if m.CurrentRow == 8 {
		probabilityCell = selectedStyle.Render(probabilityValue)
	} else {
		probabilityCell = normalStyle.Render(probabilityValue)
	}
	probabilityRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(probabilityLabel), probabilityCell)
	content.WriteString(probabilityRow)
	content.WriteString("\n\n")

	// Footer with status
	helpText := fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey())
	statusMsg := fmt.Sprintf("Modulate settings")
	content.WriteString(RenderFooter(m, 9, helpText, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}
