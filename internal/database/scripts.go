package database

import (
	"github.com/jmoiron/sqlx"
)

type Script struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Content     string `db:"content"`
	Category    string `db:"category"`
}

// CreateScript adds a new script to the database
func CreateScript(db *sqlx.DB, script Script) error {
	_, err := db.Exec(
		"INSERT INTO scripts (name, description, content, category) VALUES (?, ?, ?, ?)",
		script.Name, script.Description, script.Content, script.Category,
	)
	return err
}

// GetScripts retrieves all scripts
func GetScripts(db *sqlx.DB) ([]Script, error) {
	var scripts []Script
	err := db.Select(&scripts, "SELECT * FROM scripts ORDER BY name")
	return scripts, err
}

func GetScriptByID(db *sqlx.DB, id int64) (Script, error) {
	var script Script
	err := db.Get(&script, "SELECT * FROM scripts WHERE id=?", id)
	return script, err
}

// UpdateScript updates an existing script
func UpdateScript(db *sqlx.DB, script Script) error {
	_, err := db.Exec(
		"UPDATE scripts SET name=?, description=?, content=?, category=? WHERE id=?",
		script.Name, script.Description, script.Content, script.Category, script.ID,
	)
	return err
}

// DeleteScript deletes a script by ID
func DeleteScript(db *sqlx.DB, id int64) error {
	_, err := db.Exec("DELETE FROM scripts WHERE id=?", id)
	return err
}
