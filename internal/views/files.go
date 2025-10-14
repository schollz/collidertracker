package views

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func RenderFileMetadataView(m *model.Model) string {
	filename := filepath.Base(m.MetadataEditingFile)
	header := fmt.Sprintf("File Metadata: %s", filename)

	return renderViewWithCommonPattern(m, header, "", func(styles *ViewStyles) string {
		var content strings.Builder
		content.WriteString("\n")

		// Get current metadata or defaults
		metadata, exists := m.FileMetadata[m.MetadataEditingFile]
		if !exists {
			metadata = types.FileMetadata{BPM: 120.0, Slices: 16, Playthrough: 0, SyncToBPM: 1} // Default values
		}

		// Helper to get option text
		playthroughOptions := []string{"Sliced", "Oneshot"}
		syncToBPMOptions := []string{"No", "Yes"}

		// Metadata settings with common rendering pattern
		settings := []struct {
			label string
			value string
			row   int
		}{
			{"BPM:", fmt.Sprintf("%.2f", metadata.BPM), 0},
			{"Slices:", fmt.Sprintf("%d", metadata.Slices), 1},
			{"Playthrough:", playthroughOptions[metadata.Playthrough], 2},
			{"Sync to BPM:", syncToBPMOptions[metadata.SyncToBPM], 3},
		}

		for _, setting := range settings {
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

		content.WriteString("\n")

		// File info
		fileInfo := fmt.Sprintf("File: %s", m.MetadataEditingFile)
		content.WriteString(styles.Normal.Render(fileInfo))
		content.WriteString("\n\n")

		return content.String()
	}, fmt.Sprintf("Up/Down: Navigate | %s+Arrow: Adjust values | Shift+Down: Back to File Browser", input.GetModifierKey()), 9)
}

func RenderFileView(m *model.Model) string {
	header := fmt.Sprintf("File Browser: %s", m.CurrentDir)

	return renderViewWithCommonPattern(m, header, "", func(styles *ViewStyles) string {
		var content strings.Builder

		// File list
		visibleRows := m.GetVisibleRows()
		for i := 0; i < visibleRows && i+m.ScrollOffset < len(m.Files); i++ {
			dataIndex := i + m.ScrollOffset

			// Arrow for current selection
			arrow := " "
			if m.CurrentRow == dataIndex {
				arrow = "▶"
			}

			// File/directory name with appropriate styling
			filename := m.Files[dataIndex]
			var fileCell string
			if m.CurrentRow == dataIndex {
				fileCell = styles.Selected.Render(filename)
			} else if strings.HasSuffix(filename, "/") || filename == ".." {
				fileCell = styles.Dir.Render(filename)
			} else if IsCurrentRowFile(m, filename) {
				fileCell = styles.AssignedFile.Render(filename)
			} else {
				fileCell = styles.Normal.Render(filename)
			}

			row := fmt.Sprintf("%s %s", arrow, fileCell)
			content.WriteString(row)
			content.WriteString("\n")
		}

		return content.String()
	}, fmt.Sprintf("SPACE: Select file | %s+Right: Play/Stop | Shift+Left: Back to phrase", input.GetModifierKey()), m.GetVisibleRows())
}
