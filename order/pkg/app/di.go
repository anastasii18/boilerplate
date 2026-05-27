package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"order/pkg"
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
	InventoryClient *inventory.Client
	PaymentClient   *payment.Client

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

func (d *diContainer) NewRepo(ctx context.Context, config *Config) (*db.Repository, error) {
	if d.Repo == nil {
		database, err := db.NewDB(ctx, config.DbUri)
		if err != nil {
			return nil, err
		}
		d.Repo = db.NewRepository(database)
	}

	return d.Repo, nil
}

func (d *diContainer) NewServer(ctx context.Context, config *Config) (*http.Server, error) {
	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(pkg.AuthMiddleWare)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	orderService, err := d.NewOrderService(ctx, config)
	if err != nil {
		return nil, err
	}

	a := api.New(orderService)
	paymentService, err := d.NewPaymentClient(config)
	if err != nil {
		return nil, err
	}
	// Определяем маршруты
	r.Route("/api/v1/orders", func(r chi.Router) {
		// Получить заказ по UUID
		r.Get("/{order_uuid}", a.GetOrderHandler())
		// Создание заказа
		r.Post("/", a.CreateOrderHandler())
		// Оплата заказа
		r.Post("/{order_uuid}/pay", a.PayOrderHandler(paymentService))
		// Отменить заказ
		r.Post("/{order_uuid}/cancel", a.CancelOrderHandler())
	})

	d.Server = &http.Server{
		Addr:              net.JoinHostPort(config.HttpHost, config.HttpPort),
		Handler:           r,
		ReadHeaderTimeout: config.ReadHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}
	return d.Server, nil
}

func (d *diContainer) NewOrderService(ctx context.Context, config *Config) (service.OrderService, error) {
	if d.OrderService == nil {
		repo, errRepo := d.NewRepo(ctx, config)
		if errRepo != nil {
			return nil, errRepo
		}

		inventoryService, errInventory := d.NewInventoryClient(config)
		if errInventory != nil {
			return nil, errInventory
		}

		producerService, producerErr := d.ProducerService(config.ProduceTopicName, config.KafkaBroker)
		if producerErr != nil {
			return nil, producerErr
		}

		d.OrderService = service.NewService(repo, inventoryService, producerService)
	}

	return d.OrderService, nil
}

func (d *diContainer) NewInventoryClient(config *Config) (*inventory.Client, error) {
	if d.InventoryClient == nil {
		inventoryClient, err := inventory.NewClient(config.ServerInventoryAddress)
		if err != nil {
			return nil, err
		}
		d.InventoryClient = inventoryClient
	}

	return d.InventoryClient, nil
}

func (d *diContainer) NewPaymentClient(config *Config) (*payment.Client, error) {
	if d.PaymentClient == nil {
		paymentClient, err := payment.NewClient(config.ServerPaymentAddress)
		if err != nil {
			return nil, err
		}
		d.PaymentClient = paymentClient
	}

	return d.PaymentClient, nil
}

func (d *diContainer) ProducerService(topicName, broker string) (service.OrderProducerService, error) {
	if d.producerService == nil {
		wrappedProducer, err := d.WrappedProducer(topicName, broker)
		if err != nil {
			return nil, err
		}
		d.producerService = service.NewProducerService(wrappedProducer)
	}
	return d.producerService, nil
}

func (d *diContainer) ConsumerService(ctx context.Context, config *Config) (consumer.ShipAssembledService, error) {
	if d.consumerService == nil {
		consumerWrapped, err := d.WrappedConsumer(config.ConsumeTopicName, config.KafkaBroker, config.ConsumerGroupId)
		if err != nil {
			return nil, err
		}

		orderService, err := d.NewOrderService(ctx, config)
		if err != nil {
			return nil, err
		}

		d.consumerService = consumer.NewService(consumerWrapped, d.ShipAssembledDecoder(), orderService)
	}

	return d.consumerService, nil
}

func (d *diContainer) ShipAssembledDecoder() producer.ShipAssembledDecoder {
	if d.assembledDecoder == nil {
		d.assembledDecoder = producer.NewShipAssembledDecoder()
	}

	return d.assembledDecoder
}

func (d *diContainer) WrappedConsumer(topicName, broker, groupId string) (wrappedKafka.Consumer, error) {
	if d.wrappedConsumer == nil {
		consumerGroup, err := d.ConsumerGroup(broker, groupId)
		if err != nil {
			return nil, err
		}
		d.wrappedConsumer = wrappedKafkaConsumer.NewConsumer(
			consumerGroup,
			[]string{topicName},
			logger.Logger(),
		)
	}

	return d.wrappedConsumer, nil
}

func (d *diContainer) WrappedProducer(topicName, broker string) (wrappedKafka.Producer, error) {
	if d.wrappedProducer == nil {
		syncProducer, err := d.SyncProducer(broker)
		if err != nil {
			return nil, err
		}
		d.wrappedProducer = producer.NewProducer(
			syncProducer,
			topicName,
			logger.Logger(),
		)
	}

	return d.wrappedProducer, nil
}

func (d *diContainer) SyncProducer(broker string) (sarama.SyncProducer, error) {
	if d.syncProducer == nil {
		p, err := sarama.NewSyncProducer(
			[]string{broker},
			producer.Config(),
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
