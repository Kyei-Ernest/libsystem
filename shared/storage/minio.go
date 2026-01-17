package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig holds MinIO connection configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

// MinIOClient wraps minio.Client with helper methods
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	ctx        context.Context
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(config *MinIOConfig) (*MinIOClient, error) {
	// Initialize minio client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()

	// Ensure bucket exists
	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		// Create bucket if it doesn't exist
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOClient{
		client:     client,
		bucketName: config.BucketName,
		ctx:        ctx,
	}, nil
}

// UploadFile uploads a file to MinIO
func (m *MinIOClient) UploadFile(objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(
		m.ctx,
		m.bucketName,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

// DownloadFile retrieves a file from MinIO
func (m *MinIOClient) DownloadFile(objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(m.ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return object, nil
}

// DeleteFile removes a file from MinIO
func (m *MinIOClient) DeleteFile(objectName string) error {
	err := m.client.RemoveObject(m.ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetPresignedURL generates a pre-signed URL for temporary access
func (m *MinIOClient) GetPresignedURL(objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(m.ctx, m.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// FileExists checks if a file exists in MinIO
func (m *MinIOClient) FileExists(objectName string) (bool, error) {
	_, err := m.client.StatObject(m.ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "object not found"
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

// GetFileInfo retrieves metadata about a file
func (m *MinIOClient) GetFileInfo(objectName string) (*minio.ObjectInfo, error) {
	info, err := m.client.StatObject(m.ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return &info, nil
}

// Close closes the MinIO client connection
func (m *MinIOClient) Close() error {
	// MinIO client doesn't need explicit closing
	return nil
}
