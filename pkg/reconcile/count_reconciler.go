package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// CountReconcilerConfig 配置项
type CountReconcilerConfig struct {
	BizId     string
	BatchSize int
}

// CountReconciler 互动计数离线核对与补偿器
type CountReconciler struct {
	cfg   CountReconcilerConfig
	store Store
}

// NewCountReconciler 创建计数核对器
func NewCountReconciler(cfg CountReconcilerConfig, store Store) *CountReconciler {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &CountReconciler{
		cfg:   cfg,
		store: store,
	}
}

// ReconcileBatch 针对指定对象列表执行批次核对与自动补偿
func (r *CountReconciler) ReconcileBatch(ctx context.Context, objIds []int64) (*ReconcileReport, error) {
	startTime := time.Now()
	report := &ReconcileReport{
		BizId:        r.cfg.BizId,
		TotalScanned: len(objIds),
		Diffs:        make([]DiffItem, 0),
	}

	if len(objIds) == 0 {
		report.Duration = time.Since(startTime)
		return report, nil
	}

	// 1. 获取物理记录真值
	actualMap, err := r.store.GetActualCounts(ctx, r.cfg.BizId, objIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch actual counts: %w", err)
	}

	// 2. 获取当前业务表/计数表存储值
	currentMap, err := r.store.GetCurrentCounts(ctx, r.cfg.BizId, objIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current counts: %w", err)
	}

	// 3. 计算差异列表
	fixes := make(map[int64]int64)
	for _, id := range objIds {
		actual := actualMap[id]
		current := currentMap[id]
		if actual != current {
			report.TotalDiscrepancies++
			diff := actual - current
			report.Diffs = append(report.Diffs, DiffItem{
				ObjId:    id,
				Expected: actual,
				Actual:   current,
				Diff:     diff,
			})
			fixes[id] = actual
		}
	}

	// 4. 自动补偿落库
	if len(fixes) > 0 {
		if err := r.store.CompensateCounts(ctx, r.cfg.BizId, fixes); err != nil {
			logx.WithContext(ctx).Errorf("[CountReconciler] compensate failed: %v", err)
			return nil, fmt.Errorf("failed to compensate counts: %w", err)
		}
		report.TotalCompensated = len(fixes)
		logx.WithContext(ctx).Infof("[CountReconciler] successfully compensated %d records for biz: %s", len(fixes), r.cfg.BizId)
	}

	report.Duration = time.Since(startTime)
	return report, nil
}
