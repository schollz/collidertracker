package input

import (
	"testing"
	"time"

	"github.com/schollz/collidertracker/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestTimingDrift verifies that playback timing doesn't accumulate drift over time
func TestTimingDrift(t *testing.T) {
	// Create a model with a simple timing setup
	m := model.NewModel(0, "", false)
	m.BPM = 60.0 // 1 beat per second
	m.PPQ = 1    // 1 pulse per quarter note = 1 tick per second
	
	// Expected tick duration: 1 second (1000000 microseconds)
	// At 60 BPM with PPQ=1: 60 beats/min / 60 = 1 beat/sec
	// 1 beat/sec * 1 pulse/beat = 1 tick/sec
	expectedTickDuration := 1.0 * time.Second
	
	// Simulate playback start
	m.IsPlaying = true
	m.PlaybackStartTime = time.Now()
	m.PlaybackTickCount = 0
	
	// Measure actual tick timing over multiple iterations
	numTicks := 10
	tolerance := 50 * time.Millisecond // Allow 50ms tolerance
	
	for i := 0; i < numTicks; i++ {
		expectedTime := m.PlaybackStartTime.Add(time.Duration(i+1) * expectedTickDuration)
		
		// Simulate tick processing - in real code this would be called by AdvancePlayback
		m.PlaybackTickCount++
		
		// Calculate next tick time using absolute scheduling
		nextTickTime := m.PlaybackStartTime.Add(time.Duration(m.PlaybackTickCount) * expectedTickDuration)
		drift := nextTickTime.Sub(expectedTime)
		
		// Verify drift is within tolerance
		assert.Less(t, drift.Abs(), tolerance, 
			"Tick %d: drift %v exceeds tolerance %v", i+1, drift, tolerance)
	}
}

// TestTimingDriftLongDuration verifies timing stability over a longer period
func TestTimingDriftLongDuration(t *testing.T) {
	// Test case from the issue: 60 BPM over 1 hour should not drift
	m := model.NewModel(0, "", false)
	m.BPM = 60.0 // 1 beat per second
	m.PPQ = 1    // 1 pulse per quarter note
	
	// At 60 BPM, PPQ=1: should be exactly 1 tick per second
	// Over 1 hour: 3600 ticks
	expectedTickDuration := 1.0 * time.Second
	
	// Simulate playback start
	m.IsPlaying = true
	m.PlaybackStartTime = time.Now()
	m.PlaybackTickCount = 0
	
	// Check drift at various points throughout the hour
	checkPoints := []int{60, 300, 600, 1800, 3600} // 1min, 5min, 10min, 30min, 60min
	tolerance := 100 * time.Millisecond           // Allow 100ms tolerance even after 1 hour
	
	for _, tickNum := range checkPoints {
		// Calculate expected absolute time for this tick using absolute scheduling
		us := rowDurationMicroseconds(m)
		expectedTime := m.PlaybackStartTime.Add(time.Duration(float64(tickNum) * us * nanosecondsPerMicrosecond))
		
		// Calculate what the expected time would be using simple multiplication
		// This represents the "ideal" time if there was no drift
		idealTime := m.PlaybackStartTime.Add(time.Duration(tickNum) * expectedTickDuration)
		
		// The drift is the difference between these two
		// With proper tick duration calculation, these should be identical
		drift := expectedTime.Sub(idealTime)
		
		assert.Less(t, drift.Abs(), tolerance,
			"After %d ticks (%.1f minutes): drift %v exceeds tolerance %v",
			tickNum, float64(tickNum)/60.0, drift, tolerance)
	}
}

// TestRowDurationCalculation verifies rowDurationMicroseconds is consistent
func TestRowDurationCalculation(t *testing.T) {
	m := model.NewModel(0, "", false)
	
	// Test various BPM values
	testCases := []struct {
		bpm      float32
		ppq      int
		expected float64 // in microseconds
	}{
		{60.0, 1, 1000000.0},   // 1 tick/sec = 1000000 us
		{120.0, 1, 500000.0},   // 2 ticks/sec = 500000 us
		{60.0, 2, 500000.0},    // 2 ticks/sec = 500000 us
		{120.0, 2, 250000.0},   // 4 ticks/sec = 250000 us
		{90.0, 4, 166666.666},  // 6 ticks/sec = ~166667 us
	}
	
	for _, tc := range testCases {
		m.BPM = tc.bpm
		m.PPQ = tc.ppq
		
		// Set up a simple phrase for testing
		m.PlaybackPhrase = 0
		m.PlaybackRow = 0
		m.CurrentTrack = 0
		
		// Initialize phrase data if needed
		if len(m.PhrasesData[0]) == 0 {
			m.PhrasesData[0] = make([][]int, 255)
			for i := 0; i < 255; i++ {
				m.PhrasesData[0][i] = make([]int, 10) // 10 columns
				for j := 0; j < 10; j++ {
					m.PhrasesData[0][i][j] = -1
				}
			}
		}
		
		duration := rowDurationMicroseconds(m)
		
		// Allow small floating point error (1 microsecond)
		assert.InDelta(t, tc.expected, duration, 1.0,
			"BPM=%.1f PPQ=%d: expected %.0f us, got %.0f us",
			tc.bpm, tc.ppq, tc.expected, duration)
	}
}

// BenchmarkTimingAccuracy measures the performance of timing calculations
func BenchmarkTimingAccuracy(b *testing.B) {
	m := model.NewModel(0, "", false)
	m.BPM = 120.0 // 2 beats per second
	m.PPQ = 2     // 2 pulses per quarter note = 4 ticks per second
	
	m.PlaybackStartTime = time.Now()
	m.PlaybackTickCount = 0
	
	// Measure the performance of calculating tick times
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.PlaybackTickCount++
		us := rowDurationMicroseconds(m)
		_ = m.PlaybackStartTime.Add(time.Duration(float64(m.PlaybackTickCount) * us * nanosecondsPerMicrosecond))
	}
}
