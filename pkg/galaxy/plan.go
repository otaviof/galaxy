package galaxy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/buildkite/interpolate"
	log "github.com/sirupsen/logrus"
)

// Plan holds methods to plan releases for a given environment.
type Plan struct {
	logger *log.Entry // logger
	env    *Environment
	ctx    *Context
	envCtx *Context
}

// ContextForEnvironment narrow down context to comply with the rules defined in Environment.
func (p *Plan) ContextForEnvironment(exts []string) (*Context, error) {
	var err error

	p.logger.Info("Working out a plan...")
	if err = p.filter(exts); err != nil {
		return nil, err
	}
	if p.env.Transform.ReleasePrefix != "" || p.env.Transform.ReleaseSuffix != "" {
		if err = p.renameReleases(); err != nil {
			return nil, err
		}
	}
	p.renameNamespaces()

	return p.envCtx, nil
}

// filter based on namespace name, using skipOnNamespaces and onlyOnNamespaces, and files based in
// file name and its suffix.
func (p *Plan) filter(exts []string) error {
	var err error

	p.logger.Info("Filtering files...")
	for ns, files := range p.ctx.GetNamespaceFilesMap() {
		logger := p.logger.WithField("namespace", ns)
		logger.Infof("Planing namespace, %d files", len(files))

		if p.skipOnNamespace(ns) {
			logger.Info("Skipping namespace in environment!")
			continue
		}

		logger = logger.WithField("target-namespace", ns)
		logger.Infof("Acquiring target namespace: '%s'", ns)

		for _, file := range files {
			var suffix string

			if suffix, err = p.extractFileSuffix(file, exts); err != nil {
				return err
			}
			logger = logger.WithFields(log.Fields{"file": file, "suffix": suffix})
			logger.Info("Inspecting file...")

			// checking if file is applicable in current environment
			if p.skipOnSuffix(suffix) {
				logger.Info("Skipping file on based on suffix!")
				continue
			}

			logger.Infof("Adding file to new scope: '%s'", file)
			p.envCtx.AddReleaseFile(ns, file)
		}
	}

	return nil
}

// renameReleases execute the rename of releases passing a method along, it also exports a number
// of interpolation variables to be replacted on release name.
func (p *Plan) renameReleases() error {
	logger := p.logger.WithFields(log.Fields{
		"prefix": p.env.Transform.ReleasePrefix,
		"suffix": p.env.Transform.ReleaseSuffix,
	})
	logger.Info("Renaming releases...")

	return p.envCtx.RenameLandscaperReleases(func(namespace, name string) (string, error) {
		var err error

		placeholders := interpolate.NewSliceEnv([]string{
			fmt.Sprintf("NAMESPACE=%s", namespace),
			fmt.Sprintf("RELEASE_NAMESPACE=%s", namespace),
			fmt.Sprintf("RELEASE_NAME=%s", name),
			fmt.Sprintf("RELEASE_PREFIX=%s", p.env.Transform.ReleasePrefix),
			fmt.Sprintf("RELEASE_SUFFIX=%s", p.env.Transform.ReleaseSuffix),
			fmt.Sprintf("NAMESPACE_PREFIX=%s", p.env.Transform.NamespacePrefix),
			fmt.Sprintf("NAMESPACE_SUFFIX=%s", p.env.Transform.NamespaceSuffix),
		})
		releaseName := fmt.Sprintf("%s%s%s",
			p.env.Transform.ReleasePrefix,
			name,
			p.env.Transform.ReleaseSuffix,
		)
		if releaseName, err = interpolate.Interpolate(placeholders, releaseName); err != nil {
			return "", err
		}
		logger.WithFields(log.Fields{"namespace": namespace, "name": name}).
			Debugf("Release named '%s' is renamed to '%s'", name, releaseName)

		return releaseName, nil
	})
}

// renameNamespaces execute the rename of namespaces by passing a method along.
func (p *Plan) renameNamespaces() {
	p.logger.Infof("Renaming namespaces...")
	p.envCtx.RenameNamespaces(func(ns string) string {
		if p.env.Transform.NamespacePrefix != "" || p.env.Transform.NamespaceSuffix != "" {
			return fmt.Sprintf("%s%s%s",
				p.env.Transform.NamespacePrefix, ns, p.env.Transform.NamespaceSuffix,
			)
		}
		return ns
	})
}

// skipOnNamespace check if informed namespace is configured to be skipped in environment.
func (p *Plan) skipOnNamespace(ns string) bool {
	var s string

	for _, s = range p.env.OnlyOnNamespaces {
		if ns == s {
			return false
		}
	}
	for _, s = range p.env.SkipOnNamespaces {
		if ns == s {
			return true
		}
	}

	return false
}

// extractFileSuffix apply regex to file name in order to extract allowed extensions and suffixes,
// and return only suffix part.
func (p *Plan) extractFileSuffix(file string, exts []string) (string, error) {
	var fileRe *regexp.Regexp
	var suffixRe *regexp.Regexp
	var err error

	logger := p.logger.WithFields(log.Fields{"file": file, "exts": exts})

	extExpr := fmt.Sprintf("(.*?)\\.(%s)$", strings.Join(exts, "|")) // known extensions
	suffixExpr := ".*?-(\\w+)$"                                      // name convention

	if fileRe, err = regexp.Compile(extExpr); err != nil {
		return "", err
	}
	if suffixRe, err = regexp.Compile(suffixExpr); err != nil {
		return "", err
	}

	// removing extension from file name
	res := fileRe.FindStringSubmatch(file)
	if len(res) != 3 {
		return "", fmt.Errorf("unable to parse file '%s', using parts '%#v'", file, res)
	}
	// using the file without extension, applying regex to get suffix
	res = suffixRe.FindStringSubmatch(res[1])
	if len(res) == 2 {
		logger.Infof("Suffix: '%s'", res[1])
		return res[1], nil
	}

	logger.Infof("No suffix found for file!")
	return "", nil
}

// skipOnSuffix boolean, returns true when suffix is not in allowed slice.
func (p *Plan) skipOnSuffix(suffix string) bool {
	// when suffix string is not present in the fileSuffixes slice
	return !stringSliceContains(p.env.FileSuffixes, suffix)
}

// NewPlan creates a new Plan type instance.
func NewPlan(env *Environment, ctx *Context) *Plan {
	return &Plan{
		logger: log.WithFields(log.Fields{"type": "plan", "env": env.Name}),
		env:    env,
		ctx:    ctx,
		envCtx: NewContext(),
	}
}
