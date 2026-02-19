package service

import (
	"inventory/pkg/model"
	"inventory/pkg/repository"
	"reflect"
	"testing"
)

func TestGetPart(t *testing.T) {
	s := Service{
		InventoryRepository: repository.NewRepository(),
	}
	oneId := "fbb05498-4db6-48c8-b945-3e56f4e5ad04"
	one, _ := s.InventoryRepository.GetPart(oneId)

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Part
		wantErr bool
	}{
		{
			name:    "Success",
			args:    args{oneId},
			want:    one,
			wantErr: false,
		},
		{
			name:    "Fail",
			args:    args{"7ccd152a-efaa-4813-9519-bd9e9d4e7a06"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.GetPart(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPart() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetParts(t *testing.T) {
	s := Service{
		InventoryRepository: repository.NewRepository(),
	}
	oneId := "fbb05498-4db6-48c8-b945-3e56f4e5ad04"
	one, _ := s.InventoryRepository.GetPart(oneId)
	twoId := "bf802b57-1c7d-41ff-9cb7-ee43dbadbf98"
	two, _ := s.InventoryRepository.GetPart(twoId)
	threeId := "29a9ab94-c814-4828-9a02-b96598dbe299"
	three, _ := s.InventoryRepository.GetPart(threeId)

	type args struct {
		filter model.Filter
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Part
	}{
		{
			name: "Get parts for categories",
			args: args{model.NewFilter([]model.Category{model.CATEGORY_ENGINE}, nil, nil, nil, nil)},
			want: map[string]*model.Part{
				twoId:   two,
				threeId: three,
			},
		},
		{
			name: "Get parts for ids",
			args: args{model.NewFilter(nil, []string{"fbb05498-4db6-48c8-b945-3e56f4e5ad04"}, nil, nil, nil)},
			want: map[string]*model.Part{
				oneId: one,
			},
		},
		{
			name: "Get parts for names",
			args: args{model.NewFilter(nil, nil, []string{"two two"}, nil, nil)},
			want: map[string]*model.Part{
				twoId: two,
			},
		},
		{
			name: "Get parts for countries",
			args: args{model.NewFilter(nil, nil, nil, []string{"Moscow"}, nil)},
			want: map[string]*model.Part{
				oneId:   one,
				threeId: three,
			},
		},
		{
			name: "Get parts for tags",
			args: args{model.NewFilter(nil, nil, nil, nil, []string{"engine", "Moscow"})},
			want: map[string]*model.Part{
				threeId: three,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := s.GetParts(tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetParts() = %v, want %v", got, tt.want)
			}
		})
	}
}
