# End-to-End Testing Guide - ScribesAI

This document provides a comprehensive checklist for testing ScribesAI before release.

## Pre-Testing Setup

### Environment Preparation

- [ ] Fresh clone of repository
- [ ] All dependencies installed (PortAudio, Vosk, Go 1.24+)
- [ ] Clean data directory (`rm -rf data/ models/`)
- [ ] Valid `.env` file with ANTHROPIC_API_KEY
- [ ] Test on both supported architectures if possible:
  - [ ] macOS Intel (x86_64)
  - [ ] macOS Apple Silicon (ARM64)
  - [ ] Linux x86_64

### Build Verification

```bash
# Clean build
go clean
go build -o scribescli ./cmd/scribescli

# Verify binary
./scribescli --version  # Should show version info
file scribescli         # Verify architecture
```

## Test Scenarios

### 1. First Launch (Cold Start)

**Purpose**: Verify initialization and directory creation

**Steps**:
1. [ ] Delete `data/` and `models/` directories
2. [ ] Launch application: `./scribescli`
3. [ ] Verify directories are created automatically:
   - [ ] `data/` exists
   - [ ] `data/exports/` exists
   - [ ] `models/` exists
4. [ ] Verify database file created: `data/scribescli.db`
5. [ ] Check main menu displays correctly
6. [ ] Verify HAL eye animation works

**Expected Results**:
- No errors on startup
- All directories created automatically
- Database initialized with proper schema
- HAL eye pulsing smoothly
- Menu shows 4 options: New Recording, History, Settings, Quit

---

### 2. Settings View

**Purpose**: Verify settings are loaded from environment

**Steps**:
1. [ ] From main menu, select "Settings" (press Down, Down, Enter)
2. [ ] Verify displayed settings match `.env` file:
   - [ ] API Key shows masked (first 7 + last 4 chars, or "Not configured")
   - [ ] Sample Rate shows correct value (default: 16000 Hz)
   - [ ] Channels shows correct value and type (default: 1 - Mono)
   - [ ] Vosk Model shows path or "auto-download"
   - [ ] Database path shows correct location
3. [ ] Press 'B' to return to menu
4. [ ] Verify no errors occurred

**Expected Results**:
- All settings display correctly
- API key is properly masked
- Back navigation works smoothly

---

### 3. Model Download (First Recording)

**Purpose**: Verify automatic Vosk model download

**Steps**:
1. [ ] Ensure no Vosk model is downloaded (`rm -rf models/vosk-*`)
2. [ ] From main menu, select "New Recording"
3. [ ] Verify download view appears:
   - [ ] Shows "MODEL DOWNLOAD" header
   - [ ] Displays model info (name, language, size)
   - [ ] Shows "Downloading..." or "Preparing download..." status
4. [ ] Wait for download to complete (may take 1-5 minutes)
5. [ ] Verify automatic transition to recording view
6. [ ] Check `models/` directory contains model:
   ```bash
   ls models/vosk-model-small-en-us-0.15/
   ```

**Expected Results**:
- Download UI shows progress clearly
- No errors during download
- Model extracted correctly
- Automatic transition to recording after completion
- Model persists on disk

**Error Case**:
- [ ] Disconnect internet before download
- [ ] Verify error message is helpful and suggests manual install
- [ ] Verify app doesn't crash

---

### 4. Audio Recording (Happy Path)

**Purpose**: Verify basic recording functionality

**Steps**:
1. [ ] Select "New Recording" from menu (model should already be downloaded)
2. [ ] Verify recording view appears:
   - [ ] HAL eye turns red
   - [ ] "RECORDING" header visible
   - [ ] Timer starts at 00:00
   - [ ] Audio level meter visible
3. [ ] Speak clearly for 10-15 seconds
4. [ ] Observe real-time transcription appearing (may have slight delay)
5. [ ] Verify audio level meter responds to voice
6. [ ] Press 'S' to stop recording
7. [ ] Verify transitions to "PROCESSING" view
8. [ ] Wait for Claude AI analysis to complete
9. [ ] Verify analysis view shows:
   - [ ] Summary section
   - [ ] Key Points (bullet list)
   - [ ] Action Items (if any detected)
10. [ ] Verify WAV file saved: `ls data/recording_*.wav`
11. [ ] Verify database entry: `sqlite3 data/scribescli.db "SELECT * FROM recordings;"`

**Expected Results**:
- Recording starts immediately without errors
- Timer increments every second
- Audio levels show visual feedback
- Transcription appears during recording (partial results)
- Stop creates processing view
- Analysis completes within 5-30 seconds
- All data saved correctly

**Performance**:
- [ ] CPU usage reasonable (<30%)
- [ ] Memory stable (no leaks)
- [ ] No audio glitches or dropouts

---

### 5. Recording Without Microphone

**Purpose**: Verify error handling for missing audio device

**Steps**:
1. [ ] Disable or disconnect microphone (if possible)
2. [ ] Attempt to start recording
3. [ ] Verify error message displays:
   - [ ] Mentions PortAudio or microphone issue
   - [ ] Provides troubleshooting steps
   - [ ] Suggests checking permissions
4. [ ] Verify app doesn't crash
5. [ ] Verify can return to menu

**Expected Results**:
- Clear, helpful error message
- App remains stable
- User can continue using other features

---

### 6. Recording Without API Key

**Purpose**: Verify graceful degradation without Claude API

**Steps**:
1. [ ] Remove or comment out `ANTHROPIC_API_KEY` from `.env`
2. [ ] Restart application
3. [ ] Create a recording (speak for 5-10 seconds)
4. [ ] Stop recording
5. [ ] Verify warning message appears:
   - [ ] Mentions Claude API not available
   - [ ] Explains how to configure API key
   - [ ] Provides link to console.anthropic.com
6. [ ] Verify recording is still saved to database
7. [ ] Verify can view recording in History

**Expected Results**:
- Recording works without AI analysis
- Warning is clear and helpful
- Recording saved with transcript
- No crash or data loss

---

### 7. Transcription Quality

**Purpose**: Verify Vosk transcription accuracy

**Test Phrases** (speak clearly):
- [ ] "Hello, this is a test recording."
- [ ] "The quick brown fox jumps over the lazy dog."
- [ ] "Schedule meeting with John for next Tuesday at 3 PM."
- [ ] Technical terms: "Kubernetes deployment with Docker containers"

**Steps**:
1. [ ] Record each phrase
2. [ ] Check real-time partial transcripts
3. [ ] Check final transcript in analysis view
4. [ ] Compare with expected text

**Expected Results**:
- General speech: 80-90% accuracy
- Clear, standard phrases: 90-95% accuracy
- Technical terms: 70-80% accuracy (acceptable)
- Numbers and times: Reasonable accuracy

**Note**: Small errors are acceptable for small model. Document any systematic failures.

---

### 8. Export Functionality

**Purpose**: Verify all export formats work correctly

**Steps**:
1. [ ] Complete a recording with analysis
2. [ ] From analysis view, press 'M' for Markdown export
3. [ ] Verify success banner appears:
   - [ ] Green background
   - [ ] Shows full path to exported file
   - [ ] Checkmark (✓) visible
4. [ ] Verify file exists: `ls data/exports/meeting_*.md`
5. [ ] Open and verify Markdown format:
   ```bash
   cat data/exports/meeting_*.md
   ```
6. [ ] Repeat for JSON export (press 'J'):
   - [ ] Success banner appears
   - [ ] JSON file created
   - [ ] Valid JSON syntax: `jq . data/exports/meeting_*.json`
7. [ ] Repeat for Text export (press 'T'):
   - [ ] Success banner appears
   - [ ] Text file created
   - [ ] Readable plain text format
8. [ ] Press 'B' to go back
9. [ ] Verify success banner disappears

**Expected Results**:
- All three formats export successfully
- Success feedback is clear and visible
- Files are well-formatted and complete
- Success banner clears on navigation

**Validation**:
```bash
# Markdown should have headers, bullet points
grep "^#" data/exports/meeting_*.md

# JSON should be valid
jq . data/exports/meeting_*.json

# Text should be readable
cat data/exports/meeting_*.txt
```

---

### 9. History View

**Purpose**: Verify recordings history and playback

**Steps**:
1. [ ] Create 3-4 test recordings
2. [ ] Return to main menu
3. [ ] Select "History"
4. [ ] Verify list displays:
   - [ ] Most recent first
   - [ ] Shows date/time
   - [ ] Shows duration
   - [ ] Shows status (completed/processing)
5. [ ] Select a recording (press Up/Down, then Enter)
6. [ ] Verify shows full analysis view
7. [ ] Verify can export from history
8. [ ] Press 'B' to return to history list
9. [ ] Press 'B' again to return to menu

**Expected Results**:
- History list is accurate and complete
- Sorting is correct (newest first)
- Can navigate and view past recordings
- All features work from history view

---

### 10. Long Recording

**Purpose**: Verify stability with extended recording

**Steps**:
1. [ ] Start a new recording
2. [ ] Record for 2-5 minutes (can be ambient noise or speech)
3. [ ] Monitor:
   - [ ] Timer accuracy
   - [ ] Audio level consistency
   - [ ] Transcription updates
   - [ ] Memory usage (should stay stable)
4. [ ] Stop and verify processing completes
5. [ ] Check WAV file size is reasonable

**Expected Results**:
- No memory leaks
- Transcription handles long audio
- Processing completes successfully
- File size matches duration (~1.5 MB per minute for 16kHz mono)

---

### 11. Rapid Start/Stop

**Purpose**: Verify no race conditions or state issues

**Steps**:
1. [ ] Start recording
2. [ ] Immediately stop (within 1 second)
3. [ ] Verify processes correctly
4. [ ] Repeat 3-4 times
5. [ ] Verify all recordings saved

**Expected Results**:
- No crashes
- Very short recordings handled gracefully
- Empty or minimal transcripts acceptable
- No state corruption

---

### 12. Navigation Stress Test

**Purpose**: Verify UI state management

**Steps**:
1. [ ] Rapidly navigate through all views:
   - Menu → Settings → Back
   - Menu → History → Back
   - Menu → New Recording → Stop → Analysis → Back
2. [ ] Try unexpected key presses in each view
3. [ ] Press 'Q' from different views
4. [ ] Verify Quit always works cleanly

**Expected Results**:
- No crashes from rapid navigation
- State remains consistent
- Quit is always available
- No visual glitches

---

### 13. Database Integrity

**Purpose**: Verify database operations are safe

**Steps**:
1. [ ] Create several recordings
2. [ ] Verify database with SQLite:
   ```bash
   sqlite3 data/scribescli.db

   -- Check recordings
   SELECT id, created_at, duration, status FROM recordings;

   -- Check action items
   SELECT * FROM action_items;

   -- Check foreign keys
   PRAGMA foreign_keys;
   PRAGMA integrity_check;
   ```
3. [ ] Verify JSON fields parse correctly:
   ```sql
   SELECT id, json_extract(key_points, '$[0]') FROM recordings;
   ```

**Expected Results**:
- All data stored correctly
- Foreign keys enforced
- Integrity check passes
- JSON arrays readable

---

### 14. Error Recovery

**Purpose**: Verify graceful handling of errors

**Test Cases**:

**A. Corrupt Database**:
1. [ ] Create recording
2. [ ] Stop application
3. [ ] Corrupt database: `echo "corrupt" >> data/scribescli.db`
4. [ ] Restart application
5. [ ] Verify error message is helpful

**B. Full Disk**:
1. [ ] Simulate full disk (if possible)
2. [ ] Attempt recording
3. [ ] Verify error message about disk space

**C. Missing Model**:
1. [ ] Delete Vosk model while app is running
2. [ ] Try to start new recording
3. [ ] Verify triggers re-download or shows clear error

**Expected Results**:
- App doesn't crash on corrupted data
- Error messages suggest recovery steps
- User can continue or exit cleanly

---

### 15. Performance Benchmarks

**Purpose**: Verify acceptable performance

**Measurements**:

**Startup Time**:
```bash
time ./scribescli --version
# Should be < 1 second
```

**Cold Start**:
```bash
rm data/scribescli.db
time ./scribescli
# Should initialize in < 2 seconds
```

**Recording Overhead**:
- [ ] CPU usage during recording: < 15%
- [ ] Memory usage: < 100 MB
- [ ] Audio latency: < 100ms

**Transcription**:
- [ ] Partial results appear within 500ms of speech
- [ ] Final results within 1 second after stop

**Claude Analysis**:
- [ ] 1 minute recording analyzed in < 10 seconds
- [ ] 5 minute recording analyzed in < 30 seconds

---

## Platform-Specific Tests

### macOS Specific

- [ ] Verify microphone permission request appears on first recording
- [ ] Test with both built-in and external microphones
- [ ] Verify works with System Integrity Protection enabled
- [ ] Test on both Intel and Apple Silicon if available
- [ ] Verify library paths work: `otool -L scribescli`

### Linux Specific

- [ ] Verify works with PulseAudio
- [ ] Verify works with ALSA
- [ ] Test with different distributions (Ubuntu, Fedora, Arch)
- [ ] Verify systemd integration (if applicable)
- [ ] Check library dependencies: `ldd scribescli`

---

## Regression Testing

After each code change, verify:

- [ ] Build succeeds on all platforms
- [ ] Basic record/stop/analyze cycle works
- [ ] Export to all formats works
- [ ] Settings view loads correctly
- [ ] History view shows recordings
- [ ] No new warnings or errors in logs

---

## Known Issues / Limitations

Document any discovered issues:

1. **Vosk Small Model Accuracy**: 85-90% for clear speech, may struggle with accents
2. **Claude API Rate Limits**: Large number of rapid analyses may hit limits
3. **PortAudio Permissions**: macOS requires explicit microphone permission
4. **Technical Terms**: Vosk may misrecognize specialized terminology

---

## Test Completion Checklist

Before declaring testing complete:

- [ ] All core scenarios pass
- [ ] Tested on target platforms (macOS, Linux)
- [ ] Performance is acceptable
- [ ] Error messages are helpful
- [ ] Documentation is accurate
- [ ] No data corruption observed
- [ ] No memory leaks detected
- [ ] Build artifacts verified
- [ ] Known issues documented

---

## Sign-Off

**Tested By**: _________________
**Date**: _________________
**Platform**: _________________
**Go Version**: _________________
**Vosk Model**: vosk-model-small-en-us-0.15
**Result**: ☐ PASS  ☐ FAIL (document issues below)

**Notes**:
```
[Space for tester notes and observations]
```

---

**Last Updated**: 2025-11-14
**ScribesAI Version**: Phase 5 - Production Polish
