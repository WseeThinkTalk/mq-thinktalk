package types

type ReplyMsg struct {
	BizId        string `json:"bizId"`        // 业务ID
	TargetId     int64  `json:"targetId"`     // 评论目标id
	ReplyUserId  int64  `json:"replyUserId"`  // 评论用户ID
	BeReplyUserId int64 `json:"beReplyUserId"` // 被回复用户ID
	ParentId     int64  `json:"parentId"`     // 父评论ID
	Content      string `json:"content"`      // 内容
	OpType       int32  `json:"opType"`       // 操作类型 0:发表 1:删除
	ReplyId      int64  `json:"replyId"`      // 评论ID (删除时使用)
	UserId       int64  `json:"userId"`       // 操作用户ID (删除时使用)
}

const (
	OpTypeCreate = 0 // 发表
	OpTypeDelete = 1 // 删除
)
