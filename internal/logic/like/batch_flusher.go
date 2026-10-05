package like

import (
	"fmt"
	"sync"
	"time"
)

// LikeDelta 表示业务对象的点赞和点踩增量
type LikeDelta struct {
	BizId        string
	ObjId        int64
	LikeDelta    int64
	DislikeDelta int64
}

// BatchFlusherConfig 冲刷器配置
type BatchFlusherConfig struct {
	MaxSize  int
	Interval time.Duration
	FlushFn  func(items []*LikeDelta) error
}

// BatchFlusher 内存计数缓冲与批量冲刷器
type BatchFlusher struct {
	mu       sync.Mutex
	buffer   map[string]*LikeDelta
	maxSize  int
	interval time.Duration
	flushFn  func(items []*LikeDelta) error
	ticker   *time.Ticker
	stopCh   chan struct{}
	running  bool
}

// NewBatchFlusher 创建批量冲刷器实例
func NewBatchFlusher(cfg BatchFlusherConfig) *BatchFlusher {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 200
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 1 * time.Second
	}
	return &BatchFlusher{
		buffer:   make(map[string]*LikeDelta),
		maxSize:  cfg.MaxSize,
		interval: cfg.Interval,
		flushFn:  cfg.FlushFn,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动后台定时冲刷协程
func (f *BatchFlusher) Start() {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return
	}
	f.running = true
	f.ticker = time.NewTicker(f.interval)
	f.mu.Unlock()

	go func() {
		for {
			select {
			case <-f.ticker.C:
				f.Flush()
			case <-f.stopCh:
				f.Flush()
				return
			}
		}
	}()
}

// Stop 停止定时器并冲刷剩余数据
func (f *BatchFlusher) Stop() {
	f.mu.Lock()
	if !f.running {
		f.mu.Unlock()
		return
	}
	f.running = false
	f.ticker.Stop()
	close(f.stopCh)
	f.mu.Unlock()
}

// Add 累加增量到内存缓冲区，达到阈值时触发冲刷
func (f *BatchFlusher) Add(bizId string, objId int64, likeDelta, dislikeDelta int64) {
	key := fmt.Sprintf("%s:%d", bizId, objId)

	f.mu.Lock()
	item, ok := f.buffer[key]
	if !ok {
		item = &LikeDelta{
			BizId: bizId,
			ObjId: objId,
		}
		f.buffer[key] = item
	}
	item.LikeDelta += likeDelta
	item.DislikeDelta += dislikeDelta

	shouldFlush := len(f.buffer) >= f.maxSize
	f.mu.Unlock()

	if shouldFlush {
		f.Flush()
	}
}

// Flush 提取当前缓冲数据并执行写入
func (f *BatchFlusher) Flush() {
	f.mu.Lock()
	if len(f.buffer) == 0 {
		f.mu.Unlock()
		return
	}
	current := f.buffer
	f.buffer = make(map[string]*LikeDelta)
	f.mu.Unlock()

	items := make([]*LikeDelta, 0, len(current))
	for _, item := range current {
		items = append(items, item)
	}

	if f.flushFn != nil {
		_ = f.flushFn(items)
	}
}
