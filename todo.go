package main

import (
	"database/sql"
	"fmt"
	"time"
)

type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsDone      bool      `json:"is_done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func Add(db *sql.DB, todo Todo) (int, error) {
	sqlStatement := `INSERT INTO todos (title, description, is_done, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int
	err := db.QueryRow(sqlStatement, todo.Title, todo.Description, false, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func Update(db *sql.DB, todo Todo) error {
	sqlStatement := "UPDATE todos SET title=$1, description=$2, is_done=$3, created_at=$4, updated_at=$5 WHERE id=$6"
	res, err := db.Exec(sqlStatement, todo.Title, todo.Description, todo.IsDone, time.Now(), time.Now(), todo.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Handle error when no rows are affected
		return sql.ErrNoRows
	}
	return nil
}

func Delete(db *sql.DB, id int) error {
	sqlStatement := "DELETE from todos WHERE id=$1"
	result, err := db.Exec(sqlStatement, id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func AllList(db *sql.DB) ([]Todo, error) {
	sqlStatement := `SELECT id, title, description, is_done, created_at, updated_at FROM todos`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.IsDone, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, rows.Err()
}
