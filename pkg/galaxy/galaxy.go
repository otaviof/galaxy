package galaxy

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Galaxy holds application runtime items
type Galaxy struct {
	logger    *log.Entry // logger
	dotGalaxy *DotGalaxy // global configuration
	cfg       *Config    // runtime configuration
	original  Data       // original contexts per environment
	Modified  Data       // modified contexts per environment
}

// Data belonging to Galaxy, having environment name as key and a list of contexts
type Data map[string][]*Context

// actOnContext called during Loop method
type actOnContext func(logger *log.Entry, env string, ctx *Context) error

// Inspect directories and files per namespace, create and populate the context.
func (a *Galaxy) Inspect() error {
	if !isDir(a.dotGalaxy.Spec.Namespaces.BaseDir) {
		return fmt.Errorf("base directory not found at: %s", a.dotGalaxy.Spec.Namespaces.BaseDir)
	}

	return a.Loop(func(logger *log.Entry, env string, ctx *Context) error {
		a.original[env] = append(a.original[env], ctx)
		return nil
	})
}

// Plan manage the scope of changes, by checking which release files should be in.
func (a *Galaxy) Plan() error {
	return a.Loop(func(logger *log.Entry, envName string, ctx *Context) error {
		var env *Environment
		var modified *Context
		var err error

		if len(a.cfg.GetEnvironments()) > 0 && !stringSliceContains(a.cfg.GetEnvironments(), envName) {
			logger.Infof("Skipping environment '%s'!", envName)
			return nil
		}

		if env, err = a.dotGalaxy.GetEnvironment(envName); err != nil {
			return err
		}

		logger.Info("Planing...")
		plan := NewPlan(env, a.cfg.GetNamespaces(), ctx)
		if modified, err = plan.ContextForEnvironment(); err != nil {
			return err
		}

		a.Modified[envName] = append(a.Modified[envName], modified)
		return nil
	})
}

// Apply changes planned just before.
func (a *Galaxy) Apply() error {
	return nil
}

// Loop over environments and its contexts.
func (a *Galaxy) Loop(fn actOnContext) error {
	var exts = a.dotGalaxy.Spec.Namespaces.Extensions
	var err error

	logger := a.logger.WithField("exts", exts)
	for _, env := range a.dotGalaxy.ListEnvironments() {
		ctx := NewContext()
		logger = a.logger.WithField("env", env)

		for _, ns := range a.dotGalaxy.ListNamespaces() {
			var baseDir string

			if baseDir, err = a.dotGalaxy.GetNamespaceDir(ns); err != nil {
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
		logger:    log.WithField("type", "galaxy"),
		dotGalaxy: dotGalaxy,
		cfg:       cfg,
		original:  make(Data),
		Modified:  make(Data),
	}
}
