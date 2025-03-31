package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./notes.db")
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT,
		transcript TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func SaveNote(filename, transcript string) error {
	_, err := DB.Exec(`INSERT INTO notes (filename, transcript) VALUES (?, ?)`, filename, transcript)
	return err
}

type Note struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	Transcript string    `json:"transcript"`
	CreatedAt  time.Time `json:"created_at"`
}

func GetAllNotes() ([]Note, error) {
	rows, err := DB.Query(`SELECT id, filename, transcript, created_at FROM notes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		err := rows.Scan(&n.ID, &n.Filename, &n.Transcript, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}
