package k8sclient

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k K8sClient) ListServices(ctx context.Context, namespace string) (*v1.ServiceList, error) {
	res, err := k.Client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (k K8sClient) GetService(ctx context.Context, name, namespace string) (*v1.Service, error) {
	res, err := k.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
