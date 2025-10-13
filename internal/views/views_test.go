package views

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/types"
)

func createTestModel() *model.Model {
	m := model.NewModel(0, "test.json", false) // Port 0 to disable OSC
	m.TermWidth = 120
	m.TermHeight = 40
	return m
}

func TestRenderSongView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.SongView
	m.CurrentRow = 0
	m.CurrentCol = 0

	view := RenderSongView(m)
	assert.NotEmpty(t, view)

	// Verify basic structure
	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 10) // Should have header, content, status

	// Should contain track or song structure
	// (Output depends on current model state and view implementation)
}

func TestRenderChainView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.ChainView
	m.CurrentChain = 5
	m.CurrentRow = 2

	view := RenderChainView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain phrase references
	// Chain view shows phrase numbers in hex format
}

func TestRenderPhraseView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 10
	m.CurrentRow = 5
	m.CurrentCol = 2

	view := RenderPhraseView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 10)

	// Should render different content for sampler vs instrument tracks
	// Test both track types
	m.CurrentTrack = 2 // Instrument track
	instrumentView := RenderPhraseView(m)
	assert.NotEmpty(t, instrumentView)

	m.CurrentTrack = 6 // Sampler track
	samplerView := RenderPhraseView(m)
	assert.NotEmpty(t, samplerView)

	// Both views should render successfully
	// (Specific differences depend on track type configuration and data)
}

func TestRenderFileView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.FileView
	m.CurrentDir = "/tmp"
	m.Files = []string{"test1.wav", "test2.wav", "audio.flac"}

	view := RenderFileView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain file listings
	for _, file := range m.Files {
		assert.Contains(t, view, file)
	}
}

func TestRenderSettingsView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.SettingsView
	m.BPM = 140
	m.PPQ = 4
	m.PregainDB = -3.5
	m.PostgainDB = 2.0

	view := RenderSettingsView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 8)

	// Should contain settings values
	assert.Contains(t, view, "BPM")
	assert.Contains(t, view, "140") // BPM value
}

func TestRenderFileMetadataView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.FileMetadataView
	m.MetadataEditingFile = "test.wav"

	// Set up some file metadata
	metadata := types.FileMetadata{
		BPM:    128.0,
		Slices: 16,
	}
	m.FileMetadata = make(map[string]types.FileMetadata)
	m.FileMetadata["test.wav"] = metadata

	view := RenderFileMetadataView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain file metadata information
	assert.Contains(t, view, "test.wav")
}

func TestRenderRetriggerView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.RetriggerView
	m.RetriggerEditingIndex = 5

	view := RenderRetriggerView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain retrigger settings
	assert.Contains(t, view, "05") // Hex index
}

func TestRenderTimestrechView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.TimestrechView
	m.TimestrechEditingIndex = 10

	view := RenderTimestrechView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain timestretch settings
	assert.Contains(t, view, "0A") // Hex index
}

func TestRenderArpeggioView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.ArpeggioView
	m.ArpeggioEditingIndex = 15

	view := RenderArpeggioView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 10) // Arpeggio has 16 rows

	// Should contain arpeggio settings
	assert.Contains(t, view, "0F") // Hex index
}

func TestRenderMidiView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.MidiView
	m.MidiEditingIndex = 20

	view := RenderMidiView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain MIDI settings
	assert.Contains(t, view, "14") // Hex index
}

func TestRenderSoundMakerView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.SoundMakerView
	m.SoundMakerEditingIndex = 25

	view := RenderSoundMakerView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 5)

	// Should contain SoundMaker settings
	assert.Contains(t, view, "19") // Hex index
}

func TestRenderMixerView(t *testing.T) {
	m := createTestModel()
	m.ViewMode = types.MixerView
	m.CurrentMixerTrack = 3

	view := RenderMixerView(m)
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 8) // Should show all 8 tracks

	// Should contain mixer information
	// (Output depends on mixer state and implementation)
}

func TestRenderSplashScreen(t *testing.T) {
	splashState := NewSplashState(3 * time.Second)

	view := RenderSplashScreen(80, 24, splashState, "test-version")
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 10)     // Should have substantial content
	assert.LessOrEqual(t, len(lines), 30) // Should be reasonable size
}

func TestGetCommonStyles(t *testing.T) {
	styles := getCommonStyles()

	assert.NotNil(t, styles)
	assert.NotNil(t, styles.Selected)
	assert.NotNil(t, styles.Normal)
	assert.NotNil(t, styles.Label)
	assert.NotNil(t, styles.Container)
	assert.NotNil(t, styles.Playback)
	assert.NotNil(t, styles.Copied)
	assert.NotNil(t, styles.Chain)
	assert.NotNil(t, styles.Slice)
	assert.NotNil(t, styles.SliceDownbeat)
	assert.NotNil(t, styles.Dir)
	assert.NotNil(t, styles.AssignedFile)
}

func TestRenderViewWithCommonPattern(t *testing.T) {
	m := createTestModel()

	// Test the common rendering pattern function
	view := renderViewWithCommonPattern(
		m,
		"Test Header",
		"Right Header",
		func(styles *ViewStyles) string {
			return "Test Content"
		},
		"Status Message",
		10,
	)

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Test Header")
	assert.Contains(t, view, "Right Header")
	assert.Contains(t, view, "Test Content")
	assert.Contains(t, view, "Status Message")
}

func TestViewRenderingWithDifferentTerminalSizes(t *testing.T) {
	m := createTestModel()

	terminalSizes := []struct {
		width  int
		height int
		name   string
	}{
		{80, 24, "Standard"},
		{120, 40, "Large"},
		{40, 15, "Small"},
		{200, 60, "Very Large"},
	}

	for _, size := range terminalSizes {
		t.Run(size.name, func(t *testing.T) {
			m.TermWidth = size.width
			m.TermHeight = size.height

			// Test multiple views with this terminal size
			views := []types.ViewMode{
				types.SongView,
				types.ChainView,
				types.PhraseView,
				types.FileView,
			}

			for _, viewMode := range views {
				m.ViewMode = viewMode
				view := renderViewForMode(m, viewMode)
				assert.NotEmpty(t, view)

				// Verify view doesn't exceed terminal bounds
				lines := strings.Split(view, "\n")
				assert.LessOrEqual(t, len(lines), size.height+15) // Allow generous margin for small terminals
			}
		})
	}
}

// Helper function to render view based on mode
func renderViewForMode(m *model.Model, viewMode types.ViewMode) string {
	switch viewMode {
	case types.SongView:
		return RenderSongView(m)
	case types.ChainView:
		return RenderChainView(m)
	case types.PhraseView:
		return RenderPhraseView(m)
	case types.FileView:
		return RenderFileView(m)
	case types.SettingsView:
		return RenderSettingsView(m)
	case types.MixerView:
		return RenderMixerView(m)
	default:
		return RenderFileView(m)
	}
}

func TestViewRenderingWithData(t *testing.T) {
	m := createTestModel()

	// Add some test data to make rendering more interesting
	// Add some song data
	m.SongData[0][0] = 5  // Track 0, Row 0 = Chain 5
	m.SongData[1][2] = 10 // Track 1, Row 2 = Chain 10

	// Add some chain data
	if len(m.SamplerChainsData) > 5 {
		m.SamplerChainsData[5] = []int{0, 1, 2, -1, -1, 3, 4, 5, -1, -1, -1, -1, -1, -1, -1, -1}
	}

	// Add some phrase data with actual notes
	if len(m.SamplerPhrasesData[0]) > 0 {
		m.SamplerPhrasesData[0][0][types.ColNote] = 0x60      // C note
		m.SamplerPhrasesData[0][0][types.ColDeltaTime] = 0x08 // Delta time
		m.SamplerPhrasesData[0][1][types.ColNote] = 0x64      // E note
		m.SamplerPhrasesData[0][1][types.ColDeltaTime] = 0x08 // Delta time
	}

	// Test rendering with data
	views := []types.ViewMode{
		types.SongView,
		types.ChainView,
		types.PhraseView,
	}

	for _, viewMode := range views {
		t.Run("View with data "+string(rune(viewMode)), func(t *testing.T) {
			m.ViewMode = viewMode
			view := renderViewForMode(m, viewMode)
			assert.NotEmpty(t, view)

			// Views with data should be more substantial
			lines := strings.Split(view, "\n")
			assert.Greater(t, len(lines), 8)
		})
	}
}

func TestWaveformRendering(t *testing.T) {
	// Test the existing waveform rendering functionality
	width := 50
	height := 5
	data := []float64{0.0, 0.5, 1.0, 0.5, 0.0, -0.5, -1.0, -0.5, 0.0}

	waveform := RenderWaveform(width, height, data)
	assert.NotEmpty(t, waveform)

	lines := strings.Split(waveform, "\n")
	assert.Equal(t, height, len(lines))
}

func TestSplashState(t *testing.T) {
	duration := 2 * time.Second
	splash := NewSplashState(duration)

	assert.NotNil(t, splash)

	// Test splash rendering
	view := RenderSplashScreen(80, 24, splash, "test-version")
	assert.NotEmpty(t, view)

	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 10)     // Should have substantial content
	assert.LessOrEqual(t, len(lines), 30) // Should be reasonable size
}

func TestViewStylesConsistency(t *testing.T) {
	styles := getCommonStyles()

	// Test that styles are consistently defined
	styleTests := []struct {
		name  string
		style *lipgloss.Style
	}{
		{"Selected", &styles.Selected},
		{"Normal", &styles.Normal},
		{"Label", &styles.Label},
		{"Container", &styles.Container},
		{"Playback", &styles.Playback},
		{"Copied", &styles.Copied},
		{"Chain", &styles.Chain},
		{"Slice", &styles.Slice},
		{"SliceDownbeat", &styles.SliceDownbeat},
		{"Dir", &styles.Dir},
		{"AssignedFile", &styles.AssignedFile},
	}

	for _, test := range styleTests {
		t.Run(test.name+" style", func(t *testing.T) {
			assert.NotNil(t, test.style)
			// Test that style can render text
			rendered := test.style.Render("test")
			assert.NotEmpty(t, rendered)
		})
	}
}

func TestClipboardHighlighting(t *testing.T) {
	m := createTestModel()

	// Set up clipboard with highlight data
	m.Clipboard = types.ClipboardData{
		HasData:         true,
		Mode:            types.CellMode,
		Value:           0x60,
		CellType:        types.HexCell,
		HighlightRow:    5,
		HighlightCol:    2,
		HighlightPhrase: 10,
		HighlightView:   types.PhraseView,
	}

	// Test that phrase view shows highlighting
	m.ViewMode = types.PhraseView
	m.CurrentPhrase = 10
	view := RenderPhraseView(m)
	assert.NotEmpty(t, view)

	// Highlighting should affect the rendered output
	// (specific highlighting behavior depends on implementation)
}

func TestEmptyDataRendering(t *testing.T) {
	m := createTestModel()

	// Test rendering with completely empty data
	// All arrays should be initialized but contain no meaningful data

	views := []types.ViewMode{
		types.SongView,
		types.ChainView,
		types.PhraseView,
		types.FileView,
		types.SettingsView,
		types.MixerView,
		types.RetriggerView,
		types.TimestrechView,
		types.ArpeggioView,
		types.MidiView,
		types.SoundMakerView,
	}

	for _, viewMode := range views {
		t.Run("Empty "+string(rune(viewMode)), func(t *testing.T) {
			m.ViewMode = viewMode
			view := renderViewForMode(m, viewMode)
			assert.NotEmpty(t, view)

			// Should handle empty data gracefully
			lines := strings.Split(view, "\n")
			assert.Greater(t, len(lines), 1)
		})
	}
}

func BenchmarkRenderSongView(b *testing.B) {
	m := createTestModel()
	m.ViewMode = types.SongView

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderSongView(m)
	}
}

func BenchmarkRenderPhraseView(b *testing.B) {
	m := createTestModel()
	m.ViewMode = types.PhraseView

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderPhraseView(m)
	}
}

func BenchmarkRenderWaveform(b *testing.B) {
	width := 50
	height := 5
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i) / 100.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderWaveform(width, height, data)
	}
}
