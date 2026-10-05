package like

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	model "mq-thinktalk/internal/model/like"
	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/like"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"gorm.io/gorm"
)

var (
	flusher     *BatchFlusher
	flusherOnce sync.Once
)

func getBatchFlusher(svcCtx *svc.ServiceContext) *BatchFlusher {
	flusherOnce.Do(func() {
		flusher = NewBatchFlusher(BatchFlusherConfig{
			MaxSize:  200,
			Interval: 1 * time.Second,
			FlushFn: func(items []*LikeDelta) error {
				return flushLikeDeltas(context.Background(), svcCtx, items)
			},
		})
		flusher.Start()
	})
	return flusher
}

func flushLikeDeltas(ctx context.Context, svcCtx *svc.ServiceContext, deltas []*LikeDelta) error {
	for _, d := range deltas {
		if d == nil {
			continue
		}

		count, err := svcCtx.LikeCountModel.FindOneByBizIdObjId(ctx, d.BizId, d.ObjId)
		if err != nil && err != model.ErrNotFound {
			logx.WithContext(ctx).Errorf("[flushLikeDeltas] find like count error: %v", err)
			continue
		}
		if count == nil {
			count = &model.LikeCount{
				BizId:      d.BizId,
				ObjId:      d.ObjId,
				LikeNum:    d.LikeDelta,
				DislikeNum: d.DislikeDelta,
			}
			if count.LikeNum < 0 {
				count.LikeNum = 0
			}
			if count.DislikeNum < 0 {
				count.DislikeNum = 0
			}
			_, _ = svcCtx.LikeCountModel.Insert(ctx, count)
		} else {
			count.LikeNum += d.LikeDelta
			count.DislikeNum += d.DislikeDelta
			if count.LikeNum < 0 {
				count.LikeNum = 0
			}
			if count.DislikeNum < 0 {
				count.DislikeNum = 0
			}
			_ = svcCtx.LikeCountModel.Update(ctx, count)
		}

		// 同步到目标业务表
		if d.LikeDelta != 0 {
			if d.BizId == "article" {
				_ = svcCtx.DB.DB.WithContext(ctx).Table("article").
					Where("id = ?", d.ObjId).
					Update("like_num", gorm.Expr("GREATEST(0, CAST(like_num AS SIGNED) + ?)", d.LikeDelta)).Error
			} else if d.BizId == "reply" {
				_ = svcCtx.DB.DB.WithContext(ctx).Table("reply").
					Where("reply_id = ?", d.ObjId).
					Update("like_num", gorm.Expr("GREATEST(0, CAST(like_num AS SIGNED) + ?)", d.LikeDelta)).Error
			}
		}
	}
	return nil
}

type ThumbupLogic struct {
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	flusher *BatchFlusher
	logx.Logger
}

func NewThumbupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ThumbupLogic {
	return &ThumbupLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		flusher: getBatchFlusher(svcCtx),
		Logger:  logx.WithContext(ctx),
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

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.LikeKq, NewThumbupLogic(ctx, svcCtx)),
	}
}
