package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var app *Galaxy

func TestGalaxyNew(t *testing.T) {
	SetLogLevel("trace")

	dotGalaxy, _ = NewDotGalaxy("../../test/galaxy.yaml")
	cfg := NewConfig()
	cfg.DryRun = true
	cfg.Environments = "dev"
	cfg.Namespaces = "ns1"
	cfg.SkipSecrets = true
	app = NewGalaxy(dotGalaxy, cfg)

	assert.NotNil(t, app)
}

func TestGalaxyInspect(t *testing.T) {
	err := app.Inspect()

	assert.Nil(t, err)
}

func TestGalaxyPlan(t *testing.T) {
	err := app.Plan()

	assert.Nil(t, err)
}

func TestGalaxyApply(t *testing.T) {
	err := app.Apply()

	assert.Nil(t, err)
}
