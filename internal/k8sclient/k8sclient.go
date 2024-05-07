package k8sclient

import (
	"fmt"
	"os"

	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	assets "github.com/danielmichaels/tawny"
	"github.com/danielmichaels/tawny/internal/logger"
	traefikclientset "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	LetsEncryptStaging      = "https://acme-staging-v02.api.letsencrypt.org/directory"
	DefaultIngressRouteName = "ingressroute"
)

var (
	DefaultNamespace      = assets.AppName
	DefaultCertName       = fmt.Sprintf("%s-root-domain", assets.AppName)
	DefaultClusterIssuer  = fmt.Sprintf("%s-clusterissuer", assets.AppName)
	DefaultCertSecretName = "%s-cert-secret"
	DefaultServiceName    = "%s-%s-svc"
)

type K8sClient struct {
	DynamicClient *dynamic.DynamicClient
	Client        *kubernetes.Clientset
	logger        *logger.Logger
	restConfig    *rest.Config

	tClient  *traefikclientset.Clientset
	cmClient *cmclient.Clientset
}

func NewK8sClient(isDebug, isConsole bool) *K8sClient {
	l := logger.New("k8sclient", isDebug, isConsole)
	restConfig, err := GetKubeConfig()
	if err != nil {
		l.Fatal().Err(err).Msg("error: getting kube config")
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating kubernetes client")
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating dynamic kubernetes client")
	}
	tClient, err := traefikclientset.NewForConfig(restConfig)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating dynamic kubernetes client")
	}
	cmClient, err := cmclient.NewForConfig(restConfig)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating dynamic kubernetes client")
	}
	l.Info().Msg("k8s client created")
	return &K8sClient{
		Client:        client,
		DynamicClient: dynamicClient,
		logger:        l,
		restConfig:    restConfig,
		tClient:       tClient,
		cmClient:      cmClient,
	}
}

// GetKubeConfig checks whether KUBECONFIG environment variable is set.
// If it is set, uses it as kubeconfig.
// if it isn't, falls back to InClusterConfig.
func GetKubeConfig() (*rest.Config, error) {
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	return rest.InClusterConfig()
}
