package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Run:   runCompareCmd,
	Short: `Print out Galaxy data in a table format`,
}

func runCompareCmd(cmd *cobra.Command, args []string) {
	g := galaxyPlan()
	printer := galaxy.NewPrinter(g.Modified)
	fmt.Println(printer.Table())
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
