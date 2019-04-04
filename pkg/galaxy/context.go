package galaxy

/**
 * Context is related to the files present per namespace, as in per namespace directory.
 */

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"

	logrus "github.com/sirupsen/logrus"
)

// Context of files per namespace directory.
type Context struct {
	log     *logrus.Logger      // logger
	nsFiles map[string][]string // namespace name and file paths
}

// InspectDir look for files with informed extensions.
func (c *Context) InspectDir(ns string, dirPath string, extentions []string) error {
	var ext string
	var err error

	c.log.Infof("[Context] Inspecting namespace '%s' (directory: '%s')", ns, dirPath)

	if !isDir(dirPath) {
		return errors.New("Namespace directory is not found at: " + dirPath)
	}

	for _, ext = range extentions {
		var extExpr = fmt.Sprintf("*.%s", ext)
		var files []string
		var file string

		if files, err = filepath.Glob(path.Join(dirPath, extExpr)); err != nil {
			return err
		}

		for _, file = range files {
			c.nsFiles[ns] = append(c.nsFiles[ns], file)
		}
	}

	c.log.Infof("[Context] '%s' files: '%s'", ns, formatSlice(c.nsFiles[ns]))
	return nil
}

// GetNamespaceFilesMap expose map of namespace and its files
func (c *Context) GetNamespaceFilesMap() map[string][]string {
	return c.nsFiles
}

// AddFile to a namespace.
func (c *Context) AddFile(ns, file string) {
	c.nsFiles[ns] = append(c.nsFiles[ns], file)
}

// NewContext creates a empty new context instance.
func NewContext(log *logrus.Logger) *Context {
	return &Context{log: log, nsFiles: make(map[string][]string)}
}
