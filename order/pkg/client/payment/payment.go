package payment

import (
	"log"
	paymentV1 "shared/pkg/proto/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(serverPaymentAddress string) paymentV1.PaymentServiceClient {
	conn, err := grpc.NewClient(
		serverPaymentAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
	}

	return paymentV1.NewPaymentServiceClient(conn)
}
