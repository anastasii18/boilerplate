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

type diContainer struct {
	telegramClient  telegram.TelegramClient
	telegramBot     *bot.Bot
	telegramService *service.TGService

	assembledDecoder producer.ShipAssembledDecoder
	orderPaidDecoder wrappedKafkaConsumer.OrderPaidDecoder

	consumer        consumer.ShipAndOrderNotificationService
	wrappedConsumer wrappedKafka.Consumer
	consumerGroup   sarama.ConsumerGroup
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) TelegramService(ctx context.Context, token string) (*service.TGService, error) {
	if d.telegramService == nil {
		telegramClient, err := d.TelegramClient(ctx, token)
		if err != nil {
			return nil, err
		}

		d.telegramService = service.NewService(telegramClient)
	}

	return d.telegramService, nil
}

func (d *diContainer) TelegramClient(ctx context.Context, token string) (telegram.TelegramClient, error) {
	if d.telegramClient == nil {
		telegramBot, err := d.TelegramBot(ctx, token)
		if err != nil {
			return nil, err
		}
		d.telegramClient = telegram.NewClient(telegramBot)
	}

	return d.telegramClient, nil
}

func (d *diContainer) TelegramBot(ctx context.Context, token string) (*bot.Bot, error) {
	if d.telegramBot == nil {
		b, err := bot.New(token)
		if err != nil {
			return nil, fmt.Errorf("failed to create telegram bot: %w", err)
		}

		d.telegramBot = b
	}

	return d.telegramBot, nil
}

func (d *diContainer) NotificationConsumer(ctx context.Context, broker, groupId, token string, topicName []string) (consumer.ShipAndOrderNotificationService, error) {
	if d.consumer == nil {
		telegramService, err := d.TelegramService(ctx, token)
		if err != nil {
			return nil, err
		}

		d.consumer = consumer.NewService(
			d.WrappedConsumer(topicName, broker, groupId),
			d.ShipAssembledDecoder(),
			d.OrderPaidDecoder(),
			telegramService,
		)
	}

	return d.consumer, nil
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
