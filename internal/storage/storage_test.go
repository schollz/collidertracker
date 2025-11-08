package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDoSave(t *testing.T) {
	t.Run("successful save", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "test_save")

		m := model.NewModel(0, saveFolder, false)
		m.BPM = 140
		m.CurrentRow = 5

		DoSave(m)

		// Check that save folder was created
		_, err := os.Stat(saveFolder)
		assert.NoError(t, err)

		// Check that data.json.gz was created
		dataFile := filepath.Join(saveFolder, "data.json.gz")
		_, err = os.Stat(dataFile)
		assert.NoError(t, err)

		// Check that data file has content
		data, err := os.ReadFile(dataFile)
		assert.NoError(t, err)
		assert.True(t, len(data) > 0)
	})

	t.Run("save to invalid path", func(t *testing.T) {
		m := model.NewModel(0, "/invalid/path/that/does/not/exist/save", false)

		// Should not panic, just log error
		DoSave(m)
	})
}

func TestLoadState(t *testing.T) {
	t.Run("load existing save file", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "test_load")

		// Create and save a model with specific state
		m1 := model.NewModel(0, saveFolder, false)
		m1.BPM = 140
		m1.CurrentRow = 10
		m1.ViewMode = types.ChainView
		DoSave(m1)

		// Create new model and load state
		m2 := model.NewModel(0, saveFolder, false)
		err := LoadState(m2, 0, saveFolder)

		assert.NoError(t, err)
		assert.Equal(t, float32(140), m2.BPM)
		assert.Equal(t, 10, m2.CurrentRow)
		assert.Equal(t, types.ChainView, m2.ViewMode)
	})

	t.Run("load nonexistent file", func(t *testing.T) {
		m := model.NewModel(0, "", false)
		err := LoadState(m, 0, "/path/that/does/not/exist")

		assert.Error(t, err)
	})

	t.Run("force return to phrase view from non-main views", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "test_force_view")

		// Create and save a model in a non-main view
		m1 := model.NewModel(0, saveFolder, false)
		m1.ViewMode = types.FileView // This should be forced to PhraseView
		DoSave(m1)

		// Load state
		m2 := model.NewModel(0, saveFolder, false)
		err := LoadState(m2, 0, saveFolder)

		assert.NoError(t, err)
		assert.Equal(t, types.PhraseView, m2.ViewMode)
		assert.Equal(t, int(types.ColFilename), m2.CurrentCol)
	})
}

func TestLoadFiles(t *testing.T) {
	t.Run("load files from existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		os.WriteFile(filepath.Join(tmpDir, "test1.wav"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "test2.flac"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "test3.txt"), []byte("test"), 0644) // Should be ignored
		os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

		m := model.NewModel(0, "", false)
		m.CurrentDir = tmpDir

		LoadFiles(m)

		// Should have parent dir, subdir, and audio files
		assert.Contains(t, m.Files, "..")
		assert.Contains(t, m.Files, "subdir/")
		assert.Contains(t, m.Files, "test1.wav")
		assert.Contains(t, m.Files, "test2.flac")
		assert.NotContains(t, m.Files, "test3.txt") // Non-audio files should be excluded
	})

	t.Run("load files from root directory", func(t *testing.T) {
		m := model.NewModel(0, "", false)
		m.CurrentDir = "/"

		LoadFiles(m)

		// Should not have ".." when at root
		assert.NotContains(t, m.Files, "..")
	})

	t.Run("load files from nonexistent directory", func(t *testing.T) {
		m := model.NewModel(0, "", false)
		m.CurrentDir = "/path/that/does/not/exist"

		LoadFiles(m)

		// Should have empty files list
		assert.Equal(t, []string{}, m.Files)
	})
}

func TestAutoSave(t *testing.T) {
	t.Run("autosave debouncing", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "autosave_test")

		m := model.NewModel(0, saveFolder, false)
		m.BPM = 150

		// Call AutoSave multiple times quickly
		AutoSave(m)
		AutoSave(m)
		AutoSave(m)

		// Should not save immediately
		dataFile := filepath.Join(saveFolder, "data.json.gz")
		_, err := os.Stat(dataFile)
		assert.True(t, os.IsNotExist(err))

		// Wait for debounce timeout with polling to handle CI timing variations
		timeout := time.After(3 * time.Second)
		tick := time.Tick(100 * time.Millisecond)
		fileCreated := false
		for !fileCreated {
			select {
			case <-timeout:
				t.Fatal("Timed out waiting for data.json.gz to be created")
			case <-tick:
				if _, err := os.Stat(dataFile); err == nil {
					fileCreated = true
				}
			}
		}

		// File should exist now
		_, err = os.Stat(dataFile)
		assert.NoError(t, err)
	})
}

func TestWaveformFileResolution(t *testing.T) {
	t.Run("waveform file path is resolved on load", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "test_waveform")

		// Create waveforms directory
		waveformDir := filepath.Join(saveFolder, "waveforms")
		err := os.MkdirAll(waveformDir, 0755)
		assert.NoError(t, err)

		// Create a test audio file in save folder
		testAudioFile := filepath.Join(saveFolder, "test.wav")
		err = os.WriteFile(testAudioFile, []byte("test audio"), 0644)
		assert.NoError(t, err)

		// Create a test waveform file in the waveforms directory
		testWaveformFile := filepath.Join(waveformDir, "test_waveform.wav")
		err = os.WriteFile(testWaveformFile, []byte("test waveform"), 0644)
		assert.NoError(t, err)

		// Create a model and add file metadata with waveform file
		m1 := model.NewModel(0, saveFolder, false)
		m1.FileMetadata[testAudioFile] = types.FileMetadata{
			BPM:          120.0,
			Slices:       16,
			SliceType:    0,
			Playthrough:  0,
			SyncToBPM:    1,
			WaveformFile: testWaveformFile, // Absolute path initially
		}
		m1.SamplerPhrasesFiles = []string{testAudioFile}

		// Save the model
		DoSave(m1)

		// Verify that the waveform file path was saved as relative in data.json.gz
		// (We'll verify this by loading and checking)

		// Create a new model and load state
		m2 := model.NewModel(0, saveFolder, false)
		err = LoadState(m2, 0, saveFolder)
		assert.NoError(t, err)

		// Check that the waveform file path is correctly resolved in the loaded model
		loadedMetadata, exists := m2.FileMetadata[testAudioFile]
		assert.True(t, exists, "Metadata should exist for test audio file")
		assert.True(t, filepath.IsAbs(loadedMetadata.WaveformFile), "WaveformFile should be absolute after loading")
		assert.Equal(t, testWaveformFile, loadedMetadata.WaveformFile, "WaveformFile should match expected path")

		// Verify the waveform file exists at the resolved path
		_, err = os.Stat(loadedMetadata.WaveformFile)
		assert.NoError(t, err, "WaveformFile should exist at resolved path")
	})

	t.Run("waveform file outside save folder keeps absolute path", func(t *testing.T) {
		tmpDir := t.TempDir()
		saveFolder := filepath.Join(tmpDir, "test_waveform_external")
		externalDir := filepath.Join(tmpDir, "external")

		// Create save folder
		err := os.MkdirAll(saveFolder, 0755)
		assert.NoError(t, err)

		// Create external directory and file
		err = os.MkdirAll(externalDir, 0755)
		assert.NoError(t, err)

		externalWaveformFile := filepath.Join(externalDir, "external_waveform.wav")
		err = os.WriteFile(externalWaveformFile, []byte("external waveform"), 0644)
		assert.NoError(t, err)

		// Create a test audio file in save folder
		testAudioFile := filepath.Join(saveFolder, "test.wav")
		err = os.WriteFile(testAudioFile, []byte("test audio"), 0644)
		assert.NoError(t, err)

		// Create a model with metadata pointing to external waveform file
		m1 := model.NewModel(0, saveFolder, false)
		m1.FileMetadata[testAudioFile] = types.FileMetadata{
			BPM:          120.0,
			Slices:       16,
			SliceType:    0,
			Playthrough:  0,
			SyncToBPM:    1,
			WaveformFile: externalWaveformFile, // External absolute path
		}
		m1.SamplerPhrasesFiles = []string{testAudioFile}

		// Save the model
		DoSave(m1)

		// Load the state
		m2 := model.NewModel(0, saveFolder, false)
		err = LoadState(m2, 0, saveFolder)
		assert.NoError(t, err)

		// The external waveform file path should remain absolute
		loadedMetadata, exists := m2.FileMetadata[testAudioFile]
		assert.True(t, exists, "Metadata should exist for test audio file")
		assert.Equal(t, externalWaveformFile, loadedMetadata.WaveformFile, "External WaveformFile should remain absolute")
	})
}

func BenchmarkDoSave(b *testing.B) {
	// Create a temporary folder for testing
	tmpDir := b.TempDir()
	saveFolder := filepath.Join(tmpDir, "test_save")

	// Create a model with default data
	m := model.NewModel(0, saveFolder, false) // OSC port 0 to disable OSC

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DoSave(m)
	}
}

func BenchmarkLoadState(b *testing.B) {
	// Create a temporary folder for testing
	tmpDir := b.TempDir()
	saveFolder := filepath.Join(tmpDir, "test_load")

	// Create a model with default data and save it once
	m := model.NewModel(0, saveFolder, false) // OSC port 0 to disable OSC
	DoSave(m)

	// Verify the save folder and data file exist
	dataFile := filepath.Join(saveFolder, "data.json.gz")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		b.Fatal("Save data file was not created")
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh model for each load operation
		testModel := model.NewModel(0, saveFolder, false)
		err := LoadState(testModel, 0, saveFolder)
		if err != nil {
			b.Fatalf("LoadState failed: %v", err)
		}
	}
}
