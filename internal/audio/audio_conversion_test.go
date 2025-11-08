package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertToWaveformFile(t *testing.T) {
	// Create a temporary project directory
	tmpDir := t.TempDir()
	
	// Use a test wav file from the getbpm package
	testFile := "../getbpm/Break120.wav"
	
	// Test conversion
	waveformFile, err := ConvertToWaveformFile(testFile, tmpDir)
	if err != nil {
		t.Fatalf("ConvertToWaveformFile failed: %v", err)
	}
	
	// Verify the output file exists
	if _, err := os.Stat(waveformFile); os.IsNotExist(err) {
		t.Errorf("Waveform file was not created: %s", waveformFile)
	}
	
	// Verify the file is in the waveforms subdirectory
	expectedDir := filepath.Join(tmpDir, "waveforms")
	if !filepath.IsAbs(waveformFile) {
		t.Errorf("Waveform file path should be absolute, got: %s", waveformFile)
	}
	
	dir := filepath.Dir(waveformFile)
	if dir != expectedDir {
		t.Errorf("Expected waveform file in %s, got %s", expectedDir, dir)
	}
	
	// Test that calling again uses cached file
	waveformFile2, err := ConvertToWaveformFile(testFile, tmpDir)
	if err != nil {
		t.Fatalf("Second ConvertToWaveformFile failed: %v", err)
	}
	
	if waveformFile != waveformFile2 {
		t.Errorf("Expected same waveform file path on second call, got %s and %s", waveformFile, waveformFile2)
	}
}
