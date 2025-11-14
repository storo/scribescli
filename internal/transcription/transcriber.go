package transcription

import (
	"context"
	"io"
)

// Transcriber defines the interface for speech-to-text transcription
type Transcriber interface {
	// Initialize prepares the transcriber with a model
	Initialize(modelPath string) error

	// TranscribeStream transcribes audio from a stream in real-time
	// Returns a channel of partial and final results
	TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan Result, error)

	// TranscribeFile transcribes an entire audio file
	TranscribeFile(audioPath string) (string, error)

	// SetSampleRate configures the audio sample rate (default: 16000 Hz)
	SetSampleRate(rate int)

	// Close releases resources
	Close() error
}

// Result represents a transcription result
type Result struct {
	// Text is the transcribed text
	Text string

	// IsFinal indicates if this is a final result or partial
	IsFinal bool

	// Confidence is the recognition confidence (0.0 - 1.0)
	// Only available for final results
	Confidence float64

	// StartTime is the start time in seconds (optional)
	StartTime float64

	// EndTime is the end time in seconds (optional)
	EndTime float64
}

// TranscriptionError represents errors during transcription
type TranscriptionError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

func (e *TranscriptionError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *TranscriptionError) Unwrap() error {
	return e.Err
}
