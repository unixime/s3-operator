package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	s3v1beta1 "s3-operator/api/v1beta1"
)

type RustFS struct {
	client *s3.Client
}

func NewRustFS(endpoint, accessKeyID, secretAccessKey, region string) *RustFS {
	cfg := aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &RustFS{client: client}
}

func (r *RustFS) CreateBucket(ctx context.Context, bucket *s3v1beta1.Bucket) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket.Spec.Name),
	}

	if region := bucket.Spec.Region; region != "" && region != "us-east-1" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	if bucket.Spec.WithLock {
		input.ObjectLockEnabledForBucket = aws.Bool(true)
	}

	_, err := r.client.CreateBucket(ctx, input)
	if err != nil {
		if bucket.Spec.IgnoreExisting {
			var bae *s3types.BucketAlreadyExists
			var baoy *s3types.BucketAlreadyOwnedByYou
			if errors.As(err, &bae) || errors.As(err, &baoy) {
				return nil
			}
		}
		return fmt.Errorf("creating bucket %q: %w", bucket.Spec.Name, err)
	}

	return nil
}
