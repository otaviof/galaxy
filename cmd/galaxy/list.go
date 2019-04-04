package main

import (
	"fmt"

	galaxy "github.com/otaviof/galaxy/pkg/galaxy"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Run:   runListCmd,
	Short: "List which files are part of a environment plan.",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

// printContext loop context object to print out in terminal.
func printContext(env string, context *galaxy.Context) {
	var ns string
	var file string
	var files []string

	fmt.Printf("# Environment: %s\n", env)
	for ns, files = range context.GetNamespaceFilesMap() {
		fmt.Printf("  Namespace: %s\n", ns)
		for _, file = range files {
			fmt.Printf("    - %s\n", file)
		}
	}
}

// runListCmd executes the main actions in list sub-command.
func runListCmd(cmd *cobra.Command, args []string) {
	var dotGalaxy = loadConfig()
	var extensions = dotGalaxy.Spec.Namespaces.Extensions
	var env string
	var ns string
	var baseDir string
	var err error

	if err = setLogLevel(); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	for _, env = range dotGalaxy.ListEnvironments() {
		var context = galaxy.NewContext(log)

		for _, ns = range dotGalaxy.ListNamespaces() {
			if baseDir, err = dotGalaxy.GetNamespaceDir(ns); err != nil {
				log.Fatal(err)
			}

			if err = context.InspectDir(ns, baseDir, extensions); err != nil {
				log.Fatalf("error during inspecting context: %#v", err)
			}
		}

		printContext(env, context)
	}
}
