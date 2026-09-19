package svc

import (
	"mq-thinktalk/article/internal/config"
	"mq-thinktalk/article/internal/model"
	"mq-thinktalk/client/user/user"
	"mq-thinktalk/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	ArticleModel model.ArticleModel
	DB           *orm.DB
	BizRedis     *redis.Redis
	UserRPC      user.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds, err := redis.NewRedis(c.BizRedis)
	if err != nil {
		panic(err)
	}

	db := orm.MustNewPostgres(&orm.Config{
		DSN:          c.Datasource,
		MaxOpenConns: 10,
		MaxIdleConns: 100,
		MaxLifetime:  3600,
	})
	return &ServiceContext{
		Config:       c,
		ArticleModel: model.NewArticleModel(db.DB),
		DB:           db,
		BizRedis:     rds,
		UserRPC:      user.NewUser(zrpc.MustNewClient(c.UserRPC)),
	}
}
