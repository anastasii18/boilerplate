package grpc

import (
	"context"
	"inventory/pkg/service"
	inventoryV1 "shared/pkg/proto/inventory/v1"
)

// реализует gRPC сервис для работы с деталями для сборки космических кораблей
type Api struct {
	inventoryV1.UnimplementedInventoryServiceServer
	inventoryService service.InventoryService
}

func New(service service.InventoryService) *Api {
	return &Api{
		inventoryService: service,
	}
}

// Возвращает информацию о детали по её UUID
func (a *Api) GetPart(ctx context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	part, ok := a.inventoryService.GetPart(ctx, req.GetUuid())

	return &inventoryV1.GetPartResponse{
		Part: NewPart(part),
	}, ok
}

// Возвращает список деталей с возможностью фильтрации
func (a *Api) GetListParts(ctx context.Context, req *inventoryV1.GetListPartsRequest) (*inventoryV1.GetListPartsResponse, error) {
	parts, err := a.inventoryService.GetParts(ctx, NewPartSearch(req.GetFilter()))
	if err != nil {
		return nil, err
	}
	var result []*inventoryV1.Part
	for _, part := range parts {
		result = append(result, NewPart(part))
	}

	return &inventoryV1.GetListPartsResponse{
		Parts: result,
	}, nil
}
