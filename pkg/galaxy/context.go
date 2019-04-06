package galaxy

/**
 * Context is related to the files present per namespace, as in per namespace directory.
 */

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"path"
	"path/filepath"

	ldsc "github.com/Eneco/landscaper/pkg/landscaper"
	log "github.com/sirupsen/logrus"
)

// Context of releases per namespace directory, a context is unique per environment.
type Context struct {
	logger   *log.Entry           // logger
	releases map[string][]Release // releases per namespace (key)
}

// Release binds together a file and a Landscaper component
type Release struct {
	Name      string          // original Landscaper release name
	Namespace string          // release namespace
	File      string          // release file path
	Component *ldsc.Component // Landscaper component
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
			if err = c.AddReleaseFile(ns, file); err != nil {
				return err
			}
		}
	}

	logger.Infof("Files: '%s'", formatSlice(c.GetNamespaceFilesMap()[ns]))
	return nil
}

// AddReleaseFile adds a release file to a namespace.
func (c *Context) AddReleaseFile(ns, file string) error {
	var err error

	release := Release{Namespace: ns, File: file}
	if err = yaml.Unmarshal(readFile(file), &release.Component); err != nil {
		return err
	}

	c.releases[ns] = append(c.releases[ns], release)
	return nil
}

// RenameLandscaperReleases based on prefix and suffix, rename the existing releases.
func (c *Context) RenameLandscaperReleases(fn ReleaseRenamer) error {
	var err error

	for ns, releases := range c.releases {
		for _, release := range releases {
			if release.Component.Name, err = fn(ns, release.Component.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// RenameNamespaces loop namespaces in this context to rename it based in informed method output.
func (c *Context) RenameNamespaces(fn NamespaceRenamer) {
	var copy = make(map[string][]Release)
	for ns := range c.releases {
		name := fn(ns)
		copy[name] = c.releases[ns]
	}
	c.releases = copy
}

// GetNamespaceFilesMap expose map of namespace and its files
func (c *Context) GetNamespaceFilesMap() map[string][]string {
	filesMap := make(map[string][]string)
	for ns, releases := range c.releases {
		for _, release := range releases {
			filesMap[ns] = append(filesMap[ns], release.File)
		}
	}
	return filesMap
}

// GetNamespaceReleasesMap expose internal releases map, which consists of namespace name and a list
// of release objects.
func (c *Context) GetNamespaceReleasesMap() map[string][]Release {
	return c.releases
}

// NewContext creates a empty new context instance.
func NewContext() *Context {
	return &Context{
		logger:   log.WithField("type", "context"),
		releases: make(map[string][]Release),
	}
}
