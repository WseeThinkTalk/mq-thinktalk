package types

type ChatMsg struct {
	SenderId   int64  `json:"senderId"`
	ReceiverId int64  `json:"receiverId"`
	Content    string `json:"content"`
	MsgType    int32  `json:"msgType"`
}
