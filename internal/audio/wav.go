package audio

import (
	"encoding/binary"
	"io"
	"os"
)

// WAVHeader represents a WAV file header
type WAVHeader struct {
	// RIFF chunk descriptor
	ChunkID   [4]byte // "RIFF"
	ChunkSize uint32  // file size - 8
	Format    [4]byte // "WAVE"

	// fmt sub-chunk
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 for mono, 2 for stereo
	SampleRate    uint32  // sampling rate
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8, 16, 32

	// data sub-chunk
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // data size
}

// SaveWAV saves audio samples to a WAV file
func SaveWAV(filename string, samples []int16, sampleRate, channels int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate sizes
	bitsPerSample := uint16(16) // int16 = 16 bits
	dataSize := uint32(len(samples) * 2) // 2 bytes per sample
	fileSize := dataSize + 36 // 36 bytes for header (excluding ChunkID and ChunkSize)

	// Create header
	header := WAVHeader{
		ChunkID:   [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize: fileSize,
		Format:    [4]byte{'W', 'A', 'V', 'E'},

		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(channels),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(sampleRate * channels * int(bitsPerSample) / 8),
		BlockAlign:    uint16(channels * int(bitsPerSample) / 8),
		BitsPerSample: bitsPerSample,

		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	// Write header
	if err := binary.Write(file, binary.LittleEndian, &header); err != nil {
		return err
	}

	// Write audio data
	if err := binary.Write(file, binary.LittleEndian, samples); err != nil {
		return err
	}

	return nil
}

// LoadWAV loads audio samples from a WAV file
func LoadWAV(filename string) ([]int16, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()

	// Read header
	var header WAVHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, 0, 0, err
	}

	// Read audio data
	samples := make([]int16, header.Subchunk2Size/2) // divide by 2 because int16 is 2 bytes
	if err := binary.Read(file, binary.LittleEndian, &samples); err != nil && err != io.EOF {
		return nil, 0, 0, err
	}

	return samples, int(header.SampleRate), int(header.NumChannels), nil
}

// SaveRecording saves the recorded audio to a WAV file
func (r *Recorder) SaveRecording(filename string) error {
	samples := r.GetBuffer()
	return SaveWAV(filename, samples, SampleRate, ChannelsCount)
}
