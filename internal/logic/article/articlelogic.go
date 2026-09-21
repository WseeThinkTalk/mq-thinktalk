package article

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mq-thinktalk/internal/svc"
	types "mq-thinktalk/internal/types/article"

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

	// 处理 Canal 同步的文章数据变更
	for _, v := range msg.Data {
		status, _ := strconv.Atoi(v.Status)
		likNum, _ := strconv.ParseInt(v.LikeNum, 10, 64)

		t, err := time.ParseInLocation("2006-01-02 15:04:05", v.PublishTime, time.Local)
		if err != nil {
			t = time.Now()
		}
		publishTimeKey := articlesKey(v.AuthorId, 0)
		likeNumKey := articlesKey(v.AuthorId, 1)
		globalKey := "biz#articles#global"

		switch status {
		case types.ArticleStatusVisible:
			b, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, publishTimeKey)
			if b {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, t.Unix(), v.ID)
			}
			b, _ = l.svcCtx.BizRedis.ExistsCtx(l.ctx, likeNumKey)
			if b {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, likeNumKey, likNum, v.ID)
			}
			bGlobal, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, globalKey)
			if bGlobal {
				_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, globalKey, t.Unix(), v.ID)
			}

		default:
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, publishTimeKey, v.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, likeNumKey, v.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, globalKey, v.ID)
		}
	}

	return nil
}

func articlesKey(uid string, sortType int32) string {
	return fmt.Sprintf("biz#articles#%s#%d", uid, sortType)
}
