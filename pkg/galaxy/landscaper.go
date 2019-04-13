package galaxy

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // gcp auth
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	helmkube "k8s.io/helm/pkg/kube"
	helmversion "k8s.io/helm/pkg/version"
	podutil "k8s.io/kubernetes/pkg/api/pod"
	core "k8s.io/kubernetes/pkg/apis/core"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

// Landscaper represents upstream Landscaper.
type Landscaper struct {
	logger     *log.Entry           // logger
	cfg        *LandscaperConfig    // landscaper runtime configuration
	ns         string               // target namespace
	env        *Environment         // environment instance
	ctxs       []*Context           // slice of Context instances
	kubeClient *clientset.Clientset // kubernetes api client
	kubeCfg    *rest.Config         // kubernetes client config
	helmClient helm.Interface       // helm client
	fileState  ldsc.StateProvider   // landscaper release file state provider
	helmState  ldsc.StateProvider   // landscaper helm state provider
	executor   ldsc.Executor        // landscaper executor
}

// Apply wrapper around Landscaper Apply method.
func (l *Landscaper) Apply() error {
	var desired ldsc.Components
	var current ldsc.Components
	var result map[string][]string
	var err error

	if desired, err = l.fileState.Components(); err != nil {
		return err
	}
	if current, err = l.helmState.Components(); err != nil {
		return err
	}
	if result, err = l.executor.Apply(desired, current); err != nil {
		return err
	}

	l.logger.Debugf("results: '%#v'", result)

	return nil
}

// Bootstrap prepare Landscaper requirements and components.
func (l *Landscaper) Bootstrap(dryRun bool) error {
	var e *ldsc.Environment
	var err error

	if err = l.loadKubeClient(); err != nil {
		return err
	}
	if err = l.loadHelmClient(); err != nil {
		return err
	}

	if e, err = l.setupLandscaperEnvironment(dryRun); err != nil {
		return err
	}

	kubeSecrets := ldsc.NewKubeSecretsReadWriteDeleter(l.kubeClient.Core())
	secretsReader := ldsc.NewEnvironmentSecretsReader()

	l.fileState = ldsc.NewFileStateProvider(
		e.ComponentFiles,
		secretsReader,
		e.ChartLoader,
		e.ReleaseNamePrefix,
		e.Namespace,
		e.Environment,
		e.ConfigurationOverrideFile,
	)
	l.helmState = ldsc.NewHelmStateProvider(l.helmClient, kubeSecrets, e.ReleaseNamePrefix)
	l.executor = ldsc.NewExecutor(
		l.helmClient,
		e.ChartLoader,
		kubeSecrets,
		e.DryRun,
		e.Wait,
		int64(e.WaitTimeout/time.Second),
		e.DisabledStages,
	)

	return nil
}

// setupLandscaperEnvironment
func (l *Landscaper) setupLandscaperEnvironment(dryRun bool) (*ldsc.Environment, error) {
	var releasePrefix string
	var err error

	if l.env.Transform.ReleasePrefix != "" {
		if releasePrefix, err = l.env.Interpolate(l.env.Transform.ReleasePrefix, []string{
			fmt.Sprintf("NAMESPACE=%s", l.ns),
		}); err != nil {
			return nil, err
		}
	}

	return &ldsc.Environment{
		Context:                   l.cfg.KubeContext,
		Namespace:                 l.ns,
		Environment:               l.env.Name,
		ComponentFiles:            l.pickReleaseFiles(),
		ReleaseNamePrefix:         releasePrefix,
		HelmHome:                  l.cfg.HelmHome,
		TillerNamespace:           l.cfg.TillerNamespace,
		ChartLoader:               ldsc.NewLocalCharts(l.cfg.HelmHome),
		ConfigurationOverrideFile: l.cfg.OverrideFile,
		Wait:                      l.cfg.WaitForResources,
		WaitTimeout:               time.Duration(time.Duration(l.cfg.WaitTimeout) * time.Second),
		DisabledStages:            l.cfg.GetDisabledStages(),
	}, nil
}

// pickReleaseFiles select release components for the target namespace.
func (l *Landscaper) pickReleaseFiles() []string {
	var files []string
	var releases []Release
	var found bool

	for _, ctx := range l.ctxs {
		// checking releases for configured namespace only
		if releases, found = ctx.Releases[l.ns]; !found {
			continue
		}
		for _, release := range releases {
			l.logger.Infof("Inspecting release '%s'", release.Component.Name)
			files = append(files, release.File)
		}
	}

	return files
}

// getHelmTillerPodName using Kubernetes API client, look for Tiller's pod.
func (l *Landscaper) getHelmTillerPodName() (string, error) {
	var pods *core.PodList
	var err error

	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	options := metav1.ListOptions{LabelSelector: selector.String()}

	if pods, err = l.kubeClient.Core().Pods(l.cfg.TillerNamespace).List(options); err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("can't find tiller pod on '%s' namespace", l.cfg.TillerNamespace)
	}
	for _, pod := range pods.Items {
		if podutil.IsPodReady(&pod) {
			return pod.ObjectMeta.GetName(), nil
		}
	}

	return "", fmt.Errorf("can't find a ready tiller pod on '%s' namespace", l.cfg.TillerNamespace)
}

// getHelmTillerAddress inspect environment for Helm hostname, or establish a port-forward to tiller.
func (l *Landscaper) getHelmTillerAddress() (string, error) {
	var podName string
	var err error

	hostname := os.Getenv("HELM_HOST")
	if hostname != "" {
		l.logger.Infof("Using HELM_HOST environment variable as Tiller hostname '%s'", hostname)
		return hostname, nil
	}

	logger := l.logger.WithFields(log.Fields{
		"tillerNamespace": l.cfg.TillerNamespace,
		"tillerPort":      l.cfg.TillerPort,
	})
	logger.Infof("Setting up port-forward to reach Tiller...")

	if podName, err = l.getHelmTillerPodName(); err != nil {
		return "", err
	}
	logger.Debugf("Tiller pod name '%s'", podName)

	restClient := l.kubeClient.Core().RESTClient()
	tunnel := helmkube.NewTunnel(
		restClient, l.kubeCfg, l.cfg.TillerNamespace, podName, l.cfg.TillerPort,
	)

	if err = tunnel.ForwardPort(); err != nil {
		return "", err
	}

	return fmt.Sprintf(":%d", tunnel.Local), nil
}

// loadHelmClient creates a new instance of Helm API client by direct access or port-forward.
func (l *Landscaper) loadHelmClient() error {
	var hostname string
	var err error

	l.logger.Info("Creating a new Helm API client...")

	if hostname, err = l.getHelmTillerAddress(); err != nil {
		return err
	}

	l.logger.Infof("Connecting to Helm via '%s' (timeout %d seconds)", hostname, l.cfg.TillerTimeout)
	l.helmClient = helm.NewClient(helm.Host(hostname), helm.ConnectTimeout(l.cfg.TillerTimeout))
	if err = l.helmClient.PingTiller(); err != nil {
		return err
	}

	l.logger.Infof("Comparing Helm's Tiller version with local ('%s')", helmversion.Version)
	version, err := l.helmClient.GetVersion()
	if err != nil {
		return err
	}
	l.logger.Infof("Tiller version: '%s'", version.Version.SemVer)
	if !helmversion.IsCompatible(helmversion.Version, version.Version.SemVer) {
		return fmt.Errorf("incompatible version numbers, tiller '%s' this '%s'",
			version.Version, helmversion.Version)
	}

	return nil
}

// loadKubeClient creates a new Kubernetes API client instance for Landscaper.
func (l *Landscaper) loadKubeClient() error {
	var err error

	logger := log.WithFields(log.Fields{
		"inCluster":   l.cfg.InCluster,
		"kubeConfig":  l.cfg.KubeConfig,
		"kubeContext": l.cfg.KubeContext,
	})

	if l.cfg.InCluster {
		logger.Info("Using in-cluster Kubernetes client...")
		if l.kubeCfg, err = rest.InClusterConfig(); err != nil {
			return err
		}
	} else {
		logger.Info("Using local kube-config...")
		if l.kubeCfg, err = l.getKubeRestConfig(); err != nil {
			return err
		}
	}

	if l.kubeClient, err = clientset.NewForConfig(l.kubeCfg); err != nil {
		return err
	}
	return nil
}

// getKubeRestConfig read kube-config from home, or alternative path.
func (l *Landscaper) getKubeRestConfig() (*rest.Config, error) {
	var kubeCfg string

	if l.cfg.KubeConfig == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return nil, fmt.Errorf("environment HOME is empty, can't find '~/.kube/config' file")
		}
		kubeCfg = filepath.Join(homeDir, ".kube", "config")
	}
	l.logger.Infof("Using kubernetes configuration file: '%s'", kubeCfg)

	if !fileExists(kubeCfg) {
		return nil, fmt.Errorf("can't find kube-config file at: '%s'", kubeCfg)
	}

	return clientcmd.BuildConfigFromFlags(l.cfg.KubeContext, kubeCfg)
}

// NewLandscaper instance a new Landscaper object.
func NewLandscaper(cfg *LandscaperConfig, env *Environment, ns string, ctxs []*Context) *Landscaper {
	return &Landscaper{
		logger: log.WithField("type", "landscaper"),
		cfg:    cfg,
		env:    env,
		ns:     ns,
		ctxs:   ctxs,
	}
}
