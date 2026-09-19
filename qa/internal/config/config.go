package config

import "github.com/zeromicro/go-queue/kq"

type Config struct {
	KqConsumerConf kq.KqConf
	Mysql          struct {
		DataSource string
	}
	CacheRedis []struct {
		Host string
		Type string
		Pass string
	}
	Es struct {
		Addresses []string
		Username  string
		Password  string
	}
}
