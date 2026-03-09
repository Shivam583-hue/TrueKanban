package db

import (
	"database/sql"

	"github.com/Shivam583-hue/TrueKanban/types"
	"github.com/charmbracelet/bubbles/list"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() {
	var err error

	db, err = sql.Open("sqlite3", "./task.db")
	if err != nil {
		panic(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status INTEGER NOT NULL
	);`

	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func Insert(title string, status string) {
	_, err := db.Exec(
		"INSERT INTO tasks(title, status) VALUES(?, ?)",
		title,
		status,
	)
	if err != nil {
		panic(err)
	}
}

func Fetch(status string) []list.Item {
	rows, err := db.Query(
		"SELECT id, title, status FROM tasks WHERE status = ?",
		status,
	)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var items []list.Item
	for rows.Next() {
		var t types.Task
		err := rows.Scan(&t.Id, &t.TaskTitle, &t.Status)
		if err != nil {
			panic(err)
		}
		items = append(items, t)
	}
	return items
}

func Update(id int, newStatus string) {
	_, err := db.Exec(
		"UPDATE tasks SET status = ? WHERE id = ?",
		newStatus, id,
	)
	if err != nil {
		panic(err)
	}
}

func Delete(id int) {
	result, err := db.Exec(
		"DELETE FROM tasks WHERE id = ?",
		id,
	)
	if err != nil {
		panic(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}

	if rowsAffected == 0 {
		return
	}
}

func Close() {
	db.Close()
}
