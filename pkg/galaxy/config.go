package galaxy

import (
	"os"
	"strings"
)

// Config runtime configuration, command-line arguments.
type Config struct {
	DotGalaxyPath string // path to dot-galaxy file
	DryRun        bool   // dry-run flag
	LogLevel      string // log verboseness
	Environments  string // target environment names, comma separated
	Namespaces    string // target namespaces, comma separated

	*LandscaperConfig
}

// GetEnvironments as slice of strings based on environments.
func (c *Config) GetEnvironments() []string {
	return splitOnComma(c.Environments)
}

// GetNamespaces slice of strings based on namespaces.
func (c *Config) GetNamespaces() []string {
	return splitOnComma(c.Namespaces)
}

// LandscaperConfig runtime configuration related to Landscaper.
type LandscaperConfig struct {
	InCluster        bool   // inside a Kubernetes cluster
	KubeConfig       string // path to alternative ~/.kube/config
	KubeContext      string // kubernetes context
	HelmHome         string // helm home folder
	TillerNamespace  string // helm tiller kubernetes namespace
	TillerPort       int    // helm tiller pod service port
	TillerTimeout    int64  // helm tiller connection timeout (seconds)
	OverrideFile     string // configuration override file
	WaitForResources bool   // wait for resources flag
	WaitTimeout      int64  // wait for resources timeout
	DisabledStages   string // comma separated list of disabled stages
}

// GetDisabledStages return a slice of strings based on disabled stages.
func (l *LandscaperConfig) GetDisabledStages() []string {
	return splitOnComma(l.DisabledStages)
}

// splitOnComma using strings.Split, or empty slice in case of empty string.
func splitOnComma(str string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}

// NewConfig with default values.
func NewConfig() *Config {
	return &Config{
		LogLevel:      "error",
		DryRun:        false,
		DotGalaxyPath: ".galaxy.yaml",
		Environments:  "",
		Namespaces:    "",
		LandscaperConfig: &LandscaperConfig{
			HelmHome:        os.ExpandEnv("${HOME}/.helm"),
			TillerNamespace: "kube-system",
			TillerPort:      44134,
			TillerTimeout:   30,
			WaitTimeout:     60,
		},
	}
}
