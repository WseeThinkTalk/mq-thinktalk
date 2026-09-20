package message

import (
	"context"
	"encoding/json"
	"time"

	model "mq-thinktalk/internal/model/message"
	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/message"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ConsumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeLogic {
	return &ConsumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeLogic) Consume(ctx context.Context, key, val string) error {
	l.Infof("[Consume] consume key: %s val: %s", key, val)

	var msg types.NotificationMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Consume] unmarshal msg error: %v", err)
		return err
	}

	notif := &model.Notification{
		UserID:        msg.UserId,
		Type:          msg.Type,
		Title:         msg.Title,
		Content:       msg.Content,
		RefID:         msg.RefId,
		BizID:         msg.BizId,
		TriggerUserID: msg.TriggerUserId,
		IsRead:        0,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}

	if err := l.svcCtx.NotificationModel.Insert(ctx, notif); err != nil {
		l.Errorf("[Consume] Insert notification err: %v msg: %+v", err, msg)
		return err
	}

	l.Infof("[Consume] notification created id: %d userId: %d type: %d", notif.ID, msg.UserId, msg.Type)
	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.MessageKq, NewConsumeLogic(ctx, svcCtx)),
	}
}
