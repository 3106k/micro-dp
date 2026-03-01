package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
// MinIOPresignClient generates presigned URLs for browser-direct uploads.
// Uses MINIO_PRESIGN_ENDPOINT (external) for the minio-go client so that
// the S3 SigV4 signature is computed against the host the browser will use.
// PresignedPutObject only computes signatures locally — no actual network
// connection to MinIO is needed.
type MinIOPresignClient struct {
	client *minio.Client
	bucket string
}

func NewMinIOPresignClient() (*MinIOPresignClient, error) {
	// Use external endpoint so presigned URL signatures match the host
	// that browsers/clients will use.
	endpoint := os.Getenv("MINIO_PRESIGN_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("MINIO_ENDPOINT")
	}
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
		Region: "us-east-1", // explicit region avoids bucket location lookup
	})
	if err != nil {
		return nil, fmt.Errorf("minio presign client: %w", err)
	}

	return &MinIOPresignClient{
		client: client,
		bucket: bucket,
	}, nil
}

func (m *MinIOPresignClient) GeneratePresignedPutURL(ctx context.Context, objectKey, contentType string, expiry time.Duration) (string, time.Time, error) {
	presignedURL, err := m.client.PresignedPutObject(ctx, m.bucket, objectKey, expiry)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presigned put: %w", err)
	}

	expiresAt := time.Now().Add(expiry)
	return presignedURL.String(), expiresAt, nil
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
