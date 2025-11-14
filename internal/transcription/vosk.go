package transcription

/*
#cgo pkg-config: vosk
#cgo LDFLAGS: -lvosk
#include <vosk/vosk_api.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unsafe"
)

// VoskTranscriber implements the Transcriber interface using Vosk
type VoskTranscriber struct {
	model      *C.VoskModel
	sampleRate int
	modelPath  string
}

// voskResult represents the JSON structure returned by Vosk
type voskResult struct {
	Text       string  `json:"text"`
	Result     []word  `json:"result,omitempty"`
	Partial    string  `json:"partial,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type word struct {
	Conf  float64 `json:"conf"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Word  string  `json:"word"`
}

// NewVoskTranscriber creates a new Vosk-based transcriber
func NewVoskTranscriber() *VoskTranscriber {
	// Set Vosk log level (0 = silent, -1 = verbose)
	C.vosk_set_log_level(0)

	return &VoskTranscriber{
		sampleRate: 16000, // Default sample rate
	}
}

// Initialize loads the Vosk model
func (v *VoskTranscriber) Initialize(modelPath string) error {
	if modelPath == "" {
		return &TranscriptionError{
			Op:  "Initialize",
			Err: fmt.Errorf("model path cannot be empty"),
		}
	}

	// Check if model directory exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return &TranscriptionError{
			Op:  "Initialize",
			Err: fmt.Errorf("model directory does not exist: %s", modelPath),
		}
	}

	v.modelPath = modelPath

	// Convert Go string to C string
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	// Load model
	v.model = C.vosk_model_new(cModelPath)
	if v.model == nil {
		return &TranscriptionError{
			Op:  "Initialize",
			Err: fmt.Errorf("failed to load Vosk model from: %s", modelPath),
		}
	}

	return nil
}

// SetSampleRate sets the audio sample rate
func (v *VoskTranscriber) SetSampleRate(rate int) {
	v.sampleRate = rate
}

// TranscribeStream transcribes audio from a stream in real-time
func (v *VoskTranscriber) TranscribeStream(ctx context.Context, audioStream io.Reader) (<-chan Result, error) {
	if v.model == nil {
		return nil, &TranscriptionError{
			Op:  "TranscribeStream",
			Err: fmt.Errorf("model not initialized"),
		}
	}

	// Create recognizer
	recognizer := C.vosk_recognizer_new(v.model, C.float(v.sampleRate))
	if recognizer == nil {
		return nil, &TranscriptionError{
			Op:  "TranscribeStream",
			Err: fmt.Errorf("failed to create recognizer"),
		}
	}

	resultChan := make(chan Result, 10)

	// Start goroutine to process audio stream
	go func() {
		defer close(resultChan)
		defer C.vosk_recognizer_free(recognizer)

		buffer := make([]byte, 4000) // ~250ms at 16kHz mono

		for {
			select {
			case <-ctx.Done():
				// Context cancelled, finalize and return
				v.finalizeRecognizer(recognizer, resultChan)
				return

			default:
				n, err := audioStream.Read(buffer)
				if err != nil {
					if err == io.EOF {
						// End of stream, finalize
						v.finalizeRecognizer(recognizer, resultChan)
						return
					}
					// Other error, send error and return
					return
				}

				if n > 0 {
					// Process audio chunk
					v.processAudioChunk(recognizer, buffer[:n], resultChan)
				}
			}
		}
	}()

	return resultChan, nil
}

// processAudioChunk processes a chunk of audio data
func (v *VoskTranscriber) processAudioChunk(recognizer *C.VoskRecognizer, data []byte, resultChan chan<- Result) {
	// Accept audio data
	accepted := C.vosk_recognizer_accept_waveform(
		recognizer,
		(*C.char)(unsafe.Pointer(&data[0])),
		C.int(len(data)),
	)

	if accepted != 0 {
		// Final result available
		cResult := C.vosk_recognizer_result(recognizer)
		result := v.parseVoskJSON(C.GoString(cResult), true)
		if result.Text != "" {
			resultChan <- result
		}
	} else {
		// Partial result
		cPartial := C.vosk_recognizer_partial_result(recognizer)
		result := v.parseVoskJSON(C.GoString(cPartial), false)
		if result.Text != "" {
			resultChan <- result
		}
	}
}

// finalizeRecognizer gets the final result from the recognizer
func (v *VoskTranscriber) finalizeRecognizer(recognizer *C.VoskRecognizer, resultChan chan<- Result) {
	cFinal := C.vosk_recognizer_final_result(recognizer)
	result := v.parseVoskJSON(C.GoString(cFinal), true)
	if result.Text != "" {
		resultChan <- result
	}
}

// parseVoskJSON parses Vosk JSON output into a Result
func (v *VoskTranscriber) parseVoskJSON(jsonStr string, isFinal bool) Result {
	var vr voskResult
	if err := json.Unmarshal([]byte(jsonStr), &vr); err != nil {
		return Result{}
	}

	result := Result{
		IsFinal:    isFinal,
		Confidence: vr.Confidence,
	}

	if isFinal {
		result.Text = vr.Text
		// Calculate time bounds from word results
		if len(vr.Result) > 0 {
			result.StartTime = vr.Result[0].Start
			result.EndTime = vr.Result[len(vr.Result)-1].End
		}
	} else {
		result.Text = vr.Partial
	}

	return result
}

// TranscribeFile transcribes an entire WAV file
func (v *VoskTranscriber) TranscribeFile(audioPath string) (string, error) {
	if v.model == nil {
		return "", &TranscriptionError{
			Op:  "TranscribeFile",
			Err: fmt.Errorf("model not initialized"),
		}
	}

	// Open audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return "", &TranscriptionError{
			Op:  "TranscribeFile",
			Err: fmt.Errorf("failed to open audio file: %w", err),
		}
	}
	defer file.Close()

	// Skip WAV header (44 bytes)
	if _, err := file.Seek(44, io.SeekStart); err != nil {
		return "", &TranscriptionError{
			Op:  "TranscribeFile",
			Err: fmt.Errorf("failed to skip WAV header: %w", err),
		}
	}

	// Create context for transcription
	ctx := context.Background()

	// Transcribe
	resultChan, err := v.TranscribeStream(ctx, file)
	if err != nil {
		return "", err
	}

	// Collect all final results
	var transcript string
	for result := range resultChan {
		if result.IsFinal && result.Text != "" {
			if transcript != "" {
				transcript += " "
			}
			transcript += result.Text
		}
	}

	return transcript, nil
}

// Close releases resources
func (v *VoskTranscriber) Close() error {
	if v.model != nil {
		C.vosk_model_free(v.model)
		v.model = nil
	}
	return nil
}
