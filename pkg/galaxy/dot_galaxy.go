package galaxy

import (
	"fmt"
	"path"

	"github.com/buildkite/interpolate"
	yaml "gopkg.in/yaml.v2"
)

// DotGalaxy represents the `.galaxy.yaml` configuration file
type DotGalaxy struct {
	Spec Spec `yaml:"galaxy"`
}

// Spec configuration core, linking other types together
type Spec struct {
	Environments []Environment `yaml:"environments"`
	Namespaces   Namespaces    `yaml:"namespaces"`
}

// Environment representation, related to environment scope and transformation
type Environment struct {
	Name             string    `yaml:"name"`
	SkipOnNamespaces []string  `yaml:"skipOnNamespaces"`
	OnlyOnNamespaces []string  `yaml:"onlyOnNamespaces"`
	FileSuffixes     []string  `yaml:"fileSuffixes"`
	Transform        Transform `yaml:"transform"`
}

// Transform configuration on how to transform a release for that environment
type Transform struct {
	NamespacePrefix string `yaml:"namespacePrefix"`
	NamespaceSuffix string `yaml:"namespaceSuffix"`
	ReleasePrefix   string `yaml:"releasePrefix"`
}

// Namespaces in kubernetes, representation to where to find namespace directories and releases
type Namespaces struct {
	BaseDir    string   `yaml:"baseDir"`
	Extensions []string `yaml:"extensions"`
	Names      []string `yaml:"names"`
}

// Interpolate a string based on Environment attributes, plus whats informed.
func (e *Environment) Interpolate(str string, placeholders []string) (string, error) {
	placeholders = append(placeholders, fmt.Sprintf("RELEASE_PREFIX=%s", e.Transform.ReleasePrefix))
	placeholders = append(placeholders, fmt.Sprintf("NAMESPACE_PREFIX=%s", e.Transform.NamespacePrefix))
	placeholders = append(placeholders, fmt.Sprintf("NAMESPACE_SUFFIX=%s", e.Transform.NamespaceSuffix))

	sliceEnv := interpolate.NewSliceEnv(placeholders)

	return interpolate.Interpolate(sliceEnv, str)
}

// ListNamespaces exposes the list with namespace names.
func (d *DotGalaxy) ListNamespaces() []string {
	return d.Spec.Namespaces.Names
}

// ListEnvironments names based in known configuration.
func (d *DotGalaxy) ListEnvironments() []string {
	var list []string
	for _, env := range d.Spec.Environments {
		list = append(list, env.Name)
	}
	return list
}

// GetNamespaceDir returns the path to the namespace directory, or error
func (d *DotGalaxy) GetNamespaceDir(name string) (string, error) {
	if !stringSliceContains(d.Spec.Namespaces.Names, name) {
		return "", fmt.Errorf("namespace informed does not exist '%s'", name)
	}
	if !isDir(d.Spec.Namespaces.BaseDir) {
		return "", fmt.Errorf("baseDir is not a directory '%s'", d.Spec.Namespaces.BaseDir)
	}
	return path.Join(d.Spec.Namespaces.BaseDir, name), nil
}

// GetEnvironment return environment instance based on its name.
func (d *DotGalaxy) GetEnvironment(name string) (*Environment, error) {
	var env Environment
	for _, env = range d.Spec.Environments {
		if name == env.Name {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment is not found '%s'", name)
}

// NewDotGalaxy to load `.galaxy.yml` file.
func NewDotGalaxy(filePath string) (*DotGalaxy, error) {
	dotGalaxy := DotGalaxy{}
	if err := yaml.Unmarshal(readFile(filePath), &dotGalaxy); err != nil {
		return nil, err
	}
	return &dotGalaxy, nil
}
