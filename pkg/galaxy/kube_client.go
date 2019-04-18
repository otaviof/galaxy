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
	logger    *log.Entry           // logger
	RestCfg   *rest.Config         // kubernetes rest config
	Client    *clientset.Clientset // kubernetes clientset
	config    string               // kubernetes config
	context   string               // kubernetes context
	inCluster bool                 // running in cluster flag
}

// Load the new Kubernetes API client.
func (k *KubeClient) Load() error {
	var err error

	if k.inCluster {
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
	if k.config == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return nil, fmt.Errorf("environment HOME is empty, can't find '~/.kube/config' file")
		}
		k.config = filepath.Join(homeDir, ".kube", "config")
	}
	k.logger.Infof("Using kubernetes configuration file: '%s'", k.config)

	if !fileExists(k.config) {
		return nil, fmt.Errorf("can't find kube-config file at: '%s'", k.config)
	}

	return clientcmd.BuildConfigFromFlags(k.context, k.config)
}

// NewKubeClient instantiate a new Kubernetes API client.
func NewKubeClient(config, context string, inCluster bool) *KubeClient {
	return &KubeClient{
		logger: log.WithFields(log.Fields{
			"type":        "kubeClient",
			"kubeConfig":  config,
			"kubeContext": context,
			"inCluster":   inCluster,
		}),
		config:    config,
		context:   context,
		inCluster: inCluster,
	}
}
