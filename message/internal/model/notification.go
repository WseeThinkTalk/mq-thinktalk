package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID            int64 `gorm:"primary_key"`
	UserID        int64
	Type          int32
	Title         string
	Content       string
	RefID         int64
	BizID         string
	TriggerUserID int64
	IsRead        int
	CreateTime    time.Time
	UpdateTime    time.Time
}

func (m *Notification) TableName() string {
	return "notification"
}

type NotificationModel struct {
	db *gorm.DB
}

func NewNotificationModel(db *gorm.DB) *NotificationModel {
	return &NotificationModel{db: db}
}

func (m *NotificationModel) Insert(ctx context.Context, data *Notification) error {
	return m.db.WithContext(ctx).Create(data).Error
}
