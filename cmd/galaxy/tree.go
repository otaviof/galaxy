package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Run:   runTreeCmd,
	Short: "tree",
	Long:  ``,
}

func runTreeCmd(cmd *cobra.Command, args []string) {
	p := plan()
	t := treeprint.New()

	for env, ctxs := range p {
		branch := t.AddBranch(env)
		for _, ctx := range ctxs {
			for ns, releases := range ctx.Releases {
				branch := branch.AddBranch(ns)
				for _, release := range releases {
					branch := branch.AddBranch(fmt.Sprintf("%s (%s)",
						release.File, release.Component.Release.Chart,
					))
					branch.AddNode(fmt.Sprintf("%s (v%s)",
						release.Component.Name, release.Component.Release.Version,
					))
				}
			}
		}
	}

	fmt.Println(t.String())
}

func init() {
	// flags := treeCmd.PersistentFlags()

	rootCmd.AddCommand(treeCmd)
}
