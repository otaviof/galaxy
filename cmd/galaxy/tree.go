package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Run:   runTreeCmd,
	Short: "tree",
	Long:  ``,
}

func runTreeCmd(cmd *cobra.Command, args []string) {
	data := plan()
	printer := galaxy.NewPrinter(cfg, data)
	fmt.Println(printer.Tree())
}

func init() {
	rootCmd.AddCommand(treeCmd)
}
