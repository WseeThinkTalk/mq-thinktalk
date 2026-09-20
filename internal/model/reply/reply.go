package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Reply struct {
	ID            int64 `gorm:"primary_key"`
	BizID         string
	TargetID      int64
	ReplyUserID   int64
	BeReplyUserID int64
	ParentID      int64
	Content       string
	Status        int
	LikeNum       int
	CreateTime    time.Time
	UpdateTime    time.Time
}

func (m *Reply) TableName() string {
	return "reply"
}

type ReplyModel struct {
	db *gorm.DB
}

func NewReplyModel(db *gorm.DB) *ReplyModel {
	return &ReplyModel{db: db}
}

func (m *ReplyModel) Insert(ctx context.Context, data *Reply) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *ReplyModel) FindOne(ctx context.Context, id int64) (*Reply, error) {
	var result Reply
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *ReplyModel) UpdateFields(ctx context.Context, id int64, values map[string]interface{}) error {
	return m.db.WithContext(ctx).Model(&Reply{}).Where("id = ?", id).Updates(values).Error
}
