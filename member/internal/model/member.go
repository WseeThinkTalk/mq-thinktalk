package model

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Member struct {
	ID         int64 `gorm:"primary_key"`
	UserID     int64
	Level      int32
	ExpireTime time.Time
	Status     int32
	CreateTime time.Time
	UpdateTime time.Time
}

func (m *Member) TableName() string {
	return "member"
}

type MemberModel struct {
	db *gorm.DB
}

func NewMemberModel(db *gorm.DB) *MemberModel {
	return &MemberModel{db: db}
}

func (m *MemberModel) UpsertMember(ctx context.Context, data *Member) error {
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"level", "expire_time", "status", "update_time"}),
	}).Create(data).Error
}
