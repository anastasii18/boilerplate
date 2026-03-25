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

type TGService struct {
	telegramClient telegram.TelegramClient
}

func (s *TGService) SendMessage(ctx context.Context, msg string, chatID int64) error {
	return s.telegramClient.SendMessage(ctx, chatID, msg)
}

// NewService создает новый Telegram сервис
func NewService(telegramClient telegram.TelegramClient) *TGService {
	return &TGService{
		telegramClient: telegramClient,
	}
}

// SendOrderPaidNotification отправляет уведомление об оплате заказа
func (s *TGService) SendOrderPaidNotification(ctx context.Context, orderPaid orderService.OrderPaid, chatID int64) error {
	message, err := s.buildOrderMessage(orderPaid)
	if err != nil {
		return err
	}

	err = s.SendMessage(ctx, message, chatID)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Telegram message sent to chat", zap.Int("chat_id", int(chatID)), zap.String("message", message))
	return nil
}

// SendAssembledNotification отправляет уведомление о сборке заказа
func (s *TGService) SendAssembledNotification(ctx context.Context, shipAssembled assemblyProducer.ShipAssembled, chatID int64) error {
	message, err := s.buildAssembledMessage(shipAssembled)
	if err != nil {
		return err
	}

	err = s.SendMessage(ctx, message, chatID)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Telegram message sent to chat", zap.Int("chat_id", int(chatID)), zap.String("message", message))
	return nil
}

func (s *TGService) buildOrderMessage(orderPaid orderService.OrderPaid) (string, error) {
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

func (s *TGService) buildAssembledMessage(shipAssembled assemblyProducer.ShipAssembled) (string, error) {
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
