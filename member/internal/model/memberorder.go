package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type MemberOrder struct {
	ID            int64 `gorm:"primary_key"`
	UserID        int64
	Level         int32
	DurationDays  int32
	Amount        int64
	PayChannel    string
	TransactionID string
	Status        int32
	CreateTime    time.Time
	UpdateTime    time.Time
}

func (m *MemberOrder) TableName() string {
	return "member_order"
}

type MemberOrderModel struct {
	db *gorm.DB
}

func NewMemberOrderModel(db *gorm.DB) *MemberOrderModel {
	return &MemberOrderModel{db: db}
}

func (m *MemberOrderModel) FindByTransactionId(ctx context.Context, transactionId string) (*MemberOrder, error) {
	var result MemberOrder
	err := m.db.WithContext(ctx).Where("transaction_id = ?", transactionId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *MemberOrderModel) Insert(ctx context.Context, data *MemberOrder) error {
	return m.db.WithContext(ctx).Create(data).Error
}
