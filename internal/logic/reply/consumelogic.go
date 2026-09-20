package reply

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	model "mq-thinktalk/internal/model/reply"
	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/reply"

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

	var msg types.ReplyMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[Consume] unmarshal msg error: %v", err)
		return err
	}

	switch msg.OpType {
	case types.OpTypeCreate:
		return l.createReply(ctx, &msg)
	case types.OpTypeDelete:
		return l.deleteReply(ctx, &msg)
	default:
		l.Errorf("[Consume] unknown opType: %d", msg.OpType)
		return nil
	}
}

func (l *ConsumeLogic) createReply(ctx context.Context, msg *types.ReplyMsg) error {
	isRoot := msg.ParentId == 0

	reply := &model.Reply{
		BizID:         msg.BizId,
		TargetID:      msg.TargetId,
		ReplyUserID:   msg.ReplyUserId,
		BeReplyUserID: msg.BeReplyUserId,
		ParentID:      msg.ParentId,
		Content:       msg.Content,
		Status:        0,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.NewReplyModel(tx).Insert(ctx, reply); err != nil {
			return err
		}
		if err := model.NewReplyCountModel(tx).IncrReplyNum(ctx, msg.BizId, msg.TargetId, isRoot); err != nil {
			return err
		}
		return l.syncCommentNum(ctx, tx, msg.BizId, msg.TargetId)
	})
	if err != nil {
		l.Errorf("[createReply] transaction err: %v msg: %+v", err, msg)
		return err
	}

	// 发送评论通知
	l.sendReplyNotification(ctx, msg, reply.ID)

	l.Infof("[createReply] success replyId: %d bizId: %s targetId: %d", reply.ID, msg.BizId, msg.TargetId)
	return nil
}

func (l *ConsumeLogic) deleteReply(ctx context.Context, msg *types.ReplyMsg) error {
	reply, err := l.svcCtx.ReplyModel.FindOne(ctx, msg.ReplyId)
	if err != nil {
		l.Errorf("[deleteReply] ReplyModel.FindOne err: %v replyId: %d", err, msg.ReplyId)
		return err
	}
	if reply == nil || reply.Status == 1 {
		l.Infof("[deleteReply] reply already deleted, replyId: %d", msg.ReplyId)
		return nil
	}

	isRoot := reply.ParentID == 0

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.NewReplyModel(tx).UpdateFields(ctx, msg.ReplyId, map[string]interface{}{
			"status": 1,
		}); err != nil {
			return err
		}
		if err := model.NewReplyCountModel(tx).DecrReplyNum(ctx, reply.BizID, reply.TargetID, isRoot); err != nil {
			return err
		}
		return l.syncCommentNum(ctx, tx, reply.BizID, reply.TargetID)
	})
	if err != nil {
		l.Errorf("[deleteReply] transaction err: %v replyId: %d", err, msg.ReplyId)
		return err
	}

	l.Infof("[deleteReply] success replyId: %d", msg.ReplyId)
	return nil
}

func (l *ConsumeLogic) syncCommentNum(ctx context.Context, tx *gorm.DB, bizId string, targetId int64) error {
	if bizId != "article" {
		return nil
	}
	var count model.ReplyCount
	err := tx.WithContext(ctx).Where("biz_id = ? AND target_id = ?", bizId, targetId).First(&count).Error
	if err != nil {
		l.Errorf("[syncCommentNum] find count err: %v bizId: %s targetId: %d", err, bizId, targetId)
		return err
	}
	err = tx.WithContext(ctx).Table("article").Where("id = ?", targetId).Update("comment_num", count.ReplyNum).Error
	if err != nil {
		l.Errorf("[syncCommentNum] update article err: %v targetId: %d", err, targetId)
		return err
	}
	return nil
}

func (l *ConsumeLogic) sendReplyNotification(ctx context.Context, msg *types.ReplyMsg, replyId int64) {
	isRoot := msg.ParentId == 0

	// 确定通知接收者
	notifyUserId := msg.BeReplyUserId
	if isRoot && notifyUserId == 0 && msg.BizId == "article" {
		// 根评论（直接评论文章）：通知文章作者
		var article struct {
			AuthorId int64 `gorm:"column:author_id"`
		}
		if err := l.svcCtx.DB.DB.WithContext(ctx).Table("article").
			Select("author_id").Where("id = ?", msg.TargetId).First(&article).Error; err != nil {
			l.Errorf("[sendReplyNotification] query article author error: %v", err)
			return
		}
		notifyUserId = article.AuthorId
	}

	// 没有接收者 或 自己评论自己，跳过
	if notifyUserId == 0 || msg.ReplyUserId == notifyUserId {
		return
	}

	action := "回复了你的评论"
	if isRoot {
		action = "评论了你的笔记"
	}

	notif := map[string]interface{}{
		"userId":        notifyUserId,
		"type":          int32(3),
		"title":         action,
		"content":       msg.Content,
		"refId":         msg.TargetId,
		"bizId":         fmt.Sprintf("reply:%d", replyId),
		"triggerUserId": msg.ReplyUserId,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		l.Errorf("[sendReplyNotification] marshal error: %v", err)
		return
	}
	if err := l.svcCtx.NotificationPusher.Push(context.Background(), string(data)); err != nil {
		l.Errorf("[sendReplyNotification] push error: %v", err)
	}
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.ReplyKq, NewConsumeLogic(ctx, svcCtx)),
	}
}
