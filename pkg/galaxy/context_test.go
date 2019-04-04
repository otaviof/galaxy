package galaxy

import (
	"path"
	"testing"

	logrus "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func populatedContext(t *testing.T) *Context {
	context := NewContext(logrus.New())
	dotGalaxy, err := NewDotGalaxy("../../test/galaxy.yaml")
	assert.Nil(t, err)

	for _, ns := range dotGalaxy.Spec.Namespaces.Names {
		dirPath := path.Join(dotGalaxy.Spec.Namespaces.BaseDir, ns)
		err = context.InspectDir(ns, dirPath, dotGalaxy.Spec.Namespaces.Extensions)
		assert.Nil(t, err)
	}

	return context
}

func TestContextInspectDir(t *testing.T) {
	context := populatedContext(t)

	assert.Equal(t, 3, len(context.nsFiles["ns1"]))
	assert.Equal(t, 1, len(context.nsFiles["ns2"]))
}
