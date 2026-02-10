package grpc

import (
	"context"
	"inventory/pkg/db"
	inventoryV1 "shared/pkg/proto/inventory/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// реализует gRPC сервис для работы с деталями для сборки космических кораблей
type InventoryService struct {
	inventoryV1.UnimplementedInventoryServiceServer
	dbc *db.DB
}

func NewInventoryService(dbc *db.DB) *InventoryService {
	return &InventoryService{
		dbc: dbc,
	}
}

// Возвращает информацию о детали по её UUID
func (s *InventoryService) GetPart(ctx context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	part, ok := s.dbc.GetPart(req.GetUuid())

	if !ok {
		return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", req.GetUuid())
	}

	return &inventoryV1.GetPartResponse{
		Part: part,
	}, nil
}

// Возвращает список деталей с возможностью фильтрации
func (s *InventoryService) GetListParts(ctx context.Context, req *inventoryV1.GetListPartsRequest) (*inventoryV1.GetListPartsResponse, error) {
	filter := req.GetFilter()
	listParts := s.dbc.GetParts(db.NewFilter(filter.GetCategories(), filter.GetUuids(), filter.GetNames(), filter.GetManufacturerCountries(), filter.GetTags()))

	values := make([]*inventoryV1.Part, 0, len(listParts))
	for _, v := range listParts {
		values = append(values, v)
	}

	return &inventoryV1.GetListPartsResponse{
		Parts: values,
	}, nil
}
