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
