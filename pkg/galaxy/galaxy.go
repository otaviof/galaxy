package galaxy

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Galaxy holds application runtime items
type Galaxy struct {
	logger      *log.Entry            // logger
	dotGalaxy   *DotGalaxy            // global configuration
	cmdArgs     map[string]string     // command-line arguments
	originalCtx map[string][]*Context // original context per environment
	modifiedCtx map[string][]*Context // modified context per environment
}

// actOnContext called during Loop method
type actOnContext func(logger *log.Entry, env string, ctx *Context) error

// Inspect directories and files per namespace, create and populate the context.
func (a *Galaxy) Inspect() error {
	if !isDir(a.dotGalaxy.Spec.Namespaces.BaseDir) {
		return fmt.Errorf("base directory not found at: %s", a.dotGalaxy.Spec.Namespaces.BaseDir)
	}

	return a.Loop(func(logger *log.Entry, env string, ctx *Context) error {
		a.originalCtx[env] = append(a.originalCtx[env], ctx)
		return nil
	})
}

// Plan manage the scope of changes, by checking which release files should be in.
func (a *Galaxy) Plan() error {
	return a.Loop(func(logger *log.Entry, envName string, ctx *Context) error {
		var env *Environment
		var modified *Context
		var err error

		if env, err = a.dotGalaxy.GetEnvironment(envName); err != nil {
			return err
		}

		logger.Info("Planing...")
		plan := NewPlan(env, ctx)
		if modified, err = plan.ContextForEnvironment(); err != nil {
			return err
		}

		a.modifiedCtx[envName] = append(a.modifiedCtx[envName], modified)
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

// GetModifiedContextMap exposes the modified contexts that have been stored by Plan. It organizes
// the planing per environment, a string used a map's key.
func (a *Galaxy) GetModifiedContextMap() map[string][]*Context {
	return a.modifiedCtx
}

// NewGalaxy instantiages a new application instance.
func NewGalaxy(dotGalaxy *DotGalaxy, cmdArgs map[string]string) *Galaxy {
	return &Galaxy{
		logger:      log.WithField("type", "galaxy"),
		dotGalaxy:   dotGalaxy,
		cmdArgs:     cmdArgs,
		originalCtx: make(map[string][]*Context),
		modifiedCtx: make(map[string][]*Context),
	}
}
