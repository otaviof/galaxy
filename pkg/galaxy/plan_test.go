package galaxy

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var plan *Plan

func TestPlanNewPlan(t *testing.T) {
	SetLogLevel("trace")

	dotGalaxy, _ := NewDotGalaxy("../../test/galaxy.yaml")
	env, _ := dotGalaxy.GetEnvironment("dev")
	ctx := NewContext()
	baseDir := path.Join(dotGalaxy.Spec.Namespaces.BaseDir, "ns1")

	ctx.InspectDir("ns1", baseDir, dotGalaxy.Spec.Namespaces.Extensions)

	plan = NewPlan(env, []string{}, ctx)
}

func TestPlanSkipFile(t *testing.T) {
	for file, skip := range map[string]bool{
		"file-name@d.yaml":   false,
		"file-name@d@t.yaml": false,
		"file-name@a.yaml":   true,
		"file-name.yaml":     false,
		"file-name@x.yaml":   true,
	} {
		t.Logf("Testing name '%s' should skip '%v'", file, skip)
		skipped, err := plan.skipFile(file)
		assert.Nil(t, err)
		assert.Equal(t, skip, skipped)
	}
}

func TestPlanContextForEnvironment(t *testing.T) {
	expected := map[string][]string{
		"ns1-d": {
			"../../test/namespaces/ns1/app1.yaml",
			"../../test/namespaces/ns1/app2@d.yaml",
			"../../test/namespaces/ns1/ingress-secret.yaml",
		},
	}

	ctx, err := plan.ContextForEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, expected, ctx.GetNamespaceFilesMap())
}
