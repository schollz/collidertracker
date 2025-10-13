package views

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// SplashState represents the animation state of the splash screen
type SplashState struct {
	StartTime time.Time
	Duration  time.Duration
}

// NewSplashState creates a new splash state
func NewSplashState(duration time.Duration) *SplashState {
	return &SplashState{
		StartTime: time.Now(),
		Duration:  duration,
	}
}

// IsComplete returns true if the splash animation is done
func (s *SplashState) IsComplete() bool {
	return time.Since(s.StartTime) >= s.Duration
}

// GetProgress returns animation progress from 0.0 to 1.0
func (s *SplashState) GetProgress() float64 {
	elapsed := time.Since(s.StartTime)
	if elapsed >= s.Duration {
		return 1.0
	}
	return float64(elapsed) / float64(s.Duration)
}

// RenderSplashScreen renders the animated splash screen
func RenderSplashScreen(termWidth, termHeight int, state *SplashState, version string) string {
	var content strings.Builder

	// Handle edge cases for small or zero terminal dimensions
	if termWidth <= 0 || termHeight <= 0 {
		return "collidertracker\nby infinite digits\ngithub.com/schollz/collidertracker\nversion " + version
	}

	// Ensure minimum dimensions for proper animation
	if termWidth < 40 {
		termWidth = 40
	}
	if termHeight < 10 {
		termHeight = 10
	}

	// Get animation progress
	progress := state.GetProgress()

	// Calculate center position
	centerY := termHeight / 2
	centerX := termWidth / 2

	// Fill screen with empty lines up to center
	for i := 0; i < centerY-6; i++ {
		content.WriteString("\n")
	}

	// Render animated blocks
	blocksLine := renderAnimatedBlocks(centerX, progress)
	content.WriteString(blocksLine)
	content.WriteString("\n\n")

	// Style for the text
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center)

	urlStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Align(lipgloss.Center)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Align(lipgloss.Center)

	// Enhanced text rendering with typewriter and bounce effects
	termTime := time.Now().UnixMilli()

	// Title with enhanced animation
	if progress > 0.05 {
		titleProgress := math.Min((progress-0.05)/0.05, 1.0)

		// Bounce effect for title
		titleBounce := bounceEaseOut(titleProgress)
		titleScale := int(titleBounce * 3) // Scale from 0 to 3

		// Enhanced title style with dynamic effects
		titleStyleEnhanced := titleStyle
		if titleProgress > 0.8 {
			// Add glow effect near completion
			pulse := math.Sin(float64(termTime)/44.0)*0.3 + 0.7
			if pulse > 0.6 {
				titleStyleEnhanced = titleStyleEnhanced.Foreground(lipgloss.Color("226")) // Yellow glow
			}
		}

		titleText := "collidertracker"
		if titleScale > 0 {
			titleText = strings.Repeat(" ", titleScale) + titleText + strings.Repeat(" ", titleScale)
		}

		// Add side decorations
		leftDecor, rightDecor := renderSideDecorations(progress, 0)
		titleTextWithDecor := leftDecor + titleText + rightDecor

		titleLine := titleStyleEnhanced.Width(termWidth).Render(titleTextWithDecor)
		content.WriteString(titleLine)
		content.WriteString("\n")
	}

	// Subtitle with typewriter effect
	if progress > 0.1 {
		subtitleProgress := math.Min((progress-0.1)/0.08, 1.0)

		fullText := "by infinite digits"
		visibleChars := int(float64(len(fullText)) * subtitleProgress)
		if visibleChars > len(fullText) {
			visibleChars = len(fullText)
		}

		displayText := fullText[:visibleChars]

		// Add typing cursor effect
		if subtitleProgress < 1.0 && int(float64(termTime)/330.0)%2 == 0 {
			displayText += "▋"
		}

		// Color transition during typing
		subtitleStyleEnhanced := subtitleStyle
		if subtitleProgress > 0.5 {
			subtitleStyleEnhanced = subtitleStyleEnhanced.Foreground(lipgloss.Color("14"))
		}

		// Add side decorations
		leftDecor, rightDecor := renderSideDecorations(progress, 1)
		displayTextWithDecor := leftDecor + displayText + rightDecor

		subtitleLine := subtitleStyleEnhanced.Width(termWidth).Render(displayTextWithDecor)
		content.WriteString(subtitleLine)
		content.WriteString("\n")
	}

	// URL with character-by-character reveal
	if progress > 0.18 {
		urlProgress := math.Min((progress-0.18)/0.10, 1.0)

		fullURL := "github.com/schollz/collidertracker"
		visibleChars := int(float64(len(fullURL)) * urlProgress)
		if visibleChars > len(fullURL) {
			visibleChars = len(fullURL)
		}

		displayURL := fullURL[:visibleChars]

		// Matrix-style character morphing effect
		if urlProgress < 1.0 && len(displayURL) < len(fullURL) {
			morphChars := []string{"@", "#", "$", "%", "&", "*"}
			morphChar := morphChars[int(float64(termTime)/88.0)%len(morphChars)]
			if int(float64(termTime)/165.0)%3 == 0 {
				displayURL += morphChar
			}
		}

		urlStyleEnhanced := urlStyle
		if urlProgress > 0.7 {
			urlStyleEnhanced = urlStyleEnhanced.Bold(true)
		}

		// Add side decorations
		leftDecor, rightDecor := renderSideDecorations(progress, 2)
		displayURLWithDecor := leftDecor + displayURL + rightDecor

		urlLine := urlStyleEnhanced.Width(termWidth).Render(displayURLWithDecor)
		content.WriteString(urlLine)
		content.WriteString("\n")
	}

	// Version with scaling effect
	if progress > 0.28 {
		versionProgress := math.Min((progress-0.28)/0.06, 1.0)

		// Elastic scaling effect
		versionScale := elasticEaseOut(versionProgress)
		versionText := "version " + version

		// Add spacing for scale effect
		if versionScale < 1.0 {
			spacing := int((1.0 - versionScale) * 2)
			if spacing > 0 {
				spacedText := ""
				for _, char := range versionText {
					spacedText += string(char) + strings.Repeat(" ", spacing)
				}
				versionText = spacedText
			}
		}

		versionStyleEnhanced := subtitleStyle
		if versionProgress > 0.5 {
			versionStyleEnhanced = versionStyleEnhanced.Foreground(lipgloss.Color("8"))
		}

		// Add side decorations
		leftDecor, rightDecor := renderSideDecorations(progress, 3)
		versionTextWithDecor := leftDecor + versionText + rightDecor

		versionLine := versionStyleEnhanced.Width(termWidth).Render(versionTextWithDecor)
		content.WriteString(versionLine)
		content.WriteString("\n")
	}

	// Loading with animated dots and progress
	if progress > 0.35 {
		loadingProgress := math.Min((progress-0.35)/0.15, 1.0)

		baseText := "initializing SuperCollider"

		// Animated dots
		dotCount := int(float64(termTime)/240.0) % 4
		dots := strings.Repeat(".", dotCount)

		// Progress bar effect
		if loadingProgress > 0.3 {
			barLength := 20
			filledLength := int(float64(barLength) * loadingProgress)
			bar := "[" + strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength) + "]"
			baseText += " " + bar
		}

		loadingText := baseText + dots

		// Color intensity based on progress
		loadingStyleEnhanced := loadingStyle
		if loadingProgress > 0.7 {
			loadingStyleEnhanced = loadingStyleEnhanced.Foreground(lipgloss.Color("46")).Bold(true)
		} else if loadingProgress > 0.4 {
			loadingStyleEnhanced = loadingStyleEnhanced.Foreground(lipgloss.Color("42"))
		}

		// Add side decorations
		leftDecor, rightDecor := renderSideDecorations(progress, 4)
		loadingTextWithDecor := leftDecor + loadingText + rightDecor

		loadingLine := loadingStyleEnhanced.Width(termWidth).Render(loadingTextWithDecor)
		content.WriteString(loadingLine)
		content.WriteString("\n")

		// Add status messages during later phases
		if progress > 0.45 {
			statusProgress := (progress - 0.45) / 0.15

			var statusText string
			if statusProgress < 0.3 {
				statusText = "◦ Loading audio engine..."
			} else if statusProgress < 0.6 {
				statusText = "◦ Initializing synthesizers..."
			} else if statusProgress < 0.8 {
				statusText = "◦ Setting up MIDI connections..."
			} else {
				statusText = ""
			}

			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Align(lipgloss.Center)

			if statusProgress > 0.9 {
				statusStyle = statusStyle.Foreground(lipgloss.Color("82")).Bold(true)
			}

			statusLine := statusStyle.Width(termWidth).Render(statusText)
			content.WriteString(statusLine)
			content.WriteString("\n")
		}
	}

	// Add animations below the text as well
	content.WriteString("\n")
	blocksLineBelow := renderAnimatedBlocks(centerX, progress)
	content.WriteString(blocksLineBelow)
	content.WriteString("\n")

	// Fill remaining space
	remainingLines := termHeight - centerY - 12 // Adjust for text and bottom animation
	if remainingLines > 0 {
		for i := 0; i < remainingLines; i++ {
			content.WriteString("\n")
		}
	}

	return content.String()
}

// Particle represents a single animated element
type Particle struct {
	X, Y       float64
	VelX, VelY float64
	Char       string
	Color      string
	Phase      float64
	Age        float64
}

// renderAnimatedBlocks creates the enhanced digital convergence animation
func renderAnimatedBlocks(centerX int, progress float64) string {
	termTime := time.Now().UnixMilli()

	// Animation phases
	var phase1, phase2, phase3, phase4, phase5 float64

	if progress <= 0.25 {
		phase1 = progress / 0.25
	} else if progress <= 0.45 {
		phase1 = 1.0
		phase2 = (progress - 0.25) / 0.20
	} else if progress <= 0.70 {
		phase1, phase2 = 1.0, 1.0
		phase3 = (progress - 0.45) / 0.25
	} else if progress <= 0.90 {
		phase1, phase2, phase3 = 1.0, 1.0, 1.0
		phase4 = (progress - 0.70) / 0.20
	} else {
		phase1, phase2, phase3, phase4 = 1.0, 1.0, 1.0, 1.0
		phase5 = (progress - 0.90) / 0.10
	}

	var lines []string
	maxLines := 7 // Increased from 5 to 7 for more visual space

	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		lines = append(lines, renderAnimationLine(centerX, lineIdx, maxLines, phase1, phase2, phase3, phase4, phase5, termTime))
	}

	return strings.Join(lines, "\n")
}

// renderAnimationLine creates a single line of the complex animation
func renderAnimationLine(centerX, lineIdx, maxLines int, phase1, phase2, phase3, phase4, phase5 float64, termTime int64) string {
	var line strings.Builder

	// Different line behaviors
	isMainLine := lineIdx == maxLines/2

	if isMainLine {
		// Main convergence line with primary animation
		line.WriteString(renderMainConvergenceLine(centerX, phase1, phase2, phase3, phase4, phase5, termTime))
	} else {
		// Supporting lines with particle effects and digital rain
		line.WriteString(renderSupportLine(centerX, lineIdx, maxLines, phase1, phase2, phase3, phase4, phase5, termTime))
	}

	return line.String()
}

// renderMainConvergenceLine creates the primary animation line
func renderMainConvergenceLine(centerX int, phase1, phase2, phase3, phase4, phase5 float64, termTime int64) string {
	var line strings.Builder

	// Enhanced character sets for different phases
	scatterChars := []string{"░", "▒", "▓", "◆", "◇", "○", "●", "◎", "◉", "⬟", "⬢", "⬡", "⬙", "⬚", "⬛", "⬜", "◐", "◑", "◒", "◓"}
	blockChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	infinityChars := []string{"∞", "⧢", "⧣", "⟐", "⟑"}
	spiralChars := []string{"◠", "◡", "◜", "◝", "◞", "◟", "◌", "○", "◎"}

	// Calculate particle count and positions - INCREASED for more dynamic effect
	particleCount := 28 // Increased from 18 to 28
	particles := make([]Particle, particleCount)

	// Initialize particles based on animation phase
	for i := 0; i < particleCount; i++ {
		seed := float64(i*7 + int(termTime/110)) // Evolving seed (10% slower)

		// Phase 1: Scattered genesis
		if phase1 < 1.0 {
			randX := math.Sin(seed) * float64(centerX) * (1.2 - phase1*0.5)
			randY := math.Cos(seed*1.3) * 2 * (1.0 - phase1*0.7)

			particles[i].X = float64(centerX) + randX
			particles[i].Y = randY
			particles[i].Char = scatterChars[i%len(scatterChars)]

			// Flickering effect
			flicker := math.Sin(float64(termTime)/55.0 + seed)
			if flicker > 0.3 {
				particles[i].Color = getPhaseColor(1, seed, termTime)
			} else {
				particles[i].Color = "0" // Dim/off
			}
		}

		// Phase 2: Gravitational pull with physics
		if phase2 > 0 && phase1 >= 1.0 {
			// Apply easing functions - bounce and elastic
			easedPhase2 := elasticEaseOut(phase2)

			// Converge towards center with orbital motion
			targetX := float64(centerX)
			currentX := particles[i].X + (targetX-particles[i].X)*easedPhase2

			// Add orbital motion
			orbit := math.Sin(float64(termTime)/88.0+seed) * (1.0 - easedPhase2) * 3
			currentX += orbit

			particles[i].X = currentX
			particles[i].Y = math.Sin(float64(termTime)/132.0+seed*2) * (1.0 - easedPhase2)
			blockIndex := int(easedPhase2 * float64(len(blockChars)-1))
			if blockIndex >= len(blockChars) {
				blockIndex = len(blockChars) - 1
			}
			particles[i].Char = blockChars[blockIndex]
			particles[i].Color = getPhaseColor(2, seed, termTime)
		}

		// Phase 3: Formation sequence - infinity symbol with swirl
		if phase3 > 0 && phase2 >= 1.0 {
			// Form infinity symbol with additional swirl effect
			angle := float64(i) / float64(particleCount) * 2 * math.Pi
			infScale := 8.0 * phase3

			// Infinity symbol parametric equations
			t := angle + float64(termTime)/220.0*phase3
			infX := infScale * math.Cos(t) / (1 + math.Sin(t)*math.Sin(t))
			infY := infScale * math.Sin(t) * math.Cos(t) / (1 + math.Sin(t)*math.Sin(t)) * 0.3

			// Add swirling motion for extra particles
			if i >= particleCount/2 {
				// Secondary swirl layer
				swirlAngle := angle*2 + float64(termTime)/150.0
				swirlRadius := 6.0 + math.Sin(float64(termTime)/180.0+angle)*2.0
				infX = swirlRadius * math.Cos(swirlAngle)
				infY = swirlRadius * math.Sin(swirlAngle) * 0.4
				particles[i].Char = spiralChars[i%len(spiralChars)]
			} else {
				infIndex := int(phase3 * float64(len(infinityChars)-1))
				if infIndex >= len(infinityChars) {
					infIndex = len(infinityChars) - 1
				}
				particles[i].Char = infinityChars[infIndex]
			}

			particles[i].X = float64(centerX) + infX
			particles[i].Y = infY
			particles[i].Color = getPhaseColor(3, seed, termTime)
		}

		// Phase 4: Digital rain/matrix effect
		if phase4 > 0 && phase3 >= 1.0 {
			// Some particles create vertical streams
			if i%3 == 0 {
				particles[i].Y = math.Mod(float64(termTime)/33.0+seed*50, 6) - 3
				if len(blockChars) > 0 {
					particles[i].Char = blockChars[rand.Intn(len(blockChars))]
				} else {
					particles[i].Char = "█"
				}
				particles[i].Color = getPhaseColor(4, seed, termTime)
			}

			// Main infinity symbol solidifies
			if i < 6 {
				particles[i].Char = "∞"
				particles[i].Color = "15" // Bright white

				// Pulsing glow effect
				pulse := math.Sin(float64(termTime)/55.0)*0.3 + 0.7
				if pulse > 0.8 {
					particles[i].Color = "226" // Bright yellow
				}
			}
		}

		// Phase 5: Crystallization and final burst with sparkles
		if phase5 > 0 && phase4 >= 1.0 {
			// Lock into final perfect formation
			if i < 8 {
				particles[i].X = float64(centerX) + float64((i-4)*2)
				particles[i].Y = 0
				particles[i].Char = "∞"

				// Final burst effect
				burstIntensity := bounceEaseOut(phase5)
				if burstIntensity > 0.7 {
					particles[i].Color = "196" // Bright red
				} else {
					particles[i].Color = "15" // White
				}
			} else {
				// Add sparkle/burst particles radiating outward
				sparkleAngle := float64(i-8) / float64(particleCount-8) * 2 * math.Pi
				sparkleRadius := phase5 * 15.0 * bounceEaseOut(phase5)
				sparkleChars := []string{"✦", "✧", "✨", "⭐", "✪", "✫", "✬", "✭", "✮", "✯", "*", "·"}

				particles[i].X = float64(centerX) + math.Cos(sparkleAngle+float64(termTime)/100.0)*sparkleRadius
				particles[i].Y = math.Sin(sparkleAngle+float64(termTime)/100.0) * sparkleRadius * 0.3
				particles[i].Char = sparkleChars[i%len(sparkleChars)]

				// Multicolor sparkles
				sparkleColors := []string{"226", "220", "214", "208", "202", "196", "15", "51", "45"}
				particles[i].Color = sparkleColors[int(math.Abs(math.Sin(seed+float64(termTime)/80.0))*float64(len(sparkleColors)))%len(sparkleColors)]
			}
		}
	}

	// Render particles to line buffer
	lineBuffer := make(map[int]string)
	colorBuffer := make(map[int]string)

	for _, p := range particles {
		x := int(math.Round(p.X))
		if x >= 0 && x < centerX*2 && math.Abs(p.Y) < 0.5 { // Only render on main line
			lineBuffer[x] = p.Char
			colorBuffer[x] = p.Color
		}
	}

	// Build final line with proper spacing
	minX, maxX := centerX*2, 0
	for x := range lineBuffer {
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}

	// Add leading padding
	startPadding := minX
	if startPadding > 0 {
		line.WriteString(strings.Repeat(" ", startPadding))
	}

	// Render characters with colors
	for x := minX; x <= maxX; x++ {
		if char, exists := lineBuffer[x]; exists {
			color := colorBuffer[x]
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			if color == "15" || color == "226" {
				style = style.Bold(true)
			}
			line.WriteString(style.Render(char))
		} else {
			line.WriteString(" ")
		}
	}

	return line.String()
}

// renderSupportLine creates supporting animation lines (digital rain, particles)
func renderSupportLine(centerX, lineIdx, maxLines int, phase1, phase2, phase3, phase4, phase5 float64, termTime int64) string {
	var line strings.Builder

	// Ripple/wave effect during early phases
	if phase1 > 0 && phase2 < 1.0 {
		waveChars := []string{"~", "≈", "∼", "⌇", "⌁", "∿"}
		rippleCount := 4
		for i := 0; i < rippleCount; i++ {
			seed := float64(i*17 + lineIdx*9 + int(termTime/140))
			wavePhase := seed + float64(termTime)/200.0

			// Create expanding ripple effect from center
			rippleRadius := math.Abs(math.Sin(wavePhase)) * float64(centerX) * phase1
			distFromCenter := math.Abs(float64(lineIdx - maxLines/2))

			if distFromCenter < rippleRadius*0.15 {
				waveX := int(float64(centerX) + math.Cos(wavePhase*2)*rippleRadius*0.8)

				if waveX >= 0 && waveX < centerX*2 {
					currentLen := line.Len()
					needed := waveX - currentLen
					if needed > 0 {
						line.WriteString(strings.Repeat(" ", needed))
					}

					charIdx := int(math.Abs(math.Sin(seed)) * float64(len(waveChars)))
					if charIdx >= len(waveChars) {
						charIdx = len(waveChars) - 1
					}
					char := waveChars[charIdx]

					// Color based on phase
					color := getPhaseColor(1, seed, termTime)
					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					line.WriteString(style.Render(char))
				}
			}
		}
	}

	// Orbital particles during phase 2-3
	if phase2 > 0 && phase3 < 1.0 {
		orbitalChars := []string{"◦", "•", "◘", "○", "◎", "◉"}
		orbitCount := 5

		for i := 0; i < orbitCount; i++ {
			seed := float64(i*11 + lineIdx*7 + int(termTime/170))
			orbitAngle := seed*2 + float64(termTime)/130.0
			orbitRadius := float64(maxLines-lineIdx) * 4.0 * phase2

			orbitX := int(float64(centerX) + math.Cos(orbitAngle)*orbitRadius)

			if orbitX >= 0 && orbitX < centerX*2 {
				currentLen := line.Len()
				needed := orbitX - currentLen
				if needed > 0 {
					line.WriteString(strings.Repeat(" ", needed))
				}

				charIdx := int(math.Abs(math.Sin(seed*1.7)) * float64(len(orbitalChars)))
				if charIdx >= len(orbitalChars) {
					charIdx = len(orbitalChars) - 1
				}
				char := orbitalChars[charIdx]

				color := getPhaseColor(2, seed, termTime)
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				line.WriteString(style.Render(char))
			}
		}
	}

	// Digital rain effect for supporting lines
	if phase4 > 0 {
		rainChars := []string{"╷", "╵", "│", "┃", "┆", "┇", "┊", "┋"}
		glitchChars := []string{"▪", "▫", "▴", "▸", "▾", "◂"}

		// Create rain drops at various positions
		rainCount := 5 + int(phase4*3)
		for i := 0; i < rainCount; i++ {
			seed := float64(i*11 + lineIdx*7 + int(termTime/165))

			// Rain position
			rainX := int(math.Abs(math.Sin(seed)) * float64(centerX*2))

			// Rain character selection
			charIdx := int(math.Abs(math.Sin(seed*1.5)) * float64(len(rainChars)))
			if charIdx >= len(rainChars) {
				charIdx = len(rainChars) - 1
			}

			char := rainChars[charIdx]

			// Occasional glitch characters
			if math.Sin(seed*3.7) > 0.85 {
				glitchIdx := int(math.Abs(math.Sin(seed*4.1)) * float64(len(glitchChars)))
				if glitchIdx >= len(glitchChars) {
					glitchIdx = len(glitchChars) - 1
				}
				char = glitchChars[glitchIdx]
			}

			// Position in line
			if rainX >= 0 && rainX < centerX*2 {
				// Build line with proper spacing
				currentLen := line.Len()
				needed := rainX - currentLen
				if needed > 0 {
					line.WriteString(strings.Repeat(" ", needed))
				}

				// Color based on distance from center and time
				distFromCenter := math.Abs(float64(rainX - centerX))
				colorIntensity := (1.0 - distFromCenter/float64(centerX)) * phase4

				var color string
				if colorIntensity > 0.7 {
					color = "46" // Bright green
				} else if colorIntensity > 0.4 {
					color = "42" // Green
				} else {
					color = "22" // Dark green
				}

				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				line.WriteString(style.Render(char))
			}
		}
	}

	// Subtle particle trails for other phases
	if phase2 > 0 && phase4 == 0 {
		trailChars := []string{"·", "•", "‧", "∘"}

		for i := 0; i < 3; i++ {
			seed := float64(i*13 + lineIdx*5 + int(termTime/220))
			trailX := int(float64(centerX) + math.Sin(seed)*float64(centerX)*0.7)

			if trailX >= 0 && trailX < centerX*2 {
				currentLen := line.Len()
				needed := trailX - currentLen
				if needed > 0 {
					line.WriteString(strings.Repeat(" ", needed))
				}

				charIdx := int(math.Abs(math.Sin(seed*2.3)) * float64(len(trailChars)))
				if charIdx >= len(trailChars) {
					charIdx = len(trailChars) - 1
				}

				char := trailChars[charIdx]
				color := getPhaseColor(2, seed, termTime)

				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				line.WriteString(style.Render(char))
			}
		}
	}

	return line.String()
}

// renderSideDecorations creates decorative particles on the sides of text lines
func renderSideDecorations(progress float64, lineOffset int) (string, string) {
	termTime := time.Now().UnixMilli()

	// Decoration characters that appear on sides
	decorChars := []string{"◦", "•", "◘", "○", "◎", "◉", "✦", "✧", "✨", "⬟", "⬢", "⬡"}

	var leftDecor, rightDecor string

	// Only show decorations after initial phase
	if progress > 0.15 {
		decorProgress := math.Min((progress-0.15)/0.2, 1.0)

		// Calculate which decorations to show based on progress and time
		seed := float64(lineOffset*13 + int(termTime/180))

		// Pulsing effect
		pulse := math.Sin(float64(termTime)/100.0 + seed)

		if pulse > -0.3 && decorProgress > 0.3 {
			// Left decoration
			leftIdx := int(math.Abs(math.Sin(seed)) * float64(len(decorChars)))
			if leftIdx >= len(decorChars) {
				leftIdx = len(decorChars) - 1
			}

			// Right decoration (different from left)
			rightIdx := int(math.Abs(math.Cos(seed*1.3)) * float64(len(decorChars)))
			if rightIdx >= len(decorChars) {
				rightIdx = len(decorChars) - 1
			}

			// Color cycling
			colors := []string{"51", "87", "123", "159", "195", "226", "220", "214"}
			colorIdx := int(math.Abs(math.Sin(float64(termTime)/150.0+seed)) * float64(len(colors)))
			if colorIdx >= len(colors) {
				colorIdx = len(colors) - 1
			}
			color := colors[colorIdx]

			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			leftDecor = style.Render(decorChars[leftIdx]) + " "
			rightDecor = " " + style.Render(decorChars[rightIdx])
		}
	}

	return leftDecor, rightDecor
}

// Easing functions for smooth animations
func elasticEaseOut(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*2*math.Pi/3) + 1
}

func bounceEaseOut(t float64) float64 {
	if t < 1.0/2.75 {
		return 7.5625 * t * t
	} else if t < 2.0/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	} else {
		t -= 2.625 / 2.75
		return 7.5625*t*t + 0.984375
	}
}

// getPhaseColor returns dynamic colors for different animation phases
func getPhaseColor(phase int, seed float64, termTime int64) string {
	timeOffset := float64(termTime) / 110.0

	switch phase {
	case 1: // Scattered genesis - blues, purples, and cyans
		colors := []string{"54", "55", "56", "57", "93", "99", "105", "39", "45", "51", "81", "87"}
		idx := int(math.Abs(math.Sin(seed+timeOffset)) * float64(len(colors)))
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		return colors[idx]

	case 2: // Gravitational pull - cyans, greens, and teals
		colors := []string{"37", "43", "49", "50", "51", "87", "123", "36", "42", "48", "86", "122"}
		idx := int(math.Abs(math.Sin(seed*1.3+timeOffset)) * float64(len(colors)))
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		return colors[idx]

	case 3: // Formation sequence - magentas, reds, and pinks
		colors := []string{"161", "162", "163", "164", "200", "201", "207", "197", "198", "199", "205", "206"}
		idx := int(math.Abs(math.Sin(seed*1.7+timeOffset)) * float64(len(colors)))
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		return colors[idx]

	case 4: // Digital rain - various shades of green
		colors := []string{"22", "28", "34", "40", "46", "82", "118", "76", "70", "64", "106", "112"}
		idx := int(math.Abs(math.Sin(seed*2.1+timeOffset)) * float64(len(colors)))
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		return colors[idx]

	default: // Default white/bright
		return "15"
	}
}
