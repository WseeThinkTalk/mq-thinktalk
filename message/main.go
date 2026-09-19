package messagemq

import (
	"context"

	"mq-thinktalk/message/internal/config"
	"mq-thinktalk/message/internal/logic"
	"mq-thinktalk/message/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
