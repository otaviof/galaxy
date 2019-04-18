package galaxy

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func populatedContext(t *testing.T) *Context {
	SetLogLevel("trace")

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

	for _, release := range ctx.Releases {
		t.Logf("release: '%#v'", release)
	}

	for _, secret := range ctx.Secrets {
		t.Logf("secret: '%#v'", secret)
	}

	assert.Equal(t, 3, len(ctx.Releases["ns1"]))
	assert.Equal(t, 1, len(ctx.Secrets["ns1"]))
	assert.Equal(t, 1, len(ctx.Releases["ns2"]))
}

func TestContextRenameReleases(t *testing.T) {
	ctx := populatedContext(t)

	ctx.RenameReleases(func(ns, name string) (string, error) {
		return fmt.Sprintf("%s-%s", ns, name), nil
	})

	for ns, releases := range ctx.Releases {
		for _, release := range releases {
			assert.Contains(t, release.Component.Name, fmt.Sprintf("%s-", ns))
		}
	}
}

func TestContextRenameNamespaces(t *testing.T) {
	var ns string

	ctx := populatedContext(t)

	ctx.RenameNamespaces(func(ns string) string {
		return fmt.Sprintf("test-%s", ns)
	})

	for ns = range ctx.Releases {
		assert.Contains(t, ns, "test-")
	}
	for ns = range ctx.Secrets {
		assert.Contains(t, ns, "test-")
	}
}
