package model

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReplyCount struct {
	ID           int64 `gorm:"primary_key"`
	BizID        string
	TargetID     int64
	ReplyNum     int
	ReplyRootNum int
	CreateTime   time.Time
	UpdateTime   time.Time
}

func (m *ReplyCount) TableName() string {
	return "reply_count"
}

type ReplyCountModel struct {
	db *gorm.DB
}

func NewReplyCountModel(db *gorm.DB) *ReplyCountModel {
	return &ReplyCountModel{db: db}
}

func (m *ReplyCountModel) Insert(ctx context.Context, data *ReplyCount) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *ReplyCountModel) FindOne(ctx context.Context, id int64) (*ReplyCount, error) {
	var result ReplyCount
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *ReplyCountModel) IncrReplyNum(ctx context.Context, bizId string, targetId int64, isRoot bool) error {
	rootIncr := 0
	if isRoot {
		rootIncr = 1
	}
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}, {Name: "target_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"reply_num":      gorm.Expr("reply_count.reply_num + 1"),
			"reply_root_num": gorm.Expr("reply_count.reply_root_num + ?", rootIncr),
			"update_time":    time.Now(),
		}),
	}).Create(&ReplyCount{
		BizID:        bizId,
		TargetID:     targetId,
		ReplyNum:     1,
		ReplyRootNum: rootIncr,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}).Error
}

func (m *ReplyCountModel) DecrReplyNum(ctx context.Context, bizId string, targetId int64, isRoot bool) error {
	rootDecr := 0
	if isRoot {
		rootDecr = 1
	}
	return m.db.WithContext(ctx).
		Exec("UPDATE reply_count SET reply_num = reply_num - 1, reply_root_num = reply_root_num - ? WHERE biz_id = ? AND target_id = ? AND reply_num > 0",
			rootDecr, bizId, targetId).
		Error
}
