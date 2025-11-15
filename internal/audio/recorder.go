package audio

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate      = 16000
	ChannelsCount   = 1
	FramesPerBuffer = 1024
)

// RecordingState represents the current state of the recorder
type RecordingState int

const (
	StateIdle RecordingState = iota
	StateRecording
	StatePaused
	StateStopped
)

// Recorder manages audio recording
type Recorder struct {
	stream      *portaudio.Stream
	buffer      []int16
	bufferMutex sync.RWMutex
	state       RecordingState
	stateMutex  sync.RWMutex
	duration    time.Duration
	startTime   time.Time
	audioLevel  float32
	levelMutex  sync.RWMutex
}

// NewRecorder creates a new audio recorder
func NewRecorder() (*Recorder, error) {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("PortAudio initialization failed: %w", err)
	}

	return &Recorder{
		buffer: make([]int16, 0),
		state:  StateIdle,
	}, nil
}

// Start begins recording audio
func (r *Recorder) Start() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.state == StateRecording {
		return errors.New("already recording")
	}

	// Create input stream
	inputBuffer := make([]int16, FramesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(
		ChannelsCount, // input channels
		0,             // output channels
		SampleRate,
		FramesPerBuffer,
		inputBuffer,
	)
	if err != nil {
		return err
	}

	r.stream = stream
	r.startTime = time.Now()
	r.state = StateRecording

	// Start the stream
	if err := r.stream.Start(); err != nil {
		return err
	}

	// Start reading audio in goroutine
	go r.readAudio(inputBuffer)

	return nil
}

// readAudio continuously reads from the audio stream
func (r *Recorder) readAudio(inputBuffer []int16) {
	for {
		r.stateMutex.RLock()
		state := r.state
		r.stateMutex.RUnlock()

		if state == StateStopped {
			break
		}

		if state == StatePaused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err := r.stream.Read(); err != nil {
			// Handle error
			break
		}

		// Calculate audio level
		level := r.calculateLevel(inputBuffer)
		r.levelMutex.Lock()
		r.audioLevel = level
		r.levelMutex.Unlock()

		// Append to buffer
		r.bufferMutex.Lock()
		r.buffer = append(r.buffer, inputBuffer...)
		r.bufferMutex.Unlock()
	}
}

// calculateLevel computes the RMS (Root Mean Square) level of the audio
func (r *Recorder) calculateLevel(samples []int16) float32 {
	var sum float64
	for _, sample := range samples {
		normalized := float64(sample) / 32768.0
		sum += normalized * normalized
	}
	rms := math.Sqrt(sum / float64(len(samples)))
	return float32(rms)
}

// Pause pauses the recording
func (r *Recorder) Pause() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.state != StateRecording {
		return errors.New("not recording")
	}

	r.state = StatePaused
	return nil
}

// Resume resumes a paused recording
func (r *Recorder) Resume() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.state != StatePaused {
		return errors.New("not paused")
	}

	r.state = StateRecording
	return nil
}

// Stop stops the recording
func (r *Recorder) Stop() error {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	if r.state == StateIdle || r.state == StateStopped {
		return errors.New("not recording")
	}

	r.state = StateStopped

	if r.stream != nil {
		if err := r.stream.Stop(); err != nil {
			return err
		}
		if err := r.stream.Close(); err != nil {
			return err
		}
	}

	r.duration = time.Since(r.startTime)
	return nil
}

// GetBuffer returns a copy of the recorded audio buffer
func (r *Recorder) GetBuffer() []int16 {
	r.bufferMutex.RLock()
	defer r.bufferMutex.RUnlock()

	bufferCopy := make([]int16, len(r.buffer))
	copy(bufferCopy, r.buffer)
	return bufferCopy
}

// GetState returns the current recording state
func (r *Recorder) GetState() RecordingState {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.state
}

// GetDuration returns the current or total recording duration
func (r *Recorder) GetDuration() time.Duration {
	r.stateMutex.RLock()
	state := r.state
	r.stateMutex.RUnlock()

	if state == StateRecording || state == StatePaused {
		return time.Since(r.startTime)
	}
	return r.duration
}

// GetAudioLevel returns the current audio level (0.0 to 1.0)
func (r *Recorder) GetAudioLevel() float32 {
	r.levelMutex.RLock()
	defer r.levelMutex.RUnlock()
	return r.audioLevel
}

// Close cleans up resources
func (r *Recorder) Close() error {
	if r.stream != nil {
		r.Stop()
	}
	return portaudio.Terminate()
}
