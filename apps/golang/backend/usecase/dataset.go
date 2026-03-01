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
