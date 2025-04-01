# Voice-to-Text Note App (Go + React + Deepgram)

A lightweight full-stack app that lets users record voice notes, preview them, and transcribe them into text using Deepgram AI. Built with Go (backend) and React (frontend), this project is fast, efficient, and perfect for voice journaling, note-taking, and more.

## Features:
- Record audio directly from the browser
- Select microphone input
- Preview recordings before transcribing
- Transcribe later option (manual trigger)
- Real-time transcription via Deepgram
- Stores last 3 recordings in memory (frontend)
- Saves transcripts in SQLite (backend)
- Simple REST API (Go-based)
- CORS-enabled for local development

## Tech Stack:
Layer:     Frontend  -> Tech: React + Vite + Tailwind
Layer:     Backend   -> Tech: Go (Golang)
Layer:     Database  -> Tech: SQLite
Layer:     AI API    -> Tech: Deepgram (speech-to-text)
Layer:     Tools     -> Tech: FFmpeg (for audio conversion)

## Setup Instructions:

1. Clone the project:
   git clone https://github.com/bhavesh2103/voice-to-text-note-app.git
   cd voice-to-text-note-app

2. Backend Setup (Go):
   - Add your Deepgram API key in handlers/transcribe.go:
     const deepgramAPIKey = "YOUR_DEEPGRAM_API_KEY"
   - Then run:
     go mod tidy  
     go run main.go

   Notes:
   - Ensure FFmpeg is installed and added to your system PATH.
   - A local SQLite database (notes.db) will be created automatically on first run.

3. Frontend Setup (React):
   - Navigate to the frontend directory:
     cd voice-to-text-frontend/
   - Run:
     npm install  
     npm run dev

   Access the frontend at: http://localhost:5173  
   Backend runs at: http://localhost:8080

## API Endpoints:
Method: POST   | Endpoint: /transcribe   | Description: Upload audio and receive transcript
Method: GET    | Endpoint: /recordings   | Description: Fetch all stored transcripts from the DB

## How It Works:
1. The user records voice using their selected microphone.
2. After stopping:
   - The user can preview the recording.
   - The user can choose to "Transcribe Now" or skip for later.
3. The transcription is performed via Deepgram’s speech-to-text API.
4. The result is stored in the backend (SQLite).
5. The frontend retains the last 3 recordings in memory for preview.

## Potential Future Features:
- Add tags or titles to voice notes
- Full transcript history UI (loaded from the backend)
- Audio download / delete options
- Summarization using GPT or Claude
- Multi-user support & authentication
- PWA or Desktop app with Tauri/Electron

## Built With:
- Go
- React
- Tailwind CSS
- Deepgram
- FFmpeg

## License:
MIT License — free to use, modify, and distribute.
