package main

import (
	"fmt"
	"os"

	galaxy "github.com/otaviof/galaxy/pkg/galaxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "galaxy",
	Short: ``,
	Long:  ``,
}

var config string      // config file path
var environment string // environment name
var logLevel string    // logrus log level
var dryRun bool        // dry-run flag
var log = logrus.New() // logger

// init command-line arguments
func init() {
	var flags = rootCmd.PersistentFlags()

	flags.StringVarP(&config, "config", "c", ".galaxy.yaml", "configuration file.")
	flags.BoolVarP(&dryRun, "dry-run", "d", false, "dry-run mode.")
	flags.StringVarP(&environment, "environment", "e", "", "target environment.")
	flags.StringVarP(&logLevel, "log-level", "l", "error", "logging level.")
}

// setLogLevel interacts with logrus to set logger level.
func setLogLevel() error {
	var level logrus.Level
	var err error

	if level, err = logrus.ParseLevel(logLevel); err != nil {
		return err
	}
	log.SetLevel(level)

	return nil
}

// loadConfig reads the configuration from command-line informed place
func loadConfig() *galaxy.DotGalaxy {
	var dotGalaxy *galaxy.DotGalaxy
	var err error

	if dotGalaxy, err = galaxy.NewDotGalaxy(config); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	return dotGalaxy
}

func main() {
	var err error

	log.Out = os.Stdout

	if err = rootCmd.Execute(); err != nil {
		panic(fmt.Sprintf("[ERROR] %s", err))
	}
}
