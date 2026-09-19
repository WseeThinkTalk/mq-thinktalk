package model

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConcernedCount struct {
	ID           int64 `gorm:"primary_key"`
	BizID        string
	ObjID        int64
	ConcernedNum int
	CreateTime   time.Time
	UpdateTime   time.Time
}

func (m *ConcernedCount) TableName() string {
	return "concerned_count"
}

type ConcernedCountModel struct {
	db *gorm.DB
}

func NewConcernedCountModel(db *gorm.DB) *ConcernedCountModel {
	return &ConcernedCountModel{db: db}
}

func (m *ConcernedCountModel) IncrConcernedNum(ctx context.Context, bizId string, objId int64) error {
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}, {Name: "obj_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"concerned_num": gorm.Expr("concerned_count.concerned_num + 1"),
			"update_time":   time.Now(),
		}),
	}).Create(&ConcernedCount{
		BizID:        bizId,
		ObjID:        objId,
		ConcernedNum: 1,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}).Error
}

func (m *ConcernedCountModel) DecrConcernedNum(ctx context.Context, bizId string, objId int64) error {
	return m.db.WithContext(ctx).
		Exec("UPDATE concerned_count SET concerned_num = concerned_num - 1 WHERE biz_id = ? AND obj_id = ? AND concerned_num > 0", bizId, objId).
		Error
}
