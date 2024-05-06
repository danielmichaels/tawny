package k8sclient

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/retry"

	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	assets "github.com/danielmichaels/tawny"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k K8sClient) ListCertificates(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
func CreateCertSecretName(name string) string {
	return fmt.Sprintf(DefaultCertSecretName, name)
}

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

func (k K8sClient) GetCertificate(
	ctx context.Context,
	name, namespace string,
) (*v1.Service, error) {
	res, err := k.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (k K8sClient) CreateCertificate(
	ctx context.Context,
	name, namespace string,
	opts ...CertificateOption,
) (*cm.Certificate, error) {
	cert := NewCertificate(name, namespace, opts...)
	res, err := k.cmClient.CertmanagerV1().
		Certificates(namespace).
		Create(ctx, cert, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (k K8sClient) UpdateCertificate(
	ctx context.Context,
	name, namespace string,
	opts ...CertificateOption,
) (*cm.Certificate, error) {
	var result *cm.Certificate
	if retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		res, err := k.cmClient.CertmanagerV1().Certificates(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		update := NewCertificate(name, namespace, opts...)
		res.Spec = update.Spec
		result, err = k.cmClient.CertmanagerV1().Certificates(namespace).Update(ctx, res, metav1.UpdateOptions{})
		return err
	}); retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

type CertificateOption func(c *cm.Certificate)

func WithCertificateDomain(domain string) CertificateOption {
	return func(c *cm.Certificate) {
		c.Spec.DNSNames = append(c.Spec.DNSNames, domain)
	}
}
func WithCertificateKind(kind string) CertificateOption {
	return func(c *cm.Certificate) {
		c.Spec.IssuerRef.Kind = kind
	}
}
func WithCertificateName(name string) CertificateOption {
	return func(c *cm.Certificate) {
		c.Spec.IssuerRef.Name = name
	}
}
func NewCertificate(name, namespace string, opts ...CertificateOption) *cm.Certificate {
	labels := CreateLabels(WithName(name))
	if namespace == assets.AppName {
		labels = CreateLabels(WithName(name), WithComponent("certificate"), WithCoreLabel(true))
	}
	c := &cm.Certificate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "Certificate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	c.Spec.IssuerRef.Kind = "ClusterIssuer"
	c.Spec.IssuerRef.Name = DefaultClusterIssuer
	c.Spec.SecretName = CreateCertSecretName(name)
	for _, opt := range opts {
		opt(c)
	}
	return c
}
