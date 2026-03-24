package service

import (
	assemblyProducer "assembly/pkg/service/producer"
	"bytes"
	"context"
	"embed"
	"notification/pkg/client/telegram"
	orderService "order/pkg/service"
	"platform/pkg/logger"
	"strings"
	"text/template"
	"time"

	"go.uber.org/zap"
)

// вставить значение для демонстрации
const chatID = 1

//go:embed templates/order_notification.tmpl
var templateOrder embed.FS

//go:embed templates/assembled_notification.tmpl
var templateAssembled embed.FS

type orderPaidTemplateData struct {
	OrderUuid       string
	TransactionUuid *string
	PaymentMethod   string
	PaidAt          time.Time
}

type assembledTemplateData struct {
	OrderUuid    string
	BuildTimeSec int64
	AssembledAt  time.Time
}

var orderTemplate = template.Must(template.ParseFS(templateOrder, "templates/order_notification.tmpl"))
var assembledTemplate = template.Must(template.ParseFS(templateAssembled, "templates/assembled_notification.tmpl"))

type TelegramService interface {
	SendOrderPaidNotification(ctx context.Context, orderPaid orderService.OrderPaid) error
	SendAssembledNotification(ctx context.Context, shipAssembled assemblyProducer.ShipAssembled) error
}

type service struct {
	telegramClient telegram.TelegramClient
}

// NewService создает новый Telegram сервис
func NewService(telegramClient telegram.TelegramClient) *service {
	return &service{
		telegramClient: telegramClient,
	}
}

// SendOrderPaidNotification отправляет уведомление об оплате заказа
func (s *service) SendOrderPaidNotification(ctx context.Context, orderPaid orderService.OrderPaid) error {
	message, err := s.buildOrderMessage(orderPaid)
	if err != nil {
		return err
	}

	err = s.telegramClient.SendMessage(ctx, chatID, message)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Telegram message sent to chat", zap.Int("chat_id", chatID), zap.String("message", message))
	return nil
}

// SendAssembledNotification отправляет уведомление о сборке заказа
func (s *service) SendAssembledNotification(ctx context.Context, shipAssembled assemblyProducer.ShipAssembled) error {
	message, err := s.buildAssembledMessage(shipAssembled)
	if err != nil {
		return err
	}

	err = s.telegramClient.SendMessage(ctx, chatID, message)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Telegram message sent to chat", zap.Int("chat_id", chatID), zap.String("message", message))
	return nil
}

func (s *service) buildOrderMessage(orderPaid orderService.OrderPaid) (string, error) {
	safePaymentMethod := strings.ReplaceAll(orderPaid.PaymentMethod, "_", " ")

	data := orderPaidTemplateData{
		OrderUuid:       orderPaid.OrderUuid,
		TransactionUuid: orderPaid.TransactionUuid,
		PaymentMethod:   safePaymentMethod,
		PaidAt:          time.Now(),
	}
	var buf bytes.Buffer
	err := orderTemplate.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *service) buildAssembledMessage(shipAssembled assemblyProducer.ShipAssembled) (string, error) {
	data := assembledTemplateData{
		OrderUuid:    shipAssembled.OrderUuid,
		BuildTimeSec: shipAssembled.BuildTimeSec,
		AssembledAt:  time.Now(),
	}

	var buf bytes.Buffer
	err := assembledTemplate.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
