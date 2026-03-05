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

var STATUS_PENDING_PAYMENT = "PENDING_PAYMENT"
var STATUS_PAID = "PAID"
var STATUS_CANCELLED = "CANCELLED"

type OrderPaymentMethod int32

const (
	UNKNOWN        OrderPaymentMethod = 0
	CARD           OrderPaymentMethod = 1
	SBP            OrderPaymentMethod = 2
	CREDIT_CARD    OrderPaymentMethod = 3
	INVESTOR_MONEY OrderPaymentMethod = 4
)

var PAYMENT_METHOD_UNKNOWN = "UNKNOWN"
var PAYMENT_METHOD_CARD = "CARD"
var PAYMENT_METHOD_SBP = "SBP"
var PAYMENT_METHOD_CREDIT_CARD = "CREDIT_CARD"
var PAYMENT_METHOD_INVESTOR_MONEY = "INVESTOR_MONEY"

func (m *OrderPaymentMethod) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	s = strings.ToUpper(strings.TrimSpace(s))

	switch s {
	case PAYMENT_METHOD_CARD:
		*m = CARD
	case PAYMENT_METHOD_SBP:
		*m = SBP
	case PAYMENT_METHOD_CREDIT_CARD:
		*m = CREDIT_CARD
	case PAYMENT_METHOD_INVESTOR_MONEY:
		*m = INVESTOR_MONEY
	case PAYMENT_METHOD_UNKNOWN:
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
	case STATUS_PENDING_PAYMENT:
		*m = PENDING_PAYMENT
	case STATUS_PAID:
		*m = PAID
	case STATUS_CANCELLED:
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
	case PAYMENT_METHOD_UNKNOWN:
		*m = UNKNOWN
	case PAYMENT_METHOD_CARD:
		*m = CARD
	case PAYMENT_METHOD_SBP:
		*m = SBP
	case PAYMENT_METHOD_CREDIT_CARD:
		*m = CREDIT_CARD
	case PAYMENT_METHOD_INVESTOR_MONEY:
		*m = INVESTOR_MONEY
	default:
		return fmt.Errorf("unknown payment method: %q", str)
	}
	return nil
}

func (s OrderStatus) String() string {
	switch s {
	case PENDING_PAYMENT:
		return STATUS_PENDING_PAYMENT
	case PAID:
		return STATUS_PAID
	case CANCELLED:
		return STATUS_CANCELLED
	default:
		return STATUS_PENDING_PAYMENT
	}
}

func (m OrderPaymentMethod) String() string {
	switch m {
	case UNKNOWN:
		return PAYMENT_METHOD_UNKNOWN
	case CARD:
		return PAYMENT_METHOD_CARD
	case SBP:
		return PAYMENT_METHOD_SBP
	case CREDIT_CARD:
		return PAYMENT_METHOD_CREDIT_CARD
	case INVESTOR_MONEY:
		return PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return "UNKNOWN"
	}
}
