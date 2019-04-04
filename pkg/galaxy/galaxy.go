package galaxy

/**
 * Represents the application instance, and comprises the application execution steps. Each public
 * method represents the execution phase:
 *   - Inspect: look for directories and release files, secret definitions, etc;
 *   - Plan: based in a target environment, manage the scope;
 *   - Execute: apply changes in environment;
 */

import (
	"errors"
	"path"

	logrus "github.com/sirupsen/logrus"
)

// Galaxy holds application runtime items
type Galaxy struct {
	dotGalaxy *DotGalaxy        // global configuration
	cmdArgs   map[string]string // command-line arguments
	current   *Context          // keeps the original context of each namespace
	future    *Context          // context after modifications
	log       *logrus.Logger    // logger
}

// Inspect directories and files per namespace, create and populate the context.
func (a *Galaxy) Inspect() error {
	var err error

	baseDir := a.dotGalaxy.Spec.Namespaces.BaseDir
	exts := a.dotGalaxy.Spec.Namespaces.Extensions

	if !isDir(a.dotGalaxy.Spec.Namespaces.BaseDir) {
		return errors.New("base directory not found at: " + baseDir)
	}

	for _, ns := range a.dotGalaxy.Spec.Namespaces.Names {
		nsPath := path.Join(baseDir, ns)
		if err = a.current.InspectDir(ns, nsPath, exts); err != nil {
			return err
		}
	}

	return nil
}

// Plan manage the scope of changes, by checking which release files should be in.
func (a *Galaxy) Plan() error {
	var env *Environment
	var plan *Plan
	var err error

	// FIXME: environment name should be informed by parameter;
	if env, err = a.dotGalaxy.GetEnvironment("dev"); err != nil {
		return err
	}

	plan = NewPlan(a.log, env, a.current)

	if a.future, err = plan.ContextForEnvironment(a.dotGalaxy.Spec.Namespaces.Extensions); err != nil {
		return err
	}

	return nil
}

func (a *Galaxy) Execute() error {
	return nil
}

// NewGalaxy instantiages a new application instance.
func NewGalaxy(log *logrus.Logger, dotGalaxy *DotGalaxy, cmdArgs map[string]string) *Galaxy {
	return &Galaxy{
		log:       log,
		dotGalaxy: dotGalaxy,
		cmdArgs:   cmdArgs,
		current:   NewContext(log),
		future:    NewContext(log),
	}
}
