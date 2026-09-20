package concerned

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	model "mq-thinktalk/internal/model/concerned"
	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/concerned"

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
	return &ConsumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeLogic) Consume(ctx context.Context, key, val string) error {
	l.Infof("[Consume] consume key: %s val: %s", key, val)

	var msg types.ConcernedMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Consume] unmarshal msg error: %v", err)
		return err
	}

	switch msg.OpType {
	case types.OpTypeAdd:
		return l.addConcerned(ctx, &msg)
	case types.OpTypeCancel:
		return l.cancelConcerned(ctx, &msg)
	default:
		l.Errorf("[Consume] unknown opType: %d", msg.OpType)
		return nil
	}
}

func (l *ConsumeLogic) addConcerned(ctx context.Context, msg *types.ConcernedMsg) error {
	existing, err := l.svcCtx.ConcernedRecordModel.FindByBizIDObjIDUserID(ctx, msg.BizId, msg.ObjId, msg.UserId)
	if err != nil {
		l.Errorf("[addConcerned] find existing err: %v msg: %+v", err, msg)
		return err
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if existing != nil {
			if existing.Status == 0 {
				return nil
			}
			err := model.NewConcernedRecordModel(tx).UpdateFields(ctx, existing.ID, map[string]interface{}{
				"status": 0,
			})
			if err != nil {
				return err
			}
		} else {
			err := model.NewConcernedRecordModel(tx).Insert(ctx, &model.ConcernedRecord{
				BizID:      msg.BizId,
				ObjID:      msg.ObjId,
				UserID:     msg.UserId,
				Status:     0,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			})
			if err != nil {
				return err
			}
		}
		err = model.NewConcernedCountModel(tx).IncrConcernedNum(ctx, msg.BizId, msg.ObjId)
		if err != nil {
			return err
		}
		return l.syncCollectNum(ctx, tx, msg.BizId, msg.ObjId)
	})
	if err != nil {
		l.Errorf("[addConcerned] transaction err: %v msg: %+v", err, msg)
		return err
	}

	l.Infof("[addConcerned] success bizId: %s objId: %d userId: %d", msg.BizId, msg.ObjId, msg.UserId)

	// 发送收藏通知
	l.sendCollectNotification(ctx, msg, "收藏了你的笔记")

	return nil
}

func (l *ConsumeLogic) cancelConcerned(ctx context.Context, msg *types.ConcernedMsg) error {
	existing, err := l.svcCtx.ConcernedRecordModel.FindByBizIDObjIDUserID(ctx, msg.BizId, msg.ObjId, msg.UserId)
	if err != nil {
		l.Errorf("[cancelConcerned] find existing err: %v msg: %+v", err, msg)
		return err
	}
	if existing == nil || existing.Status == 1 {
		return nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.NewConcernedRecordModel(tx).UpdateFields(ctx, existing.ID, map[string]interface{}{
			"status": 1,
		}); err != nil {
			return err
		}
		err := model.NewConcernedCountModel(tx).DecrConcernedNum(ctx, msg.BizId, msg.ObjId)
		if err != nil {
			return err
		}
		return l.syncCollectNum(ctx, tx, msg.BizId, msg.ObjId)
	})
	if err != nil {
		l.Errorf("[cancelConcerned] transaction err: %v msg: %+v", err, msg)
		return err
	}

	l.Infof("[cancelConcerned] success bizId: %s objId: %d userId: %d", msg.BizId, msg.ObjId, msg.UserId)

	// 发送取消收藏通知
	l.sendCollectNotification(ctx, msg, "取消了对你的笔记的收藏")

	return nil
}

func (l *ConsumeLogic) syncCollectNum(ctx context.Context, tx *gorm.DB, bizId string, objId int64) error {
	if bizId != "article" {
		return nil
	}
	var count model.ConcernedCount
	err := tx.WithContext(ctx).Where("biz_id = ? AND obj_id = ?", bizId, objId).First(&count).Error
	if err != nil {
		l.Errorf("[syncCollectNum] find count err: %v bizId: %s objId: %d", err, bizId, objId)
		return err
	}
	err = tx.WithContext(ctx).Table("article").Where("id = ?", objId).Update("collect_num", count.ConcernedNum).Error
	if err != nil {
		l.Errorf("[syncCollectNum] update article err: %v objId: %d", err, objId)
		return err
	}
	return nil
}

func (l *ConsumeLogic) sendCollectNotification(ctx context.Context, msg *types.ConcernedMsg, action string) {
	if msg.BizId != "article" {
		return
	}

	// 查询文章作者
	var article struct {
		AuthorId int64 `gorm:"column:author_id"`
	}
	if err := l.svcCtx.DB.DB.WithContext(ctx).Table("article").
		Select("author_id").Where("id = ?", msg.ObjId).First(&article).Error; err != nil {
		l.Errorf("[sendCollectNotification] query article author error: %v", err)
		return
	}

	// 不给自己发通知
	if article.AuthorId == msg.UserId {
		return
	}

	notif := map[string]interface{}{
		"userId":        article.AuthorId,
		"type":          int32(6),
		"title":         action,
		"content":       action,
		"refId":         msg.ObjId,
		"bizId":         fmt.Sprintf("collect:%s:%d:%d", msg.BizId, msg.ObjId, msg.UserId),
		"triggerUserId": msg.UserId,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		l.Errorf("[sendCollectNotification] marshal error: %v", err)
		return
	}
	if err := l.svcCtx.NotificationPusher.Push(context.Background(), string(data)); err != nil {
		l.Errorf("[sendCollectNotification] push error: %v", err)
	}
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.ConcernedKq, NewConsumeLogic(ctx, svcCtx)),
	}
}
