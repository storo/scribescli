package transcription

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ModelInfo contains information about a Vosk model
type ModelInfo struct {
	Name        string
	Language    string
	Size        string
	URL         string
	Description string
}

// Available Vosk models
var AvailableModels = map[string]ModelInfo{
	"en-us-small": {
		Name:        "vosk-model-small-en-us-0.15",
		Language:    "English (US)",
		Size:        "40 MB",
		URL:         "https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip",
		Description: "Lightweight English model, fast and accurate for most use cases",
	},
	"en-us": {
		Name:        "vosk-model-en-us-0.22",
		Language:    "English (US)",
		Size:        "1.8 GB",
		URL:         "https://alphacephei.com/vosk/models/vosk-model-en-us-0.22.zip",
		Description: "Large English model, highest accuracy",
	},
	"es": {
		Name:        "vosk-model-small-es-0.42",
		Language:    "Spanish",
		Size:        "39 MB",
		URL:         "https://alphacephei.com/vosk/models/vosk-model-small-es-0.42.zip",
		Description: "Small Spanish model",
	},
	"fr": {
		Name:        "vosk-model-small-fr-0.22",
		Language:    "French",
		Size:        "41 MB",
		URL:         "https://alphacephei.com/vosk/models/vosk-model-small-fr-0.22.zip",
		Description: "Small French model",
	},
	"de": {
		Name:        "vosk-model-small-de-0.15",
		Language:    "German",
		Size:        "45 MB",
		URL:         "https://alphacephei.com/vosk/models/vosk-model-small-de-0.15.zip",
		Description: "Small German model",
	},
}

// ModelManager handles downloading and managing Vosk models
type ModelManager struct {
	modelsDir string
}

// NewModelManager creates a new model manager
func NewModelManager(modelsDir string) *ModelManager {
	return &ModelManager{
		modelsDir: modelsDir,
	}
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	BytesDownloaded int64
	TotalBytes      int64
	Percentage      float64
}

// DownloadModel downloads a Vosk model with progress tracking
func (m *ModelManager) DownloadModel(modelKey string, progressChan chan<- DownloadProgress) error {
	modelInfo, ok := AvailableModels[modelKey]
	if !ok {
		return fmt.Errorf("unknown model: %s", modelKey)
	}

	// Create models directory if it doesn't exist
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	modelPath := filepath.Join(m.modelsDir, modelInfo.Name)

	// Check if model already exists
	if _, err := os.Stat(modelPath); err == nil {
		if progressChan != nil {
			close(progressChan)
		}
		return nil // Model already downloaded
	}

	// Download model
	zipPath := filepath.Join(m.modelsDir, modelInfo.Name+".zip")
	if err := m.downloadFile(modelInfo.URL, zipPath, progressChan); err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Extract model
	if err := m.extractZip(zipPath, m.modelsDir); err != nil {
		os.Remove(zipPath)
		return fmt.Errorf("failed to extract model: %w", err)
	}

	// Clean up zip file
	os.Remove(zipPath)

	if progressChan != nil {
		close(progressChan)
	}

	return nil
}

// downloadFile downloads a file with progress tracking
func (m *ModelManager) downloadFile(url, destPath string, progressChan chan<- DownloadProgress) error {
	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get total size
	totalBytes := resp.ContentLength

	// Copy with progress tracking
	var bytesDownloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}

			bytesDownloaded += int64(n)

			// Send progress update
			if progressChan != nil && totalBytes > 0 {
				percentage := float64(bytesDownloaded) / float64(totalBytes) * 100.0
				progressChan <- DownloadProgress{
					BytesDownloaded: bytesDownloaded,
					TotalBytes:      totalBytes,
					Percentage:      percentage,
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// extractZip extracts a zip file to a destination directory
func (m *ModelManager) extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destDir, file.Name)

		// Check for Zip Slip vulnerability
		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, file.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Extract file
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		inFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, inFile)
		outFile.Close()
		inFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// GetModelPath returns the path to a downloaded model
func (m *ModelManager) GetModelPath(modelKey string) (string, error) {
	modelInfo, ok := AvailableModels[modelKey]
	if !ok {
		return "", fmt.Errorf("unknown model: %s", modelKey)
	}

	modelPath := filepath.Join(m.modelsDir, modelInfo.Name)

	// Check if model exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("model not downloaded: %s", modelKey)
	}

	return modelPath, nil
}

// ListDownloadedModels returns a list of downloaded models
func (m *ModelManager) ListDownloadedModels() ([]string, error) {
	entries, err := os.ReadDir(m.modelsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var models []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "vosk-model-") {
			models = append(models, entry.Name())
		}
	}

	return models, nil
}

// DeleteModel removes a downloaded model
func (m *ModelManager) DeleteModel(modelKey string) error {
	modelInfo, ok := AvailableModels[modelKey]
	if !ok {
		return fmt.Errorf("unknown model: %s", modelKey)
	}

	modelPath := filepath.Join(m.modelsDir, modelInfo.Name)
	return os.RemoveAll(modelPath)
}

// GetDefaultModel returns the key for the default model (small English)
func (m *ModelManager) GetDefaultModel() string {
	return "en-us-small"
}

// EnsureDefaultModel downloads the default model if not present
func (m *ModelManager) EnsureDefaultModel(progressChan chan<- DownloadProgress) error {
	defaultModel := m.GetDefaultModel()

	// Check if already downloaded
	if _, err := m.GetModelPath(defaultModel); err == nil {
		if progressChan != nil {
			close(progressChan)
		}
		return nil // Already exists
	}

	// Download default model
	return m.DownloadModel(defaultModel, progressChan)
}
