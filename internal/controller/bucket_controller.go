/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	s3v1beta1 "s3-operator/api/v1beta1"
	s3client "s3-operator/internal/client"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	S3Clients map[string]s3client.S3Client
}

// +kubebuilder:rbac:groups=s3.storage.io,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.storage.io,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.storage.io,resources=buckets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	bucket, err := r.readBucketDefinition(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Bucket resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Bucket")
		return ctrl.Result{}, err
	}

	impl := bucket.Annotations["s3/implementation"]
	s3c := r.s3ClientForImpl(impl)
	if s3c == nil {
		log.Info("No S3 client configured for implementation, skipping", "implementation", impl)
		return ctrl.Result{}, nil
	}

	if err := s3c.CreateBucket(ctx, bucket); err != nil {
		log.Error(err, "Failed to create bucket", "bucket", bucket.Spec.Name, "implementation", impl)
		return ctrl.Result{}, fmt.Errorf("creating bucket %q on %q: %w", bucket.Spec.Name, impl, err)
	}

	log.Info("Bucket reconciled successfully", "bucket", bucket.Spec.Name, "implementation", impl)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&s3v1beta1.Bucket{}).
		Named("bucket").
		Complete(r)
}

func (r *BucketReconciler) s3ClientForImpl(impl string) s3client.S3Client {
	if r.S3Clients == nil {
		return nil
	}
	switch impl {
	case "garage", "minio", "rustfs":
		return r.S3Clients[impl]
	default:
		return nil
	}
}

func (r *BucketReconciler) readBucketDefinition(ctx context.Context, req ctrl.Request) (*s3v1beta1.Bucket, error) {

	b := &s3v1beta1.Bucket{}

	if err := r.Get(ctx, req.NamespacedName, b); err != nil {
		return nil, err
	}

	return b, nil
}
