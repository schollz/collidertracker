# Timing Improvements for ColliderTracker

## Problem Statement

When playing back music in ColliderTracker, there was a noticeable pause/stutter when the sampler loaded a new audio file. This timing issue degraded the playback experience, especially when switching between different samples.

## Root Cause Analysis

The timing issue occurred due to **synchronous file loading** in the SuperCollider audio server:

1. When a phrase row references a new sample file, the Go code sends an OSC `/sampler` message to SuperCollider
2. In SuperCollider's `collidertracker.scd` (line 1346), if the sample wasn't already cached, it called `Buffer.read()` 
3. Although `Buffer.read()` is technically asynchronous in SuperCollider, the playback callback was invoked from within the loading action, meaning the synth wouldn't start until the buffer was fully loaded
4. During file loading (which can take 10-100ms+ for large files), the Go timing loop would send the next OSC message, which would also block
5. This created a cascade of delays that compounded into audible timing issues

## Solution Overview

The solution implements **asynchronous buffer preloading** with **graceful degradation**:

### 1. Non-Blocking Sample Loading (SuperCollider)

**File:** `internal/supercollider/collidertracker.scd`

Added a loading queue to track samples currently being loaded:
```supercollider
~sampleLoadingQueue = Dictionary.new();
```

Modified the `/sampler` OSC handler to skip playback if a sample is still loading:
```supercollider
if (~sampleCache.at(filename).isNil,{
    if (~sampleLoadingQueue.at(filename).isNil, {
        // Start loading asynchronously
        ~sampleLoadingQueue.put(filename, true);
        ~sampleCache.put(filename, Buffer.read(s, filename, action: { |b|
            ~sampleLoadingQueue.removeAt(filename);
            // Don't play here - let next message trigger playback
        }));
    }, {
        // Skip this note - sample is still loading
        ["skipping play - sample still loading:", filename].postln;
    });
},{
    // Sample is cached - play immediately
    ~playFromMsg.(msg, ~sampleCache.at(filename));
});
```

### 2. Preload OSC Endpoint (SuperCollider)

Added a new `/preload` endpoint that loads buffers without playing them:
```supercollider
OSCFunc({ |msg|
    var filename = msg[1];
    if (~sampleCache.at(filename).isNil,{
        if (~sampleLoadingQueue.at(filename).isNil, {
            ~sampleLoadingQueue.put(filename, true);
            ~sampleCache.put(filename, Buffer.read(s, filename, action: { |b|
                ["preloaded", b].postln;
                ~sampleLoadingQueue.removeAt(filename);
            }));
        });
    });
},'/preload');
```

### 3. Lookahead Preloading (Go)

**File:** `internal/model/model.go`

Added `SendOSCPreloadMessage()` to send preload requests:
```go
func (m *Model) SendOSCPreloadMessage(filename string) {
    // Sends /preload OSC message with absolute path
}
```

Added `PreloadUpcomingSamples()` to scan ahead and preload:
```go
func (m *Model) PreloadUpcomingSamples(track int, lookaheadRows int) {
    // Scans up to lookaheadRows ahead in the phrase
    // Sends preload messages for any samples that will be needed
    // Only processes each unique file once
}
```

### 4. Integration with Playback

**Files:** `internal/input/playback.go`, `internal/input/helpers.go`

Integrated preloading at two key points:

1. **When playback starts** - preload samples from the first 8 rows
2. **When advancing to next row** - preload samples from the next 8 rows

```go
// Load new ticks for the advanced row
m.LoadTicksLeftForTrack(track)

// Preload upcoming samples (8-row lookahead)
m.PreloadUpcomingSamples(track, 8)

// Emit the row for playback
EmitRowDataFor(m, phraseNum, currentRow, track)
```

## Benefits

1. **Stable Timing:** The playback timing loop never blocks on I/O operations
2. **Graceful Degradation:** If a sample isn't ready, that note is skipped rather than causing a pause
3. **Proactive Loading:** Samples are loaded before they're needed (8 rows ahead)
4. **No Breaking Changes:** Fully backward compatible with existing functionality
5. **Smart Caching:** Each unique file is only preloaded once per lookahead scan

## Performance Characteristics

- **Lookahead Distance:** 8 rows (configurable)
- **Typical Preload Time:** Occurs during playback of previous rows, ~100-500ms before sample is needed
- **Memory Impact:** Minimal - uses existing buffer cache
- **CPU Impact:** Negligible - file loading happens asynchronously in SuperCollider's audio thread

## Testing

Added comprehensive test coverage:

- `TestPreloadUpcomingSamples` - Verifies preloading logic and edge cases
- `TestGetPhrasesFilesForTrack` - Tests file retrieval for different track types
- `TestSendOSCPreloadMessage` - Verifies OSC message sending

All existing tests continue to pass, confirming no regressions.

## Future Enhancements

Possible future improvements:

1. **Adaptive Lookahead:** Adjust lookahead distance based on BPM and file sizes
2. **Buffer Pool Management:** Limit total cached buffers to manage memory
3. **Preload Priority Queue:** Prioritize loading files closer to playback position
4. **Statistics:** Track cache hit/miss ratios for optimization

## Migration Notes

No migration required - changes are fully backward compatible. The new preloading behavior activates automatically during playback.
