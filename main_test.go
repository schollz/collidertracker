package main

import (
	"io"
	"log"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"
	"github.com/stretchr/testify/assert"

	"github.com/schollz/collidertracker/internal/types"
)

func init() {
	log.SetOutput((io.Discard))
}

func createTestModel() *TrackerModel {
	// Create a test model with skip-jack-check enabled
	dispatcher := osc.NewStandardDispatcher()
	return initialModel(57120, "test-tracker.json", false, dispatcher, "")
}

func TestTrackerModelInit(t *testing.T) {
	tm := createTestModel()

	// Test initial state
	assert.NotNil(t, tm.model)
	assert.NotNil(t, tm.splashState)
	assert.True(t, tm.showingSplash)

	// Test Init() returns correct command
	cmd := tm.Init()
	assert.NotNil(t, cmd)
}

func TestTrackerModelUpdate(t *testing.T) {
	tm := createTestModel()

	// Test window size message
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, cmd := tm.Update(sizeMsg)
	assert.Equal(t, 120, newModel.(*TrackerModel).model.TermWidth)
	assert.Equal(t, 40, newModel.(*TrackerModel).model.TermHeight)

	// Test scReadyMsg (hide splash)
	readyMsg := scReadyMsg{}
	newModel, cmd = tm.Update(readyMsg)
	assert.False(t, newModel.(*TrackerModel).showingSplash)

	// Test key press while showing splash (should hide splash)
	tm.showingSplash = true
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, cmd = tm.Update(keyMsg)
	assert.False(t, newModel.(*TrackerModel).showingSplash)
	assert.NotNil(t, cmd) // Should return waveform tick command
}

func TestTrackerModelView(t *testing.T) {
	tm := createTestModel()

	// Test splash screen view
	tm.showingSplash = true
	tm.model.TermWidth = 80
	tm.model.TermHeight = 24
	view := tm.View()
	assert.NotEmpty(t, view)
	// Splash screen renders successfully

	// Test main application views
	tm.showingSplash = false

	// Test each view mode
	viewModes := []types.ViewMode{
		types.SongView,
		types.ChainView,
		types.PhraseView,
		types.SettingsView,
		types.FileMetadataView,
		types.RetriggerView,
		types.TimestrechView,
		types.ArpeggioView,
		types.MidiView,
		types.SoundMakerView,
		types.MixerView,
		types.FileView, // Default case
	}

	for _, viewMode := range viewModes {
		t.Run(string(rune(viewMode)), func(t *testing.T) {
			tm.model.ViewMode = viewMode
			view := tm.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestTrackerModelMessageHandling(t *testing.T) {
	tm := createTestModel()

	tests := []struct {
		name         string
		msg          tea.Msg
		expectSplash bool
		expectCmd    bool
	}{
		{
			name:         "SplashTickMsg during splash",
			msg:          SplashTickMsg{},
			expectSplash: true,
			expectCmd:    true,
		},
		{
			name:         "WaveformTickMsg after splash",
			msg:          WaveformTickMsg{},
			expectSplash: false,
			expectCmd:    true,
		},
		{
			name:         "scReadyMsg hides splash",
			msg:          scReadyMsg{},
			expectSplash: false,
			expectCmd:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to splash state for each test
			tm.showingSplash = tt.name != "WaveformTickMsg after splash"

			newModel, cmd := tm.Update(tt.msg)
			resultTM := newModel.(*TrackerModel)

			assert.Equal(t, tt.expectSplash, resultTM.showingSplash)
			if tt.expectCmd {
				assert.NotNil(t, cmd)
			}
		})
	}
}

func TestTrackerModelViewRendering(t *testing.T) {
	tm := createTestModel()
	tm.showingSplash = false
	tm.model.TermWidth = 120
	tm.model.TermHeight = 40

	// Test view rendering produces valid output for each view mode
	viewTests := []struct {
		name     string
		viewMode types.ViewMode
		setup    func(*TrackerModel)
	}{
		{
			name:     "Song View",
			viewMode: types.SongView,
			setup: func(tm *TrackerModel) {
				tm.model.CurrentRow = 0
				tm.model.CurrentCol = 0
			},
		},
		{
			name:     "Chain View",
			viewMode: types.ChainView,
			setup: func(tm *TrackerModel) {
				tm.model.CurrentRow = 5
				tm.model.CurrentCol = 1
			},
		},
		{
			name:     "Phrase View",
			viewMode: types.PhraseView,
			setup: func(tm *TrackerModel) {
				tm.model.CurrentRow = 10
				tm.model.CurrentCol = 2
			},
		},
		{
			name:     "Settings View",
			viewMode: types.SettingsView,
			setup: func(tm *TrackerModel) {
				tm.model.PreviousView = types.SongView
			},
		},
		{
			name:     "File View",
			viewMode: types.FileView,
			setup: func(tm *TrackerModel) {
				tm.model.CurrentDir = "/tmp"
			},
		},
	}

	for _, test := range viewTests {
		t.Run(test.name, func(t *testing.T) {
			tm.model.ViewMode = test.viewMode
			test.setup(tm)

			view := tm.View()
			assert.NotEmpty(t, view)

			// Verify view contains expected structure
			lines := len(strings.Split(view, "\n"))
			assert.Greater(t, lines, 5)                         // Should have multiple lines
			assert.LessOrEqual(t, lines, tm.model.TermHeight+8) // Should not exceed reasonable bounds (+8 for navigation + padding)
		})
	}
}

func TestTickFunctions(t *testing.T) {
	// Test tick command generators
	waveCmd := tickWaveform(30)
	assert.NotNil(t, waveCmd)

	waveCmd = tickWaveform(0) // Should default to 30fps
	assert.NotNil(t, waveCmd)

	splashCmd := tickSplash()
	assert.NotNil(t, splashCmd)
}

func TestTrackerModelKeyNavigation(t *testing.T) {
	tm := createTestModel()
	tm.showingSplash = false

	// Test basic navigation keys
	navigationKeys := []tea.KeyType{
		tea.KeyUp,
		tea.KeyDown,
		tea.KeyLeft,
		tea.KeyRight,
		tea.KeyTab,
		tea.KeyEnter,
		tea.KeySpace,
		tea.KeyEsc,
	}

	for _, keyType := range navigationKeys {
		t.Run(keyType.String(), func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: keyType}
			newModel, cmd := tm.Update(keyMsg)

			// Should not crash and should return a model
			assert.NotNil(t, newModel)
			// Input handling might return a command (could be nil)
			_ = cmd
		})
	}
}

func TestTrackerModelPlaybackControls(t *testing.T) {
	tm := createTestModel()
	tm.showingSplash = false

	// Test playback control keys
	playbackKeys := []string{
		"p", // Play/pause
		"s", // Stop
		"r", // Record
	}

	for _, key := range playbackKeys {
		t.Run("Key "+key, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
			newModel, cmd := tm.Update(keyMsg)

			// Should not crash
			assert.NotNil(t, newModel)
			// Command might be nil
			_ = cmd
		})
	}
}

func TestTrackerModelViewSwitching(t *testing.T) {
	tm := createTestModel()
	tm.showingSplash = false

	// Test view switching keys
	viewKeys := []struct {
		key      string
		expected types.ViewMode
	}{
		{"1", types.SongView},
		{"2", types.ChainView},
		{"3", types.PhraseView},
		{"4", types.FileView},
	}

	for _, test := range viewKeys {
		t.Run("View switch "+test.key, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(test.key[0])}}
			newModel, cmd := tm.Update(keyMsg)

			// Should not crash
			assert.NotNil(t, newModel)
			// Command might be nil
			_ = cmd

			// Check that model is valid
			resultTM := newModel.(*TrackerModel)
			assert.NotNil(t, resultTM.model)
		})
	}
}

func BenchmarkTrackerModelUpdate(b *testing.B) {
	tm := createTestModel()
	tm.showingSplash = false
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Update(keyMsg)
	}
}

func BenchmarkTrackerModelView(b *testing.B) {
	tm := createTestModel()
	tm.showingSplash = false
	tm.model.TermWidth = 120
	tm.model.TermHeight = 40

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.View()
	}
}
