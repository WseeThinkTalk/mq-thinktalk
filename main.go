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
	"mq-thinktalk/internal/logic/video"
	"mq-thinktalk/internal/svc"
	"mq-thinktalk/pkg/lib/etcdx"
	"mq-thinktalk/pkg/lib/zapx"
	"mq-thinktalk/pkg/reconcile"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

func runRemoteConfig() *config.Config {
	var c config.Config
	etcdx.MustLoadRemoteConfig("/thinktalk/config/mq", &c)
	err := c.ServiceConf.SetUp()
	if err != nil {
		panic(err)
	}
	return &c
}

func main() {
	flag.Parse()

	// 从 Etcd 配置中心拉取远程配置 (Fail-Fast)
	c := runRemoteConfig()
	if c == nil {
		return
	}

	// init logger
	writer, err := zapx.NewZapWriter()
	if err == nil {
		logx.SetWriter(writer)
	}

	logx.DisableStat()
	ctx := context.Background()
	svcCtx := svc.NewServiceContext(*c)

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

	// 9. Video Consumer - 注册视频流媒体抽帧处理消费者
	for _, v := range video.Consumers(ctx, svcCtx) {
		serviceGroup.Add(v)
	}

	// 10. Reconciliation Service - 注册离线数据对账与自动补偿治理服务
	reconcileSvc := reconcile.NewReconciliationService(30 * time.Minute)
	serviceGroup.Add(reconcileSvc)

	logx.Info("All MQ consumers registered with unified ServiceContext. Starting thinktalk-mq service group...")
	serviceGroup.Start()
}
