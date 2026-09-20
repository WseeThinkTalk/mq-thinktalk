package types

type MemberMsg struct {
	UserId        int64  `json:"user_id"`
	Level         int32  `json:"level"`
	DurationDays  int32  `json:"duration_days"`
	Amount        int64  `json:"amount"`
	PayChannel    string `json:"pay_channel"`
	TransactionId string `json:"transaction_id"`
}
