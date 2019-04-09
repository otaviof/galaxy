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
	logger *log.Entry   // logger
	env    *Environment // environment subject to planing
	ctx    *Context     // current context
	envCtx *Context     // context planned for environment
}

// ContextForEnvironment narrow down context to comply with the rules defined in Environment.
func (p *Plan) ContextForEnvironment() (*Context, error) {
	var err error

	p.logger.Info("Working out a plan...")
	if err = p.filter(); err != nil {
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
func (p *Plan) filter() error {
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
			var skip bool

			logger.Info("Inspecting file...")

			if skip, err = p.skipFile(file); err != nil {
				return err
			}
			if skip {
				logger.Info("Skipping file..")
				continue
			}

			logger.Infof("Adding file on new scope: '%s'", file)
			p.envCtx.AddFile(ns, file)
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

	return p.envCtx.RenameReleases(func(namespace, name string) (string, error) {
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
			p.env.Transform.ReleasePrefix, name, p.env.Transform.ReleaseSuffix,
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

// skipFile based on file name, checks for file suffixes (using "@" based notation).
func (p *Plan) skipFile(file string) (bool, error) {
	var suffixesRe *regexp.Regexp
	var err error

	logger := p.logger.WithField("file", file)

	if suffixesRe, err = regexp.Compile("(@\\w+)"); err != nil {
		p.logger.Errorf("Error on compiling regex: '%s'", err)
		return false, err
	}

	res := suffixesRe.FindStringSubmatch(file)

	// no suffixes are found and empty suffixes are allowed
	if len(res) == 0 && stringSliceContains(p.env.FileSuffixes, "") {
		p.logger.Debugf("Not skipping file based on empty suffix!")
		return false, nil
	}

	for _, suffix := range res {
		logger.Debugf("Found suffix '%s'", suffix)
		suffix = strings.Replace(suffix, "@", "", -1)
		if stringSliceContains(p.env.FileSuffixes, suffix) {
			p.logger.Debugf("Suffix allowed on environment: '%s'", suffix)
			return false, nil
		}
	}

	return true, nil
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
