package svc

import (
	"mq-thinktalk/like/internal/config"
	"mq-thinktalk/like/internal/model"
	"mq-thinktalk/pkg/orm"

	"github.com/zeromicro/go-queue/kq"
)

type ServiceContext struct {
	Config             config.Config
	LikeRecordModel    model.LikeRecordModel
	LikeCountModel     model.LikeCountModel
	DB                 *orm.DB
	NotificationPusher *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewPostgres(&orm.Config{
		DSN:          c.Mysql.DataSource,
		MaxOpenConns: 10,
		MaxIdleConns: 100,
		MaxLifetime:  3600,
	})
	return &ServiceContext{
		Config:             c,
		LikeRecordModel:    model.NewLikeRecordModel(db.DB),
		LikeCountModel:     model.NewLikeCountModel(db.DB),
		DB:                 db,
		NotificationPusher: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
