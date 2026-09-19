package model

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

var _ LikeRecordModel = (*customLikeRecordModel)(nil)

type (
	// LikeRecordModel is the interface for like_record operations in MQ
	LikeRecordModel interface {
		Insert(ctx context.Context, data *LikeRecord) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*LikeRecord, error)
		FindOneByBizIdObjIdUserId(ctx context.Context, bizId string, objId int64, userId int64) (*LikeRecord, error)
		Update(ctx context.Context, data *LikeRecord) error
		Delete(ctx context.Context, id int64) error
	}

	customLikeRecordModel struct {
		db *gorm.DB
	}

	LikeRecord struct {
		Id         int64     `gorm:"primaryKey;column:id"` // 主键ID
		BizId      string    `gorm:"column:biz_id"`        // 业务ID
		ObjId      int64     `gorm:"column:obj_id"`        // 点赞对象id
		UserId     int64     `gorm:"column:user_id"`       // 用户ID
		LikeType   int64     `gorm:"column:like_type"`     // 类型 0:点赞 1:点踩
		CreateTime time.Time `gorm:"column:create_time;autoCreateTime"` // 创建时间
		UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime"` // 最后修改时间
	}
)

func (LikeRecord) TableName() string {
	return "like_record"
}


// NewLikeRecordModel returns a model for the database table.
func NewLikeRecordModel(db *gorm.DB) LikeRecordModel {
	return &customLikeRecordModel{
		db: db,
	}
}

func (m *customLikeRecordModel) Insert(ctx context.Context, data *LikeRecord) (sql.Result, error) {
	err := m.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return nil, err
	}
	return sqlResult{id: data.Id, affected: 1}, nil
}

func (m *customLikeRecordModel) FindOne(ctx context.Context, id int64) (*LikeRecord, error) {
	var resp LikeRecord
	err := m.db.WithContext(ctx).First(&resp, id).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customLikeRecordModel) FindOneByBizIdObjIdUserId(ctx context.Context, bizId string, objId int64, userId int64) (*LikeRecord, error) {
	var resp LikeRecord
	err := m.db.WithContext(ctx).Where("biz_id = ? AND obj_id = ? AND user_id = ?", bizId, objId, userId).First(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customLikeRecordModel) Update(ctx context.Context, data *LikeRecord) error {
	return m.db.WithContext(ctx).Save(data).Error
}

func (m *customLikeRecordModel) Delete(ctx context.Context, id int64) error {
	return m.db.WithContext(ctx).Delete(&LikeRecord{}, id).Error
}
