package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID             int64 `gorm:"primary_key"`
	ConversationID int64
	SenderID       int64
	ReceiverID     int64
	Content        string
	MsgType        int
	IsRead         int
	Status         int
	CreateTime     time.Time
}

func (m *Message) TableName() string {
	return "message"
}

type MessageModel struct {
	db *gorm.DB
}

func NewMessageModel(db *gorm.DB) *MessageModel {
	return &MessageModel{db: db}
}

func (m *MessageModel) Insert(ctx context.Context, data *Message) error {
	return m.db.WithContext(ctx).Create(data).Error
}
