package svc

import (
	"mq-thinktalk/reply/internal/config"
	"mq-thinktalk/reply/internal/model"
	"mq-thinktalk/pkg/orm"

	"github.com/zeromicro/go-queue/kq"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *orm.DB
	ReplyModel         *model.ReplyModel
	ReplyCountModel    *model.ReplyCountModel
	NotificationPusher *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:             c,
		DB:                 db,
		ReplyModel:         model.NewReplyModel(db.DB),
		ReplyCountModel:    model.NewReplyCountModel(db.DB),
		NotificationPusher: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
