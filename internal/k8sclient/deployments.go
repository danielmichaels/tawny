package k8sclient

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k K8sClient) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	res, err := k.Client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (k K8sClient) GetDeployment(ctx context.Context, name, namespace string) (*appsv1.Deployment, error) {
	res, err := k.Client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
