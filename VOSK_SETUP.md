# Vosk Speech Recognition Setup

This guide covers installing Vosk and downloading speech recognition models for ScribesAI.

## Overview

**Vosk** is an offline speech recognition toolkit that enables ScribesAI to transcribe meeting audio in real-time without requiring an internet connection.

### Features

- ✅ Offline transcription (no internet required)
- ✅ Multi-language support (50+ languages)
- ✅ Real-time streaming recognition
- ✅ Low resource usage (~15-50 MB RAM per model)
- ✅ High accuracy with proper models

## Installation

### Step 1: Install libvosk

ScribesAI includes an automated installation script for libvosk:

```bash
./scripts/install-vosk.sh
```

This script will:
1. Detect your OS and architecture
2. Download the correct libvosk binary (v0.3.45)
3. Install to `/usr/local/lib`
4. Configure pkg-config
5. Verify the installation

#### Manual Installation

If the script fails, you can install manually:

**macOS:**
```bash
# Download for your architecture
# Intel:
curl -L https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-osx-x86_64-0.3.45.zip -o vosk.zip

# Apple Silicon:
curl -L https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-osx-arm64-0.3.45.zip -o vosk.zip

# Extract and install
unzip vosk.zip
sudo cp vosk-*/libvosk.dylib /usr/local/lib/
sudo cp vosk-*/vosk_api.h /usr/local/include/
```

**Linux:**
```bash
# Download for your architecture
# x86_64:
curl -L https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-x86_64-0.3.45.zip -o vosk.zip

# ARM64:
curl -L https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-aarch64-0.3.45.zip -o vosk.zip

# Extract and install
unzip vosk.zip
sudo cp vosk-*/libvosk.so /usr/local/lib/
sudo cp vosk-*/vosk_api.h /usr/local/include/
sudo ldconfig
```

### Step 2: Download a Speech Recognition Model

Vosk requires a language model to perform transcription. Models are available in different sizes:

- **Small models** (~40 MB): Fast, good for most use cases
- **Large models** (~1.5-2 GB): Highest accuracy, more resource-intensive

#### Recommended Model: English (US) - Small

This is the default model for ScribesAI:

```bash
# Create models directory
mkdir -p models
cd models

# Download model
curl -L https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip -o model.zip

# Extract
unzip model.zip

# Verify
ls vosk-model-small-en-us-0.15/
# Should show: am/, conf/, graph/ and other files

cd ..
```

#### Alternative Models

**English (US) - Large (1.8 GB):**
```bash
cd models
curl -L https://alphacephei.com/vosk/models/vosk-model-en-us-0.22.zip -o model.zip
unzip model.zip
cd ..
```

**Spanish:**
```bash
cd models
curl -L https://alphacephei.com/vosk/models/vosk-model-small-es-0.42.zip -o model.zip
unzip model.zip
cd ..
```

**French:**
```bash
cd models
curl -L https://alphacephei.com/vosk/models/vosk-model-small-fr-0.22.zip -o model.zip
unzip model.zip
cd ..
```

**German:**
```bash
cd models
curl -L https://alphacephei.com/vosk/models/vosk-model-small-de-0.15.zip -o model.zip
unzip model.zip
cd ..
```

**More languages:** See https://alphacephei.com/vosk/models

### Step 3: Verify Setup

Test that everything is installed correctly:

```bash
# Check libvosk
pkg-config --exists vosk && echo "✓ libvosk found" || echo "✗ libvosk not found"

# Check model
ls models/vosk-model-small-en-us-0.15 && echo "✓ Model found" || echo "✗ Model not found"

# Build ScribesAI with Vosk support
go build -o scribescli ./cmd/scribescli
```

If the build succeeds, Vosk is properly configured!

## Configuration

### Environment Variables

You can customize the model path with:

```bash
export VOSK_MODEL_PATH="./models/vosk-model-small-en-us-0.15"
```

If not set, ScribesAI will use the default model path from the model manager.

### Model Selection

To use a different model:

1. Download and extract the model to `./models/`
2. The application will auto-detect it
3. Or specify via `VOSK_MODEL_PATH`

## Usage

Once installed, ScribesAI will automatically use Vosk for transcription:

1. Start a new recording
2. Speak clearly into your microphone
3. Real-time transcription will appear during recording
4. Full transcript is saved to the database when you stop

### Transcription Quality Tips

**For best results:**

- ✅ Use a good quality microphone
- ✅ Minimize background noise
- ✅ Speak clearly at a normal pace
- ✅ Use the appropriate language model
- ✅ For technical terms, use the large model

**Sample rates:**
- ScribesAI records at 16kHz by default
- This is optimal for Vosk (matches model training data)
- Higher sample rates (44.1kHz, 48kHz) are resampled to 16kHz

## Troubleshooting

### Error: "libvosk not found"

```bash
# Check installation
ls /usr/local/lib/libvosk.*

# On Linux, update cache
sudo ldconfig

# On macOS, check library path
export DYLD_LIBRARY_PATH=/usr/local/lib:$DYLD_LIBRARY_PATH
```

### Error: "model directory does not exist"

```bash
# Verify model path
ls models/vosk-model-small-en-us-0.15

# Re-download if missing
cd models
curl -L https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip -o model.zip
unzip model.zip
cd ..
```

### Build Error: "vosk_api.h: No such file"

```bash
# Check header installation
ls /usr/local/include/vosk/vosk_api.h

# If missing, reinstall:
./scripts/install-vosk.sh
```

### Poor Transcription Quality

1. **Check microphone input:**
   ```bash
   # macOS
   system_profiler SPAudioDataType
   
   # Linux
   arecord -l
   ```

2. **Try the large model:**
   - Download `vosk-model-en-us-0.22` (1.8 GB)
   - Much better accuracy for complex audio

3. **Reduce background noise:**
   - Use a headset microphone
   - Record in a quiet room
   - Enable noise cancellation in system settings

4. **Check audio levels:**
   - Speak 6-12 inches from microphone
   - Watch the audio level meter in ScribesAI
   - Aim for 50-80% level (green zone)

### Memory Issues

If you experience high memory usage:

- Use small models instead of large models
- Close other applications during recording
- Ensure you have at least 100 MB free RAM

## Technical Details

### How It Works

1. **Audio Capture:** PortAudio records at 16kHz, mono
2. **Buffering:** Audio is buffered in 250ms chunks (4000 bytes)
3. **Streaming Recognition:** Vosk processes chunks in real-time
4. **Partial Results:** Updated every 100ms
5. **Final Results:** Generated at sentence boundaries
6. **Post-processing:** Full transcript assembled and saved

### Performance

**Typical performance (16kHz, mono):**

| Platform | CPU Usage | Latency | RAM Usage |
|----------|-----------|---------|-----------|
| macOS ARM64 | 5-8% | <50ms | 40 MB |
| macOS Intel | 8-12% | <100ms | 45 MB |
| Linux x86_64 | 6-10% | <80ms | 42 MB |

**Model sizes:**

| Model | Disk Size | RAM Usage | Accuracy |
|-------|-----------|-----------|----------|
| Small EN-US | 40 MB | 15 MB | Good |
| Large EN-US | 1.8 GB | 50 MB | Excellent |
| Small ES | 39 MB | 15 MB | Good |
| Small FR | 41 MB | 15 MB | Good |

### API Reference

ScribesAI uses the Vosk Go API bindings:

```go
// Initialize
transcriber := transcription.NewVoskTranscriber()
transcriber.Initialize("./models/vosk-model-small-en-us-0.15")

// Transcribe stream
ctx := context.Background()
resultChan, err := transcriber.TranscribeStream(ctx, audioReader)

// Process results
for result := range resultChan {
    if result.IsFinal {
        fmt.Println("Final:", result.Text)
    } else {
        fmt.Println("Partial:", result.Text)
    }
}

// Cleanup
transcriber.Close()
```

## Resources

- **Vosk Project:** https://alphacephei.com/vosk/
- **Models:** https://alphacephei.com/vosk/models
- **API Docs:** https://github.com/alphacep/vosk-api
- **Go Bindings:** https://github.com/alphacep/vosk-api/tree/master/go

## FAQ

**Q: Do I need an internet connection for transcription?**  
A: No! Vosk works completely offline once the model is downloaded.

**Q: Can I use multiple languages?**  
A: Yes, download multiple models and switch between them. (UI coming in Phase 4)

**Q: How accurate is Vosk compared to cloud services?**  
A: With large models, accuracy is comparable to cloud services for clear audio. Small models are 85-90% accurate for general speech.

**Q: Can Vosk handle technical terminology?**  
A: Large models handle technical terms better. For specialized domains, you may need custom model training (advanced).

**Q: Does Vosk support speaker diarization?**  
A: Not built-in, but planned for Phase 5 of ScribesAI.

**Q: What's the minimum system requirements?**  
A: 100 MB free RAM, any modern CPU. Even a Raspberry Pi 4 can run Vosk!

---

**Last updated:** 2025-11-14  
**Vosk version:** 0.3.45  
**Default model:** vosk-model-small-en-us-0.15
