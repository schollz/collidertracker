package input

import (
	"fmt"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/types"
)

func ModifyFileMetadataValue(m *model.Model, delta float32) {
	if m.MetadataEditingFile == "" {
		return
	}

	// Get current metadata or create default
	metadata, exists := m.FileMetadata[m.MetadataEditingFile]
	if !exists {
		metadata = types.FileMetadata{BPM: 120.0, Slices: 16, SliceType: 0, Playthrough: 0, SyncToBPM: 1} // Default values
	}

	switch types.FileMetadataRow(m.CurrentRow) {
	case types.FileMetadataRowBPM: // BPM
		modifier := createFloatModifier(
			func() float32 { return metadata.BPM },
			func(v float32) {
				metadata.BPM = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
			},
			1, 999, fmt.Sprintf("file metadata BPM for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)

	case types.FileMetadataRowSlices: // Slices
		modifier := createIntModifier(
			func() int { return metadata.Slices },
			func(v int) {
				metadata.Slices = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
				// Trigger onset detection if in Onsets mode
				if metadata.SliceType == 1 {
					m.TriggerOnsetDetection(m.MetadataEditingFile)
				} else {
					// Trigger equal slice generation if in Equal mode
					m.GenerateEqualSlices(m.MetadataEditingFile)
				}
			},
			1, 999, fmt.Sprintf("file metadata Slices for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)

	case types.FileMetadataRowSliceType: // Slice Type (0=Even, 1=Onsets)
		modifier := createIntModifier(
			func() int { return metadata.SliceType },
			func(v int) {
				metadata.SliceType = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
				// Trigger onset detection if switching to Onsets mode
				if v == 1 {
					m.TriggerOnsetDetection(m.MetadataEditingFile)
				} else {
					// Trigger equal slice generation if switching to Equal mode
					m.GenerateEqualSlices(m.MetadataEditingFile)
				}
			},
			0, 1, fmt.Sprintf("file metadata SliceType for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)

	case types.FileMetadataRowPlaythrough: // Playthrough (0=Sliced, 1=Oneshot)
		modifier := createIntModifier(
			func() int { return metadata.Playthrough },
			func(v int) {
				metadata.Playthrough = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
			},
			0, 1, fmt.Sprintf("file metadata Playthrough for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)

	case types.FileMetadataRowSyncToBPM: // Sync to BPM (0=No, 1=Yes)
		modifier := createIntModifier(
			func() int { return metadata.SyncToBPM },
			func(v int) {
				metadata.SyncToBPM = v
				m.FileMetadata[m.MetadataEditingFile] = metadata
			},
			0, 1, fmt.Sprintf("file metadata SyncToBPM for %s", m.MetadataEditingFile),
		)
		modifyValueWithBounds(modifier, delta)
	}

	storage.AutoSave(m)
}
