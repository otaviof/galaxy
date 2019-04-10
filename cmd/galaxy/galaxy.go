package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

var rootCmd = &cobra.Command{
	Use:   "galaxy",
	Short: ``,
	Long:  ``,
}

var cfg = &galaxy.Config{}

// bootstrap reads the configuration from command-line informed place, and set log-level
func bootstrap() *galaxy.DotGalaxy {
	var dotGalaxy *galaxy.DotGalaxy
	var level log.Level
	var err error

	if dotGalaxy, err = galaxy.NewDotGalaxy(cfg.DotGalaxyPath); err != nil {
		log.Fatalf("[ERROR] Parsing dot-galaxy file ('%s'): %s", cfg.DotGalaxyPath, err)
	}
	if level, err = log.ParseLevel(cfg.LogLevel); err != nil {
		log.Fatalf("[ERROR] Setting log-level ('%s'): %s", cfg.LogLevel, err)
	}
	log.SetLevel(level)

	return dotGalaxy
}

// plan execute planning phase of Galaxy.
func plan() galaxy.Data {
	var err error

	dotGalaxy := bootstrap()
	g := galaxy.NewGalaxy(dotGalaxy, cfg)

	if err = g.Plan(); err != nil {
		log.Fatal(err)
	}

	return g.Modified
}

// init command-line arguments
func init() {
	var flags = rootCmd.PersistentFlags()

	flags.StringVarP(&cfg.DotGalaxyPath, "config", "c", ".galaxy.yaml", "configuration file.")
	flags.BoolVarP(&cfg.DryRun, "dry-run", "d", false, "dry-run mode.")
	flags.StringVarP(&cfg.LogLevel, "log-level", "l", "error", "logging level.")
}

func main() {
	var err error

	if err = rootCmd.Execute(); err != nil {
		panic(fmt.Sprintf("[ERROR] %s", err))
	}
}
