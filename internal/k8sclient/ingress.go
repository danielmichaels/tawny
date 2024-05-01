package k8sclient

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k K8sClient) ListDomains(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	res, err := k.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutes",
	}).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
