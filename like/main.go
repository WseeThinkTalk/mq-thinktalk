package likemq

import (
	"context"

	"mq-thinktalk/like/internal/config"
	"mq-thinktalk/like/internal/logic"
	"mq-thinktalk/like/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
