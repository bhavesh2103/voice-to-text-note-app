package handlers

import (
	"encoding/json"
	"net/http"
	"voice-to-text-app/storage"
)

func GetNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := storage.GetAllNotes()
	if err != nil {
		http.Error(w, "Failed to fetch notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
