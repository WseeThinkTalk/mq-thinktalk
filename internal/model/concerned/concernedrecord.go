package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type ConcernedRecord struct {
	ID         int64 `gorm:"primary_key"`
	BizID      string
	ObjID      int64
	UserID     int64
	Status     int
	CreateTime time.Time
	UpdateTime time.Time
}

func (m *ConcernedRecord) TableName() string {
	return "concerned_record"
}

type ConcernedRecordModel struct {
	db *gorm.DB
}

func NewConcernedRecordModel(db *gorm.DB) *ConcernedRecordModel {
	return &ConcernedRecordModel{db: db}
}

func (m *ConcernedRecordModel) Insert(ctx context.Context, data *ConcernedRecord) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *ConcernedRecordModel) FindByBizIDObjIDUserID(ctx context.Context, bizId string, objId, userId int64) (*ConcernedRecord, error) {
	var result ConcernedRecord
	err := m.db.WithContext(ctx).Where("biz_id = ? AND obj_id = ? AND user_id = ?", bizId, objId, userId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *ConcernedRecordModel) UpdateFields(ctx context.Context, id int64, values map[string]interface{}) error {
	return m.db.WithContext(ctx).Model(&ConcernedRecord{}).Where("id = ?", id).Updates(values).Error
}
