package svc

import (
	"mq-thinktalk/concerned/internal/config"
	"mq-thinktalk/concerned/internal/model"
	"mq-thinktalk/pkg/orm"

	"github.com/zeromicro/go-queue/kq"
)

type ServiceContext struct {
	Config               config.Config
	DB                   *orm.DB
	ConcernedRecordModel *model.ConcernedRecordModel
	ConcernedCountModel  *model.ConcernedCountModel
	NotificationPusher   *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:               c,
		DB:                   db,
		ConcernedRecordModel: model.NewConcernedRecordModel(db.DB),
		ConcernedCountModel:  model.NewConcernedCountModel(db.DB),
		NotificationPusher:   kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
