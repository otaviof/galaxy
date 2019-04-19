package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/helm"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

type helmRelease struct {
	name         string
	chartName    string
	chartVersion string
}

type helmReleases map[string][]*helmRelease

const EnvName = "dev"

var k *galaxy.KubeClient
var h *galaxy.HelmClient
var cfg *galaxy.Config
var app *galaxy.Galaxy

func TestGalaxy(t *testing.T) {
	galaxy.SetLogLevel("trace")

	t.Run("prepare kubernetes and helm clients", prepare)
	t.Run("DRY-RUN dev environment", dryRunDevEnv)
	t.Run("assert nothing is deployed yet", nothingIsDeployed)
	t.Run("apply dev environment", applyDevEnv)
	t.Run("inspect dev environment releases", inspectDevEnv)
	t.Run("clean up", cleanUp)
}

func prepare(t *testing.T) {
	var err error

	cfg = galaxy.NewConfig()
	k = galaxy.NewKubeClient(cfg.KubernetesConfig)
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
	cfg.Environments = EnvName
	cfg.Namespaces = namespaces

	g := galaxy.NewGalaxy(dotGalaxy, cfg)
	err = g.Inspect()
	assert.Nil(t, err)

	return g
}

func dryRunDevEnv(t *testing.T) {
	var err error

	app = bootstrap(t, true, "")

	t.Logf("Planing %s environment", EnvName)
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

func getInstalledReleases(t *testing.T) helmReleases {
	res, err := h.Client.ListReleases()
	assert.Nil(t, err)

	installed := make(helmReleases)

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

	return installed
}

func inspectDevEnv(t *testing.T) {
	// reading installed releases
	installed := getInstalledReleases(t)

	// extracting environment from modified data
	data, found := app.Modified[EnvName]
	assert.True(t, found)

	for _, ctx := range data {
		for ns, releases := range ctx.Releases {
			t.Logf("Inspecting namespace '%s', must be on installed map", ns)
			_, found := installed[ns]
			assert.True(t, found)

			for _, release := range releases {
				name := release.Component.Name
				chartMeta := release.Component.Release.Chart
				found := false

				t.Logf("Looking for '%s' (%s) on '%s' namespace", name, chartMeta, ns)
				for _, r := range installed[ns] {
					expectedMeta := fmt.Sprintf("%s:%s", r.chartName, r.chartVersion)
					if name == r.name && strings.Contains(chartMeta, expectedMeta) {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		}
	}
}

func cleanUp(t *testing.T) {
	var installed []string
	var err error

	for _, ctx := range app.Modified[EnvName] {
		for _, releases := range ctx.Releases {
			for _, release := range releases {
				installed = append(installed, release.Component.Name)
			}
		}
	}

	for _, name := range installed {
		t.Logf("Purging release '%s'...", name)
		_, err = h.Client.DeleteRelease(name, helm.DeletePurge(true))
		assert.Nil(t, err)
	}
}
