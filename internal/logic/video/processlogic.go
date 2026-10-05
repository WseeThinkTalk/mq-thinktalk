package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	l.Infof("[VideoProcess] starting zero-disk streaming processing for video: %d, key: %s", msg.VideoId, msg.ObjectKey)

	if l.svcCtx.MinIO == nil {
		l.Errorf("[VideoProcess] MinIO client is nil")
		l.markFailed(msg.VideoId, "MinIO 存储服务未连接")
		return fmt.Errorf("minio client is nil")
	}

	// 1. 生成 10 分钟临时 Presigned GET URL，用于 FFmpeg / ffprobe 远程 HTTP Range 流式读取（零本地磁盘写入）
	presignedGetURL, err := l.svcCtx.MinIO.PresignedGetObject(ctx, msg.Bucket, msg.ObjectKey, 10*time.Minute, nil)
	if err != nil {
		l.Errorf("[VideoProcess] generate presigned get url failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}
	videoStreamURL := presignedGetURL.String()

	// 2. 调用 ffprobe 通过 HTTP Range 远程流式解析视频元数据（零本地磁盘写入）
	meta, err := ffmpeg.ProbeVideoURL(ctx, videoStreamURL)
	if err != nil {
		l.Errorf("[VideoProcess] ffprobe remote metadata failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	// 3. 调用 ffmpeg 通过 HTTP Range 远程截取第 1 秒高质量封面图并直出内存管道（零磁盘 I/O）
	coverBytes, err := ffmpeg.ExtractCoverFromURLToMemory(ctx, videoStreamURL, 1.0)
	if err != nil {
		l.Errorf("[VideoProcess] ffmpeg stream extract cover failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	// 4. 将内存中的封面图字节流直接 PutObject 推回 MinIO（无需任何临时磁盘文件）
	coverFileName := fmt.Sprintf("cover_%d.jpg", msg.VideoId)
	coverObjectKey := fmt.Sprintf("cover/auto/%s/%s", time.Now().Format("20060102"), coverFileName)
	_, err = l.svcCtx.MinIO.PutObject(ctx, msg.Bucket, coverObjectKey, bytes.NewReader(coverBytes), int64(len(coverBytes)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		l.Errorf("[VideoProcess] stream upload cover to minio failed: %v", err)
		l.markFailed(msg.VideoId, err.Error())
		return err
	}

	coverViewUrl := fmt.Sprintf("/static/%s/%s", msg.Bucket, coverObjectKey)

	// 5. 更新 Redis 状态为 ready，供前端轮询拉取
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

	l.Infof("[VideoProcess] successfully completed zero-disk stream processing for video: %d, cover: %s", msg.VideoId, coverViewUrl)
	return nil
}

func (l *VideoProcessLogic) markFailed(videoId int64, errMsg string) {
	if l.svcCtx.BizRedis != nil {
		redisKey := fmt.Sprintf("biz#video#status:%d", videoId)
		statusVal, _ := json.Marshal(map[string]interface{}{
			"status": "failed",
			"error":  errMsg,
		})
		_ = l.svcCtx.BizRedis.SetexCtx(context.Background(), redisKey, string(statusVal), 3600*24)
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
