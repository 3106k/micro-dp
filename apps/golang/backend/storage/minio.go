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

func parseEndpoint(raw string) (string, bool, error) {
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", false, err
		}
		if parsed.Host == "" {
			return "", false, fmt.Errorf("invalid endpoint: %q", raw)
		}
		return parsed.Host, strings.EqualFold(parsed.Scheme, "https"), nil
	}
	return raw, false, nil
}

func NewMinIOClient() (*MinIOClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	normalizedEndpoint, secure, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("minio endpoint: %w", err)
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

	client, err := minio.New(normalizedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	return &MinIOClient{client: client, bucket: bucket}, nil
}

// MinIOPresignClient generates presigned URLs for browser-direct uploads.
// Uses MINIO_PRESIGN_ENDPOINT for signature host matching when set.
type MinIOPresignClient struct {
	client *minio.Client
	bucket string
}

func NewMinIOPresignClient() (*MinIOPresignClient, error) {
	endpoint := os.Getenv("MINIO_PRESIGN_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("MINIO_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	signingEndpoint, secure, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("minio presign endpoint: %w", err)
	}
	region := os.Getenv("MINIO_REGION")
	if region == "" {
		region = "us-east-1"
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

	client, err := minio.New(signingEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
		Region: region,
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
