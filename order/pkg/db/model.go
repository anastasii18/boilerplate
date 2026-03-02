package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Order struct {
	OrderUuid       string             `json:"order_uuid"`
	UserUuid        string             `json:"user_uuid"`
	PartUuids       []string           `json:"part_uuids"`
	TotalPrice      float64            `json:"total_price"`
	TransactionUuid *string            `json:"transaction_uuid"`
	PaymentMethod   OrderPaymentMethod `json:"payment_method" db:"payment_method"`
	Status          OrderStatus        `json:"status" db:"status"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       *time.Time         `json:"updated_at"`
}

type OrderStatus int32

const (
	PENDING_PAYMENT OrderStatus = 0
	PAID            OrderStatus = 1
	CANCELLED       OrderStatus = 2
)

type OrderPaymentMethod int32

const (
	UNKNOWN        OrderPaymentMethod = 0
	CARD           OrderPaymentMethod = 1
	SBP            OrderPaymentMethod = 2
	CREDIT_CARD    OrderPaymentMethod = 3
	INVESTOR_MONEY OrderPaymentMethod = 4
)

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

func (m *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("status cannot be NULL")
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	switch str {
	case "PENDING_PAYMENT":
		*m = PENDING_PAYMENT
	case "PAID":
		*m = PAID
	case "CANCELLED":
		*m = CANCELLED
	default:
		return fmt.Errorf("unknown status: %q", str)
	}
	return nil
}

func (m *OrderPaymentMethod) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("payment_method cannot be NULL")
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	switch str {
	case "UNKNOWN":
		*m = UNKNOWN
	case "CARD":
		*m = CARD
	case "SBP":
		*m = SBP
	case "CREDIT_CARD":
		*m = CREDIT_CARD
	case "INVESTOR_MONEY":
		*m = INVESTOR_MONEY
	default:
		return fmt.Errorf("unknown payment method: %q", str)
	}
	return nil
}

func (s OrderStatus) String() string {
	switch s {
	case PENDING_PAYMENT:
		return "PENDING_PAYMENT"
	case PAID:
		return "PAID"
	case CANCELLED:
		return "CANCELLED"
	default:
		return "PENDING_PAYMENT"
	}
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
		return "UNKNOWN"
	}
}
