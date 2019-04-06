package galaxy

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func populatedContext(t *testing.T) *Context {
	ctx := NewContext()
	dotGalaxy, err := NewDotGalaxy("../../test/galaxy.yaml")
	assert.Nil(t, err)

	for _, ns := range dotGalaxy.Spec.Namespaces.Names {
		dirPath := path.Join(dotGalaxy.Spec.Namespaces.BaseDir, ns)
		err = ctx.InspectDir(ns, dirPath, dotGalaxy.Spec.Namespaces.Extensions)
		assert.Nil(t, err)
	}

	return ctx
}

func TestContextInspectDir(t *testing.T) {
	ctx := populatedContext(t)

	assert.Equal(t, 3, len(ctx.releases["ns1"]))
	assert.Equal(t, 1, len(ctx.releases["ns2"]))
}

func TestContextRenameLandscaperReleases(t *testing.T) {
	ctx := populatedContext(t)

	ctx.RenameLandscaperReleases(func(ns, name string) (string, error) {
		return fmt.Sprintf("%s-%s", ns, name), nil
	})

	for ns, releases := range ctx.releases {
		for _, release := range releases {
			assert.Contains(t, release.Component.Name, fmt.Sprintf("%s-", ns))
		}
	}
}

func TestContextRenameNamespaces(t *testing.T) {
	ctx := populatedContext(t)

	ctx.RenameNamespaces(func(ns string) string {
		return fmt.Sprintf("test-%s", ns)
	})

	for ns := range ctx.releases {
		assert.Contains(t, ns, "test-")
	}
}
