package model

import (
	"time"

	"gorm.io/gorm"
)

type Question struct {
	ID         int64 `gorm:"primary_key"`
	Title      string
	Content    string
	AuthorID   int64
	Status     int
	AnswerNum  int
	ViewNum    int
	TagIds     string
	CreateTime time.Time
	UpdateTime time.Time
}

func (m *Question) TableName() string {
	return "question"
}

type QuestionModel struct {
	db *gorm.DB
}

func NewQuestionModel(db *gorm.DB) *QuestionModel {
	return &QuestionModel{db: db}
}
