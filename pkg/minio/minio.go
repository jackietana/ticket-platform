package pkgMinio

import (
	pkgConfig "github.com/jackietana/ticket-platform/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(cfg *pkgConfig.MinioConfig) (*minio.Client, error) {
	client, err := minio.New(cfg.InternalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.User, cfg.Pass, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
