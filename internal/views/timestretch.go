package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

func RenderTimestrechView(m *model.Model) string {
	// Styles
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")) // Lighter background, dark text
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Main container style with padding
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2)

	// Content builder
	var content strings.Builder

	// Render header
	header := "Timestretch Settings"
	timestrechHeader := fmt.Sprintf("Timestretch %02X", m.TimestrechEditingIndex)
	content.WriteString(RenderHeader(m, header, timestrechHeader))
	content.WriteString("\n")

	// Get current timestretch settings
	settings := m.TimestrechSettings[m.TimestrechEditingIndex]

	// Start setting
	startLabel := "Start:"
	startValue := fmt.Sprintf("%.2fx", settings.Start)
	var startCell string
	if m.CurrentRow == 0 {
		startCell = selectedStyle.Render(startValue)
	} else {
		startCell = normalStyle.Render(startValue)
	}
	startRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(startLabel), startCell)
	content.WriteString(startRow)
	content.WriteString("\n")

	// End setting
	endLabel := "End:"
	endValue := fmt.Sprintf("%.2fx", settings.End)
	var endCell string
	if m.CurrentRow == 1 {
		endCell = selectedStyle.Render(endValue)
	} else {
		endCell = normalStyle.Render(endValue)
	}
	endRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(endLabel), endCell)
	content.WriteString(endRow)
	content.WriteString("\n")

	// Beats setting
	beatsLabel := "Beats:"
	beatsValue := fmt.Sprintf("%d", settings.Beats)
	var beatsCell string
	if m.CurrentRow == 2 {
		beatsCell = selectedStyle.Render(beatsValue)
	} else {
		beatsCell = normalStyle.Render(beatsValue)
	}
	beatsRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(beatsLabel), beatsCell)
	content.WriteString(beatsRow)
	content.WriteString("\n")

	// Every setting
	everyLabel := "Every:"
	everyValue := fmt.Sprintf("%d", settings.Every)
	var everyCell string
	if m.CurrentRow == 3 {
		everyCell = selectedStyle.Render(everyValue)
	} else {
		everyCell = normalStyle.Render(everyValue)
	}
	everyRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(everyLabel), everyCell)
	content.WriteString(everyRow)
	content.WriteString("\n")

	// Probability setting
	probabilityLabel := "Probability:"
	probabilityValue := fmt.Sprintf("%d%%", settings.Probability)
	var probabilityCell string
	if m.CurrentRow == 4 {
		probabilityCell = selectedStyle.Render(probabilityValue)
	} else {
		probabilityCell = normalStyle.Render(probabilityValue)
	}
	probabilityRow := fmt.Sprintf("  %-12s %s", labelStyle.Render(probabilityLabel), probabilityCell)
	content.WriteString(probabilityRow)
	content.WriteString("\n\n")

	// Footer with status
	helpText := fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey())
	statusMsg := fmt.Sprintf("Timestretch: %.2fx to %.2fx", settings.Start, settings.End)
	content.WriteString(RenderFooter(m, 8, helpText, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}
