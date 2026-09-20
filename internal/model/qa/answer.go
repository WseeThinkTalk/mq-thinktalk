package model

import (
	"time"

	"gorm.io/gorm"
)

type Answer struct {
	ID         int64 `gorm:"primary_key"`
	QuestionID int64
	AuthorID   int64
	Content    string
	IsAccepted int
	LikeNum    int
	ReplyNum   int
	Status     int
	CreateTime time.Time
	UpdateTime time.Time
}

func (m *Answer) TableName() string {
	return "answer"
}

type AnswerModel struct {
	db *gorm.DB
}

func NewAnswerModel(db *gorm.DB) *AnswerModel {
	return &AnswerModel{db: db}
}
