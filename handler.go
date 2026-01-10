package main

import (
	"database/sql"
	"net/http"
	"github.com/gin-gonic/gin"
)

func AddTodoHandler(c *gin.Context) {
	var todoInput Todo
	if err := c.ShouldBindJSON(&todoInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := Add(db, todoInput)
	throwErrorIfPresent(err, c)

	c.JSON(http.StatusOK, gin.H{"message": "Todo added successfully", "id": id})
	// Implementation for adding a todo via HTTP handler
}

func GetAllTodosHandler(c *gin.Context) {
	todos, err := AllList(db)
	throwErrorIfPresent(err, c)
	c.JSON(http.StatusOK, todos)
}

func UpdateTodoHandler(c *gin.Context) {
	var todoInput Todo
	if err := c.ShouldBindJSON(&todoInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := Update(db, todoInput)
	throwErrorIfPresent(err, c)

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func DeleteTodoHandler(c *gin.Context) {
	var req struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := Delete(db, req.ID)
	throwErrorIfPresent(err, c)

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func throwErrorIfPresent(err error, c *gin.Context) {
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
}
