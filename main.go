package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/Jeff-Kidzie/todo-go/database"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

func initDb() {
	var err error
	db, err = database.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
}

func main() {
	initDb()
	defer db.Close()
	router := gin.Default()

	//Routes
	router.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Welcome to the Todo API")
	})
	router.POST("/todos/add", AddTodoHandler)
	router.PUT("/todos/update", UpdateTodoHandler)
	router.GET("/todos/all", GetAllTodosHandler)
	router.DELETE("/todos/delete", DeleteTodoHandler)
	router.Run(":8080")
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
