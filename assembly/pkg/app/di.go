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

func (d *diContainer) ProducerService(topicName, broker string) (producer.ShipAssembledService, error) {
	if d.producerService == nil {
		wrappedProducer, producerErr := d.WrappedProducer(topicName, broker)
		if producerErr != nil {
			return nil, producerErr
		}

		d.producerService = producer.NewService(wrappedProducer)
	}

	return d.producerService, nil
}

func (d *diContainer) ConsumerService(produceTopic, consumeTopic, broker, groupId string) (consumer.OrderPaidService, error) {
	if d.consumerService == nil {
		wrappedConsumer, consumerErr := d.WrappedConsumer(consumeTopic, broker, groupId)
		if consumerErr != nil {
			return nil, consumerErr
		}

		producerService, producerErr := d.ProducerService(produceTopic, broker)
		if producerErr != nil {
			return nil, producerErr
		}

		d.consumerService = consumer.NewService(wrappedConsumer, d.OrderPaidDecoder(), producerService)
	}

	return d.consumerService, nil
}

func (d *diContainer) OrderPaidDecoder() wrappedKafkaConsumer.OrderPaidDecoder {
	if d.orderPaidDecoder == nil {
		d.orderPaidDecoder = wrappedKafkaConsumer.NewOrderPaidRecordedDecoder()
	}

	return d.orderPaidDecoder
}

func (d *diContainer) WrappedProducer(topicName, broker string) (wrappedKafka.Producer, error) {
	if d.producerService == nil {
		syncProducer, producerErr := d.SyncProducer(broker)
		if producerErr != nil {
			return nil, producerErr
		}

		d.wrappedProducer = wrappedKafkaProducer.NewProducer(
			syncProducer,
			topicName,
			logger.Logger(),
		)
	}

	return d.wrappedProducer, nil
}

func (d *diContainer) WrappedConsumer(topicName, broker, groupId string) (wrappedKafka.Consumer, error) {
	if d.consumerService == nil {
		consumerGroup, consumerErr := d.ConsumerGroup(broker, groupId)
		if consumerErr != nil {
			return nil, consumerErr
		}

		d.wrappedConsumer = wrappedKafkaConsumer.NewConsumer(
			consumerGroup,
			[]string{topicName},
			logger.Logger(),
		)
	}

	return d.wrappedConsumer, nil
}

func (d *diContainer) SyncProducer(broker string) (sarama.SyncProducer, error) {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			[]string{broker},
			wrappedKafkaProducer.Config(),
		)
		if err != nil {

			return nil, fmt.Errorf("failed to create sync producer: %w", err)
		}

		d.syncProducer = p
	}

	return d.syncProducer, nil
}

func (d *diContainer) ConsumerGroup(broker, groupId string) (sarama.ConsumerGroup, error) {
	if d.consumerGroup == nil {
		consumerGroup, err := sarama.NewConsumerGroup(
			[]string{broker},
			groupId,
			wrappedKafkaConsumer.Config(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create consumer group: %w", err)
		}

		d.consumerGroup = consumerGroup
	}

	return d.consumerGroup, nil
}
