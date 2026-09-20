package like

import (
	"context"
	"encoding/json"

	model "mq-thinktalk/internal/model/like"
	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/like"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ThumbupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewThumbupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThumbupLogic {
	return &ThumbupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ThumbupLogic) Consume(ctx context.Context, key, val string) error {
	l.Infof("[Thumbup] consume key: %s val: %s", key, val)

	var msg types.ThumbupMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Thumbup] unmarshal msg error: %v", err)
		return err
	}

	record, err := l.svcCtx.LikeRecordModel.FindOneByBizIdObjIdUserId(ctx, msg.BizId, msg.ObjId, msg.UserId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("[Thumbup] find like record error: %v", err)
		return err
	}

	if msg.LikeType == 0 {
		if record != nil {
			return l.cancelLike(ctx, record)
		}
		return nil
	}

	if record != nil {
		if record.LikeType == int64(msg.LikeType) {
			return nil
		}
		return l.switchLike(ctx, record, msg.LikeType)
	}
	return l.addLike(ctx, msg)
}

func (l *ThumbupLogic) getOrCreateCount(ctx context.Context, bizId string, objId int64) (*model.LikeCount, error) {
	count, err := l.svcCtx.LikeCountModel.FindOneByBizIdObjId(ctx, bizId, objId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("[Thumbup] find like count error: %v", err)
		return nil, err
	}
	if count == nil {
		return &model.LikeCount{BizId: bizId, ObjId: objId}, nil
	}
	return count, nil
}

func (l *ThumbupLogic) updateCount(ctx context.Context, count *model.LikeCount) error {
	if err := l.svcCtx.LikeCountModel.Update(ctx, count); err != nil {
		l.Errorf("[Thumbup] update like count error: %v", err)
		return err
	}

	// 同步到目标表
	var err error
	if count.BizId == "article" {
		err = l.svcCtx.DB.DB.WithContext(ctx).Table("article").Where("id = ?", count.ObjId).Update("like_num", count.LikeNum).Error
	} else if count.BizId == "reply" {
		err = l.svcCtx.DB.DB.WithContext(ctx).Table("reply").Where("reply_id = ?", count.ObjId).Update("like_num", count.LikeNum).Error
	}
	if err != nil {
		l.Errorf("[Thumbup] sync like_num to %s id %d error: %v", count.BizId, count.ObjId, err)
	}

	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.LikeKq, NewThumbupLogic(ctx, svcCtx)),
	}
}
