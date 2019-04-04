package galaxy

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var app *Galaxy

func TestGalaxyNew(t *testing.T) {
	var dotGalaxy, _ = NewDotGalaxy("../../test/galaxy.yaml")
	var cmdArgs = make(map[string]string)

	app = NewGalaxy(logrus.New(), dotGalaxy, cmdArgs)

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
