#!/bin/bash
# Run TUI tests without CGO dependencies
# This script helps test the database integration without requiring PortAudio

set -e

echo "Running TUI database integration tests..."
echo

# Set CGO_ENABLED=0 to skip audio-related compilation
# Note: This means we can't test audio functionality, but we can test database integration
export CGO_ENABLED=0

# Run only the database-related tests
echo "Testing storage layer..."
go test -v ./internal/storage

echo
echo "Storage tests complete!"
echo
echo "Note: TUI tests require CGO for audio support and cannot run with CGO_ENABLED=0"
echo "To run full TUI tests, ensure PortAudio is installed:"
echo "  - macOS: brew install portaudio"
echo "  - Linux: sudo apt-get install portaudio19-dev"
echo "  - WSL2: Run ./scripts/setup-wsl-audio.sh"
