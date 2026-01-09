package main

import (
	"fmt"
	"time"
)

type Todo struct {
	ID          int       `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text"`
	IsDone      bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func Add(title string, description string) Todo {
	fmt.Println("Adding a new todo item")
	todo := Todo{
		Title:       title,
		Description: description,
		IsDone:      false,
	}
	return todo
}

func Update(id int, title string, description string, isDone bool) {

}

func Delete(id int) {

}

func AllList() {

}
