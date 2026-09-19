package articlemq

import (
	"context"

	"mq-thinktalk/article/internal/config"
	"mq-thinktalk/article/internal/logic"
	"mq-thinktalk/article/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
