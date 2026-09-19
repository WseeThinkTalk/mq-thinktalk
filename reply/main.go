package replymq

import (
	"context"

	"mq-thinktalk/reply/internal/config"
	"mq-thinktalk/reply/internal/logic"
	"mq-thinktalk/reply/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
