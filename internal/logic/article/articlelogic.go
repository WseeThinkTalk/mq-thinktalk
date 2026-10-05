package article

import (
	"bytes"
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

			// 同步轻量索引文档至 Elasticsearch
			if l.svcCtx.Es != nil {
				authorId, _ := strconv.ParseInt(v.AuthorId, 10, 64)
				id, _ := strconv.ParseInt(v.ID, 10, 64)
				doc := map[string]interface{}{
					"id":           id,
					"title":        v.Title,
					"description":  v.Description,
					"author_id":    authorId,
					"status":       status,
					"publish_time": v.PublishTime,
				}
				docBytes, _ := json.Marshal(doc)
				_, _ = l.svcCtx.Es.Index(
					"thinktalk_article",
					bytes.NewReader(docBytes),
					l.svcCtx.Es.Index.WithDocumentID(v.ID),
					l.svcCtx.Es.Index.WithContext(l.ctx),
				)
			}

			// 触发推拉结合 Feed 流分水岭投递（大 V 读扩散写 Outbox，普通博主写扩散推活跃粉丝 Inbox）
			authorId, _ := strconv.ParseInt(v.AuthorId, 10, 64)
			artId, _ := strconv.ParseInt(v.ID, 10, 64)
			if l.svcCtx.DB != nil {
				dispatcher := NewFeedDispatcher(l.svcCtx.DB.DB, l.svcCtx.BizRedis, l.Logger)
				_ = dispatcher.DispatchArticleEvent(l.ctx, authorId, artId, t.Unix())
			}

		default:
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, publishTimeKey, v.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, likeNumKey, v.ID)
			_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, globalKey, v.ID)

			// 从 ES 移除索引并淘汰详情缓存
			if l.svcCtx.Es != nil {
				_, _ = l.svcCtx.Es.Delete(
					"thinktalk_article",
					v.ID,
					l.svcCtx.Es.Delete.WithContext(l.ctx),
				)
			}
			_, _ = l.svcCtx.BizRedis.DelCtx(l.ctx, fmt.Sprintf("biz#article#detail:%s", v.ID))
		}
	}

	return nil
}

func articlesKey(uid string, sortType int32) string {
	return fmt.Sprintf("biz#articles#%s#%d", uid, sortType)
}
