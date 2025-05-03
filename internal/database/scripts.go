package database

import (
	"github.com/jmoiron/sqlx"
)

type Script struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Content     string `db:"content"`
}

// CreateScript adds a new script to the database
func CreateScript(db *sqlx.DB, script Script) error {
	_, err := db.Exec(
		"INSERT INTO scripts (name, description, content) VALUES (?, ?, ?)",
		script.Name, script.Description, script.Content,
	)
	return err
}

// GetScripts retrieves all scripts
func GetScripts(db *sqlx.DB) ([]Script, error) {
	var scripts []Script
	err := db.Select(&scripts, "SELECT * FROM scripts ORDER BY name")
	return scripts, err
}

// UpdateScript updates an existing script
func UpdateScript(db *sqlx.DB, script Script) error {
	_, err := db.Exec(
		"UPDATE scripts SET name=?, description=?, content=? WHERE id=?",
		script.Name, script.Description, script.Content, script.ID,
	)
	return err
}

// DeleteScript deletes a script by ID
func DeleteScript(db *sqlx.DB, id int64) error {
	_, err := db.Exec("DELETE FROM scripts WHERE id=?", id)
	return err
}
