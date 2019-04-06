package main

import (
	"fmt"

	galaxy "github.com/otaviof/galaxy/pkg/galaxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "galaxy",
	Short: ``,
	Long:  ``,
}

type cmdLineOptions struct {
	config      string // dot-galaxy file path
	environment string // target environment name
	logLevel    string // log verboseness
	dryRun      bool   // dry-run flag
}

var opts = cmdLineOptions{}

// bootstrap reads the configuration from command-line informed place, and set log-level
func bootstrap() *galaxy.DotGalaxy {
	var dotGalaxy *galaxy.DotGalaxy
	var level log.Level
	var err error

	if dotGalaxy, err = galaxy.NewDotGalaxy(opts.config); err != nil {
		log.Fatalf("[ERROR] Parsing dot-galaxy file ('%s'): %s", opts.config, err)
	}
	if level, err = log.ParseLevel(opts.logLevel); err != nil {
		log.Fatalf("[ERROR] Setting log-level ('%s'): %s", opts.logLevel, err)
	}
	log.SetLevel(level)

	return dotGalaxy
}

func plan() map[string][]*galaxy.Context {
	var err error

	dotGalaxy := bootstrap()
	g := galaxy.NewGalaxy(dotGalaxy, map[string]string{})

	if err = g.Plan(); err != nil {
		log.Fatal(err)
	}

	return g.GetModifiedContextMap()
}

// init command-line arguments
func init() {
	var flags = rootCmd.PersistentFlags()

	flags.StringVarP(&opts.config, "config", "c", ".galaxy.yaml", "configuration file.")
	flags.BoolVarP(&opts.dryRun, "dry-run", "d", false, "dry-run mode.")
	flags.StringVarP(&opts.logLevel, "log-level", "l", "error", "logging level.")
}

func main() {
	var err error

	if err = rootCmd.Execute(); err != nil {
		panic(fmt.Sprintf("[ERROR] %s", err))
	}
}
