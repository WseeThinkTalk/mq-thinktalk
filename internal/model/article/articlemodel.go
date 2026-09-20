package model

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

var _ ArticleModel = (*customArticleModel)(nil)

type (
	// ArticleModel is the interface for article operations in MQ
	ArticleModel interface {
		Insert(ctx context.Context, data *Article) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*Article, error)
		Update(ctx context.Context, data *Article) error
		Delete(ctx context.Context, id int64) error
		UpdateLikeNum(ctx context.Context, id, likeNum int64) error
	}

	customArticleModel struct {
		db *gorm.DB
	}

	Article struct {
		Id          int64     `gorm:"primaryKey;column:id"` // 主键ID
		Title       string    `gorm:"column:title"`         // 标题
		Content     string    `gorm:"column:content"`       // 内容
		Cover       string    `gorm:"column:cover"`         // 封面
		Description string    `gorm:"column:description"`   // 描述
		AuthorId    int64     `gorm:"column:author_id"`     // 作者ID
		Status      int64     `gorm:"column:status"`        // 状态 0:待审核 1:审核不通过 2:可见 3:用户删除
		CommentNum  int64     `gorm:"column:comment_num"`   // 评论数
		LikeNum     int64     `gorm:"column:like_num"`      // 点赞数
		CollectNum  int64     `gorm:"column:collect_num"`   // 收藏数
		ViewNum     int64     `gorm:"column:view_num"`      // 浏览数
		ShareNum    int64     `gorm:"column:share_num"`     // 分享数
		TagIds      string    `gorm:"column:tag_ids"`       // 标签ID
		PublishTime time.Time `gorm:"column:publish_time"`  // 发布时间
		CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"` // 创建时间
		UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"` // 最后修改时间
	}
)

func (Article) TableName() string {
	return "article"
}

// sqlResult implements sql.Result
type sqlResult struct {
	id       int64
	affected int64
}

func (r sqlResult) LastInsertId() (int64, error) { return r.id, nil }
func (r sqlResult) RowsAffected() (int64, error) { return r.affected, nil }

// NewArticleModel returns a model for the database table.
func NewArticleModel(db *gorm.DB) ArticleModel {
	return &customArticleModel{
		db: db,
	}
}

func (m *customArticleModel) Insert(ctx context.Context, data *Article) (sql.Result, error) {
	err := m.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return nil, err
	}
	return sqlResult{id: data.Id, affected: 1}, nil
}

func (m *customArticleModel) FindOne(ctx context.Context, id int64) (*Article, error) {
	var resp Article
	err := m.db.WithContext(ctx).First(&resp, id).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customArticleModel) Update(ctx context.Context, data *Article) error {
	return m.db.WithContext(ctx).Save(data).Error
}

func (m *customArticleModel) Delete(ctx context.Context, id int64) error {
	return m.db.WithContext(ctx).Delete(&Article{}, id).Error
}

func (m *customArticleModel) UpdateLikeNum(ctx context.Context, id, likeNum int64) error {
	return m.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).Update("like_num", likeNum).Error
}
