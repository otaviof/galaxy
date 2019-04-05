package galaxy

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func populatedContext(t *testing.T) *Context {
	context := NewContext()
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

	assert.Equal(t, 3, len(context.releases["ns1"]))
	assert.Equal(t, 1, len(context.releases["ns2"]))
}

func TestContextRenameLandscaperReleases(t *testing.T) {
	context := populatedContext(t)

	context.RenameLandscaperReleases(func(ns, name string) (string, error) {
		return fmt.Sprintf("%s-%s", ns, name), nil
	})

	for ns, releases := range context.releases {
		for _, release := range releases {
			assert.Contains(t, release.Component.Name, fmt.Sprintf("%s-", ns))
		}
	}
}

func TestContextRenameNamespaces(t *testing.T) {
	context := populatedContext(t)

	context.RenameNamespaces(func(ns string) string {
		return fmt.Sprintf("test-%s", ns)
	})

	for ns := range context.releases {
		assert.Contains(t, ns, "test-")
	}
}
