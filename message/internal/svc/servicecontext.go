package svc

import (
	"mq-thinktalk/message/internal/config"
	"mq-thinktalk/message/internal/model"
	"mq-thinktalk/pkg/orm"
)

type ServiceContext struct {
	Config            config.Config
	DB                *orm.DB
	NotificationModel *model.NotificationModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:            c,
		DB:                db,
		NotificationModel: model.NewNotificationModel(db.DB),
	}
}
