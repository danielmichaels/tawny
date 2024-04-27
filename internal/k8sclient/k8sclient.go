package k8sclient

import (
	"github.com/danielmichaels/tawny/internal/logger"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type K8sClient struct {
	DynamicClient *dynamic.DynamicClient
	Client        *kubernetes.Clientset
	logger        *logger.Logger
}

func NewK8sClient(isDebug, isConsole bool) *K8sClient {
	l := logger.New("k8sclient", isDebug, isConsole)
	config, err := GetKubeConfig()
	if err != nil {
		l.Fatal().Err(err).Msg("error: getting kube config")
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating kubernetes client")
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		l.Fatal().Err(err).Msg("error: creating dynamic kubernetes client")
	}
	l.Info().Msg("k8s client created")
	return &K8sClient{Client: client, DynamicClient: dynamicClient, logger: l}
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
