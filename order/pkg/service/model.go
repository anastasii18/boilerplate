package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	OrderUuid       string             `json:"order_uuid"`
	UserUuid        string             `json:"user_uuid"`
	PartUuids       []string           `json:"part_uuids"`
	TotalPrice      decimal.Decimal    `json:"total_price"`
	TransactionUuid *string            `json:"transaction_uuid"`
	PaymentMethod   OrderPaymentMethod `json:"payment_method"`
	Status          OrderStatus        `json:"status"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       *time.Time         `json:"updated_at"`
}

type OrderStatus int32

const (
	PENDING_PAYMENT OrderStatus = 0
	PAID            OrderStatus = 1
	CANCELLED       OrderStatus = 2
	COMPLETED       OrderStatus = 3
)

type OrderPaymentMethod int32

const (
	UNKNOWN        OrderPaymentMethod = 0
	CARD           OrderPaymentMethod = 1
	SBP            OrderPaymentMethod = 2
	CREDIT_CARD    OrderPaymentMethod = 3
	INVESTOR_MONEY OrderPaymentMethod = 4
)

type PayBody struct {
	PaymentMethod OrderPaymentMethod `json:"payment_method"`
}

func (m OrderPaymentMethod) String() string {
	switch m {
	case UNKNOWN:
		return "UNKNOWN"
	case CARD:
		return "CARD"
	case SBP:
		return "SBP"
	case CREDIT_CARD:
		return "CREDIT_CARD"
	case INVESTOR_MONEY:
		return "INVESTOR_MONEY"
	default:
		return fmt.Sprintf("OrderPaymentMethod(%d)", m)
	}
}

func (m *OrderPaymentMethod) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	s = strings.ToUpper(strings.TrimSpace(s))

	switch s {
	case "CARD":
		*m = CARD
	case "SBP":
		*m = SBP
	case "CREDIT_CARD", "CREDITCARD":
		*m = CREDIT_CARD
	case "INVESTOR_MONEY", "INVESTORMONEY":
		*m = INVESTOR_MONEY
	case "UNKNOWN":
		*m = UNKNOWN
	default:
		return fmt.Errorf("unknown payment method: %q", s)
	}
	return nil
}
