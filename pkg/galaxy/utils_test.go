package galaxy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtilsExists(t *testing.T) {
	assert.True(t, exists("../../test/galaxy.yaml"))
}

func TestUtilsReadFile(t *testing.T) {
	assert.True(t, len(readFile("../../test/galaxy.yaml")) > 0)
}

func TestUtilsIsDir(t *testing.T) {
	assert.True(t, isDir("../../test"))
}

func TestUtilsStringSliceContains(t *testing.T) {
	assert.True(t, stringSliceContains([]string{"a", "b", "c"}, "b"))
}
