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

	// 1. Article Consumers - 注册文章消息消费者
	for _, v := range article.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 2. Chat Consumer - 注册聊天消息消费者
	for _, v := range chat.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 3. Concerned Consumer - 注册关注消息消费者
	for _, v := range concerned.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 4. Like Consumer - 注册点赞消息消费者
	for _, v := range like.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 5. Member Consumer - 注册会员消息消费者
	for _, v := range member.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 6. Message Consumer - 注册通知消息消费者
	for _, v := range message.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 7. QA Consumer - 注册问答消息消费者
	for _, v := range qa.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 8. Reply Consumer - 注册评论消息消费者
	for _, v := range reply.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	logx.Info("All 8 MQ consumers registered with unified ServiceContext. Starting thinktalk-mq service group...")
	serviceGroup.Start()
}
