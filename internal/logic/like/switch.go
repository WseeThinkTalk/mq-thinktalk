package like

import (
	"context"
	"time"

	model "mq-thinktalk/internal/model/like"
)

func (l *ThumbupLogic) switchLike(ctx context.Context, record *model.LikeRecord, newLikeType int32) error {
	oldLikeType := record.LikeType
	record.LikeType = int64(newLikeType)
	record.UpdateTime = time.Now()

	if err := l.svcCtx.LikeRecordModel.Update(ctx, record); err != nil {
		l.Errorf("[Thumbup] update like record error: %v", err)
		return err
	}

	// 增量累加至批量冲刷器
	if oldLikeType == 1 && newLikeType == 2 {
		l.flusher.Add(record.BizId, record.ObjId, -1, 1)
	} else if oldLikeType == 2 && newLikeType == 1 {
		l.flusher.Add(record.BizId, record.ObjId, 1, -1)
	}

	return nil
}
