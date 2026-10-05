package reconcile

import (
	"context"
	"time"
)

// DiffItem 对账差异条目
type DiffItem struct {
	ObjId    int64 `json:"objId"`
	Expected int64 `json:"expected"`
	Actual   int64 `json:"actual"`
	Diff     int64 `json:"diff"`
}

// ReconcileReport 对账与补偿执行审计报告
type ReconcileReport struct {
	BizId              string        `json:"bizId"`
	TotalScanned       int           `json:"totalScanned"`
	TotalDiscrepancies int           `json:"totalDiscrepancies"`
	TotalCompensated   int           `json:"totalCompensated"`
	Duration           time.Duration `json:"duration"`
	Diffs              []DiffItem    `json:"diffs,omitempty"`
}

// Store 对账数据持久化适配接口
type Store interface {
	GetActualCounts(ctx context.Context, bizId string, objIds []int64) (map[int64]int64, error)
	GetCurrentCounts(ctx context.Context, bizId string, objIds []int64) (map[int64]int64, error)
	CompensateCounts(ctx context.Context, bizId string, fixes map[int64]int64) error
}
