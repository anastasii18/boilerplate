package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	api "order/pkg/api/v1"
	"order/pkg/client/inventory"
	"order/pkg/client/payment"
	"order/pkg/db"
	"order/pkg/service"
	"order/pkg/service/consumer"
	wrappedKafka "platform/pkg/kafka"
	wrappedKafkaConsumer "platform/pkg/kafka/consumer"
	"platform/pkg/kafka/producer"
	"platform/pkg/logger"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type diContainer struct {
	Repo            *db.Repository
	Server          *http.Server
	OrderService    service.OrderService
	InventoryClient inventory.Client
	PaymentClient   payment.Client

	consumerService consumer.ShipAssembledService
	producerService service.OrderProducerService

	wrappedConsumer wrappedKafka.Consumer
	wrappedProducer wrappedKafka.Producer

	syncProducer     sarama.SyncProducer
	consumerGroup    sarama.ConsumerGroup
	assembledDecoder producer.ShipAssembledDecoder
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) NewRepo(ctx context.Context, config *Config) *db.Repository {
	database, err := db.NewDB(ctx, config.DbUri)
	if err != nil {
		panic(err)
	}
	d.Repo = db.NewRepository(database)

	return d.Repo
}

func (d *diContainer) NewServer(ctx context.Context, config *Config) *http.Server {
	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	a := api.New(d.OrderService)
	// Определяем маршруты
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Получить заказ по UUID
		r.Get("/{order_uuid}", a.GetOrderHandler(ctx))
		// Создание заказа
		r.Post("/", a.CreateOrderHandler(ctx))
		// Оплата заказа
		r.Post("/{order_uuid}/pay", a.PayOrderHandler(ctx, d.NewPaymentClient(ctx, config)))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", a.CancelOrderHandler(ctx))
	})

	d.Server = &http.Server{
		Addr:              net.JoinHostPort("localhost", config.HttpPort),
		Handler:           r,
		ReadHeaderTimeout: config.ReadHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}
	return d.Server
}

func (d *diContainer) NewOrderService(ctx context.Context, config *Config) service.OrderService {
	d.OrderService = service.NewService(d.NewRepo(ctx, config), d.NewInventoryClient(ctx, config), d.ProducerService(config.ProduceTopicName, config.KafkaBroker))
	return d.OrderService
}

func (d *diContainer) NewInventoryClient(ctx context.Context, config *Config) inventory.Client {
	inventoryClient, err := inventory.NewClient(config.ServerInventoryAddress)
	if err != nil {
		panic(err)
	}
	d.InventoryClient = inventoryClient

	return d.InventoryClient
}

func (d *diContainer) NewPaymentClient(ctx context.Context, config *Config) payment.Client {
	paymentClient, err := payment.NewClient(config.ServerPaymentAddress)
	if err != nil {
		panic(err)
	}
	d.PaymentClient = paymentClient

	return d.PaymentClient
}

func (d *diContainer) ProducerService(topicName, broker string) service.OrderProducerService {
	if d.producerService == nil {
		d.producerService = service.NewProducerService(d.WrappedProducer(topicName, broker))
	}
	return d.producerService
}

func (d *diContainer) ConsumerService(topicName, broker, groupId string) consumer.ShipAssembledService {
	if d.consumerService == nil {
		d.consumerService = consumer.NewService(d.WrappedConsumer(topicName, broker, groupId), d.ShipAssembledDecoder(), d.OrderService)
	}

	return d.consumerService
}

func (d *diContainer) ShipAssembledDecoder() producer.ShipAssembledDecoder {
	if d.assembledDecoder == nil {
		d.assembledDecoder = producer.NewShipAssembledDecoder()
	}

	return d.assembledDecoder
}

func (d *diContainer) WrappedConsumer(topicName, broker, groupId string) wrappedKafka.Consumer {
	if d.wrappedConsumer == nil {
		d.wrappedConsumer = wrappedKafkaConsumer.NewConsumer(
			d.ConsumerGroup(broker, groupId),
			[]string{topicName},
			logger.Logger(),
		)
	}

	return d.wrappedConsumer
}

func (d *diContainer) WrappedProducer(topicName, broker string) wrappedKafka.Producer {
	if d.wrappedProducer == nil {
		d.wrappedProducer = producer.NewProducer(
			d.SyncProducer(broker),
			topicName,
			logger.Logger(),
		)
	}

	return d.wrappedProducer
}

func (d *diContainer) SyncProducer(broker string) sarama.SyncProducer {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			[]string{broker},
			producer.Config(),
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
