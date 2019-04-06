package main

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Run:   runCompareCmd,
	Short: ``,
	Long:  ``,
}

func runCompareCmd(cmd *cobra.Command, args []string) {
	var lines []string

	p := plan()

	lines = append(lines, "ENVIRONMENT | NAMESPACE | RELEASE | VERSION | CHART | FILE")
	for env, ctxs := range p {
		for _, ctx := range ctxs {
			for ns, releases := range ctx.GetNamespaceReleasesMap() {
				for _, release := range releases {
					lines = append(lines, fmt.Sprintf(
						"%s | %s | %s | %s | %s | %s",
						env,
						ns,
						release.Component.Name,
						release.Component.Release.Version,
						release.Component.Release.Chart,
						release.File,
					))
				}
			}
		}
	}

	fmt.Println(columnize.SimpleFormat(lines))
}

func init() {
	flags := compareCmd.PersistentFlags()

	flags.String("releases", "", "show release information")
	flags.String("files", "", "show files")

	rootCmd.AddCommand(compareCmd)
}
