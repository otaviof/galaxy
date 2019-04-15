package galaxy

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Galaxy holds application runtime items
type Galaxy struct {
	logger        *log.Entry                   // logger
	dotGalaxy     *DotGalaxy                   // global configuration
	cfg           *Config                      // runtime configuration
	original      Data                         // original contexts per env
	Modified      Data                         // modified contexts per env
	envOriginalNs map[string]map[string]string // mapping original namespace names per env
}

// Data belonging to Galaxy, having environment name as key and a list of contexts
type Data map[string][]*Context

// actOnContext called during Loop method
type actOnContext func(logger *log.Entry, env string, ctx *Context) error

// Inspect directories and files per namespace, create and populate the context.
func (g *Galaxy) Inspect() error {
	if !isDir(g.dotGalaxy.Spec.Namespaces.BaseDir) {
		return fmt.Errorf("base directory not found at: %s", g.dotGalaxy.Spec.Namespaces.BaseDir)
	}

	return g.Loop(func(logger *log.Entry, env string, ctx *Context) error {
		g.original[env] = append(g.original[env], ctx)
		return nil
	})
}

// Plan manage the scope of changes, by checking which release files should be in.
func (g *Galaxy) Plan() error {
	g.logger.Infof("Planing for namespaces '%s' on environments '%s'",
		g.cfg.GetNamespaces(), g.cfg.GetEnvironments(),
	)
	return g.Loop(func(logger *log.Entry, envName string, ctx *Context) error {
		var env *Environment
		var modified *Context
		var err error

		if len(g.cfg.GetEnvironments()) > 0 && !stringSliceContains(g.cfg.GetEnvironments(), envName) {
			logger.Infof("Skipping environment '%s'!", envName)
			return nil
		}

		if env, err = g.dotGalaxy.GetEnvironment(envName); err != nil {
			return err
		}

		logger.Info("Planing...")
		plan := NewPlan(env, g.cfg.GetNamespaces(), ctx)
		if modified, err = plan.ContextForEnvironment(); err != nil {
			return err
		}

		// saving original namespace names
		g.envOriginalNs[envName] = plan.OriginalNs
		// saving planned data
		g.Modified[envName] = append(g.Modified[envName], modified)
		return nil
	})
}

// Apply changes planned just before.
func (g *Galaxy) Apply() error {
	var e *Environment
	var data []*Context
	var originalNs map[string]string
	var found bool
	var err error

	g.logger.Infof("DRY-RUN: '%v', Environment: '%s'", g.cfg.DryRun, g.cfg.GetEnvironments())

	if len(g.cfg.GetEnvironments()) != 1 {
		return fmt.Errorf("a single environment must be informed")
	}
	envName := g.cfg.GetEnvironments()[0]
	logger := g.logger.WithFields(log.Fields{"env": envName, "dryRun": g.cfg.DryRun})
	logger.Infof("Applying changes for environment...")

	if data, found = g.Modified[envName]; !found {
		return fmt.Errorf("environment '%s' is not found on planned data", envName)
	}
	logger.Debugf("data: '%#v'", data)
	logger.Debugf("data len: '%d'", len(data))

	for _, c := range data {
		logger.Debugf("context: '%#v'", c)
	}

	if originalNs, found = g.envOriginalNs[envName]; !found {
		return fmt.Errorf("environment '%s' is not found on original namespace names map", envName)
	}
	if e, err = g.dotGalaxy.GetEnvironment(envName); err != nil {
		return err
	}

	l := NewLandscaper(g.cfg.LandscaperConfig, e, data)
	for ns, originalNs := range originalNs {
		logger.Infof("Handling namespace '%s', original name '%s'", ns, originalNs)
		if err = l.Bootstrap(ns, originalNs, g.cfg.DryRun); err != nil {
			return err
		}
		if err = l.Apply(); err != nil {
			return err
		}
	}
	return nil
}

// Loop over environments and its contexts.
func (g *Galaxy) Loop(fn actOnContext) error {
	var exts = g.dotGalaxy.Spec.Namespaces.Extensions
	var err error

	logger := g.logger.WithField("exts", exts)
	for _, env := range g.dotGalaxy.ListEnvironments() {
		ctx := NewContext()
		logger = g.logger.WithField("env", env)

		for _, ns := range g.dotGalaxy.ListNamespaces() {
			var baseDir string

			if baseDir, err = g.dotGalaxy.GetNamespaceDir(ns); err != nil {
				return err
			}
			logger.Infof("Inspecting namespace '%s', directory '%s'", ns, baseDir)
			if err = ctx.InspectDir(ns, baseDir, exts); err != nil {
				logger.Fatalf("error during inspecting context: %#v", err)
				return err
			}
		}

		if err = fn(logger, env, ctx); err != nil {
			return err
		}
	}
	return nil
}

// NewGalaxy instantiages a new application instance.
func NewGalaxy(dotGalaxy *DotGalaxy, cfg *Config) *Galaxy {
	return &Galaxy{
		logger:        log.WithFields(log.Fields{"type": "galaxy", "dryRun": cfg.DryRun}),
		dotGalaxy:     dotGalaxy,
		cfg:           cfg,
		original:      make(Data),
		Modified:      make(Data),
		envOriginalNs: make(map[string]map[string]string),
	}
}
