package main

import (
	"database/sql"
	"fmt"

	"github.com/Jeff-Kidzie/todo-go/database"
)

func main() {
	fmt.Println("Hello World")
	db := database.Connect()
	showPrompt(db)
}

func showPrompt(db *sql.DB) {
	fmt.Println("1. Show All Todos")
	fmt.Println("2. Add Todo")
	fmt.Println("3. Update Todo")
	fmt.Println("4. Delete Todo")
	fmt.Println("5. Exit")
	var choice int
	fmt.Print("Enter your choice: ")
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Invalid input. Please enter a number between 1 and 5.")
		showPrompt(db)
		return
	}

	switch choice {
	case 1:
		showAllTodo(db)
	case 2:
		AddTodo(db)
	case 3:
		UpdateTodo(db)
	case 4:
		DeleteTodo(db)
	case 5:
		fmt.Println("Exiting...")
		return
	default:
		fmt.Println("Invalid choice. Please try again.")
	}
	showPrompt(db)
}

func UpdateTodo(db *sql.DB) {
	panic("unimplemented")
}

func showAllTodo(db *sql.DB) {
	todos, err := AllList(db)
	if err != nil {
		fmt.Println(err)
	}
	if len(todos) == 0 {
		fmt.Println("No todos found")
		return
	}
	fmt.Println("id | Title | Description | Is Done | Created At | Updated At")
	for _, todo := range todos {
		fmt.Printf("%d | %s | %s | %t | %s | %s\n", todo.ID, todo.Title, todo.Description, todo.IsDone, todo.CreatedAt.Format("2006-01-02 15:04:05"), todo.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("-------------------------------")
}

func AddTodo(db *sql.DB) {
	var title, description string
	fmt.Print("Enter title: ")
	fmt.Scan(&title)
	fmt.Print("Enter description: ")
	fmt.Scan(&description)

	todo := Todo{
		Title:       title,
		Description: description,
	}

	id, err := Add(db, todo)
	if err != nil {
		fmt.Println("Error adding todo:", err)
		return
	}
	fmt.Printf("Todo added with ID: %d\n", id)
}

func DeleteTodo(db *sql.DB) {
	var id int
	fmt.Print("Enter ID of the todo to delete: ")
	fmt.Scan(&id)

	err := Delete(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No todo found with the given ID.")
		} else {
			fmt.Println("Error deleting todo:", err)
		}
		return
	}
	fmt.Println("Todo deleted successfully.")
}
