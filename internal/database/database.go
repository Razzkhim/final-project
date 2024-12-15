package database

import (
	"database/sql"
	"final-project/internal/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func InitDB() (*sql.DB, error) {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		appPath, err := os.Executable()
		if err != nil {
			return nil, err
		}
		dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db")
	}
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("error while open db: %w", err)
	}

	if install {
		if err = createTable(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func createTable(db *sql.DB) error {
	query := `
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title TEXT NOT NULL DEFAULT "",
			comment TEXT DEFAULT "",
			repeat VARCHAR(128) DEFAULT ""
		);
		CREATE INDEX scheduler_date ON scheduler (date)
	`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to create table: %v", err)
		return err
	}
	return nil
}

func AddTask(db *sql.DB, task *models.Task) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	task.ID = id
	return nil
}

func GetTasksWithSearch(db *sql.DB, search string) ([]models.Task, error) {
	var query string
	var args []interface{}

	if search != "" {
		parsedDate, err := time.Parse("02.01.2006", search)
		if err == nil {
			query = `
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE date = ?
				ORDER BY date
				LIMIT 50
			`
			args = append(args, parsedDate.Format("20060102"))
		} else {
			searchPattern := "%" + search + "%"
			query = `
				SELECT id, date, title, comment, repeat
				FROM scheduler
				WHERE title LIKE ? OR comment LIKE ?
				ORDER BY date
				LIMIT 50
			`
			args = append(args, searchPattern, searchPattern)
		}
	} else {
		query = `
			SELECT id, date, title, comment, repeat
			FROM scheduler
			ORDER BY date
			LIMIT 50
		`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("failed to parse tasks: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

func GetTask(db *sql.DB, id int64) (*models.Task, error) {
	query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`

	row := db.QueryRow(query, id)

	var task models.Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}

	return &task, nil
}

func UpdateTask(db *sql.DB, task *models.Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func DeleteTask(db *sql.DB, id int64) error {
	query := `DELETE FROM scheduler WHERE id = ?`

	res, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func UpdateTaskDate(db *sql.DB, id int64, nextDate string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`

	res, err := db.Exec(query, nextDate, id)
	if err != nil {
		return fmt.Errorf("failed to update task date: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
