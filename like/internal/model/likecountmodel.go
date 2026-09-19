package model

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

var _ LikeCountModel = (*customLikeCountModel)(nil)

type (
	// LikeCountModel is the interface for like_count operations in MQ
	LikeCountModel interface {
		Insert(ctx context.Context, data *LikeCount) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*LikeCount, error)
		FindOneByBizIdObjId(ctx context.Context, bizId string, objId int64) (*LikeCount, error)
		Update(ctx context.Context, data *LikeCount) error
		Delete(ctx context.Context, id int64) error
	}

	customLikeCountModel struct {
		db *gorm.DB
	}

	LikeCount struct {
		Id         int64     `gorm:"primaryKey;column:id"` // 主键ID
		BizId      string    `gorm:"column:biz_id"`        // 业务ID
		ObjId      int64     `gorm:"column:obj_id"`        // 点赞对象id
		LikeNum    int64     `gorm:"column:like_num"`      // 点赞数
		DislikeNum int64     `gorm:"column:dislike_num"`   // 点踩数
		CreateTime time.Time `gorm:"column:create_time;autoCreateTime"` // 创建时间
		UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime"` // 最后修改时间
	}
)

func (LikeCount) TableName() string {
	return "like_count"
}

// sqlResult implements sql.Result
type sqlResult struct {
	id       int64
	affected int64
}

func (r sqlResult) LastInsertId() (int64, error) { return r.id, nil }
func (r sqlResult) RowsAffected() (int64, error) { return r.affected, nil }

// NewLikeCountModel returns a model for the database table.
func NewLikeCountModel(db *gorm.DB) LikeCountModel {
	return &customLikeCountModel{
		db: db,
	}
}

func (m *customLikeCountModel) Insert(ctx context.Context, data *LikeCount) (sql.Result, error) {
	err := m.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return nil, err
	}
	return sqlResult{id: data.Id, affected: 1}, nil
}

func (m *customLikeCountModel) FindOne(ctx context.Context, id int64) (*LikeCount, error) {
	var resp LikeCount
	err := m.db.WithContext(ctx).First(&resp, id).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customLikeCountModel) FindOneByBizIdObjId(ctx context.Context, bizId string, objId int64) (*LikeCount, error) {
	var resp LikeCount
	err := m.db.WithContext(ctx).Where("biz_id = ? AND obj_id = ?", bizId, objId).First(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customLikeCountModel) Update(ctx context.Context, data *LikeCount) error {
	return m.db.WithContext(ctx).Save(data).Error
}

func (m *customLikeCountModel) Delete(ctx context.Context, id int64) error {
	return m.db.WithContext(ctx).Delete(&LikeCount{}, id).Error
}
