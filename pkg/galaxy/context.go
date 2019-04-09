package galaxy

import (
	"fmt"
	"path"
	"path/filepath"

	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	vh "github.com/otaviof/vault-handler/pkg/vault-handler"
)

// Context of releases per namespace directory, a context is unique per environment.
type Context struct {
	logger   *log.Entry                  // logger
	Releases map[string][]Release        // releases per namespace (key)
	Secrets  map[string][]SecretManifest // secret manifests per namespace (key)
}

// Release binds together a file and a Landscaper component
type Release struct {
	Namespace string          // release namespace
	File      string          // release file path
	Component *ldsc.Component // Landscaper component
}

// SecretManifest vault-handler manifest to copy secrets from Vault to Kubernetes.
type SecretManifest struct {
	Namespace string       // release namespace
	File      string       // release file path
	Manifest  *vh.Manifest // vault-handler manifest
}

// ReleaseRenamer method to rename releases in this context
type ReleaseRenamer func(ns, name string) (string, error)

// NamespaceRenamer method to rename namespaces in this context
type NamespaceRenamer func(ns string) string

// InspectDir look for files with informed extensions.
func (c *Context) InspectDir(ns string, dirPath string, exts []string) error {
	var err error

	logger := c.logger.WithFields(log.Fields{"namespace": ns, "dir": dirPath, "exts": exts})
	logger.Infof("Inspecting namespace: '%s'", ns)

	if !isDir(dirPath) {
		return fmt.Errorf("Namespace directory is not found at: '%s'", dirPath)
	}

	for _, ext := range exts {
		var files []string

		extExpr := fmt.Sprintf("*.%s", ext)

		if files, err = filepath.Glob(path.Join(dirPath, extExpr)); err != nil {
			return err
		}
		for _, file := range files {
			logger.Infof("Inspecting file: '%s'", file)
			if err = c.AddFile(ns, file); err != nil {
				return err
			}
		}
	}

	logger.Infof("Files: '%s'", formatSlice(c.GetNamespaceFilesMap()[ns]))
	return nil
}

// AddFile as Landscaper release or Vault-Handler secret manifest. It will try to parse payload first
// as a Landscaper file, and if on errors, it tries as a secret manifest.
func (c *Context) AddFile(ns, file string) error {
	var component *ldsc.Component
	var manifest *vh.Manifest
	var err error

	logger := c.logger.WithFields(log.Fields{"namespace": ns, "file": file})
	logger.Debugf("Adding file '%s' on namespace '%s'", file, ns)

	// trying as a landscaper file first
	if err = yaml.UnmarshalStrict(readFile(file), &component); err == nil {
		logger.Debug("Landscaper release file")
		c.Releases[ns] = append(c.Releases[ns], Release{
			Namespace: ns, File: file, Component: component,
		})
		return nil
	}
	logger.Debugf("Error on parsing file as Landscaper's: '%s'", err)

	logger.Debug("Trying to handle file as a secret manifest...")
	// trying as a secret manifest afterwards
	if err = yaml.UnmarshalStrict(readFile(file), &manifest); err == nil {
		logger.Debug("Valid Vault-Handler secrets manifest file!")
		c.Secrets[ns] = append(c.Secrets[ns], SecretManifest{
			Namespace: ns, File: file, Manifest: manifest,
		})
		return nil
	}
	logger.Debugf("Error on parsing file as Vault-Handler's: '%s'", err)

	return fmt.Errorf("unable to parse file as landscaper release or vault-handler manifest")
}

// RenameReleases based on prefix and suffix, rename the existing releases.
func (c *Context) RenameReleases(fn ReleaseRenamer) error {
	var err error

	for ns, releases := range c.Releases {
		for _, release := range releases {
			if release.Component.Name, err = fn(ns, release.Component.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// RenameNamespaces loop namespaces in this context to rename it based in informed method output,
// applied to releases and secrets in this context.
func (c *Context) RenameNamespaces(fn NamespaceRenamer) {
	var r = make(map[string][]Release)
	var s = make(map[string][]SecretManifest)

	for k, v := range c.Releases {
		r[fn(k)] = v
	}
	c.Releases = r

	for k, v := range c.Secrets {
		s[fn(k)] = v
	}
	c.Secrets = s
}

// GetNamespaceFilesMap expose map of namespace and its files
func (c *Context) GetNamespaceFilesMap() map[string][]string {
	filesMap := make(map[string][]string)

	for ns, releases := range c.Releases {
		for _, release := range releases {
			filesMap[ns] = append(filesMap[ns], release.File)
		}
	}
	for ns, secrets := range c.Secrets {
		for _, secret := range secrets {
			filesMap[ns] = append(filesMap[ns], secret.File)
		}
	}

	return filesMap
}

// NewContext creates a empty new context instance.
func NewContext() *Context {
	return &Context{
		logger:   log.WithField("type", "context"),
		Releases: make(map[string][]Release),
		Secrets:  make(map[string][]SecretManifest),
	}
}
