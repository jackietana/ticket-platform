package repository

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

const BUCKET_NAME = "tickets"

type MinioStorage struct {
	minioClient *minio.Client
}

func NewMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{client}
}

func (s *MinioStorage) InitStorage(ctx context.Context) error {
	err := s.minioClient.MakeBucket(ctx, BUCKET_NAME, minio.MakeBucketOptions{})
	if err != nil {
		exist, errBucketExist := s.minioClient.BucketExists(ctx, BUCKET_NAME)
		if errBucketExist == nil && exist {
			log.Printf("bucket %s already exists", BUCKET_NAME)
			return nil
		} else {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	log.Printf("successfully created %s", BUCKET_NAME)
	return nil
}

func (s *MinioStorage) UploadTicket(ctx context.Context, filename string, content []byte) error {
	reader := bytes.NewReader(content)
	size := int64(len(content))

	_, err := s.minioClient.PutObject(ctx, BUCKET_NAME, filename, reader, size, minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return fmt.Errorf("failed to put object to minio: %w", err)
	}

	return nil
}
