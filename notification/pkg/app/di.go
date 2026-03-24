package app

import (
	"context"
	"fmt"
	"notification/pkg/client/telegram"
	"notification/pkg/service"
	"notification/pkg/service/consumer"
	wrappedKafka "platform/pkg/kafka"
	wrappedKafkaConsumer "platform/pkg/kafka/consumer"
	"platform/pkg/kafka/producer"
	"platform/pkg/logger"

	"github.com/IBM/sarama"
	"github.com/go-telegram/bot"
)

const (
	// вставить значение для демонстрации
	telegramBotToken = ""
)

type diContainer struct {
	telegramClient  telegram.TelegramClient
	telegramBot     *bot.Bot
	telegramService service.TelegramService

	assembledDecoder producer.ShipAssembledDecoder
	orderPaidDecoder wrappedKafkaConsumer.OrderPaidDecoder

	consumer        consumer.ShipAndOrderNotificationService
	wrappedConsumer wrappedKafka.Consumer
	consumerGroup   sarama.ConsumerGroup
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) TelegramService(ctx context.Context) service.TelegramService {
	if d.telegramService == nil {
		d.telegramService = service.NewService(
			d.TelegramClient(ctx),
		)
	}

	return d.telegramService
}

func (d *diContainer) TelegramClient(ctx context.Context) telegram.TelegramClient {
	if d.telegramClient == nil {
		d.telegramClient = telegram.NewClient(d.TelegramBot(ctx))
	}

	return d.telegramClient
}

func (d *diContainer) TelegramBot(ctx context.Context) *bot.Bot {
	if d.telegramBot == nil {
		b, err := bot.New(telegramBotToken)
		if err != nil {
			panic(fmt.Sprintf("failed to create telegram bot: %s\n", err.Error()))
		}

		d.telegramBot = b
	}

	return d.telegramBot
}

func (d *diContainer) NotificationConsumer(ctx context.Context, broker, groupId string, topicName []string) consumer.ShipAndOrderNotificationService {
	if d.consumer == nil {
		d.consumer = consumer.NewService(
			d.WrappedConsumer(topicName, broker, groupId),
			d.ShipAssembledDecoder(),
			d.OrderPaidDecoder(),
			d.TelegramService(ctx),
		)
	}

	return d.consumer
}

func (d *diContainer) WrappedConsumer(topicName []string, broker, groupId string) wrappedKafka.Consumer {
	if d.wrappedConsumer == nil {
		d.wrappedConsumer = wrappedKafkaConsumer.NewConsumer(
			d.ConsumerGroup(broker, groupId),
			topicName,
			logger.Logger(),
		)
	}

	return d.wrappedConsumer
}

func (d *diContainer) ConsumerGroup(broker, groupId string) sarama.ConsumerGroup {
	if d.consumerGroup == nil {
		consumerGroup, err := sarama.NewConsumerGroup(
			[]string{broker},
			groupId,
			wrappedKafkaConsumer.Config(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create consumer group: %s\n", err.Error()))
		}

		d.consumerGroup = consumerGroup
	}

	return d.consumerGroup
}

func (d *diContainer) ShipAssembledDecoder() producer.ShipAssembledDecoder {
	if d.assembledDecoder == nil {
		d.assembledDecoder = producer.NewShipAssembledDecoder()
	}

	return d.assembledDecoder
}

func (d *diContainer) OrderPaidDecoder() wrappedKafkaConsumer.OrderPaidDecoder {
	if d.orderPaidDecoder == nil {
		d.orderPaidDecoder = wrappedKafkaConsumer.NewOrderPaidRecordedDecoder()
	}

	return d.orderPaidDecoder
}
