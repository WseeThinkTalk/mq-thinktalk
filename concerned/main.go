package concernedmq

import (
	"context"

	"mq-thinktalk/concerned/internal/config"
	"mq-thinktalk/concerned/internal/logic"
	"mq-thinktalk/concerned/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
