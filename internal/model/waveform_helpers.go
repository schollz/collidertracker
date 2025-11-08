package model

import (
	"math"
	"sort"

	"github.com/schollz/collidertracker/internal/types"
)

// Waveform manipulation helper functions

// AddWaveformMarker adds a new marker at the midpoint of the current view
func (m *Model) AddWaveformMarker() {
	file := m.WaveformFile
	if file == "" {
		return
	}
	
	// Get file metadata
	metadata, exists := m.FileMetadata[file]
	if !exists {
		metadata = types.FileMetadata{
			BPM:         120.0,
			Slices:      16,
			Playthrough: 0,
			SyncToBPM:   1,
			SliceType:   1, // Onsets mode when manually editing
			Onsets:      []float64{},
		}
	}
	
	// Add marker at midpoint
	midpoint := (m.WaveformStart + m.WaveformEnd) / 2.0
	metadata.Onsets = append(metadata.Onsets, midpoint)
	
	// Sort markers by time
	sort.Float64s(metadata.Onsets)
	
	// Find the newly created marker and select it
	for i, time := range metadata.Onsets {
		if math.Abs(time-midpoint) < 0.0001 {
			m.WaveformSelectedSlice = i
			break
		}
	}
	
	// Update slices count
	metadata.Slices = len(metadata.Onsets)
	
	// Save back to model
	m.FileMetadata[file] = metadata
}

// DeleteSelectedWaveformMarker deletes the currently selected marker
func (m *Model) DeleteSelectedWaveformMarker() {
	file := m.WaveformFile
	if file == "" || m.WaveformSelectedSlice < 0 {
		return
	}
	
	metadata, exists := m.FileMetadata[file]
	if !exists || m.WaveformSelectedSlice >= len(metadata.Onsets) {
		return
	}
	
	// Remove the marker
	metadata.Onsets = append(metadata.Onsets[:m.WaveformSelectedSlice], 
		metadata.Onsets[m.WaveformSelectedSlice+1:]...)
	
	// Update slices count
	metadata.Slices = len(metadata.Onsets)
	
	// Adjust selection
	if len(metadata.Onsets) == 0 {
		m.WaveformSelectedSlice = -1
	} else if m.WaveformSelectedSlice >= len(metadata.Onsets) {
		m.WaveformSelectedSlice = len(metadata.Onsets) - 1
	}
	
	// Save back to model
	m.FileMetadata[file] = metadata
}

// SelectNextWaveformMarker selects the next visible marker (Tab)
func (m *Model) SelectNextWaveformMarker() {
	file := m.WaveformFile
	if file == "" {
		return
	}
	
	metadata, exists := m.FileMetadata[file]
	if !exists || len(metadata.Onsets) == 0 {
		m.WaveformSelectedSlice = -1
		return
	}
	
	// Find markers in current view
	visibleMarkers := []int{}
	for i, time := range metadata.Onsets {
		if time >= m.WaveformStart && time <= m.WaveformEnd {
			visibleMarkers = append(visibleMarkers, i)
		}
	}
	
	if len(visibleMarkers) == 0 {
		m.WaveformSelectedSlice = -1
	} else if m.WaveformSelectedSlice == -1 {
		// Select first visible marker
		m.WaveformSelectedSlice = visibleMarkers[0]
	} else {
		// Find current in visible list and select next
		currentIdx := -1
		for i, idx := range visibleMarkers {
			if idx == m.WaveformSelectedSlice {
				currentIdx = i
				break
			}
		}
		if currentIdx == -1 {
			// Current marker not visible, select first
			m.WaveformSelectedSlice = visibleMarkers[0]
		} else {
			// Cycle to next
			nextIdx := (currentIdx + 1) % len(visibleMarkers)
			m.WaveformSelectedSlice = visibleMarkers[nextIdx]
		}
	}
}

// JogWaveformMarker moves the selected marker left or right
func (m *Model) JogWaveformMarker(direction float64, fast bool) {
	file := m.WaveformFile
	if file == "" || m.WaveformSelectedSlice < 0 {
		return
	}
	
	metadata, exists := m.FileMetadata[file]
	if !exists || m.WaveformSelectedSlice >= len(metadata.Onsets) {
		return
	}
	
	// Calculate step size
	viewDuration := m.WaveformEnd - m.WaveformStart
	stepPercent := 0.005 // 0.5%
	if fast {
		stepPercent = 0.05 // 5%
	}
	step := viewDuration * stepPercent * direction
	
	// Move marker
	metadata.Onsets[m.WaveformSelectedSlice] += step
	
	// Clamp to valid range using cached duration
	if metadata.Onsets[m.WaveformSelectedSlice] < 0 {
		metadata.Onsets[m.WaveformSelectedSlice] = 0
	}
	if metadata.Onsets[m.WaveformSelectedSlice] > m.WaveformDuration {
		metadata.Onsets[m.WaveformSelectedSlice] = m.WaveformDuration
	}
	
	// Re-sort markers
	sort.Float64s(metadata.Onsets)
	
	// Save back to model
	m.FileMetadata[file] = metadata
}

// JogWaveformView moves the view left or right
func (m *Model) JogWaveformView(direction float64, fast bool) {
	duration := m.WaveformEnd - m.WaveformStart
	stepPercent := 0.005 // 0.5%
	if fast {
		stepPercent = 0.05 // 5%
	}
	step := duration * stepPercent * direction
	
	m.WaveformStart += step
	m.WaveformEnd += step
	
	// Clamp to valid range using cached duration
	if m.WaveformStart < 0 {
		m.WaveformStart = 0
		m.WaveformEnd = duration
	}
	if m.WaveformEnd > m.WaveformDuration {
		m.WaveformEnd = m.WaveformDuration
		m.WaveformStart = m.WaveformEnd - duration
		if m.WaveformStart < 0 {
			m.WaveformStart = 0
		}
	}
}

// ZoomWaveformView zooms in or out (zoomIn = true for zoom in, false for zoom out)
func (m *Model) ZoomWaveformView(zoomIn bool) {
	duration := m.WaveformEnd - m.WaveformStart
	center := (m.WaveformStart + m.WaveformEnd) / 2.0
	
	var newDuration float64
	if zoomIn {
		newDuration = duration * 0.8 // Zoom in by 20%
	} else {
		newDuration = duration * 1.25 // Zoom out by 25%
	}
	
	// Don't zoom out beyond total duration (using cached duration)
	if newDuration > m.WaveformDuration {
		newDuration = m.WaveformDuration
	}
	
	m.WaveformStart = center - newDuration/2.0
	m.WaveformEnd = center + newDuration/2.0
	
	// Clamp to valid range
	if m.WaveformStart < 0 {
		m.WaveformStart = 0
		m.WaveformEnd = newDuration
	}
	if m.WaveformEnd > m.WaveformDuration {
		m.WaveformEnd = m.WaveformDuration
		m.WaveformStart = m.WaveformEnd - newDuration
		if m.WaveformStart < 0 {
			m.WaveformStart = 0
		}
	}
}
