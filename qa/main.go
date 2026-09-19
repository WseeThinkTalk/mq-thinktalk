package qamq

import (
	"context"

	"mq-thinktalk/qa/internal/config"
	"mq-thinktalk/qa/internal/logic"
	"mq-thinktalk/qa/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
