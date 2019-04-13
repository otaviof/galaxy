package galaxy

import (
	"fmt"
	"strings"

	"github.com/ryanuber/columnize"
	log "github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

// Printer is a helper to display galaxy related data in command-line.
type Printer struct {
	logger *log.Entry // logger
	data   Data       // galaxy data
}

// actOnSecret to be executed against each secret entry.
type actOnSecret func(ns string, secret SecretManifest)

// actOnRelease to be executed against each release entry.
type actOnRelease func(ns string, release Release)

// Tree formated version of secrets and releases.
func (p *Printer) Tree() string {
	t := treeprint.New()
	trunk := make(map[string]treeprint.Tree)
	branches := make(map[string]treeprint.Tree)

	p.loopData(func(logger *log.Entry, env string, ctx *Context) error {
		if _, exists := trunk[env]; !exists {
			trunk[env] = t.AddBranch(env)
		}

		p.loopSecrets(ctx, func(ns string, secret SecretManifest) {
			if _, exists := branches[ns]; !exists {
				branches[ns] = trunk[env].AddBranch(ns)
			}

			branch := branches[ns].AddBranch(fmt.Sprintf("%s (%s)",
				secret.File, p.formatSecretTypes(secret),
			))
			branch.AddNode(p.formatSecretData(secret))
		})

		p.loopReleases(ctx, func(ns string, release Release) {
			if _, exists := branches[ns]; !exists {
				branches[ns] = trunk[env].AddBranch(ns)
			}

			branch := branches[ns].AddBranch(fmt.Sprintf("%s (%s)",
				release.File, release.Component.Release.Chart,
			))
			branch.AddNode(fmt.Sprintf("%s (v%s)",
				release.Component.Name, release.Component.Release.Version,
			))
		})
		return nil
	})

	return t.String()
}

// Table formatted data.
func (p *Printer) Table() string {
	lines := []string{}
	lines = append(lines, "ENVIRONMENT | NAMESPACE | TYPE | ITEM | DETAILS | FILE")
	p.loopData(func(logger *log.Entry, env string, ctx *Context) error {
		p.loopSecrets(ctx, func(ns string, secret SecretManifest) {
			lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s | %s",
				env,
				ns,
				"secret",
				p.formatSecretTypes(secret),
				p.formatSecretData(secret),
				secret.File,
			))
		})
		p.loopReleases(ctx, func(ns string, release Release) {
			lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s | %s",
				env,
				ns,
				"release",
				fmt.Sprintf("%s:%s", release.Component.Name, release.Component.Release.Version),
				release.Component.Release.Chart,
				release.File,
			))
		})
		return nil
	})
	return columnize.SimpleFormat(lines)
}

// loopData present in this instance
func (p *Printer) loopData(fn actOnContext) error {
	for env, ctxs := range p.data {
		for _, ctx := range ctxs {
			if err := fn(p.logger, env, ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// loopSecrets present in informed data.
func (p *Printer) loopSecrets(ctx *Context, fn actOnSecret) {
	for ns, secrets := range ctx.Secrets {
		for _, secret := range secrets {
			fn(ns, secret)
		}
	}
}

// loopReleases present in informed data.
func (p *Printer) loopReleases(ctx *Context, fn actOnRelease) {
	for ns, releases := range ctx.Releases {
		for _, release := range releases {
			fn(ns, release)
		}
	}
}

// formatSecretTypes format types found in secret manifest.
func (p *Printer) formatSecretTypes(secret SecretManifest) string {
	var types []string
	for _, data := range secret.Manifest.Secrets {
		types = append(types, data.Type)
	}

	return strings.Join(types, ", ")
}

// formatSecretData format secrets found in data part of manifest.
func (p *Printer) formatSecretData(secret SecretManifest) string {
	var secrets []string

	for group, secretData := range secret.Manifest.Secrets {
		for _, data := range secretData.Data {
			secrets = append(secrets, fmt.Sprintf("%s.%s", group, data.Name))
		}
	}

	return strings.Join(secrets, ", ")
}

// NewPrinter creates new Printer instance.
func NewPrinter(data Data) *Printer {
	return &Printer{logger: log.WithField("type", "printer"), data: data}
}
