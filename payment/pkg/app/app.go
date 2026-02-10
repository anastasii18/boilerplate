package app

import (
	"fmt"
	"log"
	"net"
	rpc "payment/pkg/grpc"
	paymentV1 "shared/pkg/proto/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Port int
}

type App struct {
	Config         *Config
	paymentService *rpc.PaymentService
	Server         *grpc.Server
}

func New(config *Config) *App {
	return &App{Config: config, paymentService: rpc.NewPaymentService(), Server: grpc.NewServer()}
}

func (a *App) createServer() {
	paymentV1.RegisterPaymentServiceServer(a.Server, a.paymentService)
}

func (a *App) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Config.Port))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}

	// Регистрируем наш сервис
	a.createServer()
	// Включаем рефлексию для отладки
	reflection.Register(a.Server)

	go func() {
		log.Printf("🚀 gRPC server listening on %d\n", a.Config.Port)
		err = a.Server.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()
}
