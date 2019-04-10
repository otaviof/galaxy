package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var app *Galaxy

func TestGalaxyNew(t *testing.T) {
	dotGalaxy, _ = NewDotGalaxy("../../test/galaxy.yaml")
	app = NewGalaxy(dotGalaxy, NewConfig())

	assert.NotNil(t, app)
}

func TestGalaxyInspect(t *testing.T) {
	var err error

	err = app.Inspect()

	assert.Nil(t, err)
}

func TestGalaxyPlan(t *testing.T) {
	var err error

	err = app.Plan()

	assert.Nil(t, err)
}
