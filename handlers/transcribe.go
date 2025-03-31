package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"voice-to-text-app/storage"
)

const deepgramAPIKey = "" // add your key

func init() {
	// Create recordings directory if it doesn't exist
	if err := os.MkdirAll("recordings", 0755); err != nil {
		log.Printf("Failed to create recordings directory: %v", err)
	}
}

func saveFile(data []byte, filename string) error {
	path := filepath.Join("recordings", filename)
	return os.WriteFile(path, data, 0644)
}

func detectAudioFormat(data []byte) string {
	// Check file signatures (magic numbers)
	if len(data) < 12 {
		return "unknown"
	}

	// WebM starts with 0x1A 0x45 0xDF 0xA3
	if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return "webm"
	}

	// MP3 starts with ID3 or 0xFF 0xFB
	if (data[0] == 0x49 && data[1] == 0x44 && data[2] == 0x33) || // ID3
		(data[0] == 0xFF && (data[1]&0xFB) == 0xFB) { // MPEG sync
		return "mp3"
	}

	// WAV starts with RIFF....WAVE
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x41 && data[10] == 0x56 && data[11] == 0x45 {
		return "wav"
	}

	return "unknown"
}

func convertAudioToWav(audioData []byte) ([]byte, error) {
	format := detectAudioFormat(audioData)
	log.Printf("Detected audio format: %s", format)

	// If it's already WAV and meets our requirements, return as-is
	if format == "wav" {
		return audioData, nil
	}

	// Create temporary files for input and output
	tempDir := os.TempDir()
	inputPath := filepath.Join(tempDir, fmt.Sprintf("input.%s", format))
	outputPath := filepath.Join(tempDir, "output.wav")

	// Write audio data to temp file
	if err := os.WriteFile(inputPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp input file: %v", err)
	}
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	// Convert to WAV using system ffmpeg with high quality settings
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-acodec", "pcm_s24le", // 24-bit audio
		"-ac", "2", // Stereo
		"-ar", "48000", // 48kHz sample rate
		"-af", "highpass=f=50,lowpass=f=20000", // Basic audio filtering
		"-y", // Overwrite output file
		outputPath,
	)

	// Capture any error output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %v, stderr: %s", err, stderr.String())
	}

	// Read the converted WAV file
	wavData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read converted wav file: %v", err)
	}

	log.Printf("Converted audio size: %d bytes", len(wavData))
	return wavData, nil
}

func HandleTranscription(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read audio file from form
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Failed to read audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s, size: %d bytes", header.Filename, header.Size)

	// Read file into buffer
	audioData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read audio data", http.StatusInternalServerError)
		return
	}

	// Save incoming audio file
	timestamp := time.Now().Format("20060102-150405")
	incomingFilename := fmt.Sprintf("incoming_%s.webm", timestamp)
	if err := saveFile(audioData, incomingFilename); err != nil {
		log.Printf("Failed to save incoming file: %v", err)
	}

	// Convert and save WAV
	wavData, err := convertAudioToWav(audioData)
	if err != nil {
		http.Error(w, "Failed to convert audio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	convertedFilename := fmt.Sprintf("converted_%s.wav", timestamp)
	if err := saveFile(wavData, convertedFilename); err != nil {
		log.Printf("Failed to save converted file: %v", err)
	}

	// Get transcript
	transcript, err := sendToDeepgram(wavData)
	if err != nil {
		http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to database

	// Save to database
	if err := storage.SaveNote(convertedFilename, transcript); err != nil {
		log.Printf("Failed to save to database: %v", err)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"transcript":    transcript,
		"timestamp":     timestamp,
		"incomingFile":  incomingFilename,
		"convertedFile": convertedFilename,
	})
}

func sendToDeepgram(audio []byte) (string, error) {
	url := "https://api.deepgram.com/v1/listen?language=en-US&model=general&tier=enhanced"

	req, err := http.NewRequest("POST", url, bytes.NewReader(audio))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+deepgramAPIKey)
	req.Header.Set("Content-Type", "audio/wav")

	log.Printf("Sending WAV audio of size: %d bytes", len(audio))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Deepgram request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and log the raw response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Log full response for debugging
	log.Printf("Response Status: %d", resp.StatusCode)
	log.Printf("Response Headers: %v", resp.Header)
	log.Printf("Deepgram Response: %s", string(respBody))

	// Parse response
	var result struct {
		Results struct {
			Channels []struct {
				Alternatives []struct {
					Transcript string  `json:"transcript"`
					Confidence float64 `json:"confidence"`
				} `json:"alternatives"`
			} `json:"channels"`
		} `json:"results"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(result.Results.Channels) == 0 ||
		len(result.Results.Channels[0].Alternatives) == 0 {
		return "", fmt.Errorf("no transcription results found")
	}

	transcript := result.Results.Channels[0].Alternatives[0].Transcript
	confidence := result.Results.Channels[0].Alternatives[0].Confidence

	log.Printf("Transcript confidence: %f", confidence)

	if transcript == "" {
		return "", fmt.Errorf("empty transcript received (confidence: %f)", confidence)
	}

	return transcript, nil
}
func HandleGetRecordings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	notes, err := storage.GetAllNotes()
	if err != nil {
		http.Error(w, "Failed to get recordings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
