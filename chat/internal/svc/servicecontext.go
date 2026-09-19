package svc

import (
	"mq-thinktalk/chat/internal/config"
	"mq-thinktalk/chat/internal/model"
	"mq-thinktalk/pkg/orm"
)

type ServiceContext struct {
	Config            config.Config
	DB                *orm.DB
	ConversationModel *model.ConversationModel
	MessageModel      *model.MessageModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:            c,
		DB:                db,
		ConversationModel: model.NewConversationModel(db.DB),
		MessageModel:      model.NewMessageModel(db.DB),
	}
}
