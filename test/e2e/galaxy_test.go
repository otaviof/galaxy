package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

type helmRelease struct {
	name         string
	chartName    string
	chartVersion string
}

var k *galaxy.KubeClient
var h *galaxy.HelmClient
var cfg *galaxy.Config
var app *galaxy.Galaxy

func TestGalaxy(t *testing.T) {
	log.SetLevel(log.TraceLevel)

	prepare(t)

	t.Run("DRY-RUN dev environment", dryRunDevEnv)
	// t.Run("assert nothing is deployed yet", nothingIsDeployed)
	t.Run("apply dev environment", applyDevEnv)
	t.Run("inspect dev environment releases", inspectDevEnv)
}

func prepare(t *testing.T) {
	var err error

	cfg = galaxy.NewConfig()
	cfg.KubeConfig = os.Getenv("KUBECONFIG")

	k = galaxy.NewKubeClient(cfg.KubeConfig, cfg.KubeContext, cfg.InCluster)
	err = k.Load()
	assert.Nil(t, err)

	h = galaxy.NewHelmClient(cfg.HelmHome, cfg.TillerNamespace, cfg.TillerPort, cfg.TillerTimeout, k)
	err = h.Load()
	assert.Nil(t, err)
}

func bootstrap(t *testing.T, dryRun bool, namespaces string) *galaxy.Galaxy {
	dotGalaxy, err := galaxy.NewDotGalaxy("../../test/galaxy.yaml")
	assert.Nil(t, err)

	cfg.DryRun = dryRun
	cfg.Environments = "dev"
	cfg.Namespaces = namespaces

	g := galaxy.NewGalaxy(dotGalaxy, cfg)
	err = g.Inspect()
	assert.Nil(t, err)

	return g
}

func dryRunDevEnv(t *testing.T) {
	var err error

	app = bootstrap(t, true, "")

	t.Log("Planing 'dev' environment")
	err = app.Plan()
	assert.Nil(t, err)

	t.Logf("Applying changes (dry-run: '%v')", cfg.DryRun)
	err = app.Apply()
	assert.Nil(t, err)
}

func nothingIsDeployed(t *testing.T) {
	res, err := h.Client.ListReleases()
	assert.Nil(t, err)
	count := res.GetCount()

	t.Logf("Found '%d' releases in Helm", count)
	assert.Equal(t, int64(0), count)
}

func applyDevEnv(t *testing.T) {
	app := bootstrap(t, false, "")
	_ = app.Plan()

	t.Logf("Applying changes (dry-run: '%v')", cfg.DryRun)
	err := app.Apply()
	assert.Nil(t, err)
}

func inspectDevEnv(t *testing.T) {
	res, err := h.Client.ListReleases()
	assert.Nil(t, err)

	// map of namespace name and array of releases
	installed := make(map[string][]*helmRelease)

	// organizing helm releases in a map
	for _, r := range res.GetReleases() {
		metadata := r.GetChart().GetMetadata()
		ns := r.GetNamespace()
		name := r.GetName()
		t.Logf("Helm release '%s' on '%s' namespace", name, ns)
		installed[ns] = append(installed[ns], &helmRelease{
			name:         name,
			chartName:    metadata.GetName(),
			chartVersion: metadata.GetVersion(),
		})
	}

	// extracting environment from modified data
	data, found := app.Modified["dev"]
	assert.True(t, found)

	for _, ctx := range data {
		for ns, releases := range ctx.Releases {
			t.Logf("Inspecting namespace '%s', must be on installed map", ns)
			_, found := installed[ns]
			assert.True(t, found)

			for _, release := range releases {
				var exists bool

				name := release.Component.Name
				chartMeta := release.Component.Release.Chart

				t.Logf("Looking for '%s' (%s) on '%s' namespace", name, chartMeta, ns)
				for _, r := range installed[ns] {
					expectedMeta := fmt.Sprintf("%s:%s", r.chartName, r.chartVersion)
					if name == r.name && strings.Contains(chartMeta, expectedMeta) {
						exists = true
						break
					}
				}
				assert.True(t, exists)
			}
		}
	}
}
