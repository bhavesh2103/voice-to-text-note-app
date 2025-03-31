package main

import (
	"log"
	"net/http"

	"voice-to-text-app/handlers"
	"voice-to-text-app/storage"
)

func main() {
	// Initialize SQLite database
	storage.InitDB()

	// Register routes with CORS wrapper
	http.HandleFunc("/transcribe", withCORS(handlers.HandleTranscription))
	http.HandleFunc("/notes", withCORS(handlers.GetNotes))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// withCORS adds CORS headers to allow frontend access
func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h(w, r)
	}
}
