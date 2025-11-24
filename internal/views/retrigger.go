package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
)

func RenderRetriggerView(m *model.Model) string {
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
	header := "Retrigger Settings"
	retriggerHeader := fmt.Sprintf("Retrigger %02X", m.RetriggerEditingIndex)
	content.WriteString(RenderHeader(m, header, retriggerHeader))
	content.WriteString("\n")

	// Get current retrigger settings
	settings := m.RetriggerSettings[m.RetriggerEditingIndex]

	// Times setting
	timesLabel := "Times:"
	timesValue := fmt.Sprintf("%d", settings.Times)
	var timesCell string
	if m.CurrentRow == 0 {
		timesCell = selectedStyle.Render(timesValue)
	} else {
		timesCell = normalStyle.Render(timesValue)
	}
	timesRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(timesLabel), timesCell)
	content.WriteString(timesRow)
	content.WriteString("\n")

	// Starting Rate setting
	startLabel := "Starting Rate:"
	startValue := fmt.Sprintf("%.2f/beat", settings.Start)
	var startCell string
	if m.CurrentRow == 1 {
		startCell = selectedStyle.Render(startValue)
	} else {
		startCell = normalStyle.Render(startValue)
	}
	startRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(startLabel), startCell)
	content.WriteString(startRow)
	content.WriteString("\n")

	// Final Rate setting
	endLabel := "Final Rate:"
	endValue := fmt.Sprintf("%.2f/beat", settings.End)
	var endCell string
	if m.CurrentRow == 2 {
		endCell = selectedStyle.Render(endValue)
	} else {
		endCell = normalStyle.Render(endValue)
	}
	endRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(endLabel), endCell)
	content.WriteString(endRow)
	content.WriteString("\n")

	// Beats setting
	beatsLabel := "Beats:"
	beatsValue := fmt.Sprintf("%d", settings.Beats)
	var beatsCell string
	if m.CurrentRow == 3 {
		beatsCell = selectedStyle.Render(beatsValue)
	} else {
		beatsCell = normalStyle.Render(beatsValue)
	}
	beatsRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(beatsLabel), beatsCell)
	content.WriteString(beatsRow)
	content.WriteString("\n")

	// Volume change setting
	volumeLabel := "Volume dB:"
	var volumeValue string
	if settings.VolumeDB >= 0 {
		volumeValue = fmt.Sprintf("+%.1f", settings.VolumeDB)
	} else {
		volumeValue = fmt.Sprintf("%.1f", settings.VolumeDB)
	}
	var volumeCell string
	if m.CurrentRow == 4 {
		volumeCell = selectedStyle.Render(volumeValue)
	} else {
		volumeCell = normalStyle.Render(volumeValue)
	}
	volumeRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(volumeLabel), volumeCell)
	content.WriteString(volumeRow)
	content.WriteString("\n")

	// Pitch change setting
	pitchLabel := "Pitch:"
	var pitchValue string
	if settings.PitchChange >= 0 {
		pitchValue = fmt.Sprintf("+%.1f", settings.PitchChange)
	} else {
		pitchValue = fmt.Sprintf("%.1f", settings.PitchChange)
	}
	var pitchCell string
	if m.CurrentRow == 5 {
		pitchCell = selectedStyle.Render(pitchValue)
	} else {
		pitchCell = normalStyle.Render(pitchValue)
	}
	pitchRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(pitchLabel), pitchCell)
	content.WriteString(pitchRow)
	content.WriteString("\n")

	// Final pitch to start setting
	finalPitchLabel := "Final pitch to start:"
	finalPitchValue := "No"
	if settings.FinalPitchToStart == 1 {
		finalPitchValue = "Yes"
	}
	var finalPitchCell string
	if m.CurrentRow == 6 {
		finalPitchCell = selectedStyle.Render(finalPitchValue)
	} else {
		finalPitchCell = normalStyle.Render(finalPitchValue)
	}
	finalPitchRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(finalPitchLabel), finalPitchCell)
	content.WriteString(finalPitchRow)
	content.WriteString("\n")

	// Final volume to start setting
	finalVolumeLabel := "Final volume to start:"
	finalVolumeValue := "No"
	if settings.FinalVolumeToStart == 1 {
		finalVolumeValue = "Yes"
	}
	var finalVolumeCell string
	if m.CurrentRow == 7 {
		finalVolumeCell = selectedStyle.Render(finalVolumeValue)
	} else {
		finalVolumeCell = normalStyle.Render(finalVolumeValue)
	}
	finalVolumeRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(finalVolumeLabel), finalVolumeCell)
	content.WriteString(finalVolumeRow)
	content.WriteString("\n")

	// Every setting
	everyLabel := "Every:"
	everyValue := fmt.Sprintf("%d", settings.Every)
	var everyCell string
	if m.CurrentRow == 8 {
		everyCell = selectedStyle.Render(everyValue)
	} else {
		everyCell = normalStyle.Render(everyValue)
	}
	everyRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(everyLabel), everyCell)
	content.WriteString(everyRow)
	content.WriteString("\n")

	// Probability setting
	probabilityLabel := "Probability:"
	probabilityValue := fmt.Sprintf("%d%%", settings.Probability)
	var probabilityCell string
	if m.CurrentRow == 9 {
		probabilityCell = selectedStyle.Render(probabilityValue)
	} else {
		probabilityCell = normalStyle.Render(probabilityValue)
	}
	probabilityRow := fmt.Sprintf("  %-14s %s", labelStyle.Render(probabilityLabel), probabilityCell)
	content.WriteString(probabilityRow)
	content.WriteString("\n\n")

	// Footer with status
	helpText := fmt.Sprintf("arrows: navigate | %s+arrows: adjust", input.GetModifierKey())
	statusMsg := fmt.Sprintf("Retrigger: %d times, %.2f/beat to %.2f/beat", settings.Times, settings.Start, settings.End)
	content.WriteString(RenderFooter(m, 13, helpText, statusMsg))

	// Apply container padding
	return containerStyle.Render(content.String())
}
