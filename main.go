package main

import (
	"context"
	"flag"

	articlemq "mq-thinktalk/article"
	chatmq "mq-thinktalk/chat"
	concernedmq "mq-thinktalk/concerned"
	likemq "mq-thinktalk/like"
	membermq "mq-thinktalk/member"
	messagemq "mq-thinktalk/message"
	qamq "mq-thinktalk/qa"
	replymq "mq-thinktalk/reply"
	"mq-thinktalk/pkg/env"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/mq.yaml", "the config file")

type Config struct {
	service.ServiceConf

	DataSource string
	Mysql      struct {
		DataSource string
	}
	BizRedis   redis.RedisConf
	CacheRedis cache.CacheConf
	UserRPC    zrpc.RpcClientConf
	Es         struct {
		Addresses []string
		Username  string
		Password  string
	}
	KqPusherConf struct {
		Brokers []string
		Topic   string
	}

	ArticleKq      kq.KqConf
	ArticleEventKq kq.KqConf
	ChatKq         kq.KqConf
	ConcernedKq    kq.KqConf
	LikeKq         kq.KqConf
	MemberKq       kq.KqConf
	MessageKq      kq.KqConf
	QaKq           kq.KqConf
	ReplyKq        kq.KqConf
}

func main() {
	flag.Parse()

	env.LoadEnv()

	var c Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	err := c.ServiceConf.SetUp()
	if err != nil {
		panic(err)
	}

	logx.DisableStat()
	ctx := context.Background()
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 1. Article Consumers
	for _, s := range articlemq.Consumers(ctx, articlemq.Config{
		ServiceConf:           c.ServiceConf,
		KqConsumerConf:        c.ArticleKq,
		ArticleKqConsumerConf: c.ArticleEventKq,
		Datasource:            c.DataSource,
		BizRedis:              c.BizRedis,
		UserRPC:               c.UserRPC,
	}) {
		serviceGroup.Add(s)
	}

	// 2. Chat Consumer
	for _, s := range chatmq.Consumers(ctx, chatmq.Config{
		KqConsumerConf: c.ChatKq,
		Mysql:          c.Mysql,
		CacheRedis: []struct {
			Host string
			Type string
			Pass string
		}{
			{
				Host: c.BizRedis.Host,
				Type: c.BizRedis.Type,
				Pass: c.BizRedis.Pass,
			},
		},
	}) {
		serviceGroup.Add(s)
	}

	// 3. Concerned Consumer
	for _, s := range concernedmq.Consumers(ctx, concernedmq.Config{
		KqConsumerConf: c.ConcernedKq,
		Mysql:          c.Mysql,
		CacheRedis: []struct {
			Host string
			Type string
			Pass string
		}{
			{
				Host: c.BizRedis.Host,
				Type: c.BizRedis.Type,
				Pass: c.BizRedis.Pass,
			},
		},
		KqPusherConf: c.KqPusherConf,
	}) {
		serviceGroup.Add(s)
	}

	// 4. Like Consumer
	for _, s := range likemq.Consumers(ctx, likemq.Config{
		KqConsumerConf: c.LikeKq,
		Mysql: sqlx.SqlConf{
			DataSource: c.Mysql.DataSource,
		},
		CacheRedis:   c.CacheRedis,
		KqPusherConf: c.KqPusherConf,
	}) {
		serviceGroup.Add(s)
	}

	// 5. Member Consumer
	for _, s := range membermq.Consumers(ctx, membermq.Config{
		KqConsumerConf: c.MemberKq,
		Mysql:          c.Mysql,
	}) {
		serviceGroup.Add(s)
	}

	// 6. Message Consumer
	for _, s := range messagemq.Consumers(ctx, messagemq.Config{
		KqConsumerConf: c.MessageKq,
		Mysql:          c.Mysql,
		CacheRedis: []struct {
			Host string
			Type string
			Pass string
		}{
			{
				Host: c.BizRedis.Host,
				Type: c.BizRedis.Type,
				Pass: c.BizRedis.Pass,
			},
		},
	}) {
		serviceGroup.Add(s)
	}

	// 7. QA Consumer
	for _, s := range qamq.Consumers(ctx, qamq.Config{
		KqConsumerConf: c.QaKq,
		Mysql:          c.Mysql,
		CacheRedis: []struct {
			Host string
			Type string
			Pass string
		}{
			{
				Host: c.BizRedis.Host,
				Type: c.BizRedis.Type,
				Pass: c.BizRedis.Pass,
			},
		},
		Es: c.Es,
	}) {
		serviceGroup.Add(s)
	}

	// 8. Reply Consumer
	for _, s := range replymq.Consumers(ctx, replymq.Config{
		KqConsumerConf: c.ReplyKq,
		Mysql:          c.Mysql,
		CacheRedis: []struct {
			Host string
			Type string
			Pass string
		}{
			{
				Host: c.BizRedis.Host,
				Type: c.BizRedis.Type,
				Pass: c.BizRedis.Pass,
			},
		},
		KqPusherConf: c.KqPusherConf,
	}) {
		serviceGroup.Add(s)
	}

	logx.Info("All 8 MQ consumers registered. Starting thinktalk-mq service group...")
	serviceGroup.Start()
}
