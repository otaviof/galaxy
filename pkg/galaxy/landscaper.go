package galaxy

import (
	"fmt"
	"time"

	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	log "github.com/sirupsen/logrus"
)

// Landscaper represents upstream Landscaper.
type Landscaper struct {
	logger     *log.Entry         // logger
	cfg        *LandscaperConfig  // landscaper runtime configuration
	kubeCfg    *KubernetesConfig  // kubernetes related configuration
	env        *Environment       // environment instance
	ctxs       []*Context         // slice of context instances
	kubeClient *KubeClient        // kubernetes api client
	helmClient *HelmClient        // helm api client
	fileState  ldsc.StateProvider // landscaper release file state provider
	helmState  ldsc.StateProvider // landscaper helm state provider
	executor   ldsc.Executor      // landscaper executor
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
func (l *Landscaper) Bootstrap(ns, originalNs string, dryRun bool) error {
	var e *ldsc.Environment
	var err error

	l.logger.Infof("Bootstraping Landscaper for namespace '%s' (originally '%s')", ns, originalNs)

	if err = l.loadKubeClient(); err != nil {
		return err
	}
	if err = l.loadHelmClient(); err != nil {
		return err
	}

	if e, err = l.setup(ns, originalNs, dryRun); err != nil {
		return err
	}

	kubeSecrets := ldsc.NewKubeSecretsReadWriteDeleter(l.kubeClient.Client.Core())
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
	l.helmState = ldsc.NewHelmStateProvider(l.helmClient.Client, kubeSecrets, e.ReleaseNamePrefix)
	l.executor = ldsc.NewExecutor(
		l.helmClient.Client,
		e.ChartLoader,
		kubeSecrets,
		e.DryRun,
		e.Wait,
		int64(e.WaitTimeout/time.Second),
		e.DisabledStages,
	)

	return nil
}

// setup Landscaper environment and release prefix.
func (l *Landscaper) setup(ns, originalNs string, dryRun bool) (*ldsc.Environment, error) {
	var releasePrefix string
	var err error

	if l.env.Transform.ReleasePrefix != "" {
		if releasePrefix, err = l.env.Interpolate(l.env.Transform.ReleasePrefix, []string{
			fmt.Sprintf("NAMESPACE=%s", originalNs),
		}); err != nil {
			return nil, err
		}
	}

	return &ldsc.Environment{
		DryRun:                    dryRun,
		Context:                   l.kubeCfg.KubeContext,
		Namespace:                 ns,
		Environment:               l.env.Name,
		ComponentFiles:            l.pickReleaseFiles(ns),
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
func (l *Landscaper) pickReleaseFiles(ns string) []string {
	var files []string
	var releases []Release
	var found bool

	for _, ctx := range l.ctxs {
		// checking releases for configured namespace only
		if releases, found = ctx.Releases[ns]; !found {
			continue
		}
		for _, release := range releases {
			l.logger.Infof("Inspecting release '%s'", release.Component.Name)
			files = append(files, release.File)
		}
	}

	return files
}

// loadHelmClient creates a new instance of Helm API client.
func (l *Landscaper) loadHelmClient() error {
	l.helmClient = NewHelmClient(
		l.cfg.HelmHome, l.cfg.TillerNamespace, l.cfg.TillerPort, l.cfg.TillerTimeout, l.kubeClient,
	)
	return l.helmClient.Load()
}

// loadKubeClient creates a new Kubernetes API client instance.
func (l *Landscaper) loadKubeClient() error {
	l.kubeClient = NewKubeClient(l.kubeCfg)
	return l.kubeClient.Load()
}

// NewLandscaper instance a new Landscaper object.
func NewLandscaper(
	cfg *LandscaperConfig, kubeCfg *KubernetesConfig, env *Environment, ctxs []*Context,
	forceColors bool) *Landscaper {

	if forceColors {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	return &Landscaper{
		logger:  log.WithField("type", "landscaper"),
		cfg:     cfg,
		kubeCfg: kubeCfg,
		env:     env,
		ctxs:    ctxs,
	}
}
