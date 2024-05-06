package k8sclient

import (
	"context"
	"fmt"

	assets "github.com/danielmichaels/tawny"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func serviceNameGenerator(name string) string {
	return fmt.Sprintf(DefaultServiceName, assets.AppName, name)
}

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
