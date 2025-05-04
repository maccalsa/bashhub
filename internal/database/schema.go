package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var Schema = `
CREATE TABLE IF NOT EXISTS scripts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	content TEXT NOT NULL
);`

func ConnectDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "scripts.db")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(Schema)

	// db.MustExec("ALTER TABLE scripts ADD COLUMN category TEXT DEFAULT 'General';")
	// db.MustExec("INSERT INTO scripts (name, description, content) VALUES ('Hello World', 'Test script', 'echo Hello World'), ('Date', 'Print current date', 'date');")

	return db
}
