package galaxy

import (
	"strings"
)

// Config runtime configuration, command-line arguments.
type Config struct {
	DotGalaxyPath string // path to dot-galaxy file
	DryRun        bool   // dry-run flag
	LogLevel      string // log verboseness
	Environments  string // target environment names, comma separated
}

// GetEnvironments as slice of strings.:w
func (c *Config) GetEnvironments() []string {
	return strings.Split(c.Environments, ",")
}

// NewConfig with default values.
func NewConfig() *Config {
	return &Config{
		Environments:  "",
		LogLevel:      "error",
		DryRun:        false,
		DotGalaxyPath: ".galaxy.yaml",
	}
}
