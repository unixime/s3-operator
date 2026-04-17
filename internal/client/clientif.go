package client

import (
	"context"
	s3v1beta1 "s3-operator/api/v1beta1"
)

type S3Client interface {
	CreateBucket(ctx context.Context, bucket *s3v1beta1.Bucket) error
}
