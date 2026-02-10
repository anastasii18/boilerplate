package main

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	inventoryV1 "shared/pkg/proto/inventory/v1"
)

const grpcPort = 50051

// реализует gRPC сервис для работы с деталями для сборки космических кораблей
type inventoryService struct {
	inventoryV1.UnimplementedInventoryServiceServer

	mu    sync.RWMutex
	parts map[string]*inventoryV1.Part
}

// Возвращает информацию о детали по её UUID
func (s *inventoryService) GetPart(ctx context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := s.parts[req.GetUuid()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", req.GetUuid())
	}

	return &inventoryV1.GetPartResponse{
		Part: part,
	}, nil
}

// Возвращает список деталей с возможностью фильтрации
func (s *inventoryService) GetListParts(ctx context.Context, req *inventoryV1.GetListPartsRequest) (*inventoryV1.GetListPartsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter := req.GetFilter()

	parts := make(map[string]*inventoryV1.Part)
	maps.Copy(parts, s.parts)

	if filter.GetUuids() != nil {
		for key := range parts {
			if !slices.Contains(filter.GetUuids(), key) {
				delete(parts, key)
			}
		}
	}

	if filter.GetNames() != nil {
		for key, value := range parts {
			if !slices.Contains(filter.GetNames(), value.GetName()) {
				delete(parts, key)
			}
		}
	}

	if filter.GetCategories() != nil {
		for key, value := range parts {
			if !slices.Contains(filter.GetCategories(), value.GetCategory()) {
				delete(parts, key)
			}
		}
	}

	if filter.GetManufacturerCountries() != nil {
		for key, value := range parts {
			if !slices.Contains(filter.GetManufacturerCountries(), value.GetManufacturer().Country) {
				delete(parts, key)
			}
		}
	}

	if filter.GetTags() != nil {
		for key, value := range parts {
			if !isIntersect(filter.GetTags(), value.GetTags()) {
				delete(parts, key)
			}
		}
	}

	values := make([]*inventoryV1.Part, 0, len(parts))
	for _, v := range parts {
		values = append(values, v)
	}

	return &inventoryV1.GetListPartsResponse{
		Parts: values,
	}, nil
}

func isIntersect(slice1, slice2 []string) bool {
	for _, i := range slice1 {
		if !slices.Contains(slice2, i) {
			return false
		}
	}
	return true
}

func CreatePartsMap() map[string]*inventoryV1.Part {
	parts := make(map[string]*inventoryV1.Part)

	parts["fbb05498-4db6-48c8-b945-3e56f4e5ad04"] = &inventoryV1.Part{
		Uuid:          "fbb05498-4db6-48c8-b945-3e56f4e5ad04",
		Name:          "test name",
		Description:   "test description",
		Price:         112.33,
		StockQuantity: 38,
		Category:      inventoryV1.Category_CATEGORY_FUEL,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
		Tags:          []string{"fuel", "Moscow"},
	}

	parts["bf802b57-1c7d-41ff-9cb7-ee43dbadbf98"] = &inventoryV1.Part{
		Uuid:          "bf802b57-1c7d-41ff-9cb7-ee43dbadbf98",
		Name:          "two two",
		Description:   "test description",
		Price:         45.45,
		StockQuantity: 7,
		Category:      inventoryV1.Category_CATEGORY_ENGINE,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Rostov", Website: "https://rostov.com"},
		Tags:          []string{"engine", "Rostov"},
	}

	parts["29a9ab94-c814-4828-9a02-b96598dbe299"] = &inventoryV1.Part{
		Uuid:          "29a9ab94-c814-4828-9a02-b96598dbe299",
		Name:          "three three",
		Description:   "test description",
		Price:         66.77,
		StockQuantity: 90,
		Category:      inventoryV1.Category_CATEGORY_ENGINE,
		Manufacturer:  &inventoryV1.Manufacturer{Name: "test name", Country: "Moscow", Website: "https://moscow.com"},
		Tags:          []string{"engine", "Moscow"},
	}

	return parts
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			log.Printf("failed to close listener: %v\n", cerr)
		}
	}()

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис
	service := &inventoryService{parts: CreatePartsMap()}

	inventoryV1.RegisterInventoryServiceServer(s, service)

	// Включаем рефлексию для отладки
	reflection.Register(s)

	go func() {
		log.Printf("🚀 gRPC server listening on %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("✅ Server stopped")
}
