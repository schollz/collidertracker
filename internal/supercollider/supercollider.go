package supercollider

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed collidertracker.scd
var embeddedSamplerSCD []byte

//go:embed DX7.afx
var embeddedDX7AFX []byte

//go:embed DX7.scd
var embeddedDX7SCD []byte

var (
	startedBySelf   = false
	tempSamplerFile = ""
	tempDX7AFXFile  = ""
	tempDX7SCDFile  = ""
	sclangProcess   *exec.Cmd
	cleanupCalled   = false
	detectedPort    = int32(0) // Port detected from SuperCollider output, 0 means not detected yet (atomic access)
)

// Pre-compiled regex patterns for port detection (compiled once for performance)
var (
	portPatternTryingUsing = regexp.MustCompile(`(?i)(?:trying|using)\s+port\s+(\d+)`)
	portPatternAddress     = regexp.MustCompile(`(?:address|on)\s+[\d.]+:(\d+)`)
	portPatternBooting     = regexp.MustCompile(`(?i)booting\s+(\d+)`)
)

// portDetectingWriter is an io.Writer that monitors SuperCollider's output
// for port selection messages and extracts the actual port being used
type portDetectingWriter struct {
	logWriter io.Writer
}

// tryExtractPort attempts to extract a port number from the given matches and stores it atomically
func tryExtractPort(matches []string, line string) bool {
	if len(matches) > 1 {
		if port, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			atomic.StoreInt32(&detectedPort, int32(port))
			log.Printf("Detected SuperCollider using port %d from output: %s", port, strings.TrimSpace(line))
			return true
		}
	}
	return false
}

// Write implements io.Writer and scans each line for port information
func (w *portDetectingWriter) Write(p []byte) (n int, err error) {
	// First write to the log
	n, err = w.logWriter.Write(p)
	
	// Then scan for port information
	// SuperCollider may output messages like:
	// "Cannot bind to UDP port 57120, trying port 57121"
	// "Starting server 'localhost' on address 127.0.0.1:57121"
	// "booting 57121"
	line := string(p)
	
	// Try each pattern until one matches, then return early
	// Pattern 1: "trying port XXXXX" or "using port XXXXX"
	if matches := portPatternTryingUsing.FindStringSubmatch(line); tryExtractPort(matches, line) {
		return n, err
	}
	
	// Pattern 2: "address 127.0.0.1:XXXXX" or similar IP:port patterns
	if matches := portPatternAddress.FindStringSubmatch(line); tryExtractPort(matches, line) {
		return n, err
	}
	
	// Pattern 3: "booting XXXXX"
	if matches := portPatternBooting.FindStringSubmatch(line); tryExtractPort(matches, line) {
		return n, err
	}
	
	return n, err
}

func IsJackEnabled() bool {
	// Check for common JACK daemon process names
	jackProcessNames := []string{"jackd", "jackdbus", "jackdmp"}

	for _, processName := range jackProcessNames {
		if isProcessRunning(processName) {
			return true
		}
	}

	return false
}

func IsSuperColliderEnabled() bool {
	return isProcessRunning("sclang")
}

func StartSuperCollider() error {
	return StartSuperColliderWithRecording(false)
}

func StartSuperColliderWithRecording(enableRecording bool) error {
	if IsSuperColliderEnabled() {
		return nil // Already running (started externally)
	}

	// Find sclang executable
	sclangPath, err := findSclangPath()
	if err != nil {
		return fmt.Errorf("sclang not found: %v", err)
	}

	// Create temporary files from embedded SuperCollider files
	// Create collidertracker.scd with optional recording modification
	tempFile, err := os.CreateTemp("", "sampler-*.scd")
	if err != nil {
		return fmt.Errorf("failed to create temporary sampler file: %v", err)
	}

	// Modify the embedded content if recording is enabled
	scdContent := embeddedSamplerSCD
	if enableRecording {
		log.Printf("enableRecording is true")
		// Replace "//Server.default.record;" with "Server.default.record;"
		modified := strings.Replace(string(embeddedSamplerSCD), "//Server.default.record;", "Server.default.record;", 1)
		scdContent = []byte(modified)
	}

	_, err = tempFile.Write(scdContent)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return fmt.Errorf("failed to write sampler content: %v", err)
	}
	tempFile.Close()
	tempSamplerFile = tempFile.Name()

	// Get the directory of the sampler file to place DX7 files alongside
	tempDir := filepath.Dir(tempSamplerFile)

	// Create DX7.afx in the same directory
	dx7AFXPath := filepath.Join(tempDir, "DX7.afx")
	err = os.WriteFile(dx7AFXPath, embeddedDX7AFX, 0644)
	if err != nil {
		os.Remove(tempSamplerFile)
		return fmt.Errorf("failed to write DX7.afx: %v", err)
	}
	tempDX7AFXFile = dx7AFXPath

	// Create DX7.scd in the same directory
	dx7SCDPath := filepath.Join(tempDir, "DX7.scd")
	err = os.WriteFile(dx7SCDPath, embeddedDX7SCD, 0644)
	if err != nil {
		os.Remove(tempSamplerFile)
		os.Remove(tempDX7AFXFile)
		return fmt.Errorf("failed to write DX7.scd: %v", err)
	}
	tempDX7SCDFile = dx7SCDPath

	// Start sclang with the temporary scd file
	sclangProcess = exec.Command(sclangPath, tempSamplerFile)

	// On Windows, set working directory to sclang's directory so it can find scsynth
	if runtime.GOOS == "windows" {
		sclangProcess.Dir = filepath.Dir(sclangPath)
	}

	// Set up platform-specific process attributes
	setupProcessGroup(sclangProcess)

	// Create a port-detecting writer that wraps the log writer
	// This will monitor SuperCollider's output for port selection messages
	portWriter := &portDetectingWriter{logWriter: log.Writer()}
	
	// Redirect SuperCollider output through the port detector
	sclangProcess.Stdout = portWriter
	sclangProcess.Stderr = portWriter

	// Start the process but don't wait for it to complete
	err = sclangProcess.Start()
	if err != nil {
		os.Remove(tempSamplerFile)
		os.Remove(tempDX7AFXFile)
		os.Remove(tempDX7SCDFile)
		tempSamplerFile = ""
		tempDX7AFXFile = ""
		tempDX7SCDFile = ""
		return fmt.Errorf("failed to start SuperCollider: %v", err)
	}

	// Mark that we started it
	startedBySelf = true

	// Wait a moment and check if it's actually running
	time.Sleep(2 * time.Second)
	if !IsSuperColliderEnabled() {
		// Clean up if it failed to start
		if sclangProcess != nil && sclangProcess.Process != nil {
			sclangProcess.Process.Kill()
		}
		os.Remove(tempSamplerFile)
		os.Remove(tempDX7AFXFile)
		os.Remove(tempDX7SCDFile)
		tempSamplerFile = ""
		tempDX7AFXFile = ""
		tempDX7SCDFile = ""
		startedBySelf = false
		return fmt.Errorf("SuperCollider failed to start properly")
	}

	return nil
}

func StartSuperColliderWithProgress(readyChannel <-chan struct{}) error {
	if IsSuperColliderEnabled() {
		return nil // Already running (started externally)
	}

	// Create and run the progress bar TUI
	model := NewStartupProgressModel(readyChannel)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if there was an error during startup
	if progressModel, ok := finalModel.(StartupProgressModel); ok {
		return progressModel.Error()
	}

	return nil
}

// GetDetectedPort returns the port that SuperCollider actually started on.
// Returns 0 if no port has been detected yet.
// Thread-safe via atomic operations.
func GetDetectedPort() int {
	return int(atomic.LoadInt32(&detectedPort))
}

// ResetDetectedPort resets the detected port (useful for testing or restarting)
// Thread-safe via atomic operations.
func ResetDetectedPort() {
	atomic.StoreInt32(&detectedPort, 0)
}

func Cleanup() {
	// Prevent multiple cleanup calls
	if cleanupCalled {
		return
	}
	cleanupCalled = true

	if startedBySelf {
		// Stop SuperCollider process if we started it
		if sclangProcess != nil && sclangProcess.Process != nil {
			// Use platform-specific process termination
			killProcessGroup(sclangProcess)
			// Wait for the process to actually stop (with timeout)
			done := make(chan error, 1)
			go func() {
				done <- sclangProcess.Wait()
			}()
			select {
			case <-done:
				// Process finished normally
			case <-time.After(2 * time.Second):
				// Timeout - process may still be running but we've done our best
				log.Printf("Timeout waiting for sclang process to terminate")
			}
		}
		startedBySelf = false
		sclangProcess = nil
	}

	// Reset detected port
	ResetDetectedPort()

	// Remove temporary files if we created them
	if tempSamplerFile != "" {
		os.Remove(tempSamplerFile)
		tempSamplerFile = ""
	}
	if tempDX7AFXFile != "" {
		os.Remove(tempDX7AFXFile)
		tempDX7AFXFile = ""
	}
	if tempDX7SCDFile != "" {
		os.Remove(tempDX7SCDFile)
		tempDX7SCDFile = ""
	}
}

func WasStartedBySelf() bool {
	return startedBySelf
}

// GetSynthDefNames extracts all SynthDef names from the embedded SuperCollider file
func GetSynthDefNames() []string {
	return ExtractSynthDefNames(string(embeddedSamplerSCD))
}

// ExtractSynthDefNames extracts synthdef names from SuperCollider code
// It looks for patterns like SynthDef("name", ... and SynthDef(\name, ...
func ExtractSynthDefNames(scdContent string) []string {
	// Regex pattern to match both quoted strings and symbols
	// SynthDef("name" or SynthDef(\name
	pattern := `SynthDef\s*\(\s*(?:"([^"]+)"|\\([^,\s\)]+))`

	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(scdContent, -1)

	var names []string
	for _, match := range matches {
		if match[1] != "" {
			// Quoted string match (group 1)
			names = append(names, match[1])
		} else if match[2] != "" {
			// Symbol match (group 2)
			names = append(names, match[2])
		}
	}

	namesToRemove := map[string]bool{
		"sampler":       true,
		"externalInput": true,
		"playback":      true,
		"diskout":       true,
		"out":           true,
	}
	var filteredNames []string
	for _, name := range names {
		if !namesToRemove[name] {
			filteredNames = append(filteredNames, name)
		}
	}

	return filteredNames
}

func findSclangPath() (string, error) {
	// First try to find sclang in PATH
	if path, err := exec.LookPath("sclang"); err == nil {
		return path, nil
	}

	// Platform-specific fallback locations
	var possiblePaths []string

	switch runtime.GOOS {
	case "windows":
		// Search for SuperCollider folders with version numbers in Program Files
		programFilesDirs := []string{
			"C:\\Program Files",
			"C:\\Program Files (x86)",
		}

		for _, baseDir := range programFilesDirs {
			if scDir := findSuperColliderDir(baseDir); scDir != "" {
				possiblePaths = append(possiblePaths, filepath.Join(scDir, "sclang.exe"))
			}
		}

		// Fallback to exact paths (in case someone has a custom installation)
		possiblePaths = append(possiblePaths,
			"C:\\Program Files\\SuperCollider\\sclang.exe",
			"C:\\Program Files (x86)\\SuperCollider\\sclang.exe",
		)

		// Also check user's local app data
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			possiblePaths = append(possiblePaths, filepath.Join(localAppData, "SuperCollider", "sclang.exe"))
		}
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			possiblePaths = append(possiblePaths, filepath.Join(programData, "SuperCollider", "sclang.exe"))
		}

	case "darwin":
		// Common installation paths on macOS
		possiblePaths = []string{
			"/Applications/SuperCollider.app/Contents/MacOS/sclang",
			"/Applications/SuperCollider/SuperCollider.app/Contents/MacOS/sclang",
		}
		// Also check user's Applications
		if homeDir, err := os.UserHomeDir(); err == nil {
			possiblePaths = append(possiblePaths,
				filepath.Join(homeDir, "Applications", "SuperCollider.app", "Contents", "MacOS", "sclang"),
				filepath.Join(homeDir, "Applications", "SuperCollider", "SuperCollider.app", "Contents", "MacOS", "sclang"),
			)
		}

	case "linux":
		// Common installation paths on Linux
		possiblePaths = []string{
			"/usr/bin/sclang",
			"/usr/local/bin/sclang",
			"/opt/supercollider/bin/sclang",
		}
		// Also check user's local bin
		if homeDir, err := os.UserHomeDir(); err == nil {
			possiblePaths = append(possiblePaths,
				filepath.Join(homeDir, ".local", "bin", "sclang"),
				filepath.Join(homeDir, "bin", "sclang"),
			)
		}
	}

	// Check each possible path
	for _, path := range possiblePaths {
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("sclang executable not found in PATH or common installation locations")
}

func isProcessRunning(processName string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Exact image match on Windows is already supported by tasklist
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq "+processName+".exe")
	default: // darwin, linux, etc.
		// Use -x for exact match of the process name (no substring matches like "jackdbus")
		cmd = exec.Command("pgrep", "-x", processName)
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		out := strings.ToLower(string(output))
		return strings.Contains(out, strings.ToLower(processName+".exe"))
	}

	// pgrep returns PIDs if found; empty output means not running
	return len(strings.TrimSpace(string(output))) > 0
}

func HasRequiredExtensions() bool {
	extensions := []string{"Fverb.sc", "AnalogTape.sc", "MiBraids.sc"}

	for _, ext := range extensions {
		if !hasExtension(ext) {
			return false
		}
	}

	// Also check for Open303
	if !hasOpen303() {
		return false
	}

	// Also check for Juno60
	if !hasJuno60() {
		return false
	}

	return true
}

func hasExtension(filename string) bool {
	extensionDirs := getSuperColliderExtensionDirs()

	for _, dir := range extensionDirs {
		// Check direct file path
		if fileExists(filepath.Join(dir, filename)) {
			return true
		}

		// Check in subdirectories recursively
		found := false
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == filename {
				found = true
				return filepath.SkipDir
			}
			return nil
		})

		if found {
			return true
		}
	}
	return false
}

func getSuperColliderExtensionDirs() []string {
	var dirs []string

	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions"))
		}
		dirs = append(dirs, "/Library/Application Support/SuperCollider/Extensions")
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(homeDir, ".local/share/SuperCollider/Extensions"))
		}
		dirs = append(dirs, "/usr/share/SuperCollider/Extensions")
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "SuperCollider/Extensions"))
		}
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			dirs = append(dirs, filepath.Join(programData, "SuperCollider/Extensions"))
		}
	}

	return dirs
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// findSuperColliderDir searches for a SuperCollider installation directory
// in the given base directory, looking for folders that start with "SuperCollider"
func findSuperColliderDir(baseDir string) string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "SuperCollider") {
			scDir := filepath.Join(baseDir, entry.Name())
			// Verify sclang.exe exists in this directory
			if fileExists(filepath.Join(scDir, "sclang.exe")) {
				return scDir
			}
		}
	}

	return ""
}

func DownloadRequiredExtensions() error {
	extensionDir := getLocalExtensionDir()
	if extensionDir == "" {
		return fmt.Errorf("could not determine local extension directory")
	}

	// Create extension directory if it doesn't exist
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("failed to create extension directory: %v", err)
	}

	// Check for PortedPlugins extensions
	if !hasExtension("Fverb.sc") || !hasExtension("AnalogTape.sc") {
		fmt.Println("Downloading PortedPlugins extensions...")
		downloadURL := getPortedPluginsURL()
		if downloadURL == "" {
			return fmt.Errorf("unsupported platform for PortedPlugins: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		if err := downloadAndExtract(downloadURL, extensionDir); err != nil {
			return fmt.Errorf("failed to download PortedPlugins: %v", err)
		}
		fmt.Println("PortedPlugins downloaded successfully")
	}

	// Check for mi-UGens extensions
	if !hasExtension("MiBraids.sc") {
		fmt.Println("Downloading mi-UGens extensions...")
		downloadURL := getMiUGensURL()
		if downloadURL == "" {
			return fmt.Errorf("unsupported platform for mi-UGens: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		if err := downloadAndExtract(downloadURL, extensionDir); err != nil {
			return fmt.Errorf("failed to download mi-UGens: %v", err)
		}
		fmt.Println("mi-UGens downloaded successfully")
	}

	// Check for Open303
	if !hasOpen303() {
		fmt.Println("Downloading Open303...")
		downloadURL := getOpen303URL()
		if downloadURL == "" {
			return fmt.Errorf("unsupported platform for Open303: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		open303Dir := getOpen303InstallDir()
		if open303Dir == "" {
			return fmt.Errorf("could not determine Open303 installation directory")
		}

		// Create Open303 directory if it doesn't exist
		if err := os.MkdirAll(open303Dir, 0755); err != nil {
			return fmt.Errorf("failed to create Open303 directory: %v", err)
		}

		if err := downloadAndExtractOpen303(downloadURL, open303Dir); err != nil {
			return fmt.Errorf("failed to download Open303: %v", err)
		}
		fmt.Println("Open303 downloaded successfully")
	}

	// Check for Juno60
	if !hasJuno60() {
		fmt.Println("Downloading Juno60...")
		downloadURL := getJuno60URL()
		if downloadURL == "" {
			return fmt.Errorf("unsupported platform for Juno60: %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		juno60Dir := getJuno60InstallDir()
		if juno60Dir == "" {
			return fmt.Errorf("could not determine Juno60 installation directory")
		}

		// Create Juno60 directory if it doesn't exist
		if err := os.MkdirAll(juno60Dir, 0755); err != nil {
			return fmt.Errorf("failed to create Juno60 directory: %v", err)
		}

		if err := downloadAndExtractJuno60(downloadURL, juno60Dir); err != nil {
			return fmt.Errorf("failed to download Juno60: %v", err)
		}
		fmt.Println("Juno60 downloaded successfully")
	}

	if HasRequiredExtensions() {
		fmt.Println("All required extensions are now available")
		return nil
	}

	return fmt.Errorf("failed to install all required extensions")
}

func getPortedPluginsURL() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
			return "https://github.com/schollz/portedplugins/releases/download/v0.4.6/PortedPlugins-RaspberryPi.zip"
		}
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-Linux.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-macOS-ARM.zip"
		}
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-macOS.zip"
	case "windows":
		return "https://github.com/schollz/portedplugins/releases/download/v0.4.5/PortedPlugins-Windows.zip"
	}
	return ""
}

func getMiUGensURL() string {
	switch runtime.GOOS {
	case "linux":
		return "https://github.com/v7b1/mi-UGens/releases/download/v0.0.8/mi-UGens-Linux.zip"
	case "darwin":
		return "https://github.com/v7b1/mi-UGens/releases/download/v0.0.8/mi-UGens-macOS.zip"
	case "windows":
		return "https://github.com/v7b1/mi-UGens/releases/download/v0.0.8/mi-UGens-Windows.zip"
	}
	return ""
}

func getOpen303URL() string {
	// Using the latest release from https://github.com/schollz/open303
	const baseURL = "https://github.com/schollz/open303/releases/latest/download/"

	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return baseURL + "Open303-Linux-arm64.zip"
		}
		return baseURL + "Open303-Linux-x64.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return baseURL + "Open303-macOS-arm64.zip"
		}
		// x64 macOS - use arm64 as it will work via Rosetta
		return baseURL + "Open303-macOS-arm64.zip"
	case "windows":
		return baseURL + "Open303-Windows-x64.zip"
	}
	return ""
}

func getJuno60URL() string {
	// Using the latest release from https://github.com/schollz/juno-60
	const baseURL = "https://github.com/schollz/juno-60/releases/latest/download/"

	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return baseURL + "Juno60-Linux-arm64.zip"
		}
		return baseURL + "Juno60-Linux-x64.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return baseURL + "Juno60-macOS-arm64.zip"
		}
		// x64 macOS - use arm64 as it will work via Rosetta
		return baseURL + "Juno60-macOS-arm64.zip"
	case "windows":
		return baseURL + "Juno60-Windows-x64.zip"
	}
	return ""
}

func getOpen303InstallDir() string {
	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions/Open303")
		}
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, ".local/share/SuperCollider/Extensions/Open303")
		}
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "SuperCollider/Extensions/Open303")
		}
	}
	return ""
}

func getJuno60InstallDir() string {
	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions/Juno60")
		}
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, ".local/share/SuperCollider/Extensions/Juno60")
		}
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "SuperCollider/Extensions/Juno60")
		}
	}
	return ""
}

func hasOpen303() bool {
	installDir := getOpen303InstallDir()
	if installDir == "" {
		return false
	}

	// Check for Open303.sc file recursively in the Extensions directory
	extensionDirs := getSuperColliderExtensionDirs()

	for _, dir := range extensionDirs {
		// Check direct file path
		if fileExists(filepath.Join(dir, "Open303.sc")) {
			return true
		}

		// Check in subdirectories recursively
		found := false
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "Open303.sc" {
				found = true
				return filepath.SkipDir
			}
			return nil
		})

		if found {
			return true
		}
	}
	return false
}

func hasJuno60() bool {
	installDir := getJuno60InstallDir()
	if installDir == "" {
		return false
	}

	// Check for Juno60.sc file recursively in the Extensions directory
	extensionDirs := getSuperColliderExtensionDirs()

	for _, dir := range extensionDirs {
		// Check direct file path
		if fileExists(filepath.Join(dir, "Juno60.sc")) {
			return true
		}

		// Check in subdirectories recursively
		found := false
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "Juno60.sc" {
				found = true
				return filepath.SkipDir
			}
			return nil
		})

		if found {
			return true
		}
	}
	return false
}

func getLocalExtensionDir() string {
	switch runtime.GOOS {
	case "darwin":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, "Library/Application Support/SuperCollider/Extensions")
		}
	case "linux":
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, ".local/share/SuperCollider/Extensions")
		}
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "SuperCollider/Extensions")
		}
	}
	return ""
}

func downloadAndExtract(url, destDir string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "portedplugins-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy response body to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}

	// Close temp file before reading
	tmpFile.Close()

	// Extract zip file
	return extractZip(tmpFile.Name(), destDir)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// Create destination directory
	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		// Create the directories for this file
		destPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.FileInfo().Mode())
			continue
		}

		// Create the directories for this file
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		// Open file in zip
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		// Create destination file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file: %v", err)
		}

		// Copy file contents
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}
	}

	return nil
}

// downloadAndExtractOpen303 downloads and extracts Open303, similar to downloadAndExtract
// but with special handling to set executable permissions on the binary
func downloadAndExtractOpen303(url, destDir string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "open303-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy response body to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}

	// Close temp file before reading
	tmpFile.Close()

	// Extract zip file and set executable permissions
	if err := extractZipWithExecutable(tmpFile.Name(), destDir, "Open303"); err != nil {
		return err
	}

	return nil
}

// downloadAndExtractJuno60 downloads and extracts Juno60, similar to downloadAndExtractOpen303
// but with special handling to set executable permissions on the binary
func downloadAndExtractJuno60(url, destDir string) error {
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "juno60-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy response body to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %v", err)
	}

	// Close temp file before reading
	tmpFile.Close()

	// Extract zip file and set executable permissions
	if err := extractZipWithExecutable(tmpFile.Name(), destDir, "Juno60"); err != nil {
		return err
	}

	return nil
}

func extractZipWithExecutable(src, dest, executableName string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	// Create destination directory
	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		destPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.FileInfo().Mode())
			continue
		}

		// Ensure parent directories exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		// Open file in zip
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		// Determine file mode - make executable if it's the main binary
		fileMode := f.FileInfo().Mode()
		if strings.Contains(f.Name, executableName) {
			// Make executable for owner, group, and others (0755)
			fileMode = 0755
		}

		// Create destination file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file: %v", err)
		}

		// Copy file contents
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}
	}

	return nil
}
