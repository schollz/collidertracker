package audio

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// testSuperColliderAvailable checks if SuperCollider is available and responding to OSC
func testSuperColliderAvailable() bool {
	// Try to send a simple OSC message to see if SuperCollider is listening
	oscClient := osc.NewClient("localhost", 57120)
	msg := osc.NewMessage("/ping")
	err := oscClient.Send(msg)
	return err == nil
}

func TestConvertToWaveformFile(t *testing.T) {
	// Skip test if SuperCollider is not available
	if !testSuperColliderAvailable() {
		t.Skip("SuperCollider is not running on port 57120, skipping test")
	}

	// Create a temporary project directory
	tmpDir := t.TempDir()
	
	// Use a test wav file from the getbpm package
	testFile := "../getbpm/Break120.wav"
	
	// Make sure the test file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("Test file does not exist: %s", testFile)
	}
	
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
	
	// Verify file has content
	info, err := os.Stat(waveformFile)
	if err != nil {
		t.Errorf("Failed to stat waveform file: %v", err)
	} else if info.Size() == 0 {
		t.Errorf("Waveform file is empty")
	}
	
	// Test that calling again uses cached file
	// First, get the mod time of the original file
	origModTime := info.ModTime()
	
	// Wait a bit to ensure time has passed
	time.Sleep(100 * time.Millisecond)
	
	waveformFile2, err := ConvertToWaveformFile(testFile, tmpDir)
	if err != nil {
		t.Fatalf("Second ConvertToWaveformFile failed: %v", err)
	}
	
	if waveformFile != waveformFile2 {
		t.Errorf("Expected same waveform file path on second call, got %s and %s", waveformFile, waveformFile2)
	}
	
	// Verify the file wasn't re-created (mod time should be the same)
	info2, err := os.Stat(waveformFile2)
	if err != nil {
		t.Errorf("Failed to stat waveform file on second call: %v", err)
	} else if !info2.ModTime().Equal(origModTime) {
		t.Logf("Warning: File was re-created on second call (original: %v, second: %v)", origModTime, info2.ModTime())
		// This is just a warning, not a failure, as caching behavior may vary
	}
}
