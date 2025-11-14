# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ScribesAI** is a Terminal User Interface (TUI) application for recording meetings, automatic transcription, and AI-powered analysis with a HAL 9000-inspired interface.

### Supported Platforms

**Primary Targets:**
- 🐧 **Linux** (x86_64): Ubuntu, Debian, Fedora, Arch
- 🍎 **macOS** (Intel & Apple Silicon): macOS 11+

**Not Supported:**
- ❌ Windows (CGO + PortAudio complexity)

See [PLATFORMS.md](PLATFORMS.md) for detailed platform information.

### Technology Stack

- **Language**: Go 1.24+
- **Platforms**: Linux (x86_64) + macOS (x86_64, ARM64)
- **TUI Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lipgloss (github.com/charmbracelet/lipgloss)
- **Audio**: PortAudio (github.com/gordonklaus/portaudio) - Requires CGO
- **Transcription**: Vosk (github.com/alphacep/vosk-api/go) - Offline multi-language
- **AI**: Claude API (github.com/anthropics/anthropic-sdk-go)
- **Database**: SQLite (modernc.org/sqlite) - Pure Go implementation
- **Config**: godotenv (github.com/joho/godotenv)

## Architecture

### Directory Structure

```
scribescli/
├── cmd/scribescli/          # Application entry point
├── internal/                # Private application code
│   ├── audio/              # Audio recording and WAV file handling
│   ├── transcription/      # Vosk integration (TODO)
│   ├── ai/                 # Claude API client and prompts
│   ├── tui/                # Bubble Tea TUI implementation
│   ├── storage/            # SQLite database operations
│   └── export/             # Export to Markdown/JSON/Text
├── pkg/models/             # Shared data models
├── data/                   # Runtime data (recordings, database)
└── models/                 # Transcription models (Vosk)
```

### Key Components

#### 1. Audio Module (`internal/audio/`)

- **recorder.go**: Manages audio recording with PortAudio
  - Real-time audio capture
  - State management (recording, paused, stopped)
  - Audio level calculation (RMS)
  - Thread-safe buffer management

- **wav.go**: WAV file handling
  - WAV header creation
  - Save recordings to .wav files
  - Load WAV files for playback

#### 2. AI Module (`internal/ai/`)

- **claude.go**: Claude API integration
  - `AnalyzeMeeting()`: Full analysis (summary, key points, action items)
  - `GenerateSummary()`: Quick summary generation
  - Structured prompt engineering
  - Response parsing into structured data

**Prompt Structure**:
```
1. RESUMEN: 100-200 words summary
2. PUNTOS CLAVE: 3-7 bullet points
3. ACCIONABLES: Tasks with priority and assignee
```

#### 3. TUI Module (`internal/tui/`)

- **model.go**: Bubble Tea model
  - State machine: Menu → Recording → Processing → Analysis
  - Key bindings and navigation
  - Component lifecycle management

- **views.go**: View rendering
  - `viewMenu()`: Main menu with HAL eye
  - `viewRecording()`: Recording interface with live transcript
  - `viewAnalysis()`: Analysis results display
  - `viewHistory()`: Recordings history
  - `viewSettings()`: Configuration

- **styles.go**: HAL 9000 color scheme
  - Red (#FF0000): HAL's iconic color
  - CRT Green (#00FF41): Terminal text
  - Amber (#FFB000): Warnings/active state
  - Cyan (#00FFFF): Highlights

- **haleye.go**: Animated HAL 9000 eye
  - Frame-based animation (8 frames)
  - State-dependent colors and symbols
  - Pulsing effect
  - Contextual quotes

#### 4. Storage Module (`internal/storage/`)

- **database.go**: SQLite operations
  - Tables: `recordings`, `action_items`
  - CRUD operations
  - JSON serialization for arrays (key_points, tags)
  - Foreign key constraints

#### 5. Export Module (`internal/export/`)

- **exporter.go**: Multiple export formats
  - Markdown: For documentation/Notion/Obsidian
  - JSON: For integration/APIs
  - Text: Plain text with formatting

## Development Guidelines

### Code Style

1. **Package organization**: Follow Go standard project layout
2. **Error handling**: Always wrap errors with context
3. **Concurrency**: Use mutexes for shared state (audio buffer, recorder state)
4. **Naming**:
   - Exported functions: PascalCase
   - Private functions: camelCase
   - Constants: PascalCase or UPPER_SNAKE_CASE

### Common Patterns

#### Adding a New View

1. Add view state to `ViewState` enum in `model.go`
2. Implement `view<Name>()` function in `views.go`
3. Add navigation logic in `Update()` method
4. Add key bindings if needed

#### Adding a New Export Format

1. Implement `ExportTo<Format>()` in `exporter.go`
2. Add format constant to `models/recording.go`
3. Update export menu in TUI

#### Database Schema Changes

1. Update `initSchema()` in `storage/database.go`
2. Add migration logic if needed
3. Update model structs in `pkg/models/`
4. Update CRUD operations

### Testing Strategy

- **Unit tests**: For pure functions (parsing, formatting)
- **Integration tests**: For database operations
- **Manual testing**: For TUI and audio (requires human interaction)

### Environment Variables

Required:
- `ANTHROPIC_API_KEY`: Claude API key

Optional:
- `VOSK_MODEL_PATH`: Path to Vosk model
- `SAMPLE_RATE`: Audio sample rate (default: 16000)
- `CHANNELS`: Audio channels (default: 1)
- `DB_PATH`: Database path (default: ./data/scribescli.db)

## Common Tasks

### Building the Application

```bash
go build -o scribescli ./cmd/scribescli
```

### Running Tests

```bash
go test ./...
```

### Adding Dependencies

```bash
go get <package>
go mod tidy
```

### Debugging TUI Issues

1. Use `tea.LogToFile()` for debug output
2. Add debug view with state inspection
3. Test with `--debug` flag to disable alt screen

## TODOs and Future Improvements

### High Priority

- [ ] Implement Vosk transcription integration
- [ ] Add real-time transcription display during recording
- [ ] Implement export menu in analysis view
- [ ] Add error recovery and retry logic

### Medium Priority

- [ ] Speaker diarization
- [ ] Automatic language detection
- [ ] Streaming mode for long meetings
- [ ] PDF export
- [ ] Calendar integration

### Low Priority

- [ ] Voice synthesis (HAL speaks)
- [ ] Custom theme support
- [ ] Plugin system
- [ ] Web UI companion

## Known Issues

1. **PortAudio dependency**: Requires system-level installation
   - Solution: Document installation instructions per OS

2. **Vosk models**: Large download size
   - Solution: Implement auto-download with progress bar

3. **Build without PortAudio**: Fails if portaudio19-dev not installed
   - Solution: Add build tags for audio-less mode

## Debugging Tips

### Audio Not Recording

1. Check PortAudio installation: `pkg-config --libs portaudio-2.0`
2. List audio devices: `arecord -l` (Linux) or `system_profiler SPAudioDataType` (macOS)
3. Verify permissions: Microphone access required

### Claude API Errors

1. Verify API key in `.env`
2. Check API quota/limits
3. Review request/response in logs

### TUI Rendering Issues

1. Check terminal size: `echo $COLUMNS x $LINES`
2. Test in different terminals (some don't support all Unicode)
3. Disable animations if flickering occurs

## Performance Considerations

1. **Audio buffer**: Circular buffer to prevent memory growth
2. **Database**: Use indexes for common queries
3. **TUI updates**: Throttle animation frames (100ms)
4. **AI calls**: Cache results to avoid re-processing

## Security Considerations

1. **API Key**: Never commit `.env` to git
2. **Audio files**: Store in user-controlled directory
3. **Database**: Use parameterized queries (SQL injection prevention)
4. **Exports**: Sanitize filenames to prevent path traversal

## Contributing Guidelines

When modifying this project:

1. **Maintain HAL 9000 aesthetic**: Keep the retro-futuristic theme
2. **Test audio thoroughly**: Audio bugs are hard to debug
3. **Document prompts**: AI prompts should be well-commented
4. **Follow Go conventions**: Run `go fmt` and `golint`
5. **Update CLAUDE.md**: Document significant changes here

## Resources

- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Vosk Models](https://alphacephei.com/vosk/models)
- [Claude API Docs](https://docs.anthropic.com/claude/reference)
- [HAL 9000 Reference](https://en.wikipedia.org/wiki/HAL_9000)

---

*"I'm putting myself to the fullest possible use, which is all I think that any conscious entity can ever hope to do."* - HAL 9000
- siempre que sea posible usar TDD
- siempre despues de un cambio commitea y pushea, no pushes directo  siempre sale con un branch desde develop