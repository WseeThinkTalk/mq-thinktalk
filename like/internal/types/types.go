package types

type ThumbupMsg struct {
	BizId    string `json:"bizId"`    // 业务id
	ObjId    int64  `json:"objId"`    // 点赞对象id
	UserId   int64  `json:"userId"`   // 用户id
	LikeType int32  `json:"likeType"` // 点赞类型
}
