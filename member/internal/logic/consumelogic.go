package logic

import (
	"context"
	"encoding/json"
	"time"

	"mq-thinktalk/member/internal/model"
	"mq-thinktalk/member/internal/svc"
	"mq-thinktalk/member/internal/types"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"gorm.io/gorm"
)

type ConsumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeLogic {
	return &ConsumeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ConsumeLogic) Consume(ctx context.Context, key, val string) error {
	l.Infof("[Consume] key: %s val: %s", key, val)

	var msg types.MemberMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Consume] unmarshal err: %v", err)
		return err
	}

	now := time.Now()

	existing, err := l.svcCtx.MemberOrderModel.FindByTransactionId(ctx, msg.TransactionId)
	if err != nil {
		l.Errorf("[Consume] FindByTransactionId err: %v txId: %s", err, msg.TransactionId)
		return err
	}
	if existing != nil {
		l.Infof("[Consume] duplicate transaction ignored: %s", msg.TransactionId)
		return nil
	}

	expireTime := now.Add(time.Duration(msg.DurationDays) * 24 * time.Hour)

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		order := &model.MemberOrder{
			UserID:        msg.UserId,
			Level:         msg.Level,
			DurationDays:  msg.DurationDays,
			Amount:        msg.Amount,
			PayChannel:    msg.PayChannel,
			TransactionID: msg.TransactionId,
			Status:        1,
			CreateTime:    now,
			UpdateTime:    now,
		}
		if err := model.NewMemberOrderModel(tx).Insert(ctx, order); err != nil {
			return err
		}

		member := &model.Member{
			UserID:     msg.UserId,
			Level:      msg.Level,
			ExpireTime: expireTime,
			Status:     1,
			CreateTime: now,
			UpdateTime: now,
		}
		return model.NewMemberModel(tx).UpsertMember(ctx, member)
	})
	if err != nil {
		l.Errorf("[Consume] transaction err: %v msg: %+v", err, msg)
		return err
	}

	l.Infof("[Consume] member upgraded userId: %d level: %d", msg.UserId, msg.Level)
	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.KqConsumerConf, NewConsumeLogic(ctx, svcCtx)),
	}
}
