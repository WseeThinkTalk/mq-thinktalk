package like

import (
	"context"
	"encoding/json"
	"fmt"

	model "mq-thinktalk/internal/model/like"
)

func (l *ThumbupLogic) cancelLike(ctx context.Context, record *model.LikeRecord) error {
	if err := l.svcCtx.LikeRecordModel.Delete(ctx, record.Id); err != nil {
		l.Errorf("[Thumbup] delete like record error: %v", err)
		return err
	}

	// 增量累加至批量冲刷器
	if record.LikeType == 1 {
		l.flusher.Add(record.BizId, record.ObjId, -1, 0)
	} else if record.LikeType == 2 {
		l.flusher.Add(record.BizId, record.ObjId, 0, -1)
	}

	// 发送取消通知给文章/评论作者
	l.sendCancelNotification(ctx, record)

	return nil
}

func (l *ThumbupLogic) sendCancelNotification(ctx context.Context, record *model.LikeRecord) {
	// record.LikeType=1 → 取消点赞通知 type=1
	if record.LikeType != 1 {
		return
	}
	notifType := int32(1)
	action := "取消了对你的笔记的点赞"

	// 查询作者ID
	var authorId int64
	if record.BizId == "article" {
		var article struct {
			AuthorId int64 `gorm:"column:author_id"`
		}
		if err := l.svcCtx.DB.DB.WithContext(ctx).Table("article").
			Select("author_id").Where("id = ?", record.ObjId).First(&article).Error; err != nil {
			l.Errorf("[sendCancelNotification] query article author error: %v", err)
			return
		}
		authorId = article.AuthorId
	} else if record.BizId == "reply" {
		var reply struct {
			UserId int64 `gorm:"column:user_id"`
		}
		if err := l.svcCtx.DB.DB.WithContext(ctx).Table("reply").
			Select("user_id").Where("reply_id = ?", record.ObjId).First(&reply).Error; err != nil {
			l.Errorf("[sendCancelNotification] query reply author error: %v", err)
			return
		}
		authorId = reply.UserId
	} else {
		return
	}

	// 不给自己发通知
	if authorId == record.UserId {
		return
	}

	notif := map[string]interface{}{
		"userId":        authorId,
		"type":          notifType,
		"title":         action,
		"content":       action,
		"refId":         record.ObjId,
		"bizId":         fmt.Sprintf("cancel_like:%s:%d:%d", record.BizId, record.ObjId, record.UserId),
		"triggerUserId": record.UserId,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		l.Errorf("[sendCancelNotification] marshal error: %v", err)
		return
	}
	if err := l.svcCtx.NotificationPusher.Push(context.Background(), string(data)); err != nil {
		l.Errorf("[sendCancelNotification] push error: %v", err)
	}
}
