package galaxy

// Config runtime configuration, command-line arguments.
type Config struct {
	Environment   string // target environment name
	LogLevel      string // log verboseness
	DryRun        bool   // dry-run flag
	DotGalaxyPath string // path to dot-galaxy file
}

// NewConfig with default values.
func NewConfig() *Config {
	return &Config{
		Environment:   "",
		LogLevel:      "error",
		DryRun:        false,
		DotGalaxyPath: ".galaxy.yaml",
	}
}
