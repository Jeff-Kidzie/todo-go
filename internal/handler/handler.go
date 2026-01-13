package handler

import (
	"database/sql"
	"github.com/Jeff-Kidzie/todo-go/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	db *sql.DB
}

func AddTodoHandler(h *Handler,c *gin.Context) {
	var todoInput models.Todo
	if err := c.ShouldBindJSON(&todoInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := models.Add(h.db, todoInput)
	throwErrorIfPresent(err, c)

	c.JSON(http.StatusOK, gin.H{"message": "Todo added successfully", "id": id})
	// Implementation for adding a todo via HTTP handler
}

func GetAllTodosHandler(h *Handler,c *gin.Context) {
	todos, err := models.AllList(h.db)
	throwErrorIfPresent(err, c)
	if len(todos) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No todos found"})
		return
	}
	c.JSON(http.StatusOK, todos)
}

func UpdateTodoHandler(h *Handler,c *gin.Context) {
	var todoInput models.Todo
	if err := c.ShouldBindJSON(&todoInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.Update(h.db, todoInput)
	throwErrorIfPresent(err, c)

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func DeleteTodoHandler(h *Handler,c *gin.Context) {
	var req struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.Delete(h.db, req.ID)
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
