package payment

import (
	"context"
	"fmt"
	"order/pkg/service"
	paymentV1 "shared/pkg/proto/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	paymentClient paymentV1.PaymentServiceClient
}

func NewClient(serverPaymentAddress string) (*Client, error) {
	conn, err := grpc.NewClient(
		serverPaymentAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return &Client{}, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		paymentClient: paymentV1.NewPaymentServiceClient(conn),
	}, nil
}

func (c *Client) PayOrder(ctx context.Context, orderId, userId string, paymentMethod service.OrderPaymentMethod) (*paymentV1.PayOrderResponse, error) {
	payOrderRequest := &paymentV1.PayOrderRequest{
		OrderUuid:     orderId,
		UserUuid:      userId,
		PaymentMethod: paymentV1.PaymentMethod(paymentMethod),
	}

	payOrderResponse, err := c.paymentClient.PayOrder(ctx, payOrderRequest)
	if err != nil {
		return nil, err
	}
	return payOrderResponse, nil
}
