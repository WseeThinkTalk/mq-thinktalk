package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	DB         struct {
		DataSource   string
		MaxOpenConns int `json:",default=20"`
		MaxIdleConns int `json:",default=50"`
		MaxLifetime  int `json:",default=3600"`
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
	VideoProcessKq kq.KqConf `json:",optional"`

	MinIO struct {
		Endpoint        string
		AccessKeyID     string
		AccessKeySecret string
		BucketName      string
		Location        string
		UseSSL          bool
	} `json:",optional"`
}
