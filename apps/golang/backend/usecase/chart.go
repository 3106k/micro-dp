package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type ChartService struct {
	charts   domain.ChartRepository
	datasets domain.DatasetRepository
	minio    *storage.MinIOClient
}

func NewChartService(charts domain.ChartRepository, datasets domain.DatasetRepository, minio *storage.MinIOClient) *ChartService {
	return &ChartService{charts: charts, datasets: datasets, minio: minio}
}

type ChartDataResult struct {
	Labels   []string
	Datasets []ChartDatasetResult
}

type ChartDatasetResult struct {
	Label string
	Data  []float32
}

func (s *ChartService) Create(ctx context.Context, name, chartType, datasetID, measure, dimension string, configJSON *string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify dataset exists and belongs to tenant
	if _, err := s.datasets.FindByID(ctx, tenantID, datasetID); err != nil {
		return nil, err
	}

	c := &domain.Chart{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       name,
		ChartType:  chartType,
		DatasetID:  datasetID,
		Measure:    measure,
		Dimension:  dimension,
		ConfigJSON: configJSON,
	}
	if err := s.charts.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.charts.FindByID(ctx, tenantID, c.ID)
}

func (s *ChartService) Get(ctx context.Context, id string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.charts.FindByID(ctx, tenantID, id)
}

func (s *ChartService) List(ctx context.Context) ([]domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.charts.ListByTenant(ctx, tenantID)
}

func (s *ChartService) Update(ctx context.Context, id, name, chartType, datasetID, measure, dimension string, configJSON *string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	c, err := s.charts.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Verify dataset exists if changed
	if datasetID != c.DatasetID {
		if _, err := s.datasets.FindByID(ctx, tenantID, datasetID); err != nil {
			return nil, err
		}
	}

	c.Name = name
	c.ChartType = chartType
	c.DatasetID = datasetID
	c.Measure = measure
	c.Dimension = dimension
	c.ConfigJSON = configJSON

	if err := s.charts.Update(ctx, c); err != nil {
		return nil, err
	}
	return s.charts.FindByID(ctx, tenantID, id)
}

func (s *ChartService) Delete(ctx context.Context, id string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	if _, err := s.charts.FindByID(ctx, tenantID, id); err != nil {
		return err
	}
	return s.charts.Delete(ctx, tenantID, id)
}

func (s *ChartService) GetData(ctx context.Context, chartID, period string, startDate, endDate *time.Time) (*ChartDataResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	chart, err := s.charts.FindByID(ctx, tenantID, chartID)
	if err != nil {
		return nil, err
	}

	dataset, err := s.datasets.FindByID(ctx, tenantID, chart.DatasetID)
	if err != nil {
		return nil, err
	}

	// List parquet files from MinIO
	prefix := dataset.StoragePath
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	keys, err := s.minio.ListObjectKeys(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("list parquet files: %w", err)
	}

	// Filter to .parquet files only
	var parquetKeys []string
	for _, k := range keys {
		if strings.HasSuffix(k, ".parquet") {
			parquetKeys = append(parquetKeys, k)
		}
	}

	if len(parquetKeys) == 0 {
		return &ChartDataResult{Labels: []string{}, Datasets: []ChartDatasetResult{}}, nil
	}

	// Download parquet files to temp dir
	tmpDir, err := os.MkdirTemp("", "micro-dp-chart-data-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for i, key := range parquetKeys {
		dest := filepath.Join(tmpDir, fmt.Sprintf("data_%d.parquet", i))
		if err := s.minio.DownloadToFile(ctx, key, dest); err != nil {
			return nil, fmt.Errorf("download %s: %w", key, err)
		}
	}

	// Open in-memory DuckDB
	ddb, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer ddb.Close()

	glob := filepath.Join(tmpDir, "*.parquet")
	dimension := quoteIdent(chart.Dimension)
	measure := quoteIdent(chart.Measure)

	// Build query with optional period filter
	start, end := periodToDateRange(period, startDate, endDate)
	var query string
	if start != nil {
		endVal := time.Now().UTC()
		if end != nil {
			endVal = *end
		}
		query = fmt.Sprintf(
			`SELECT CAST(%s AS VARCHAR) AS label, SUM(CAST(%s AS DOUBLE)) AS value FROM read_parquet('%s', union_by_name=true) WHERE TRY_CAST(%s AS DATE) BETWEEN '%s' AND '%s' GROUP BY label ORDER BY label`,
			dimension, measure, glob, dimension,
			start.Format("2006-01-02"), endVal.Format("2006-01-02"),
		)
	} else {
		query = fmt.Sprintf(
			`SELECT CAST(%s AS VARCHAR) AS label, SUM(CAST(%s AS DOUBLE)) AS value FROM read_parquet('%s', union_by_name=true) GROUP BY label ORDER BY label`,
			dimension, measure, glob,
		)
	}

	rows, err := ddb.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query chart data: %w", err)
	}
	defer rows.Close()

	var labels []string
	var data []float32
	for rows.Next() {
		var label string
		var value float64
		if err := rows.Scan(&label, &value); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		labels = append(labels, label)
		data = append(data, float32(value))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if labels == nil {
		labels = []string{}
	}
	if data == nil {
		data = []float32{}
	}

	return &ChartDataResult{
		Labels: labels,
		Datasets: []ChartDatasetResult{
			{Label: chart.Measure, Data: data},
		},
	}, nil
}

func periodToDateRange(period string, startDate, endDate *time.Time) (start, end *time.Time) {
	now := time.Now().UTC()
	switch period {
	case "last_7_days":
		s := now.AddDate(0, 0, -7)
		return &s, &now
	case "last_30_days":
		s := now.AddDate(0, 0, -30)
		return &s, &now
	case "last_90_days":
		s := now.AddDate(0, 0, -90)
		return &s, &now
	case "custom":
		return startDate, endDate
	default:
		return nil, nil
	}
}

// quoteIdent quotes a SQL identifier with double quotes, escaping embedded double quotes.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
