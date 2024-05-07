package k8sclient

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k K8sClient) ListPods(ctx context.Context, namespace string) (*v1.PodList, error) {
	res, err := k.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
