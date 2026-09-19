// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Config struct {
	KqConsumerConf kq.KqConf
	Mysql          sqlx.SqlConf
	CacheRedis     cache.CacheConf
	KqPusherConf   struct {
		Brokers []string
		Topic   string
	}
}
