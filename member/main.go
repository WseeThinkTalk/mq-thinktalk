package membermq

import (
	"context"

	"mq-thinktalk/member/internal/config"
	"mq-thinktalk/member/internal/logic"
	"mq-thinktalk/member/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type Config = config.Config

func Consumers(ctx context.Context, c Config) []service.Service {
	svcCtx := svc.NewServiceContext(c)
	return logic.Consumers(ctx, svcCtx)
}
