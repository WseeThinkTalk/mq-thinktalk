package types

type ConcernedMsg struct {
	BizId  string `json:"bizId"`
	ObjId  int64  `json:"objId"`
	UserId int64  `json:"userId"`
	OpType int32  `json:"opType"`
}

const (
	OpTypeAdd    = 0
	OpTypeCancel = 1
)
