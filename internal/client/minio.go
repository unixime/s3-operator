package client

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	s3v1beta1 "s3-operator/api/v1beta1"
)

type MinIO struct {
	client *minio.Client
}

func NewMinIO(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinIO, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing minio client: %w", err)
	}

	return &MinIO{client: client}, nil
}

func (m *MinIO) CreateBucket(ctx context.Context, bucket *s3v1beta1.Bucket) error {
	opts := minio.MakeBucketOptions{
		Region:        bucket.Spec.Region,
		ObjectLocking: bucket.Spec.WithLock,
	}

	err := m.client.MakeBucket(ctx, bucket.Spec.Name, opts)
	if err != nil {
		if bucket.Spec.IgnoreExisting {
			exists, existsErr := m.client.BucketExists(ctx, bucket.Spec.Name)
			if existsErr == nil && exists {
				return nil
			}
		}
		return fmt.Errorf("creating bucket %q: %w", bucket.Spec.Name, err)
	}

	return nil
}
