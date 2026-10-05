package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// VideoMetadata 视频元数据
type VideoMetadata struct {
	Duration float64 `json:"duration"` // 时长（秒）
	Width    int     `json:"width"`    // 宽度
	Height   int     `json:"height"`   // 高度
	Bitrate  int64   `json:"bitrate"`  // 码率
}

type ffprobeJSON struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

// ParseFfprobeOutput 解析 ffprobe JSON 输出
func ParseFfprobeOutput(data []byte) (*VideoMetadata, error) {
	var parsed ffprobeJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ffprobe output: %w", err)
	}

	dur, _ := strconv.ParseFloat(parsed.Format.Duration, 64)
	bitrate, _ := strconv.ParseInt(parsed.Format.BitRate, 10, 64)

	meta := &VideoMetadata{
		Duration: dur,
		Bitrate:  bitrate,
	}

	for _, stream := range parsed.Streams {
		if stream.CodecType == "video" {
			meta.Width = stream.Width
			meta.Height = stream.Height
			break
		}
	}

	return meta, nil
}

// BuildExtractCoverArgs 构建高质量首帧抽帧参数
func BuildExtractCoverArgs(inputPath, outputPath string, atSeconds float64) []string {
	timeStr := fmt.Sprintf("%02d:%02d:%02d", int(atSeconds)/3600, (int(atSeconds)%3600)/60, int(atSeconds)%60)
	return []string{
		"-ss", timeStr,
		"-i", inputPath,
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	}
}

// ProbeVideo 调用系统 ffprobe 提取视频元数据（强依赖环境）
func ProbeVideo(ctx context.Context, videoPath string) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed (%v): %s", err, stderr.String())
	}

	return ParseFfprobeOutput(stdout.Bytes())
}

// ExtractFirstFrameCover 调用系统 ffmpeg 提取首帧图片
func ExtractFirstFrameCover(ctx context.Context, videoPath, outputPath string) error {
	args := BuildExtractCoverArgs(videoPath, outputPath, 1.0)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg extract cover failed (%v): %s", err, stderr.String())
	}

	return nil
}
