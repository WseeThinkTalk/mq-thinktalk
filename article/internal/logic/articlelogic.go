package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mq-thinktalk/article/internal/svc"
	"mq-thinktalk/article/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleLogic {
	return &ArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleLogic) Consume(ctx context.Context, _, val string) error {
	logx.Infof("Consume msg val: %s", val)
	var msg *types.CanalArticleMsg
	err := json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("Consume val: %s error: %v", val, err)
		return err
	}

	return l.articleOperate(msg)
}

func (l *ArticleLogic) articleOperate(msg *types.CanalArticleMsg) error {
	if len(msg.Data) == 0 {
		return nil
	}

	for _, d := range msg.Data {
		status, _ := strconv.Atoi(d.Status)
		likNum, _ := strconv.ParseInt(d.LikeNum, 10, 64)

		t, err := time.ParseInLocation("2006-01-02 15:04:05", d.PublishTime, time.Local)
		if err != nil {
			t = time.Now()
		}
		publishTimeKey := articlesKey(d.AuthorId, 0)
		likeNumKey := articlesKey(d.AuthorId, 1)
		globalKey := "biz#articles#global"

		switch status {
		case types.ArticleStatusVisible:
			b, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, publishTimeKey)
			if b {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, t.Unix(), d.ID)
			}
			b, _ = l.svcCtx.BizRedis.ExistsCtx(l.ctx, likeNumKey)
			if b {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, likeNumKey, likNum, d.ID)
			}
			bGlobal, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, globalKey)
			if bGlobal {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, globalKey, t.Unix(), d.ID)
			}

		default:
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, publishTimeKey, d.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, likeNumKey, d.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, globalKey, d.ID)
		}
	}

	return nil
}

func articlesKey(uid string, sortType int32) string {
	return fmt.Sprintf("biz#articles#%s#%d", uid, sortType)
}
