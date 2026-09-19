package model

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Conversation struct {
	ID              int64 `gorm:"primary_key"`
	UserID          int64
	TargetUserID    int64
	LastMessage     string
	LastMessageTime time.Time
	UnreadCount     int
	CreateTime      time.Time
	UpdateTime      time.Time
}

func (m *Conversation) TableName() string {
	return "conversation"
}

type ConversationModel struct {
	db *gorm.DB
}

func NewConversationModel(db *gorm.DB) *ConversationModel {
	return &ConversationModel{db: db}
}

func (m *ConversationModel) FindByUserIDAndTargetID(ctx context.Context, userId, targetId int64) (*Conversation, error) {
	var result Conversation
	err := m.db.WithContext(ctx).Where("user_id = ? AND target_user_id = ?", userId, targetId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *ConversationModel) UpsertConversation(ctx context.Context, conv *Conversation) error {
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "target_user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_message":      conv.LastMessage,
			"last_message_time": conv.LastMessageTime,
			"unread_count":      gorm.Expr("conversation.unread_count + 1"),
			"update_time":       conv.UpdateTime,
		}),
	}).Create(conv).Error
}
