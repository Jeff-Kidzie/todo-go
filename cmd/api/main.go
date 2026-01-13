package main

import (
	"database/sql"
	"fmt"
	"github.com/Jeff-Kidzie/todo-go/database"
	"github.com/Jeff-Kidzie/todo-go/internal/handler"
	"github.com/gin-gonic/gin"
	"net/http"
)

var sqlDb *sql.DB

func initDb() {
	var err error
	sqlDb, err = database.Connect()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
}

func main() {
	initDb()
	defer DB.Close()
	router := gin.Default()

	//Routes
	router.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Welcome to the Todo API")
	})
	handler := handler.Handler {
		db : sqlDb
	}
	router.POST("/todos/add", handler.AddTodoHandler())
	router.PUT("/todos/update", handler.UpdateTodoHandler)
	router.GET("/todos/all", handler.GetAllTodosHandler)
	router.DELETE("/todos/delete", handler.DeleteTodoHandler)
	router.Run(":8080")
}
