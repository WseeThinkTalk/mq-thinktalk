package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mq-thinktalk/internal/svc"
	"mq-thinktalk/pkg/ffmpeg"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type VideoProcessMsg struct {
	VideoId   int64  `json:"videoId"`
	ObjectKey string `json:"objectKey"`
	Bucket    string `json:"bucket"`
	VideoUrl  string `json:"videoUrl"`
}

type VideoProcessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVideoProcessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoProcessLogic {
	return &VideoProcessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VideoProcessLogic) Consume(ctx context.Context, key, val string) error {
	var msg VideoProcessMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		l.Errorf("[VideoProcess] unmarshal message failed: %v", err)
		return err
	}

	l.Infof("[VideoProcess] starting processing for video: %d, key: %s", msg.VideoId, msg.ObjectKey)

	if l.svcCtx.MinIO == nil {
		l.Errorf("[VideoProcess] MinIO client is nil")
		l.markFailed(msg.VideoId, "MinIO 存储服务未连接")
		return fmt.Errorf("minio client is nil")
	}

	// 1. 下载原视频到临时目录
	tmpDir := os.TempDir()
	localVideoPath := filepath.Join(tmpDir, fmt.Sprintf("vid_%d_%s", msg.VideoId, filepath.Base(msg.ObjectKey)))
	defer os.Remove(localVideoPath)

	err := l.svcCtx.MinIO.FGetObject(ctx, msg.Bucket, msg.ObjectKey, localVideoPath, minio.GetObjectOptions{})
	if err != nil {
		l.Errorf("[VideoProcess] download video from minio failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	// 2. 调用 ffprobe 提取元数据
	meta, err := ffmpeg.ProbeVideo(ctx, localVideoPath)
	if err != nil {
		l.Errorf("[VideoProcess] ffprobe metadata failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	// 3. 调用 ffmpeg 截取第 1 秒高质量封面图
	coverFileName := fmt.Sprintf("cover_%d.jpg", msg.VideoId)
	localCoverPath := filepath.Join(tmpDir, coverFileName)
	defer os.Remove(localCoverPath)

	if err := ffmpeg.ExtractFirstFrameCover(ctx, localVideoPath, localCoverPath); err != nil {
		l.Errorf("[VideoProcess] ffmpeg extract cover failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	// 4. 将生成的封面图推回 MinIO
	coverObjectKey := fmt.Sprintf("cover/auto/%s/%s", time.Now().Format("20060102"), coverFileName)
	_, err = l.svcCtx.MinIO.FPutObject(ctx, msg.Bucket, coverObjectKey, localCoverPath, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		l.Errorf("[VideoProcess] upload cover to minio failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	coverViewUrl := fmt.Sprintf("/static/%s/%s", msg.Bucket, coverObjectKey)

	// 5. 更新 Redis/DB 状态为 ready，供前端轮询拉取
	if l.svcCtx.BizRedis != nil {
		redisKey := fmt.Sprintf("biz#video#status:%d", msg.VideoId)
		statusVal, _ := json.Marshal(map[string]interface{}{
			"status":   "ready",
			"videoUrl": msg.VideoUrl,
			"coverUrl": coverViewUrl,
			"duration": int(meta.Duration),
			"width":    meta.Width,
			"height":   meta.Height,
		})
		_ = l.svcCtx.BizRedis.SetexCtx(ctx, redisKey, string(statusVal), 86400*7)
	}

	l.Infof("[VideoProcess] successfully completed for video: %d, cover: %s", msg.VideoId, coverViewUrl)
	return nil
}

func (l *VideoProcessLogic) markFailed(videoId int64, errMsg string) {
	if l.svcCtx.BizRedis != nil {
		redisKey := fmt.Sprintf("biz#video#status:%d", videoId)
		statusVal, _ := json.Marshal(map[string]interface{}{
			"status": "failed",
			"errMsg": errMsg,
		})
		_ = l.svcCtx.BizRedis.SetexCtx(l.ctx, redisKey, string(statusVal), 86400*7)
	}
}

// Consumers 注册视频流媒体消费服务
func Consumers(ctx context.Context, svcCtx *svc.ServiceContext) []service.Service {
	if len(svcCtx.Config.VideoProcessKq.Brokers) == 0 {
		return nil
	}
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.VideoProcessKq, NewVideoProcessLogic(ctx, svcCtx)),
	}
}
