package inventory

import (
	"log"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(serverInventoryAddress string) inventoryV1.InventoryServiceClient {
	conn, err := grpc.NewClient(
		serverInventoryAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to connect: %v\n", err)
		return nil
	}

	return inventoryV1.NewInventoryServiceClient(conn)
}
