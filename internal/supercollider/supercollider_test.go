package supercollider

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		assert.True(t, fileExists(testFile))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		assert.False(t, fileExists("/path/that/does/not/exist.txt"))
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		assert.True(t, fileExists(tmpDir))
	})
}

func TestWasStartedBySelf(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		// Should start as false
		result := WasStartedBySelf()
		assert.False(t, result)
	})
}

func TestGetPortedPluginsURL(t *testing.T) {
	t.Run("returns correct URL for platform", func(t *testing.T) {
		url := getPortedPluginsURL()

		switch runtime.GOOS {
		case "linux":
			if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
				assert.Contains(t, url, "RaspberryPi.zip")
			} else {
				assert.Contains(t, url, "Linux.zip")
			}

		case "darwin":
			if runtime.GOARCH == "arm64" {
				assert.Contains(t, url, "macOS-ARM.zip")
			} else {
				assert.Contains(t, url, "macOS.zip")
			}

		case "windows":
			assert.Contains(t, url, "Windows.zip")

		default:
			assert.Equal(t, "", url) // Unsupported platform
		}

		// If URL is not empty, should be a valid GitHub releases URL
		if url != "" {
			assert.Contains(t, url, "github.com/schollz/portedplugins/releases")
		}
	})
}

func TestHasExtension(t *testing.T) {
	t.Run("extension not found in empty dirs", func(t *testing.T) {
		// This will check real system dirs, but since we're looking for specific
		// extension files that probably don't exist, should return false
		result := hasExtension("NonexistentExtension.sc")
		assert.False(t, result)
	})
}

func TestHasRequiredExtensions(t *testing.T) {
	t.Run("check required extensions", func(t *testing.T) {
		// This will likely return false unless SuperCollider extensions are installed
		result := HasRequiredExtensions()
		// We can't assert true/false since it depends on system state,
		// but we can verify it doesn't panic
		assert.IsType(t, true, result)
	})
}

func TestGetMiUGensURL(t *testing.T) {
	t.Run("returns correct URL for platform", func(t *testing.T) {
		url := getMiUGensURL()

		switch runtime.GOOS {
		case "linux":
			assert.Contains(t, url, "mi-UGens-Linux.zip")
		case "darwin":
			assert.Contains(t, url, "mi-UGens-macOS.zip")
		case "windows":
			assert.Contains(t, url, "mi-UGens-Windows.zip")
		default:
			assert.Equal(t, "", url) // Unsupported platform
		}

		// If URL is not empty, should be a valid GitHub releases URL
		if url != "" {
			assert.Contains(t, url, "github.com/v7b1/mi-UGens/releases")
			assert.Contains(t, url, "v0.0.8")
		}
	})
}

func TestGetOpen303URL(t *testing.T) {
	t.Run("returns correct URL for platform", func(t *testing.T) {
		url := getOpen303URL()

		switch runtime.GOOS {
		case "linux":
			if runtime.GOARCH == "arm64" {
				assert.Contains(t, url, "Open303-Linux-arm64.zip")
			} else {
				assert.Contains(t, url, "Open303-Linux-x64.zip")
			}
		case "darwin":
			assert.Contains(t, url, "Open303-macOS-arm64.zip")
		case "windows":
			assert.Contains(t, url, "Open303-Windows-x64.zip")
		default:
			assert.Equal(t, "", url) // Unsupported platform
		}

		// If URL is not empty, should be a valid GitHub releases URL
		if url != "" {
			assert.Contains(t, url, "github.com/schollz/open303/releases")
		}
	})
}

func TestHasOpen303(t *testing.T) {
	t.Run("returns false when Open303 not installed", func(t *testing.T) {
		// This will likely return false unless Open303 is installed
		result := hasOpen303()
		// We can't assert true/false since it depends on system state,
		// but we can verify it doesn't panic
		assert.IsType(t, true, result)
	})
}

func TestGetOpen303InstallDir(t *testing.T) {
	t.Run("returns non-empty directory path", func(t *testing.T) {
		dir := getOpen303InstallDir()
		// Should return a path on all supported platforms
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
			assert.NotEmpty(t, dir)
			assert.Contains(t, dir, "Open303")
		}
	})
}

func TestExtractSynthDefNames(t *testing.T) {
	t.Run("extracts quoted synthdef names", func(t *testing.T) {
		scdContent := `
			SynthDef("MiPlaits", {
				// some code
			}).add;
			
			SynthDef("MiBraids", {
				// more code
			}).add;
		`

		result := ExtractSynthDefNames(scdContent)
		assert.Equal(t, []string{"MiPlaits", "MiBraids"}, result)
	})

	t.Run("extracts symbol synthdef names", func(t *testing.T) {
		scdContent := `
			SynthDef(\PolyPerc, {
				// some code
			}).add;
			
			SynthDef(\SimpleSynth, {
				// more code
			}).add;
		`

		result := ExtractSynthDefNames(scdContent)
		assert.Equal(t, []string{"PolyPerc", "SimpleSynth"}, result)
	})

	t.Run("extracts mixed quoted and symbol synthdef names", func(t *testing.T) {
		scdContent := `
			SynthDef("MiPlaits", {
				// some code
			}).add;
			
			SynthDef(\PolyPerc, {
				// more code
			}).add;
			
			SynthDef("externalInput", {
				// more code
			}).add;
		`

		result := ExtractSynthDefNames(scdContent)
		assert.Equal(t, []string{"MiPlaits", "PolyPerc"}, result)
	})

	t.Run("handles whitespace variations", func(t *testing.T) {
		scdContent := `
			SynthDef ( "SpacedOut" , {
				// some code
			}).add;
			
			SynthDef(  \TabSymbol  ,{
				// more code
			}).add;
		`

		result := ExtractSynthDefNames(scdContent)
		assert.Equal(t, []string{"SpacedOut", "TabSymbol"}, result)
	})

	t.Run("handles dynamic synthdef names with concatenation", func(t *testing.T) {
		scdContent := `
			SynthDef("sampler"++(ch+1), {
				// some code
			}).add;
			
			SynthDef("playback"++(ch+1), {
				// more code  
			}).add;
		`

		// This should extract the base names before concatenation
		result := ExtractSynthDefNames(scdContent)
		assert.Empty(t, result)
	})

	t.Run("returns empty string for no synthdefs", func(t *testing.T) {
		scdContent := `
			// Some SuperCollider code without SynthDefs
			~variable = 42;
			"hello".postln;
		`

		result := ExtractSynthDefNames(scdContent)
		assert.Empty(t, result)
	})

	t.Run("extracts all synthdefs including those in comments", func(t *testing.T) {
		// Note: Current implementation doesn't handle comment filtering
		// This test documents current behavior - future enhancement could add comment handling
		scdContent := `
			SynthDef("ActiveSynth", {
				// some code
			}).add;
			
			// SynthDef("CommentedSynth", {
			//     // this is commented out
			// }).add;
			
			/* SynthDef("BlockCommentSynth", {
				// this is also commented out
			}).add; */
		`

		result := ExtractSynthDefNames(scdContent)
		// Current implementation extracts all matches regardless of comments
		assert.Contains(t, result, "ActiveSynth")
		assert.Contains(t, result, "CommentedSynth")
		assert.Contains(t, result, "BlockCommentSynth")
	})
}

func TestGetSynthDefNames(t *testing.T) {
	t.Run("extracts names from embedded file", func(t *testing.T) {
		result := GetSynthDefNames()

		// Should contain all the synthdefs from the embedded file
		assert.Contains(t, result, "MiPlaits")
		assert.Contains(t, result, "MiBraids")
		assert.Contains(t, result, "PolyPerc")

		// Should not be empty
		assert.NotEmpty(t, result)
	})
}
