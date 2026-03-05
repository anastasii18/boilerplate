package inventory

import (
	"context"
	"fmt"
	inventoryApi "inventory/pkg/grpc"
	inventoryService "inventory/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	serviceClient inventoryV1.InventoryServiceClient
}

func NewClient(serverInventoryAddress string) (Client, error) {
	conn, err := grpc.NewClient(
		serverInventoryAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return Client{}, fmt.Errorf("failed to connect: %w", err)
	}

	return Client{inventoryV1.NewInventoryServiceClient(conn)}, nil
}

func (c *Client) GetListParts(ctx context.Context, uuids []string) (*inventoryV1.GetListPartsResponse, error) {
	getListPartsRequest := &inventoryV1.GetListPartsRequest{
		Filter: &inventoryV1.PartsFilter{
			Uuids: uuids,
		},
	}

	parts, err := c.serviceClient.GetListParts(ctx, getListPartsRequest)
	if err != nil {
		return nil, err
	}
	return parts, nil
}

func (c *Client) GetInventoryParts(parts *inventoryV1.GetListPartsResponse) []*inventoryService.Part {
	var newParts []*inventoryService.Part
	for _, part := range parts.Parts {
		newParts = append(newParts, inventoryApi.NewServicePart(part))
	}

	return newParts
}

func (c *Client) GetInventoryPartsForIDs(ctx context.Context, partUuids []string) ([]*inventoryService.Part, error) {
	if partUuids == nil {
		return []*inventoryService.Part{}, nil
	}

	parts, err := c.GetListParts(ctx, partUuids)

	if err != nil {
		return nil, fmt.Errorf("ошибка получения деталей: %w", err)
	}
	newParts := c.GetInventoryParts(parts)

	return newParts, nil
}
