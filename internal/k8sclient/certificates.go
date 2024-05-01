package k8sclient

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k K8sClient) ListCertificates(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	res, err := k.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (k K8sClient) GetCertificates(ctx context.Context, name, namespace string) (*v1.Service, error) {
	res, err := k.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
