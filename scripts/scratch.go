package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	client := k8sClient()
	dynClient := k8sDynClient()

	err := createDeployment(
		client,
		"default",
		"echo",
		2,
		"ealen/echo-server",
	)
	if err != nil {
		slog.Error("deployment", "error", err)
	}
	err = createService(
		client,
		"default",
		"echo-svc",
		3000,
	)
	if err != nil {
		slog.Error("service", "error", err)
	}
	err = createDynamicIngressRoute(
		dynClient,
		"default",
		"echo-ingress",
		3000,
	)
	if err != nil {
		slog.Error("ingress", "error", err)
	}
}

func createDeployment(
	client *kubernetes.Clientset,
	ns string,
	deploymentName string,
	replicas int32,
	image string,
) error {
	dc := client.AppsV1().Deployments(ns)

	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									Name:          deploymentName,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 8888,
								},
							},
						},
					},
				},
			},
		},
	}

	slog.Info("deployment", "deployment", deploymentName, "status", "creating")
	result, err := dc.Create(context.Background(), dp, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	slog.Info("deployment", "deployment", result.GetObjectMeta().GetName(), "status", "created")
	return nil
}
func createService(client *kubernetes.Clientset, ns string, serviceName string, port int32) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"app": "echo",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 80},
				},
			},
			Selector: map[string]string{
				"app": "echo",
			},
		},
	}

	slog.Info("service", "service", serviceName, "status", "creating")
	service, err := client.CoreV1().Services(ns).Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	slog.Info("service", "service", service.GetObjectMeta().GetName(), "status", "created")
	return nil
}
func Ptr[T any](v T) *T {
	return &v
}

func createDynamicIngressRoute(
	client *dynamic.DynamicClient,
	ns string,
	ingressName string,
	port int32,
) error {
	ingress := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "traefik.io/v1alpha1",
			"kind":       "IngressRoute",
			"metadata": map[string]interface{}{
				"name":      ingressName,
				"namespace": ns,
			},
			"spec": map[string]interface{}{
				"entryPoints": []interface{}{
					"web",
				},
				"routes": []interface{}{
					map[string]interface{}{
						"match": "Host(`echo.k3s.lcl`) && PathPrefix(`/`)",
						"kind":  "Rule",
						"services": []interface{}{
							map[string]interface{}{
								"name":      "echo-svc",
								"port":      port,
								"namespace": "default",
								"kind":      "Service",
							},
						},
					},
				},
			},
		},
	}
	slog.Info("ingress", "ingress", ingressName, "status", "creating")
	result, err := client.Resource(schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutes",
	}).Namespace(ns).Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	slog.Info("ingress", "ingress", result.GetName(), "status", "created")
	return nil
}
func createIngress(client *kubernetes.Clientset, ns string, ingressName string, port int32) error {
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: ns,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: Ptr("traefik"),
			Rules: []networkingv1.IngressRule{
				{
					Host: "echo.k3s.lcl",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: Ptr(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "echo-svc",
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	slog.Info("ingress", "ingress", ingressName, "status", "creating")
	result, err := client.NetworkingV1().
		Ingresses(ingress.Namespace).
		Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	slog.Info("ingress", "ingress", result.GetObjectMeta().GetName(), "status", "created")
	return nil
}

func k8sClient() *kubernetes.Clientset {
	//userHomeDir, err := os.UserHomeDir()
	//if err != nil {
	//	fmt.Printf("error getting user home dir: %v\n", err)
	//	os.Exit(1)
	//}
	//kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfigPath := os.Getenv("KUBECONFIG")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("Error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}
	return clientset
}
func k8sDynClient() *dynamic.DynamicClient {
	//userHomeDir, err := os.UserHomeDir()
	//if err != nil {
	//	fmt.Printf("error getting user home dir: %v\n", err)
	//	os.Exit(1)
	//}
	//kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfigPath := os.Getenv("KUBECONFIG")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("Error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("error getting kubernetes config: %v\n", err)
		os.Exit(1)
	}
	return clientset
}
