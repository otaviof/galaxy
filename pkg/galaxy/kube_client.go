package galaxy

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

// KubeClient wrapper for Kubernetes API client.
type KubeClient struct {
	logger  *log.Entry           // logger
	cfg     *KubernetesConfig    // configuration parameters
	RestCfg *rest.Config         // kubernetes rest config
	Client  *clientset.Clientset // kubernetes clientset
}

// Load the new Kubernetes API client.
func (k *KubeClient) Load() error {
	var err error

	if k.cfg.InCluster {
		k.logger.Info("Using in-cluster Kubernetes client...")
		if k.RestCfg, err = rest.InClusterConfig(); err != nil {
			return err
		}
	} else {
		k.logger.Info("Using local kube-config...")
		if k.RestCfg, err = k.getKubeRestConfig(); err != nil {
			return err
		}
	}

	if k.Client, err = clientset.NewForConfig(k.RestCfg); err != nil {
		return err
	}
	return nil
}

// getKubeRestConfig load REST client config from home or alternative location.
func (k *KubeClient) getKubeRestConfig() (*rest.Config, error) {
	if k.cfg.KubeConfig == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return nil, fmt.Errorf("environment HOME is empty, can't find '~/.kube/config' file")
		}
		k.cfg.KubeConfig = filepath.Join(homeDir, ".kube", "config")
	}
	k.logger.Infof("Using kubernetes configuration file: '%s'", k.cfg.KubeConfig)

	if !fileExists(k.cfg.KubeConfig) {
		return nil, fmt.Errorf("can't find kube-config file at: '%s'", k.cfg.KubeConfig)
	}

	return clientcmd.BuildConfigFromFlags(k.cfg.KubeContext, k.cfg.KubeConfig)
}

// NewKubeClient instantiate a new Kubernetes API client.
func NewKubeClient(cfg *KubernetesConfig) *KubeClient {
	return &KubeClient{
		logger: log.WithFields(log.Fields{
			"type":        "kubeClient",
			"kubeConfig":  cfg.KubeConfig,
			"kubeContext": cfg.KubeContext,
			"inCluster":   cfg.InCluster,
		}),
		cfg: cfg,
	}
}
