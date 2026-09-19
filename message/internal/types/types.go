package types

type NotificationMsg struct {
	UserId        int64  `json:"userId"`        // 接收用户ID
	Type          int32  `json:"type"`          // 通知类型
	Title         string `json:"title"`         // 通知标题
	Content       string `json:"content"`       // 通知内容
	RefId         int64  `json:"refId"`         // 关联资源ID
	BizId         string `json:"bizId"`         // 业务ID
	TriggerUserId int64  `json:"triggerUserId"` // 触发用户ID
}
