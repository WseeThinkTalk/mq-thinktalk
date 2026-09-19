package chatmq

import (
	"context"

	"mq-thinktalk/chat/internal/config"
	"mq-thinktalk/chat/internal/logic"
	"mq-thinktalk/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
