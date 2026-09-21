package article

import (
	"context"
	"encoding/json"
	"strconv"

	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/article"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ArticleLikeNumLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleLikeNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleLikeNumLogic {
	return &ArticleLikeNumLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleLikeNumLogic) Consume(ctx context.Context, _, val string) error {
	var msg *types.CanalLikeMsg
	err := json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("Consume val: %s error: %v", val, err)
		return err
	}

	return l.updateArticleLikeNum(l.ctx, msg)
}

func (l *ArticleLikeNumLogic) updateArticleLikeNum(ctx context.Context, msg *types.CanalLikeMsg) error {
	if len(msg.Data) == 0 {
		return nil
	}

	// 处理 Canal 同步的点赞数数据变更
	for _, v := range msg.Data {
		if v.BizID != types.ArticleBizID {
			continue
		}
		id, err := strconv.ParseInt(v.ObjID, 10, 64)
		if err != nil {
			logx.Errorf("strconv.ParseInt id: %s error: %v", v.ID, err)
			continue
		}
		likeNum, err := strconv.ParseInt(v.LikeNum, 10, 64)
		if err != nil {
			logx.Errorf("strconv.ParseInt likeNum: %s error: %v", v.LikeNum, err)
			continue
		}
		err = l.svcCtx.ArticleModel.UpdateLikeNum(ctx, id, likeNum)
		if err != nil {
			logx.Errorf("UpdateLikeNum id: %d like: %d", id, likeNum)
		}
	}

	return nil
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.ArticleKq, NewArticleLikeNumLogic(ctx, svcCtx)),
		kq.MustNewQueue(svcCtx.Config.ArticleEventKq, NewArticleLogic(ctx, svcCtx)),
	}
}
