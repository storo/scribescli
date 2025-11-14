# Building ScribesAI

## Prerequisites

ScribesAI requires the following system dependencies:

### macOS

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install pkg-config portaudio
```

### Linux (Debian/Ubuntu)

```bash
sudo apt-get update
sudo apt-get install -y pkg-config portaudio19-dev
```

### Linux (Fedora)

```bash
sudo dnf install -y pkg-config portaudio-devel
```

### Linux (Arch)

```bash
sudo pacman -S pkg-config portaudio
```

## Building

Once dependencies are installed:

```bash
# Build the application
go build -o scribescli ./cmd/scribescli

# Or build with version information
VERSION="1.0.0"
go build -ldflags "-X main.Version=$VERSION" -o scribescli ./cmd/scribescli
```

## Running

```bash
# Create .env file with your API key
echo "ANTHROPIC_API_KEY=your-api-key-here" > .env

# Run the application
./scribescli

# Or with debug logging
./scribescli --debug
```

## Troubleshooting

### "pkg-config: executable file not found"

This means pkg-config is not installed. Follow the prerequisites section for your OS.

### "Package portaudio-2.0 was not found"

This means PortAudio is not installed. Follow the prerequisites section for your OS.

### WSL2 Audio Issues

If you're running on WSL2, audio capture may not work. You can still use the application for viewing history and analysis, but recording will fail. For full functionality, run on native Linux or macOS.

## Development

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Vet code
go vet ./...
```
