package galaxy

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

/**
 * Works on a Context to enforce rules defined by a given Environment.
 */

// Plan holds methods to plan releases for a given environment.
type Plan struct {
	log *logrus.Logger // logger
	env *Environment
	ctx *Context
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
func (p *Plan) extractFileSuffix(file string, extensions []string) (string, error) {
	var fileRe *regexp.Regexp
	var suffixRe *regexp.Regexp
	var res []string
	var err error

	extExpr := fmt.Sprintf("(.*?)\\.(%s)$", strings.Join(extensions, "|")) // known extensions
	suffixExpr := ".*?-(\\w+)$"                                            // name convention

	if fileRe, err = regexp.Compile(extExpr); err != nil {
		return "", err
	}
	if suffixRe, err = regexp.Compile(suffixExpr); err != nil {
		return "", err
	}

	// removing extension from file name
	res = fileRe.FindStringSubmatch(file)
	if len(res) != 3 {
		return "", errors.New(
			"Unable to parse file '" + file + "' parts=" + fmt.Sprintf("%#v", res))
	}
	// using the file without extension, applying regex to get suffix
	res = suffixRe.FindStringSubmatch(res[1])
	if len(res) == 2 {
		p.log.Infof("[Plan] file='%s', suffix='%s'", file, res[1])
		return res[1], nil
	}

	p.log.Infof("[Plan] file='%s', suffix=''", file)
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
func (p *Plan) ContextForEnvironment(extensions []string) (*Context, error) {
	var suffix string
	var err error

	envContext := NewContext(p.log)

	p.log.Infof("[Plan] Working out a plan for '%s' environment.", p.env.Name)

	for ns, files := range p.ctx.GetNamespaceFilesMap() {
		p.log.Infof("[Plan] Namespace: '%s', files: '%s'", ns, formatSlice(files))

		if p.skipOnNamespace(ns) {
			p.log.Infof("[Plan] Skipping namespace '%s' in environment!", ns)
			continue
		}

		// altering the namespace name for environment
		ns = p.namespaceName(ns)
		p.log.Infof("[Plan] Target namespace: '%s'", ns)

		for _, file := range files {
			if suffix, err = p.extractFileSuffix(file, extensions); err != nil {
				return nil, err
			}
			p.log.Infof("[Plan] Inspecting file: '%s' (suffix='%s')", file, suffix)

			// checking if file is applicable in current environment
			if p.skipOnSuffix(suffix) {
				p.log.Info("[Plan] Skipping!")
				continue
			}

			p.log.Infof("[Plan] Adding to new scope: '%s'", file)
			envContext.AddFile(ns, file)
		}
	}

	log.Printf("[Plan] New-Context: '%#v'", envContext)
	return envContext, nil
}

// NewPlan creates a new Plan type instance.
func NewPlan(log *logrus.Logger, env *Environment, ctx *Context) *Plan {
	return &Plan{log: log, env: env, ctx: ctx}
}
