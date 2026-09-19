package svc

import (
	"mq-thinktalk/qa/internal/config"
	"mq-thinktalk/qa/internal/model"
	"mq-thinktalk/pkg/es"
	"mq-thinktalk/pkg/orm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *orm.DB
	QuestionModel *model.QuestionModel
	AnswerModel   *model.AnswerModel
	Es            *es.Es
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN: c.Mysql.DataSource,
	})

	return &ServiceContext{
		Config:        c,
		DB:            db,
		QuestionModel: model.NewQuestionModel(db.DB),
		AnswerModel:   model.NewAnswerModel(db.DB),
		Es: es.MustNewEs(&es.Config{
			Addresses: c.Es.Addresses,
			Username:  c.Es.Username,
			Password:  c.Es.Password,
		}),
	}
}
