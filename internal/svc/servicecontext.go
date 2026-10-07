package svc

import (
	"mq-thinktalk/client/user/user"
	"mq-thinktalk/internal/config"
	articlemodel "mq-thinktalk/internal/model/article"
	chatmodel "mq-thinktalk/internal/model/chat"
	concernedmodel "mq-thinktalk/internal/model/concerned"
	likemodel "mq-thinktalk/internal/model/like"
	membermodel "mq-thinktalk/internal/model/member"
	messagemodel "mq-thinktalk/internal/model/message"
	qamodel "mq-thinktalk/internal/model/qa"
	replymodel "mq-thinktalk/internal/model/reply"
	"mq-thinktalk/pkg/es"
	"mq-thinktalk/pkg/orm"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *orm.DB
	BizRedis           *redis.Redis
	MinIO              *minio.Client
	NotificationPusher *kq.Pusher
	UserRPC            user.User
	Es                 *es.Es

	ArticleModel         articlemodel.ArticleModel
	ConversationModel    *chatmodel.ConversationModel
	MessageModel         *chatmodel.MessageModel
	ConcernedRecordModel *concernedmodel.ConcernedRecordModel
	ConcernedCountModel  *concernedmodel.ConcernedCountModel
	LikeRecordModel      likemodel.LikeRecordModel
	LikeCountModel       likemodel.LikeCountModel
	MemberModel          *membermodel.MemberModel
	MemberOrderModel     *membermodel.MemberOrderModel
	NotificationModel    *messagemodel.NotificationModel
	QuestionModel        *qamodel.QuestionModel
	AnswerModel          *qamodel.AnswerModel
	ReplyModel           *replymodel.ReplyModel
	ReplyCountModel      *replymodel.ReplyCountModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewPostgres(&orm.Config{
		DSN:          c.DB.DataSource,
		MaxOpenConns: c.DB.MaxOpenConns,
		MaxIdleConns: c.DB.MaxIdleConns,
		MaxLifetime:  c.DB.MaxLifetime,
	})

	rds := redis.MustNewRedis(c.BizRedis)

	sc := &ServiceContext{
		Config:             c,
		DB:                 db,
		BizRedis:           rds,
		NotificationPusher: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),

		ArticleModel:         articlemodel.NewArticleModel(db.DB),
		ConversationModel:    chatmodel.NewConversationModel(db.DB),
		MessageModel:         chatmodel.NewMessageModel(db.DB),
		ConcernedRecordModel: concernedmodel.NewConcernedRecordModel(db.DB),
		ConcernedCountModel:  concernedmodel.NewConcernedCountModel(db.DB),
		LikeRecordModel:      likemodel.NewLikeRecordModel(db.DB),
		LikeCountModel:       likemodel.NewLikeCountModel(db.DB),
		MemberModel:          membermodel.NewMemberModel(db.DB),
		MemberOrderModel:     membermodel.NewMemberOrderModel(db.DB),
		NotificationModel:    messagemodel.NewNotificationModel(db.DB),
		QuestionModel:        qamodel.NewQuestionModel(db.DB),
		AnswerModel:          qamodel.NewAnswerModel(db.DB),
		ReplyModel:           replymodel.NewReplyModel(db.DB),
		ReplyCountModel:      replymodel.NewReplyCountModel(db.DB),
	}

	if len(c.UserRPC.Endpoints) > 0 || c.UserRPC.Target != "" {
		sc.UserRPC = user.NewUser(zrpc.MustNewClient(c.UserRPC))
	}

	if len(c.Es.Addresses) > 0 {
		sc.Es = es.MustNewEs(&es.Config{
			Addresses: c.Es.Addresses,
			Username:  c.Es.Username,
			Password:  c.Es.Password,
		})
	}

	if c.MinIO.Endpoint != "" {
		minioClient, err := minio.New(c.MinIO.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.MinIO.AccessKeyID, c.MinIO.AccessKeySecret, ""),
			Secure: c.MinIO.UseSSL,
		})
		if err == nil {
			sc.MinIO = minioClient
		}
	}

	return sc
}
