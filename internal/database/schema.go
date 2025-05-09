package database

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var Schema = `
CREATE TABLE IF NOT EXISTS scripts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	content TEXT NOT NULL,
	language TEXT DEFAULT 'bash',
	category TEXT DEFAULT 'General'
);`

func getDBPath() string {
    var baseDir string

    switch runtime.GOOS {
    case "darwin":
        baseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "bashhub")
    case "linux":
        baseDir = filepath.Join(os.Getenv("HOME"), ".config", "bashhub")
    default:
        log.Fatal("Unsupported OS")
    }

    if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
        log.Fatalf("Failed to create config directory: %v", err)
    }

    return filepath.Join(baseDir, "bashhub.db")
}


func ConnectDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", getDBPath())
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(Schema)

	return db
}
