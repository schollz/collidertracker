package audio

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/audiomorph"
	"github.com/schollz/collidertracker/internal/getbpm"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/types"
)

// ConvertToWaveformFile converts an audio file to 16-bit mono .wav format for waveform visualization
// Returns the path to the converted file, or an error if conversion fails
func ConvertToWaveformFile(inputPath string, projectDir string) (string, error) {
	// Create waveforms subdirectory in project directory
	waveformDir := filepath.Join(projectDir, "waveforms")
	if err := os.MkdirAll(waveformDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create waveforms directory: %w", err)
	}

	// Generate output filename: use basename + _waveform.wav
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	outputPath := filepath.Join(waveformDir, nameWithoutExt+"_waveform.wav")

	// Check if converted file already exists and is newer than source
	if info, err := os.Stat(outputPath); err == nil {
		sourceInfo, err := os.Stat(inputPath)
		if err == nil && info.ModTime().After(sourceInfo.ModTime()) {
			// Converted file exists and is up to date
			log.Printf("Using existing waveform file: %s", outputPath)
			return outputPath, nil
		}
	}

	// Use audiomorph to decode the audio file
	audio, err := audiomorph.DecodeFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to decode audio file: %w", err)
	}

	// Encode to WAV file
	if err := audiomorph.EncodeFile(audio, outputPath, audiomorph.OptionUseChannels([]int{0})); err != nil {
		return "", fmt.Errorf("failed to encode WAV file: %w", err)
	}

	log.Printf("Converted audio file for waveform: %s -> %s", inputPath, outputPath)
	return outputPath, nil
}

func PlayFile(m *model.Model) {
	if len(m.Files) == 0 || m.CurrentRow >= len(m.Files) {
		return
	}

	filename := m.Files[m.CurrentRow]

	// Don't play directories
	if strings.HasSuffix(filename, "/") || filename == ".." {
		return
	}

	// Get full path for the file
	fullPath := filepath.Join(m.CurrentDir, filename)

	// Check if this specific file is currently playing
	if m.CurrentlyPlayingFile == fullPath {
		// Same file is playing, so stop it
		m.CurrentlyPlayingFile = ""
		m.IsPlaying = false
		m.SendOSCPlaybackMessage(fullPath, false)
		log.Printf("File playback stopped: %s", filename)
	} else {
		// Different file or no file playing, so start playing this one
		// First stop any currently playing file
		if m.CurrentlyPlayingFile != "" {
			m.SendOSCPlaybackMessage(m.CurrentlyPlayingFile, false)
			log.Printf("Stopping previously playing file: %s", m.CurrentlyPlayingFile)
		}
		// Now start the new file
		m.CurrentlyPlayingFile = fullPath
		m.IsPlaying = true
		m.SendOSCPlaybackMessage(fullPath, true)
		log.Printf("File playback started: %s", filename)
	}
}

func SelectFile(m *model.Model) {
	if len(m.Files) == 0 || m.CurrentRow >= len(m.Files) {
		return
	}

	selected := m.Files[m.CurrentRow]

	// Handle directory navigation
	if selected == ".." {
		m.CurrentDir = filepath.Dir(m.CurrentDir)
		storage.LoadFiles(m)
		m.CurrentRow = 0
		m.ScrollOffset = 0
		return
	}

	if strings.HasSuffix(selected, "/") {
		m.CurrentDir = filepath.Join(m.CurrentDir, strings.TrimSuffix(selected, "/"))
		storage.LoadFiles(m)
		m.CurrentRow = 0
		m.ScrollOffset = 0
		return
	}

	// Select audio file - store the full path
	fullPath := filepath.Join(m.CurrentDir, selected)
	fileIndex := m.AppendPhrasesFile(fullPath)
	phrasesData := m.GetCurrentPhrasesData()
	(*phrasesData)[m.CurrentPhrase][m.FileSelectRow][int(types.ColFilename)] = fileIndex

	// Convert file for waveform visualization
	waveformFile, err := ConvertToWaveformFile(fullPath, m.SaveFolder)
	if err != nil {
		log.Printf("Warning: Failed to create waveform file for %s: %v", fullPath, err)
		// Continue anyway - waveform visualization will be unavailable but file can still be used
		waveformFile = "" // Empty string indicates no waveform file available
	}

	// Set initial metadata using getbpm.GetBPM
	// Use the waveform WAV file for BPM detection if available (works better than FLAC)
	// BPM should be float, slices should be 2x beats (rounded to int)
	var bpm float64
	var beats float64
	bpmDetectionFile := fullPath
	if waveformFile != "" {
		bpmDetectionFile = waveformFile
	}
	beats, bpm, err = getbpm.GetBPM(bpmDetectionFile)
	if err == nil {
		slices := int(2 * math.Round(beats))
		m.FileMetadata[fullPath] = types.FileMetadata{
			BPM:          float32(bpm),
			Slices:       slices,
			SliceType:    0, // Default: Even
			Playthrough:  0, // Default: Sliced
			SyncToBPM:    1, // Default: Yes
			WaveformFile: waveformFile,
		}
		// Generate equal slices for the default Even mode
		m.GenerateEqualSlices(fullPath)
	} else {
		log.Printf("Could not get BPM for %s: %v", fullPath, err)
		// Still store the waveform file path even if BPM detection failed
		m.FileMetadata[fullPath] = types.FileMetadata{
			BPM:          120.0, // Default BPM
			Slices:       16,    // Default slices
			SliceType:    0,
			Playthrough:  0,
			SyncToBPM:    1,
			WaveformFile: waveformFile,
		}
	}

	// Track this as the last edited row so "S" key will work
	m.LastEditRow = m.FileSelectRow

	log.Printf("Selected file %s (full path: %s) for phrase %d row %d", selected, fullPath, m.CurrentPhrase, m.FileSelectRow)
	storage.AutoSave(m)
}
