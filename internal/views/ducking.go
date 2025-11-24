package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

func RenderDuckingView(m *model.Model) string {
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
	header := "Ducking Settings"
	duckingHeader := fmt.Sprintf("Ducking %02X", m.DuckingEditingIndex)
	content.WriteString(RenderHeader(m, header, duckingHeader))
	content.WriteString("\n")

	// Get current ducking settings
	settings := m.DuckingSettings[m.DuckingEditingIndex]

	// Type setting
	typeLabel := "Type:"
	typeNames := []string{"none", "ducking", "ducked"}
	typeValue := "none"
	if settings.Type >= 0 && settings.Type < len(typeNames) {
		typeValue = typeNames[settings.Type]
	}
	var typeCell string
	if m.CurrentRow == 0 {
		typeCell = selectedStyle.Render(typeValue)
	} else {
		typeCell = normalStyle.Render(typeValue)
	}
	typeRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(typeLabel), typeCell)
	content.WriteString(typeRow)
	content.WriteString("\n")

	// Bus setting
	busLabel := "Bus:"
	busValue := fmt.Sprintf("%d", settings.Bus)
	var busCell string
	if m.CurrentRow == 1 {
		busCell = selectedStyle.Render(busValue)
	} else {
		busCell = normalStyle.Render(busValue)
	}
	busRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(busLabel), busCell)
	content.WriteString(busRow)
	content.WriteString("\n")

	// Depth setting
	depthLabel := "Depth:"
	depthValue := fmt.Sprintf("%.2f", settings.Depth)
	var depthCell string
	if m.CurrentRow == 2 {
		depthCell = selectedStyle.Render(depthValue)
	} else {
		depthCell = normalStyle.Render(depthValue)
	}
	depthRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(depthLabel), depthCell)
	content.WriteString(depthRow)
	content.WriteString("\n")

	// Only show Attack, Release, and Thresh when type is 2 (ducked)
	if settings.Type == 2 {
		// Attack setting
		attackLabel := "Attack:"
		attackValue := fmt.Sprintf("%.2fs", settings.Attack)
		var attackCell string
		if m.CurrentRow == 3 {
			attackCell = selectedStyle.Render(attackValue)
		} else {
			attackCell = normalStyle.Render(attackValue)
		}
		attackRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(attackLabel), attackCell)
		content.WriteString(attackRow)
		content.WriteString("\n")

		// Release setting
		releaseLabel := "Release:"
		releaseValue := fmt.Sprintf("%.2fs", settings.Release)
		var releaseCell string
		if m.CurrentRow == 4 {
			releaseCell = selectedStyle.Render(releaseValue)
		} else {
			releaseCell = normalStyle.Render(releaseValue)
		}
		releaseRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(releaseLabel), releaseCell)
		content.WriteString(releaseRow)
		content.WriteString("\n")

		// Thresh setting
		threshLabel := "Thresh:"
		threshValue := fmt.Sprintf("%.2f", settings.Thresh)
		var threshCell string
		if m.CurrentRow == 5 {
			threshCell = selectedStyle.Render(threshValue)
		} else {
			threshCell = normalStyle.Render(threshValue)
		}
		threshRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(threshLabel), threshCell)
		content.WriteString(threshRow)
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Footer with status
	helpText := fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey())
	statusMsg := fmt.Sprintf("Ducking settings")
	footerPad := 6
	if settings.Type == 2 {
		footerPad = 9
	}
	content.WriteString(RenderFooter(m, footerPad, helpText, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}
