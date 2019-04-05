package galaxy

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

/**
 * Works on a Context to enforce rules defined by a given Environment.
 */

// Plan holds methods to plan releases for a given environment.
type Plan struct {
	logger *log.Entry // logger
	env    *Environment
	ctx    *Context
}

// skipOnNamespace check if informed namespace is configured to be skipped in environment.
func (p *Plan) skipOnNamespace(ns string) bool {
	var skipNs string
	for _, skipNs = range p.env.SkipOnNamespaces {
		if ns == skipNs {
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

// namespaceName alters the name of the namespace based in environment configuration.
func (p *Plan) namespaceName(ns string) string {
	if p.env.Transform.NamespaceSuffix != "" {
		return fmt.Sprintf("%s-%s", ns, p.env.Transform.NamespaceSuffix)
	}
	return ns
}

// ContextForEnvironment narrow down context to comply with the rules defined in Environment.
func (p *Plan) ContextForEnvironment(exts []string) (*Context, error) {
	var suffix string
	var err error

	p.logger.Info("Working out a plan...")
	envContext := NewContext()

	for ns, files := range p.ctx.GetNamespaceFilesMap() {
		logger := p.logger.WithField("namespace", ns)
		logger.Infof("Planing namespace, %d files", len(files))

		if p.skipOnNamespace(ns) {
			logger.Info("Skipping namespace in environment!")
			continue
		}

		// altering the namespace name for environment
		ns = p.namespaceName(ns)
		logger = logger.WithField("target-namespace", ns)
		logger.Infof("Acquiring target namespace: '%s'", ns)

		for _, file := range files {
			if suffix, err = p.extractFileSuffix(file, exts); err != nil {
				return nil, err
			}
			logger = logger.WithFields(log.Fields{"file": file, "suffix": suffix})
			logger.Info("Inspecting file")

			// checking if file is applicable in current environment
			if p.skipOnSuffix(suffix) {
				logger.Info("Skipping file on based on suffix!")
				continue
			}

			logger.Infof("Adding file to new scope: '%s'", file)
			envContext.AddFile(ns, file)
		}
	}

	return envContext, nil
}

// NewPlan creates a new Plan type instance.
func NewPlan(env *Environment, ctx *Context) *Plan {
	return &Plan{
		logger: log.WithFields(log.Fields{"type": "plan", "env": env.Name}),
		env:    env,
		ctx:    ctx,
	}
}
