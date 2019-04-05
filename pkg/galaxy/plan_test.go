package galaxy

import (
	"log"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var plan *Plan

func TestPlanNewPlan(t *testing.T) {
	dotGalaxy, _ := NewDotGalaxy("../../test/galaxy.yaml")
	env, _ := dotGalaxy.GetEnvironment("dev")
	ctx := NewContext()
	baseDir := path.Join(dotGalaxy.Spec.Namespaces.BaseDir, "ns1")

	ctx.InspectDir("ns1", baseDir, dotGalaxy.Spec.Namespaces.Extensions)

	plan = NewPlan(env, ctx)
}

func TestPlanExtractFileSuffix(t *testing.T) {
	for file, input := range map[string][]string{
		"file-d.yaml":                    {"secret", "yaml", "d"},
		"file-x.yaml":                    {"secret", "yaml", "x"},
		"file.yaml":                      {"secret", "yaml", ""},
		"/slash/dir/file.yaml":           {"secret", "yaml", ""},
		"file-a-b-c-d-d.yaml":            {"secret", "yaml", "d"},
		"file-a-b-c-d-x.yaml":            {"secret", "yaml", "x"},
		"/slash/dir/file-a-b-c-d-d.yaml": {"secret", "yaml", "d"},
		"/slash/dir/file-a-b-c-x-x.yaml": {"secret", "yaml", "x"},
		"file.a.b-c.d.d.yaml":            {"secret", "yaml", ""},
		"/slash/dir/file.a.b.c.d.d.yaml": {"secret", "yaml", ""},
	} {
		log.Printf("input='%#v'", input)
		res, err := plan.extractFileSuffix(file, []string{input[0], input[1]})
		assert.Nil(t, err)
		assert.Equal(t, input[2], res)
	}
}

func TestPlanSkipOnSuffix(t *testing.T) {
	assert.True(t, plan.skipOnSuffix("a"))
	assert.False(t, plan.skipOnSuffix("d"))
	assert.False(t, plan.skipOnSuffix(""))
}

func TestPlanContextForEnvironment(t *testing.T) {
	var extensions = []string{"secret", "yaml"}
	var expected = map[string][]string{
		"ns1-d": {
			"../../test/namespaces/ns1/app1.yaml",
			"../../test/namespaces/ns1/app2-d.yaml",
		},
	}
	var context *Context
	var err error

	context, err = plan.ContextForEnvironment(extensions)

	assert.Nil(t, err)
	assert.Equal(t, expected, context.GetNamespaceFilesMap())
}
