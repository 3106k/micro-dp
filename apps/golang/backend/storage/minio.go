package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client *minio.Client
	bucket string
}

func NewMinIOClient() (*MinIOClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "micro-dp"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	return &MinIOClient{client: client, bucket: bucket}, nil
}

// MinIOPresignClient generates presigned URLs for browser-direct uploads.
// The minio-go client connects via MINIO_ENDPOINT (Docker-internal), then
// the generated URL's host is rewritten to MINIO_PRESIGN_ENDPOINT (external)
// so that browsers can reach MinIO from outside Docker.
type MinIOPresignClient struct {
	client           *minio.Client
	bucket           string
	presignEndpoint  string // external host:port for URL rewriting
	internalEndpoint string // internal host:port used by minio-go
}

func NewMinIOPresignClient() (*MinIOPresignClient, error) {
	internalEndpoint := os.Getenv("MINIO_ENDPOINT")
	if internalEndpoint == "" {
		internalEndpoint = "localhost:9000"
	}
	presignEndpoint := os.Getenv("MINIO_PRESIGN_ENDPOINT")
	if presignEndpoint == "" {
		presignEndpoint = internalEndpoint
	}
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "micro-dp"
	}

	client, err := minio.New(internalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio presign client: %w", err)
	}

	return &MinIOPresignClient{
		client:           client,
		bucket:           bucket,
		presignEndpoint:  presignEndpoint,
		internalEndpoint: internalEndpoint,
	}, nil
}

func (m *MinIOPresignClient) GeneratePresignedPutURL(ctx context.Context, objectKey, contentType string, expiry time.Duration) (string, time.Time, error) {
	presignedURL, err := m.client.PresignedPutObject(ctx, m.bucket, objectKey, expiry)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presigned put: %w", err)
	}

	// Rewrite internal host to external presign endpoint if different.
	result := presignedURL.String()
	if m.presignEndpoint != m.internalEndpoint {
		parsed, err := url.Parse(result)
		if err == nil {
			parsed.Host = m.presignEndpoint
			if !strings.Contains(m.presignEndpoint, "://") {
				parsed.Scheme = "http"
			}
			result = parsed.String()
		}
	}

	expiresAt := time.Now().Add(expiry)
	return result, expiresAt, nil
}

func (m *MinIOClient) DownloadToFile(ctx context.Context, objectKey, destPath string) error {
	if err := m.client.FGetObject(ctx, m.bucket, objectKey, destPath, minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("download to file: %w", err)
	}
	return nil
}

func (m *MinIOClient) PutParquet(ctx context.Context, objectKey string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("put parquet: %w", err)
	}
	return nil
}
