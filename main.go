package main

import (
	"context"
	"flag"

	"mq-thinktalk/internal/config"
	"mq-thinktalk/internal/logic/article"
	"mq-thinktalk/internal/logic/chat"
	"mq-thinktalk/internal/logic/concerned"
	"mq-thinktalk/internal/logic/like"
	"mq-thinktalk/internal/logic/member"
	"mq-thinktalk/internal/logic/message"
	"mq-thinktalk/internal/logic/qa"
	"mq-thinktalk/internal/logic/reply"
	"mq-thinktalk/internal/svc"
	"mq-thinktalk/pkg/env"
	"mq-thinktalk/pkg/lib/zapx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/mq.yaml", "the config file")

func main() {
	flag.Parse()

	env.LoadEnv()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	err := c.ServiceConf.SetUp()
	if err != nil {
		panic(err)
	}

	// init logger
	writer, err := zapx.NewZapWriter()
	if err == nil {
		logx.SetWriter(writer)
	}

	logx.DisableStat()
	ctx := context.Background()
	svcCtx := svc.NewServiceContext(c)

	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 1. Article Consumers
	for _, s := range article.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 2. Chat Consumer
	for _, s := range chat.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 3. Concerned Consumer
	for _, s := range concerned.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 4. Like Consumer
	for _, s := range like.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 5. Member Consumer
	for _, s := range member.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 6. Message Consumer
	for _, s := range message.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 7. QA Consumer
	for _, s := range qa.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	// 8. Reply Consumer
	for _, s := range reply.Consumers(ctx, svcCtx) {
		serviceGroup.Add(s)
	}

	logx.Info("All 8 MQ consumers registered with unified ServiceContext. Starting thinktalk-mq service group...")
	serviceGroup.Start()
}
