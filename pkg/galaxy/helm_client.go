package galaxy

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/helm/pkg/helm"
	helmkube "k8s.io/helm/pkg/kube"
	helmversion "k8s.io/helm/pkg/version"
	podutil "k8s.io/kubernetes/pkg/api/pod"
	core "k8s.io/kubernetes/pkg/apis/core"
)

// HelmClient is wrapper for loading a Helm API client instance.
type HelmClient struct {
	logger     *log.Entry     // logger
	Client     helm.Interface // helm client
	home       string         // helm home directory
	ns         string         // namespace name
	port       int            // tiller port number
	timeout    int64          // tiller timeout
	kubeClient *KubeClient    // kubernetes client
}

// Load configuration and instantiate client.
func (h *HelmClient) Load() error {
	var address string
	var err error

	h.logger.Info("Creating a new Helm API client...")

	if address, err = h.getHelmTillerAddress(); err != nil {
		return err
	}

	h.logger.Infof("Connecting to Helm via '%s' (timeout %d seconds)", address, h.timeout)
	h.Client = helm.NewClient(helm.Host(address), helm.ConnectTimeout(h.timeout))
	if err = h.Client.PingTiller(); err != nil {
		return err
	}

	h.logger.Infof("Comparing Helm's Tiller version with local ('%s')", helmversion.Version)
	version, err := h.Client.GetVersion()
	if err != nil {
		return err
	}
	h.logger.Infof("Tiller version: '%s'", version.Version.SemVer)
	if !helmversion.IsCompatible(helmversion.Version, version.Version.SemVer) {
		return fmt.Errorf("incompatible version numbers, tiller '%s' this '%s'",
			version.Version, helmversion.Version)
	}

	return nil
}

// getHelmTillerAddress inspect environment for Helm hostname, or establish a port-forward to tiller.
func (h *HelmClient) getHelmTillerAddress() (string, error) {
	var podName string
	var err error

	hostname := os.Getenv("HELM_HOST")
	if hostname != "" {
		h.logger.Infof("Using HELM_HOST environment variable as Tiller hostname '%s'", hostname)
		return hostname, nil
	}

	h.logger.Infof("Setting up port-forward to reach Tiller...")

	if podName, err = h.getHelmTillerPodName(); err != nil {
		return "", err
	}
	h.logger.Debugf("Tiller pod name '%s'", podName)

	restClient := h.kubeClient.Client.Core().RESTClient()
	tunnel := helmkube.NewTunnel(restClient, h.kubeClient.RestCfg, h.ns, podName, h.port)

	if err = tunnel.ForwardPort(); err != nil {
		return "", err
	}

	return fmt.Sprintf(":%d", tunnel.Local), nil
}

// getHelmTillerPodName using Kubernetes API client, look for Tiller's pod.
func (h *HelmClient) getHelmTillerPodName() (string, error) {
	var pods *core.PodList
	var err error

	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	options := metav1.ListOptions{LabelSelector: selector.String()}

	if pods, err = h.kubeClient.Client.Core().Pods(h.ns).List(options); err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("can't find tiller pod on '%s' namespace", h.ns)
	}
	for _, pod := range pods.Items {
		if podutil.IsPodReady(&pod) {
			return pod.ObjectMeta.GetName(), nil
		}
	}

	return "", fmt.Errorf("can't find a ready tiller pod on '%s' namespace", h.ns)
}

// NewHelmClient new type instance.
func NewHelmClient(home, ns string, port int, timeout int64, kubeClient *KubeClient) *HelmClient {
	return &HelmClient{
		logger: log.WithFields(log.Fields{
			"type":      "helmClient",
			"home":      home,
			"namespace": ns,
			"timeout":   timeout,
			"port":      port,
		}),
		home:       home,
		ns:         ns,
		port:       port,
		timeout:    timeout,
		kubeClient: kubeClient,
	}
}
