package galaxy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var landscaper *Landscaper

func TestLandscaperNewLandscaper(t *testing.T) {
	SetLogLevel("trace")

	dotGalaxy, _ := NewDotGalaxy("../../test/galaxy.yaml")
	g := NewGalaxy(dotGalaxy, NewConfig())
	g.Plan()
	env, _ := dotGalaxy.GetEnvironment("dev")

	cfg := NewConfig()
	cfg.KubeConfig = os.Getenv("KUBECONFIG")

	landscaper = NewLandscaper(cfg.LandscaperConfig, env, g.Modified["dev"])
}

func TestLandscaperBootstrap(t *testing.T) {
	err := landscaper.Bootstrap("ns1-d", "ns1", true)
	assert.Nil(t, err)

	assert.NotNil(t, landscaper.kubeClient)
	assert.NotNil(t, landscaper.helmClient)
}

func TestLandscaperApply(t *testing.T) {
	err := landscaper.Apply()
	assert.Nil(t, err)
}
