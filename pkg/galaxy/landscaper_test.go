package galaxy

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var landscaper *Landscaper

func TestLandscaperNewLandscaper(t *testing.T) {
	log.SetLevel(log.TraceLevel)

	dotGalaxy, _ := NewDotGalaxy("../../test/galaxy.yaml")
	g := NewGalaxy(dotGalaxy, NewConfig())
	g.Plan()
	env, _ := dotGalaxy.GetEnvironment("dev")

	cfg := NewConfig()
	cfg.KubeConfig = os.Getenv("KUBECONFIG")

	landscaper = NewLandscaper(cfg.LandscaperConfig, env, "ns1", g.Modified["ns1-d"])
}

func TestLandscaperBootstrap(t *testing.T) {
	err := landscaper.Bootstrap(true)
	assert.Nil(t, err)

	assert.NotNil(t, landscaper.kubeCfg)
	assert.NotNil(t, landscaper.kubeClient)
	assert.NotNil(t, landscaper.helmClient)
}

func TestLandscaperApply(t *testing.T) {
	err := landscaper.Apply()
	assert.Nil(t, err)
}
