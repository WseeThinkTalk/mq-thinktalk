package like

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBatchFlusher_AddAndFlush(t *testing.T) {
	var flushCount int64
	var flushedItems int64

	flusher := NewBatchFlusher(BatchFlusherConfig{
		MaxSize:  5,
		Interval: 200 * time.Millisecond,
		FlushFn: func(items []*LikeDelta) error {
			atomic.AddInt64(&flushCount, 1)
			atomic.AddInt64(&flushedItems, int64(len(items)))
			return nil
		},
	})
	flusher.Start()
	defer flusher.Stop()

	// 添加不同对象的点赞增量
	flusher.Add("article", 101, 1, 0)
	flusher.Add("article", 102, 1, 0)
	flusher.Add("article", 101, 1, 0) // 同一对象累加

	// 等待定时窗口触发 flush
	time.Sleep(350 * time.Millisecond)

	if atomic.LoadInt64(&flushCount) == 0 {
		t.Fatalf("expected at least 1 flush, got 0")
	}
	if atomic.LoadInt64(&flushedItems) != 2 {
		t.Fatalf("expected 2 unique aggregated items, got %d", atomic.LoadInt64(&flushedItems))
	}
}

func TestBatchFlusher_ThresholdTrigger(t *testing.T) {
	var flushCount int64

	flusher := NewBatchFlusher(BatchFlusherConfig{
		MaxSize:  3,
		Interval: 5 * time.Second, // 较长时间，确保由数量阈值触发
		FlushFn: func(items []*LikeDelta) error {
			atomic.AddInt64(&flushCount, 1)
			return nil
		},
	})
	flusher.Start()
	defer flusher.Stop()

	flusher.Add("article", 1, 1, 0)
	flusher.Add("article", 2, 1, 0)
	flusher.Add("article", 3, 1, 0) // 达到 3 个不同 Key，立即触发

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt64(&flushCount) != 1 {
		t.Fatalf("expected threshold flush triggered, got %d", atomic.LoadInt64(&flushCount))
	}
}
