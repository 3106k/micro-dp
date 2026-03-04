package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type DatasetService struct {
	datasets domain.DatasetRepository
	minio    *storage.MinIOClient
}

func NewDatasetService(datasets domain.DatasetRepository, minio *storage.MinIOClient) *DatasetService {
	return &DatasetService{datasets: datasets, minio: minio}
}

func (s *DatasetService) Get(ctx context.Context, id string) (*domain.Dataset, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.datasets.FindByID(ctx, tenantID, id)
}

func (s *DatasetService) List(ctx context.Context, filter domain.DatasetListFilter) ([]domain.Dataset, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.datasets.ListByTenant(ctx, tenantID, filter)
}

type UpdateColumnInput struct {
	Name         string
	Description  string
	SemanticType string
	Tags         []string
}

func (s *DatasetService) UpdateColumns(ctx context.Context, id string, inputs []UpdateColumnInput) (*domain.Dataset, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	ds, err := s.datasets.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	cols, err := ds.ParseColumns()
	if err != nil {
		return nil, fmt.Errorf("parse columns: %w", err)
	}

	colMap := make(map[string]int, len(cols))
	for i, c := range cols {
		colMap[c.Name] = i
	}

	for _, input := range inputs {
		idx, ok := colMap[input.Name]
		if !ok {
			return nil, fmt.Errorf("column %q: %w", input.Name, domain.ErrColumnNotFound)
		}
		cols[idx].Description = input.Description
		cols[idx].SemanticType = input.SemanticType
		cols[idx].Tags = input.Tags
	}

	if err := ds.SetColumns(cols); err != nil {
		return nil, fmt.Errorf("set columns: %w", err)
	}
	if err := s.datasets.Update(ctx, ds); err != nil {
		return nil, fmt.Errorf("update dataset: %w", err)
	}
	return ds, nil
}

func (s *DatasetService) GetRows(ctx context.Context, id string, limit, offset int) (*storage.ParquetRowsResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	ds, err := s.datasets.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if ds.StoragePath == "" {
		return nil, fmt.Errorf("dataset has no storage path")
	}
	if s.minio == nil {
		return nil, fmt.Errorf("storage client not available")
	}

	tmpDir, err := os.MkdirTemp("", "micro-dp-dataset-preview-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "data.parquet")
	if err := s.minio.DownloadToFile(ctx, ds.StoragePath, localPath); err != nil {
		return nil, fmt.Errorf("download parquet: %w", err)
	}

	return storage.ReadParquetRows(ctx, localPath, limit, offset)
}
