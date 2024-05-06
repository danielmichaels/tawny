package k8sclient

import (
	"context"
	"fmt"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/danielmichaels/tawny/internal/ptr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

func generatePrivKeyRef(name string) cmmeta.SecretKeySelector {
	label := fmt.Sprintf("domain-cert-%s", name)
	return cmmeta.SecretKeySelector{
		Key:                  label,
		LocalObjectReference: cmmeta.LocalObjectReference{Name: label},
	}
}

func (k K8sClient) CreateClusterIssuer(
	ctx context.Context,
	opts ...ClusterIssuerOption,
) (*cm.ClusterIssuer, error) {
	desired := NewClusterIssuer(opts...)
	res, err := k.cmClient.CertmanagerV1().
		ClusterIssuers().
		Create(ctx, desired, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (k K8sClient) UpdateClusterIssuer(
	ctx context.Context,
	name string,
	opts ...ClusterIssuerOption,
) (*cm.ClusterIssuer, error) {
	var result *cm.ClusterIssuer
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		res, err := k.cmClient.CertmanagerV1().ClusterIssuers().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		update := NewClusterIssuer(opts...)
		res.Spec = update.Spec
		result, err = k.cmClient.CertmanagerV1().
			ClusterIssuers().
			Update(ctx, res, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

type ClusterIssuerOption func(*cm.ClusterIssuer)

func WithClusterIssuerACMEDNS01(
	email, server, name, apikeyName, apikeyRef string,
	dnsZones []string,
) ClusterIssuerOption {
	return func(i *cm.ClusterIssuer) {
		if i.Spec.IssuerConfig.ACME == nil {
			i.Spec.IssuerConfig.ACME = &cmacme.ACMEIssuer{
				Email:  email,
				Server: server,
				PrivateKey: cmmeta.SecretKeySelector{
					//Key:                  name,
					LocalObjectReference: cmmeta.LocalObjectReference{Name: name},
				},
			}
		}
		dns01Solver := cmacme.ACMEChallengeSolver{
			Selector: &cmacme.CertificateDNSNameSelector{DNSZones: dnsZones},
			DNS01: &cmacme.ACMEChallengeSolverDNS01{
				Cloudflare: &cmacme.ACMEIssuerDNS01ProviderCloudflare{
					Email: email,
					APIToken: &cmmeta.SecretKeySelector{
						Key:                  apikeyRef,
						LocalObjectReference: cmmeta.LocalObjectReference{Name: apikeyName},
					},
				},
			},
		}
		i.Spec.IssuerConfig.ACME.Solvers =
			append(i.Spec.IssuerConfig.ACME.Solvers, dns01Solver)
	}
}

func WithClusterIssuerACMEHTTP01(email, name string) ClusterIssuerOption {
	return func(i *cm.ClusterIssuer) {
		if i.Spec.IssuerConfig.ACME == nil {
			i.Spec.IssuerConfig.ACME = &cmacme.ACMEIssuer{
				Email:      email,
				Server:     LetsEncryptStaging, // todo replace with prod
				PrivateKey: generatePrivKeyRef(name),
			}
		}
		http01Solver := cmacme.ACMEChallengeSolver{
			Selector: &cmacme.CertificateDNSNameSelector{DNSNames: []string{name}},
			HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
				Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
					ServiceType:      v1.ServiceTypeClusterIP,
					IngressClassName: ptr.Ptr("traefik"),
				},
			},
		}
		i.Spec.IssuerConfig.ACME.Solvers =
			append(i.Spec.IssuerConfig.ACME.Solvers, http01Solver)
	}
}
func WithClusterIssuerSelfSigned() ClusterIssuerOption {
	return func(i *cm.ClusterIssuer) {
		i.Spec.IssuerConfig.SelfSigned = &cm.SelfSignedIssuer{}
	}
}
func WithClusterIssuerCustomName(name string) ClusterIssuerOption {
	return func(i *cm.ClusterIssuer) {
		i.ObjectMeta.Name = name
	}
}

func NewClusterIssuer(opts ...ClusterIssuerOption) *cm.ClusterIssuer {
	c := &cm.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "ClusterIssuer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultClusterIssuer,
			Labels: CreateLabels(
				WithName(DefaultClusterIssuer),
				WithComponent("clusterissuer"),
				WithCoreLabel(true),
			),
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
