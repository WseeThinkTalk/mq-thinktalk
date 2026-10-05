package like

import (
	"context"
	"encoding/json"
	"fmt"

	model "mq-thinktalk/internal/model/like"
	types "mq-thinktalk/internal/types/like"
)

func (l *ThumbupLogic) addLike(ctx context.Context, msg types.ThumbupMsg) error {
	_, err := l.svcCtx.LikeRecordModel.Insert(ctx, &model.LikeRecord{
		BizId:    msg.BizId,
		ObjId:    msg.ObjId,
		UserId:   msg.UserId,
		LikeType: int64(msg.LikeType),
	})
	if err != nil {
		l.Errorf("[Thumbup] insert like record error: %v", err)
		return err
	}

	// 增量累加至批量冲刷器，避免并发高频单行锁争用
	if msg.LikeType == 1 {
		l.flusher.Add(msg.BizId, msg.ObjId, 1, 0)
	} else if msg.LikeType == 2 {
		l.flusher.Add(msg.BizId, msg.ObjId, 0, 1)
	}

	// 发送通知给文章/评论作者
	l.sendLikeNotification(ctx, msg)

	return nil
}

func (l *ThumbupLogic) sendLikeNotification(ctx context.Context, msg types.ThumbupMsg) {
	// LikeType=1 → 点赞通知 type=1
	if msg.LikeType != 1 {
		return
	}
	notifType := int32(1)
	action := "点赞了你的笔记"

	// 查询作者ID
	var authorId int64
	if msg.BizId == "article" {
		var article struct {
			AuthorId int64 `gorm:"column:author_id"`
		}
		if err := l.svcCtx.DB.DB.WithContext(ctx).Table("article").
			Select("author_id").Where("id = ?", msg.ObjId).First(&article).Error; err != nil {
			l.Errorf("[sendLikeNotification] query article author error: %v", err)
			return
		}
		authorId = article.AuthorId
	} else if msg.BizId == "reply" {
		var reply struct {
			UserId int64 `gorm:"column:user_id"`
		}
		if err := l.svcCtx.DB.DB.WithContext(ctx).Table("reply").
			Select("user_id").Where("reply_id = ?", msg.ObjId).First(&reply).Error; err != nil {
			l.Errorf("[sendLikeNotification] query reply author error: %v", err)
			return
		}
		authorId = reply.UserId
	} else {
		return
	}

	// 不给自己发通知
	if authorId == msg.UserId {
		return
	}

	notif := map[string]interface{}{
		"userId":        authorId,
		"type":          notifType,
		"title":         action,
		"content":       action,
		"refId":         msg.ObjId,
		"bizId":         fmt.Sprintf("like:%s:%d:%d", msg.BizId, msg.ObjId, msg.UserId),
		"triggerUserId": msg.UserId,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		l.Errorf("[sendLikeNotification] marshal error: %v", err)
		return
	}
	if err := l.svcCtx.NotificationPusher.Push(context.Background(), string(data)); err != nil {
		l.Errorf("[sendLikeNotification] push error: %v", err)
	}
}
