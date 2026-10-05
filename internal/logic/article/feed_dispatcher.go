package article

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

const (
	BigVFollowerThreshold = 5000      // 大 V 粉丝分水岭
	MaxInboxCapacity      = 800       // 粉丝收件箱保留上限
	MaxOutboxCapacity     = 200       // 大 V 发件箱保留上限
	ColdUserThresholdSec  = 7 * 86400 // 超过 7 天未活跃判定为冷用户
)

type FeedDispatcher struct {
	db       *gorm.DB
	bizRedis *redis.Redis
	logger   logx.Logger
}

func NewFeedDispatcher(db *gorm.DB, bizRedis *redis.Redis, logger logx.Logger) *FeedDispatcher {
	return &FeedDispatcher{
		db:       db,
		bizRedis: bizRedis,
		logger:   logger,
	}
}

// IsColdUser 判断用户是否为超过 7 天未上线的冷用户
func IsColdUser(lastActiveUnix, nowUnix int64) bool {
	if lastActiveUnix <= 0 {
		return false // 未记录过活跃时间的用户默认不丢弃，按普通用户处理
	}
	return nowUnix-lastActiveUnix > ColdUserThresholdSec
}

// DispatchArticleEvent 当文章状态为 Visible (可见) 时执行推拉结合分水岭投递
func (d *FeedDispatcher) DispatchArticleEvent(ctx context.Context, authorID, articleID, publishTime int64) error {
	if d.bizRedis == nil {
		return nil
	}

	artIdStr := strconv.FormatInt(articleID, 10)

	// 1. 查询作者粉丝总数
	var followerCount int64
	if d.db != nil {
		_ = d.db.WithContext(ctx).Table("concerned_count").
			Where("biz_id = ? AND obj_id = ?", "user", authorID).
			Pluck("concerned_num", &followerCount).Error
	}

	// 2. 分水岭判定：大 V 仅写入自身发件箱 (Outbox)，规避写放大
	if followerCount > BigVFollowerThreshold {
		outboxKey := fmt.Sprintf("biz#feed#outbox:%d", authorID)
		_, err := d.bizRedis.ZaddCtx(ctx, outboxKey, publishTime, artIdStr)
		if err != nil {
			return err
		}
		_, _ = d.bizRedis.ZremrangebyrankCtx(ctx, outboxKey, 0, -MaxOutboxCapacity-1)
		_ = d.bizRedis.ExpireCtx(ctx, outboxKey, 86400*30)
		d.logger.Infof("[FeedDispatcher] Author %d is Big V (fans=%d), pushed to outbox", authorID, followerCount)
		return nil
	}

	// 3. 普通博主：查询活跃粉丝列表并执行写扩散 (Push)
	if d.db == nil {
		return nil
	}

	var followerIDs []int64
	err := d.db.WithContext(ctx).Table("concerned_record").
		Where("biz_id = ? AND obj_id = ? AND status = 1", "user", authorID).
		Pluck("user_id", &followerIDs).Error
	if err != nil {
		d.logger.Errorf("[FeedDispatcher] query followers error: %v", err)
		return err
	}

	if len(followerIDs) == 0 {
		return nil
	}

	nowUnix := time.Now().Unix()
	pushedCount := 0

	// 4. 粉丝活跃度过滤（冷热隔离）：超过 7 天不上线的僵尸用户跳过写扩散
	for _, fID := range followerIDs {
		activeKey := fmt.Sprintf("user:active:%d", fID)
		activeVal, err := d.bizRedis.GetCtx(ctx, activeKey)
		if err == nil && activeVal != "" {
			lastActive, _ := strconv.ParseInt(activeVal, 10, 64)
			if IsColdUser(lastActive, nowUnix) {
				// 冷用户跳过推入收件箱，节约极昂贵的 Redis 内存
				continue
			}
		}

		inboxKey := fmt.Sprintf("biz#feed#inbox:%d", fID)
		_, _ = d.bizRedis.ZaddCtx(ctx, inboxKey, publishTime, artIdStr)
		_, _ = d.bizRedis.ZremrangebyrankCtx(ctx, inboxKey, 0, -MaxInboxCapacity-1)
		_ = d.bizRedis.ExpireCtx(ctx, inboxKey, 86400*14)
		pushedCount++
	}

	d.logger.Infof("[FeedDispatcher] Author %d article %d pushed to %d/%d active followers",
		authorID, articleID, pushedCount, len(followerIDs))
	return nil
}
