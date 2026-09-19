package svc

import (
	"mq-thinktalk/member/internal/config"
	"mq-thinktalk/member/internal/model"
	"mq-thinktalk/pkg/orm"
)

type ServiceContext struct {
	Config           config.Config
	DB               *orm.DB
	MemberModel      *model.MemberModel
	MemberOrderModel *model.MemberOrderModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:           c,
		DB:               db,
		MemberModel:      model.NewMemberModel(db.DB),
		MemberOrderModel: model.NewMemberOrderModel(db.DB),
	}
}
