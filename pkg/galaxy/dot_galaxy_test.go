package galaxy

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dotGalaxy *DotGalaxy

func TestDotGalaxyNewDotGalaxy(t *testing.T) {
	var err error

	dotGalaxy, err = NewDotGalaxy("../../test/galaxy.yaml")
	log.Printf("Galaxy: '%#v'", dotGalaxy)

	assert.Nil(t, err)
	assert.Equal(t, "../../test/namespaces", dotGalaxy.Spec.Namespaces.BaseDir)
}

func TestDotGalaxyListNamespaces(t *testing.T) {
	var list []string

	list = dotGalaxy.ListNamespaces()

	assert.Equal(t, []string{"ns1", "ns2", "ns3", "ns4"}, list)
}

func TestDotGalaxyListEnvironments(t *testing.T) {
	var list []string

	list = dotGalaxy.ListEnvironments()

	assert.Equal(t, []string{"dev", "tst"}, list)
}

func TestDotGalaxyGetNamespaceDir(t *testing.T) {
	var dir string
	var err error

	dir, err = dotGalaxy.GetNamespaceDir("ns1")

	assert.Nil(t, err)
	assert.Equal(t, "../../test/namespaces/ns1", dir)
}

func TestDotGalaxyGetEnvironment(t *testing.T) {
	var env *Environment
	var err error

	env, err = dotGalaxy.GetEnvironment("dev")

	assert.Nil(t, err)
	assert.Equal(t, "dev", env.Name)
}
