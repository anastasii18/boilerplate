package app

import (
	"assembly/pkg/service/consumer"
	"assembly/pkg/service/producer"
	"fmt"
	wrappedKafka "platform/pkg/kafka"
	wrappedKafkaConsumer "platform/pkg/kafka/consumer"
	wrappedKafkaProducer "platform/pkg/kafka/producer"
	"platform/pkg/logger"

	"github.com/IBM/sarama"
)

type diContainer struct {
	consumerService  consumer.OrderPaidService
	producerService  producer.ShipAssembledService
	orderPaidDecoder wrappedKafkaConsumer.OrderPaidDecoder

	wrappedConsumer wrappedKafka.Consumer
	wrappedProducer wrappedKafka.Producer

	syncProducer  sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) ProducerService(topicName, broker string) producer.ShipAssembledService {
	if d.producerService == nil {
		d.producerService = producer.NewService(d.WrappedProducer(topicName, broker))
	}

	return d.producerService
}

func (d *diContainer) ConsumerService(produceTopic, consumeTopic, broker, groupId string) consumer.OrderPaidService {
	if d.consumerService == nil {
		d.consumerService = consumer.NewService(d.WrappedConsumer(consumeTopic, broker, groupId), d.OrderPaidDecoder(), d.ProducerService(produceTopic, broker))
	}

	return d.consumerService
}

func (d *diContainer) OrderPaidDecoder() wrappedKafkaConsumer.OrderPaidDecoder {
	if d.orderPaidDecoder == nil {
		d.orderPaidDecoder = wrappedKafkaConsumer.NewOrderPaidRecordedDecoder()
	}

	return d.orderPaidDecoder
}

func (d *diContainer) WrappedProducer(topicName, broker string) wrappedKafka.Producer {
	if d.producerService == nil {
		d.wrappedProducer = wrappedKafkaProducer.NewProducer(
			d.SyncProducer(broker),
			topicName,
			logger.Logger(),
		)
	}

	return d.wrappedProducer
}

func (d *diContainer) WrappedConsumer(topicName, broker, groupId string) wrappedKafka.Consumer {
	if d.consumerService == nil {
		d.wrappedConsumer = wrappedKafkaConsumer.NewConsumer(
			d.ConsumerGroup(broker, groupId),
			[]string{topicName},
			logger.Logger(),
		)
	}

	return d.wrappedConsumer
}

func (d *diContainer) SyncProducer(broker string) sarama.SyncProducer {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			[]string{broker},
			wrappedKafkaProducer.Config(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create sync producer: %s\n", err.Error()))
		}

		d.syncProducer = p
	}

	return d.syncProducer
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
