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

// Context of releases per namespace directory.
type Context struct {
	logger   *log.Entry           // logger
	releases map[string][]Release // releases per namespace (key)
}

// Release binds together a file and a Landscaper component
type Release struct {
	Namespace string
	File      string
	Component *ldsc.Component
}

// InspectDir look for files with informed extensions.
func (c *Context) InspectDir(ns string, dirPath string, extentions []string) error {
	var err error

	logger := c.logger.WithFields(log.Fields{"namespace": ns, "dir": dirPath, "exts": extentions})
	logger.Infof("Inspecting namespace: '%s'", ns)

	if !isDir(dirPath) {
		return fmt.Errorf("Namespace directory is not found at: '%s'", dirPath)
	}

	for _, ext := range extentions {
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

// AddFile to a namespace.
func (c *Context) AddFile(ns, file string) error {
	var err error

	release := Release{Namespace: ns, File: file}
	if err = yaml.Unmarshal(readFile(file), &release.Component); err != nil {
		return err
	}

	c.releases[ns] = append(c.releases[ns], release)
	return nil
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
