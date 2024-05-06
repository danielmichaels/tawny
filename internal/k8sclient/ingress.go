package k8sclient

import (
	"context"
	"fmt"

	assets "github.com/danielmichaels/tawny"
	"github.com/rs/zerolog/log"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
)

const (
	EntryPointWeb       = "web"
	EntryPointWebSecure = "websecure"
)

func ingressRouteNameGenerator(name, namespace, entryPoint string) string {
	a := fmt.Sprintf("%s-%s-%s-%s", namespace, name, entryPoint, DefaultIngressRouteName)
	fmt.Printf("Name: %q\n", a)
	return fmt.Sprintf("%s-%s-%s-%s", namespace, name, entryPoint, DefaultIngressRouteName)
}

func (k K8sClient) ListDomains(
	ctx context.Context,
	namespace string,
) (*unstructured.UnstructuredList, error) {
	// todo make typed; change name to list ingresses
	gvr := NewGVR("traefik.io/v1alpha1/ingressroutes")
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

func (k K8sClient) CreateIngress(
	ctx context.Context,
	name, namespace string,
	opts ...IngressRouteOption,
) (*traefikv1alpha1.IngressRoute, error) {
	ingressRoute := NewIngressRoute(name, namespace, opts...)
	ingress, err := k.tClient.TraefikV1alpha1().
		IngressRoutes(namespace).
		Create(ctx, ingressRoute, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return ingress, nil
}

func (k K8sClient) UpdateIngress(
	ctx context.Context,
	name, namespace string,
	opts ...IngressRouteOption,
) (*traefikv1alpha1.IngressRoute, error) {
	var result *traefikv1alpha1.IngressRoute
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		res, err := k.tClient.TraefikV1alpha1().IngressRoutes(namespace).Get(ctx, ingressRouteNameGenerator(name, namespace, "websecure"), metav1.GetOptions{})
		if err != nil {
			return err
		}
		update := NewIngressRoute(name, namespace, opts...)
		res.Spec = update.Spec
		result, err = k.tClient.TraefikV1alpha1().IngressRoutes(namespace).Update(ctx, res, metav1.UpdateOptions{})
		return err
	}); retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

type IngressRouteOption func(*traefikv1alpha1.IngressRoute)

func WithIngressRouteEntryPoint(entryType string) IngressRouteOption {
	switch entryType {
	case EntryPointWeb:
	case EntryPointWebSecure:
	default:
		log.Error().Msgf("Invalid entry point type: %s", entryType)
		return nil
	}
	return func(i *traefikv1alpha1.IngressRoute) {
		i.Spec.EntryPoints = append(i.Spec.EntryPoints, entryType)
	}
}

func WithIngressRouteTLS(secretName string) IngressRouteOption {
	tls := &traefikv1alpha1.TLS{
		SecretName: secretName,
	}
	fmt.Println(tls)
	return func(i *traefikv1alpha1.IngressRoute) {
		i.Spec.TLS = tls
	}
}

func WithIngressRouteRule(
	match, svcName, svcNamespace string,
	middlewareRefs []string,
	svcPort int32,
) IngressRouteOption {
	var middlewares []traefikv1alpha1.MiddlewareRef
	if len(middlewareRefs) != 0 {
		for _, mw := range middlewareRefs {
			middlewares = append(middlewares, traefikv1alpha1.MiddlewareRef{
				Name: mw,
			})
		}
	}
	route := traefikv1alpha1.Route{
		Match: fmt.Sprintf("Host(`%s`)", match),
		Kind:  "Rule",
		Services: []traefikv1alpha1.Service{
			{LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
				Name:      svcName,
				Kind:      "Service",
				Namespace: svcNamespace,
				Port:      intstr.IntOrString{IntVal: svcPort},
			}},
		},
		Middlewares: middlewares,
	}
	return func(i *traefikv1alpha1.IngressRoute) {
		i.Spec.Routes = append(i.Spec.Routes, route)
	}
}

func NewIngressRoute(
	name, namespace string,
	opts ...IngressRouteOption,
) *traefikv1alpha1.IngressRoute {
	labels := CreateLabels(WithName(name), WithComponent("ingressroute"))
	if namespace == assets.AppName {
		labels = CreateLabels(WithName(name), WithComponent("ingressroute"), WithCoreLabel(true))
	}
	i := &traefikv1alpha1.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "traefik.io/v1alpha1",
			Kind:       "IngressRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressRouteNameGenerator(name, namespace, EntryPointWeb),
			Namespace: namespace,
			Labels:    labels,
		},
	}
	for _, opt := range opts {
		opt(i)
	}

	if i.Spec.TLS.SecretName != "" {
		i.ObjectMeta.Name = ingressRouteNameGenerator(name, namespace, EntryPointWebSecure)
	}
	return i
}
