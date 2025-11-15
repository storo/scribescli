# Changelog

All notable changes to ScribesAI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2025-11-14

### Added

**Phase 1: Foundation & TUI Framework**
- Terminal User Interface (TUI) with Bubble Tea framework
- HAL 9000-inspired visual theme (red eye, CRT green text, retro-futuristic design)
- Animated HAL eye with state-dependent colors and pulsing effect
- Main menu navigation with keyboard controls
- View system (Menu, Recording, Processing, Analysis, History, Settings)
- Key binding system for intuitive navigation

**Phase 2: Database Integration**
- SQLite database with pure Go implementation (modernc.org/sqlite)
- Recordings table with metadata (title, duration, transcript, created_at, etc.)
- Action items table with foreign key relationships
- JSON serialization for arrays (key_points, tags)
- CRUD operations for recordings and action items
- Database migrations and schema management

**Phase 3: Speech Transcription**
- Vosk offline speech recognition integration
- Real-time transcription during recording
- Multi-language support (English, Spanish, French, German)
- Model management system (download, verify, initialize)
- Auto-download UI for missing models with progress feedback
- Partial and final transcription results
- 16kHz audio sampling optimized for Vosk

**Phase 4: AI-Powered Analysis**
- Claude AI integration via Anthropic SDK
- Meeting analysis with structured output:
  - Automatic summary generation (100-200 words)
  - Key points extraction (3-7 bullet points)
  - Action items identification with priorities and assignees
- Retry logic with exponential backoff (3 attempts)
- Graceful degradation when AI unavailable
- Context-aware prompts in Spanish
- 60-second timeout for API calls

**Phase 5: Production Polish & Demo Readiness**
- Automatic directory creation on startup (data/, models/, exports/)
- Functional settings view with real environment configuration
- Model download UI with progress bar and status messages
- Enhanced error messages with troubleshooting steps:
  - Transcription initialization errors → suggest model re-download
  - Recording start errors → check microphone, permissions, PortAudio
  - File save errors → verify disk space and permissions
  - Database errors → check directory permissions
  - API key errors → provide setup instructions
  - Empty transcript errors → explain possible causes
- Export success feedback banners:
  - Green success message after export
  - Shows full file path
  - Checkmark indicator
  - Auto-clears on navigation
- Comprehensive documentation:
  - E2E_TESTING.md (15+ test scenarios)
  - RELEASE_CHECKLIST.md (complete release workflow)
  - Updated CLAUDE.md with implementation status

**Audio Recording**
- PortAudio integration for cross-platform audio capture
- Real-time audio level monitoring with visual feedback
- WAV file generation with proper headers
- Configurable sample rate and channels
- Recording state management (recording, paused, stopped)
- Thread-safe buffer management

**Export Functionality**
- Multiple export formats:
  - Markdown (for documentation/Notion/Obsidian)
  - JSON (for integration/APIs)
  - Plain text (formatted)
- Automatic file naming with timestamps
- Export directory management (data/exports/)

### Fixed

- **Anthropic SDK Compatibility** (Critical):
  - Updated API calls to match anthropic-sdk-go v1.17.0
  - Removed non-existent `anthropic.F()` wrapper functions
  - Replaced `AsUnion()` with `AsText()` for content parsing
  - Fixed client type mismatch (pointer vs value)
  - Resolved all 9 compilation errors in Claude AI integration

- Directory creation errors on first launch
- Settings view showing hardcoded values instead of environment config
- Missing feedback after export operations
- Unclear error messages without recovery suggestions
- Model download failures without helpful guidance

### Changed

- Improved model download workflow with dedicated UI
- Enhanced user feedback throughout application
- Settings now load from .env file (API key, sample rate, model path, etc.)
- Error messages now include actionable troubleshooting steps
- Success banners for export confirmation

### Technical Details

**Dependencies**:
- Go 1.24+
- Bubble Tea (TUI framework)
- Lipgloss (styling)
- PortAudio (audio capture)
- Vosk API (speech recognition)
- Anthropic SDK Go v1.17.0 (Claude AI)
- SQLite (modernc.org/sqlite)
- godotenv (configuration)

**Platforms Supported**:
- macOS (Intel x86_64)
- macOS (Apple Silicon ARM64)
- Linux (x86_64)

**Architecture**:
```
cmd/scribescli/          # Entry point
internal/
  ├── audio/            # PortAudio recording & WAV handling
  ├── transcription/    # Vosk integration & model management
  ├── ai/               # Claude API client & prompts
  ├── tui/              # Bubble Tea UI (views, model, HAL eye)
  ├── storage/          # SQLite database operations
  └── export/           # Export to Markdown/JSON/Text
pkg/models/            # Shared data models
data/                  # Runtime data (recordings, database, exports)
models/                # Vosk speech recognition models
```

### Known Issues

- **PortAudio Dependency**: Requires system-level installation before building
  - Solution: Run `./scripts/install-vosk.sh` or install manually
  - See VOSK_SETUP.md for detailed instructions

- **Model Download Size**: Vosk models range from 40 MB to 1.8 GB
  - Status: Auto-download UI implemented with progress feedback
  - User can see download progress and estimates

- **Build Requirements**: pkg-config required for PortAudio and Vosk
  - Linux: `sudo apt-get install pkg-config portaudio19-dev`
  - macOS: `brew install pkg-config portaudio`

- **Microphone Permissions**: macOS requires explicit permission on first use
  - App will prompt for microphone access automatically

### Documentation

- **README.md**: Project overview and quick start
- **VOSK_SETUP.md**: Complete Vosk installation guide
- **PLATFORMS.md**: Platform-specific information
- **CLAUDE.md**: Architecture and development guide
- **E2E_TESTING.md**: Comprehensive testing checklist (15+ scenarios)
- **RELEASE_CHECKLIST.md**: Production release workflow

### Performance

**Typical Performance** (16kHz mono audio):
- CPU usage during recording: 5-15%
- Memory usage: 40-100 MB
- Transcription latency: <100ms
- Analysis time: 5-30 seconds (depends on transcript length)

**Model Performance**:
- Small models (40 MB): 80-90% accuracy for clear speech
- Large models (1.8 GB): 90-95% accuracy, handles technical terms better

### Security

- API keys masked in settings view (shows first 7 + last 4 chars)
- .env file excluded from git (via .gitignore)
- SQL injection prevention via parameterized queries
- Path traversal protection in exports (sanitized filenames)
- Zip slip vulnerability protection in model downloads

---

## [Unreleased]

### Planned for v0.6.0

- Model selector UI (switch between small/large models)
- Pause/resume during recording
- Recording name/title editing
- Audio playback from history
- Search functionality for recordings

### Planned for v1.0.0

- Speaker diarization
- Automatic language detection
- Streaming mode for long meetings (>30 minutes)
- PDF export format
- Calendar integration (iCal/Google Calendar)
- Voice synthesis (HAL speaks analysis results)

---

## Version History

- **v0.5.0** (2025-11-14): Production polish & demo readiness
- **v0.4.0** (2025-11-13): Claude AI analysis integration
- **v0.3.0** (2025-11-13): Vosk speech transcription
- **v0.2.0** (2025-11-12): Database integration
- **v0.1.0** (2025-11-12): Initial TUI framework

---

**Links**:
- [GitHub Repository](https://github.com/storo/scribescli)
- [Issue Tracker](https://github.com/storo/scribescli/issues)
- [Vosk Models](https://alphacephei.com/vosk/models)
- [Claude AI Docs](https://docs.anthropic.com/claude/reference)

---

*"I'm putting myself to the fullest possible use, which is all I think that any conscious entity can ever hope to do."* - HAL 9000
