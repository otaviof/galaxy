package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg *Config

func TestConfigNew(t *testing.T) {
	cfg = NewConfig()

	t.Logf("Config: '%#v'", cfg)

	assert.Equal(t, "error", cfg.LogLevel)
	assert.Equal(t, int64(60), cfg.WaitTimeout)

	// making sure sub-struct is present
	assert.NotNil(t, cfg.LandscaperConfig)
}

func TestConfigGetDisabledStages(t *testing.T) {
	cfg.DisabledStages = "one"
	assert.Equal(t, []string{"one"}, cfg.GetDisabledStages())

	cfg.DisabledStages = "one,two"
	assert.Equal(t, []string{"one", "two"}, cfg.GetDisabledStages())
}

func TestConfigGetEnvironments(t *testing.T) {
	cfg.Environments = "one"
	assert.Equal(t, []string{"one"}, cfg.GetEnvironments())

	cfg.Environments = "one,two"
	assert.Equal(t, []string{"one", "two"}, cfg.GetEnvironments())
}

func TestConfigGetNamespaces(t *testing.T) {
	cfg.Namespaces = "one"
	assert.Equal(t, []string{"one"}, cfg.GetNamespaces())

	cfg.Namespaces = "one,two"
	assert.Equal(t, []string{"one", "two"}, cfg.GetNamespaces())
}
