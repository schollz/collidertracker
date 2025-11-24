package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hypebeast/go-osc/osc"
	"github.com/spf13/cobra"

	"github.com/schollz/collidertracker/internal/hacks"
	"github.com/schollz/collidertracker/internal/input"
	"github.com/schollz/collidertracker/internal/midiconnector"
	"github.com/schollz/collidertracker/internal/model"
	"github.com/schollz/collidertracker/internal/project"
	"github.com/schollz/collidertracker/internal/storage"
	"github.com/schollz/collidertracker/internal/supercollider"
	"github.com/schollz/collidertracker/internal/types"
	"github.com/schollz/collidertracker/internal/views"
)

var (
	Version = "dev"

	// Command-line configuration
	config struct {
		port            int
		project         string
		projectProvided bool // Track if --project flag was explicitly provided
		record          bool
		debug           string
		skipSC          bool
		vim             bool
	}
)

type scReadyMsg struct{}

var rootCmd = &cobra.Command{
	Use:   "collidertracker",
	Short: "A modern music tracker for SuperCollider",
	Long: `ColliderTracker is a modern, terminal-based music tracker that integrates 
with SuperCollider for real-time audio synthesis and sampling.

Features:
• Real-time audio synthesis with SuperCollider
• Sample-based music composition
• MIDI integration
• Live audio recording and playback
• Retrigger and time-stretch effects`,
	Version: Version,
	Run:     runColliderTracker,
}

func init() {
	rootCmd.PersistentFlags().IntVar(&config.port, "port", 57120,
		"OSC port for SuperCollider communication")
	rootCmd.PersistentFlags().StringVarP(&config.project, "project", "p", "save",
		"Project directory for songs and audio files")
	rootCmd.PersistentFlags().BoolVarP(&config.record, "record", "r", false,
		"Enable automatic session recording")
	rootCmd.PersistentFlags().StringVarP(&config.debug, "log", "l", "",
		"Write debug logs to specified file (empty disables)")
	rootCmd.PersistentFlags().BoolVarP(&config.skipSC, "skip-sc", "s", false,
		"Skip SuperCollider detection and management entirely")
	rootCmd.PersistentFlags().BoolVar(&config.vim, "vim", false,
		"Enable vim-style cursor movement (h/j/k/l)")

	// Set up a callback to track when --project is explicitly provided
	rootCmd.PersistentFlags().Lookup("project").Changed = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


// checkAndUpdatePortIfNeeded checks if SuperCollider detected a different port
// and updates the OSC client if necessary
func checkAndUpdatePortIfNeeded(tm *TrackerModel) {
	// Wait a moment for SuperCollider to output its port information
	time.Sleep(2 * time.Second)
	// Check if SuperCollider detected a different port
	if detectedPort := supercollider.GetDetectedPort(); detectedPort > 0 && detectedPort != config.port {
		log.Printf("SuperCollider started on port %d (expected %d), updating OSC configuration", detectedPort, config.port)
		tm.model.UpdateOSCPort(detectedPort)
	}
}

func restartWithProject() {
	// This function restarts the ColliderTracker with the new project
	// without going through cobra command parsing again

	// Check JACK and SuperCollider requirements (same as in runColliderTracker)

	// Check for required SuperCollider extensions before starting
	if !supercollider.HasRequiredExtensions() {
		dialog := supercollider.NewInstallDialogModel()
		p := tea.NewProgram(dialog, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			log.Printf("Error running install dialog: %v", err)
			os.Exit(1)
		}

		if result, ok := finalModel.(supercollider.InstallDialogModel); ok {
			if !result.ShouldInstall() {
				os.Exit(1)
			}
			if result.Error() != nil {
				log.Printf("Failed to install SuperCollider extensions: %v", result.Error())
				os.Exit(1)
			}
		} else {
			log.Printf("Unexpected model type returned from install dialog")
			os.Exit(1)
		}
	}

	// Set up debug logging early
	if config.debug != "" {
		f, err := tea.LogToFile(config.debug, "debug")
		if err != nil {
			log.Printf("Fatal: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		// Set log flags to include file and line number for VS Code clickable links
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		// send log to io.Discard
		log.SetOutput(io.Discard)
	}

	log.Println("Debug logging enabled")
	log.Printf("OSC port configured: %d", config.port)

	// Create readiness channel for SuperCollider startup detection
	readyChannel := make(chan struct{}, 1)

	// Set up OSC dispatcher early to detect SuperCollider readiness
	d := osc.NewStandardDispatcher()
	var tm *TrackerModel // Will be set after model creation
	var initialPreferencesSent = false

	d.AddMsgHandler("/cpuusage", func(msg *osc.Message) {
		log.Printf("SuperCollider CPU Usage: %v", msg.Arguments[0])

		// Send initial preferences on first CPU message (when SC is confirmed ready)
		if !initialPreferencesSent && tm != nil {
			log.Printf("Sending initial preferences to SuperCollider")
			tm.model.SendOSCPregainMessage()
			tm.model.SendOSCPostgainMessage()
			tm.model.SendOSCBiasMessage()
			tm.model.SendOSCSaturationMessage()
			tm.model.SendOSCDriveMessage()
			tm.model.SendOSCInputLevelMessage()
			tm.model.SendOSCReverbSendMessage()
			tm.model.SendOSCTapeMessage()
			tm.model.SendOSCShimmerMessage()

			// Send track set levels too
			for track := 0; track < 8; track++ {
				tm.model.SendOSCTrackSetLevelMessage(track)
			}
			initialPreferencesSent = true
		}

		// Signal that SuperCollider is ready (non-blocking)
		select {
		case readyChannel <- struct{}{}:
		default:
		}
	})

	d.AddMsgHandler("/track_volume", func(msg *osc.Message) {
		if tm != nil {
			for i := 0; i < len(tm.model.TrackVolumes); i++ {
				tm.model.TrackVolumes[i] = msg.Arguments[i].(float32)
			}
		}
	})
	// Build program
	tm = initialModel(config.port, config.project, config.vim, d)

	p := tea.NewProgram(tm, tea.WithAltScreen())

	// Start OSC server after p is created but before p.Run()
	server := &osc.Server{Addr: fmt.Sprintf(":%d", config.port+1), Dispatcher: d}
	go func() {
		log.Printf("Starting OSC server on port %d", config.port+1)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting OSC server: %v", err)
		}
	}()

	// Fast SuperCollider detection and startup
	if !config.skipSC {
		go func() {
			// First, quickly check if sclang process is running
			if !supercollider.IsSuperColliderEnabled() {
				// No sclang process found - start SuperCollider immediately
				log.Printf("No sclang process found, starting SuperCollider")
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			checkAndUpdatePortIfNeeded(tm)
				return
			}

			// sclang is running - wait briefly to see if it has ColliderTracker loaded
			log.Printf("Found sclang process, checking if ColliderTracker is loaded...")
			timeout := time.NewTimer(1 * time.Second)
			defer timeout.Stop()

			select {
			case <-readyChannel:
				// SuperCollider with ColliderTracker is already running
				log.Printf("Found existing SuperCollider instance with ColliderTracker")
				return
			case <-timeout.C:
				// sclang is running but no ColliderTracker - start our own instance
				log.Printf("sclang running but no ColliderTracker detected, starting new instance")
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			checkAndUpdatePortIfNeeded(tm)
			}
		}()
	} else {
		log.Printf("Skipping SuperCollider detection and management entirely (--skip-sc flag provided)")
	}

	// When SC signals readiness via /cpuusage, hide the splash
	go func() {
		if config.skipSC {
			p.Send(scReadyMsg{}) // skip splash if skipping SC management
		} else {
			<-readyChannel
			log.Printf("Received SuperCollider ready; hiding splash")
			p.Send(scReadyMsg{})
		}
	}()

	// hack to make sure Ctrl+V works on Windows
	hacks.StoreWinClipboard()

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Check if we should return to project selection again (recursive)
	if finalModel != nil {
		if trackerModel, ok := finalModel.(*TrackerModel); ok && trackerModel.model.ReturnToProjectSelector {
			log.Printf("Returning to project selection...")
			// Clean up current session
			supercollider.Cleanup()

			// Run project selector again
			selectedPath, cancelled, isNewProject := project.RunProjectSelector()
			if !cancelled {
				if isNewProject {
					// User chose to create new project with provided name
					config.project = selectedPath
				} else {
					// User selected an existing project
					config.project = selectedPath
				}
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			}
		}
	}

	// Always call cleanup when the program exits normally (e.g., Ctrl+Q)
	supercollider.Cleanup()
}

func runColliderTracker(cmd *cobra.Command, args []string) {
	// Start CPU profiling for the first 30 seconds
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
	} else {
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
		} else {
			go func() {
				time.Sleep(30 * time.Second)
				pprof.StopCPUProfile()
				cpuFile.Close()
			}()
		}
	}

	// Set up cleanup on exit
	setupCleanupOnExit()

	// Check if --project flag was explicitly provided
	config.projectProvided = cmd.PersistentFlags().Changed("project")

	// If no project was specified, show project selector
	if !config.projectProvided {
		selectedPath, cancelled, isNewProject := project.RunProjectSelector()
		if cancelled {
			os.Exit(0)
		}

		if isNewProject {
			// User chose to create new project with provided name
			config.project = selectedPath
		} else {
			// User selected an existing project
			config.project = selectedPath
		}
	}

	// Check for required SuperCollider extensions before starting
	if !supercollider.HasRequiredExtensions() {
		dialog := supercollider.NewInstallDialogModel()
		p := tea.NewProgram(dialog, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			log.Printf("Error running install dialog: %v", err)
			os.Exit(1)
		}

		if result, ok := finalModel.(supercollider.InstallDialogModel); ok {
			if !result.ShouldInstall() {
				os.Exit(1)
			}
			if result.Error() != nil {
				log.Printf("Failed to install SuperCollider extensions: %v", result.Error())
				os.Exit(1)
			}
		} else {
			log.Printf("Unexpected model type returned from install dialog")
			os.Exit(1)
		}
	}

	// Set up debug logging early
	if config.debug != "" {
		f, err := tea.LogToFile(config.debug, "debug")
		if err != nil {
			log.Printf("Fatal: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		// Set log flags to include file and line number for VS Code clickable links
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		// send log to io.Discard
		log.SetOutput(io.Discard)
	}

	log.Println("Debug logging enabled")
	log.Printf("OSC port configured: %d", config.port)

	// Create readiness channel for SuperCollider startup detection
	readyChannel := make(chan struct{}, 1)

	// Set up OSC dispatcher early to detect SuperCollider readiness
	d := osc.NewStandardDispatcher()
	var tm *TrackerModel // Will be set after model creation
	var initialPreferencesSent = false

	d.AddMsgHandler("/cpuusage", func(msg *osc.Message) {
		log.Printf("SuperCollider CPU Usage: %v", msg.Arguments[0])

		// Send initial preferences on first CPU message (when SC is confirmed ready)
		if !initialPreferencesSent && tm != nil {
			log.Printf("Sending initial preferences to SuperCollider")
			tm.model.SendOSCPregainMessage()
			tm.model.SendOSCPostgainMessage()
			tm.model.SendOSCBiasMessage()
			tm.model.SendOSCSaturationMessage()
			tm.model.SendOSCDriveMessage()
			tm.model.SendOSCInputLevelMessage()
			tm.model.SendOSCReverbSendMessage()
			tm.model.SendOSCTapeMessage()
			tm.model.SendOSCShimmerMessage()

			// Send track set levels too
			for track := 0; track < 8; track++ {
				tm.model.SendOSCTrackSetLevelMessage(track)
			}
			initialPreferencesSent = true
		}

		// Signal that SuperCollider is ready (non-blocking)
		select {
		case readyChannel <- struct{}{}:
		default:
		}
	})

	d.AddMsgHandler("/sampler_playhead", func(msg *osc.Message) {
		if tm != nil {
			trackID := int(msg.Arguments[0].(float32))
			gate := int(msg.Arguments[1].(float32))
			pos := float64(msg.Arguments[2].(float32))
			sliceStart := float64(msg.Arguments[3].(float32))
			sliceEnd := float64(msg.Arguments[4].(float32))
			log.Printf("Track %d playhead: gate=%d pos=%.2f sliceStart=%.2f sliceEnd=%.2f", trackID, gate, pos, sliceStart, sliceEnd)
			// Update model with playhead data
			tm.model.PlayheadTrackID = trackID
			tm.model.PlayheadGate = gate
			tm.model.PlayheadPos = pos
			tm.model.PlayheadSliceStart = sliceStart
			tm.model.PlayheadSliceEnd = sliceEnd
			tm.model.PlayheadLastUpdate = time.Now()
		}
	})

	d.AddMsgHandler("/track_volume", func(msg *osc.Message) {
		if tm != nil {
			for i := 0; i < len(tm.model.TrackVolumes); i++ {
				tm.model.TrackVolumes[i] = msg.Arguments[i].(float32)
			}
		}
	})
	// Build program
	tm = initialModel(config.port, config.project, config.vim, d)

	p := tea.NewProgram(tm, tea.WithAltScreen())

	// Start OSC server after p is created but before p.Run()
	server := &osc.Server{Addr: fmt.Sprintf(":%d", config.port+1), Dispatcher: d}
	go func() {
		log.Printf("Starting OSC server on port %d", config.port+1)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting OSC server: %v", err)
		}
	}()

	// Fast SuperCollider detection and startup
	if !config.skipSC {
		go func() {
			// First, quickly check if sclang process is running
			if !supercollider.IsSuperColliderEnabled() {
				// No sclang process found - start SuperCollider immediately
				log.Printf("No sclang process found, starting SuperCollider")
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			checkAndUpdatePortIfNeeded(tm)
				return
			}

			// sclang is running - wait briefly to see if it has ColliderTracker loaded
			log.Printf("Found sclang process, checking if ColliderTracker is loaded...")
			timeout := time.NewTimer(1 * time.Second)
			defer timeout.Stop()

			select {
			case <-readyChannel:
				// SuperCollider with ColliderTracker is already running
				log.Printf("Found existing SuperCollider instance with ColliderTracker")
				return
			case <-timeout.C:
				// sclang is running but no ColliderTracker - start our own instance
				log.Printf("sclang running but no ColliderTracker detected, starting new instance")
				if err := supercollider.StartSuperColliderWithRecording(config.record); err != nil {
					log.Printf("Failed to start SuperCollider: %v", err)
				}
			checkAndUpdatePortIfNeeded(tm)
			}
		}()
	} else {
		log.Printf("Skipping SuperCollider detection and management entirely (--skip-sc flag provided)")
	}

	// When SC signals readiness via /cpuusage, hide the splash
	go func() {
		if config.skipSC {
			p.Send(scReadyMsg{}) // skip splash if skipping SC management
		} else {
			<-readyChannel
			log.Printf("Received SuperCollider ready; hiding splash")
			p.Send(scReadyMsg{})
		}
	}()

	// hack to make sure Ctrl+V works on Windows
	hacks.StoreWinClipboard()

	finalModel, err := p.Run()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Check if we should return to project selection
	if finalModel != nil {
		if trackerModel, ok := finalModel.(*TrackerModel); ok && trackerModel.model.ReturnToProjectSelector {
			log.Printf("Returning to project selection...")
			// Clean up current session
			supercollider.Cleanup()

			// Run project selector again
			selectedPath, cancelled, isNewProject := project.RunProjectSelector()
			if !cancelled {
				if isNewProject {
					// User chose to create new project with provided name
					config.project = selectedPath
				} else {
					// User selected an existing project
					config.project = selectedPath
				}
				config.projectProvided = true // Mark as provided to skip selector
				// Restart the main function logic
				restartWithProject()
				return
			}
		}
	}

	// Always call cleanup when the program exits normally (e.g., Ctrl+Q)
	supercollider.Cleanup()
}

func initialModel(oscPort int, saveFolder string, vimMode bool, dispatcher *osc.StandardDispatcher) *TrackerModel {
	m := model.NewModel(oscPort, saveFolder, vimMode)

	// Try to load saved state
	if err := storage.LoadState(m, oscPort, saveFolder); err == nil {
		log.Printf("Loaded saved state successfully from %s", saveFolder)
	} else {
		log.Printf("No saved state found or error loading from %s: %v", saveFolder, err)
		// Load files for new model
		storage.LoadFiles(m)
	}

	// Note: Preference OSC messages are now sent when first CPU message is received
	// to ensure SuperCollider is ready to receive them

	// Add waveform handler to the existing OSC dispatcher
	dispatcher.AddMsgHandler("/waveform", func(msg *osc.Message) {
		sample := float64(msg.Arguments[0].(float32)) // expected in [-1,+1]
		m.LastWaveform = sample
		// available content width inside the padded container (2 spaces each side)
		maxCols := m.TermWidth - 4
		if maxCols < 1 {
			maxCols = 1
		}
		m.PushWaveformSample(sample, maxCols*2/3)
	})
	// Add track waveform handler to the existing OSC dispatcher
	dispatcher.AddMsgHandler("/track_waveform", func(msg *osc.Message) {
		// available content width inside the padded container (2 spaces each side)
		maxCols := m.TermWidth - 4
		if maxCols < 1 {
			maxCols = 1
		}
		maxCols = maxCols * 2 / 3
		for i := 0; i < len(m.TrackWaveformBuf); i++ {
			m.PushTrackWaveformSample(i, float64(msg.Arguments[i].(float32)), maxCols)
		}
	})

	m.AvailableMidiDevices = midiconnector.Devices()
	for _, device := range m.AvailableMidiDevices {
		log.Printf("MIDI device found: %+v", device)
	}

	// Set default MIDI device to first available device (only for unset devices)
	if len(m.AvailableMidiDevices) > 0 {
		firstDevice := m.AvailableMidiDevices[0]
		// Only update MIDI settings that are still set to "None" (preserve user selections)
		for i := 0; i < 255; i++ {
			if m.MidiSettings[i].Device == "None" {
				m.MidiSettings[i].Device = firstDevice
				// Channel is already set to "1" by default in initializeDefaultData()
			}
		}
		log.Printf("Default MIDI device set to: %s (for unset devices only)", firstDevice)
	}

	return &TrackerModel{
		model:         m,
		splashState:   views.NewSplashState(36 * time.Second / 10), // 3.6 seconds (20% slower)
		showingSplash: true,                                        // splash is ALWAYS shown until SC ready
	}
}

// TrackerModel wraps the model and implements the tea.Model interface
type TrackerModel struct {
	model         *model.Model
	splashState   *views.SplashState
	showingSplash bool
}

// WaveformTickMsg is a special message that fires at a steady UI rate (30fps)
// to refresh/redraw waveform and UI without advancing playback.
type WaveformTickMsg struct{}

// SplashTickMsg drives the splash screen animation
type SplashTickMsg struct{}

// tickWaveform schedules the next WaveformTickMsg at the requested fps.
func tickWaveform(fps int) tea.Cmd {
	if fps <= 0 {
		fps = 30
	}
	interval := time.Second / time.Duration(fps)
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return WaveformTickMsg{}
	})
}

// tickSplash schedules the next SplashTickMsg for smooth animation
func tickSplash() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return SplashTickMsg{}
	})
}

func (tm *TrackerModel) Init() tea.Cmd {
	if tm.showingSplash {
		// Start splash screen animation at 60fps
		return tickSplash()
	}
	// Start a 30fps UI loop so the waveform redraws smoothly.
	// Playback advancement stays on its own schedule (input.TickMsg).
	return tickWaveform(30)
}

func (tm *TrackerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tm.model.TermHeight = msg.Height
		tm.model.TermWidth = msg.Width
		// keep the appropriate loop going
		if tm.showingSplash {
			return tm, nil
		}
		return tm, nil

	case SplashTickMsg:
		// Keep animating the splash; do NOT auto-dismiss on duration.
		// We'll exit the splash only on scReadyMsg or a keypress.
		return tm, tickSplash()

	case WaveformTickMsg:
		// Redraw UI/waveform at 30fps. Do NOT advance playback here.
		// Reschedule the next UI tick.
		if tm.showingSplash {
			return tm, nil
		}
		return tm, tickWaveform(30)

	case input.TickMsg:
		// Tempo/engine ticks: only advance playback here, at your musical rate.
		if tm.model.IsPlaying {
			input.AdvancePlayback(tm.model)
			// Reschedule the next tempo tick according to your input package.
			return tm, input.Tick(tm.model)
		}
		return tm, nil

	case scReadyMsg:
		// SC is ready — leave the splash screen
		tm.showingSplash = false
		return tm, nil

	case tea.KeyMsg:
		// Skip splash screen on any key press
		if tm.showingSplash {
			tm.showingSplash = false
			return tm, tickWaveform(30)
		}
		// Keys may toggle playback, change views, etc.
		return tm, input.HandleKeyInput(tm.model, msg)
	}

	return tm, nil
}

func (tm TrackerModel) View() string {
	if tm.showingSplash {
		return views.RenderSplashScreen(tm.model.TermWidth, tm.model.TermHeight, tm.splashState, Version)
	}

	switch tm.model.ViewMode {
	case types.SongView:
		return views.RenderSongView(tm.model)
	case types.ChainView:
		return views.RenderChainView(tm.model)
	case types.PhraseView:
		return views.RenderPhraseView(tm.model)
	case types.SettingsView:
		return views.RenderSettingsView(tm.model)
	case types.FileMetadataView:
		return views.RenderFileMetadataView(tm.model)
	case types.RetriggerView:
		return views.RenderRetriggerView(tm.model)
	case types.TimestrechView:
		return views.RenderTimestrechView(tm.model)
	case types.ModulateView:
		return views.RenderModulateView(tm.model)
	case types.ArpeggioView:
		return views.RenderArpeggioView(tm.model)
	case types.MidiView:
		return views.RenderMidiView(tm.model)
	case types.SoundMakerView:
		return views.RenderSoundMakerView(tm.model)
	case types.DuckingView:
		return views.RenderDuckingView(tm.model)
	case types.MixerView:
		return views.RenderMixerView(tm.model)
	case types.WaveformView:
		return views.RenderWaveformView(tm.model)
	default: // FileView
		return views.RenderFileView(tm.model)
	}
}

func setupCleanupOnExit() {
	// Handle cleanup on various exit signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-c
		supercollider.Cleanup()
		os.Exit(0)
	}()
}
