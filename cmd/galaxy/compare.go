package main

import (
	"fmt"
	"strings"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"

	"github.com/otaviof/galaxy/pkg/galaxy"
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

	lines = append(lines, "ENVIRONMENT | NAMESPACE | TYPE | ITEM | DETAILS | FILE")
	for env, ctxs := range p {
		for _, ctx := range ctxs {
			for ns, secrets := range ctx.Secrets {
				for _, secret := range secrets {
					lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s | %s",
						env,
						ns,
						"secret",
						formatSecretTypes(secret),
						formatSecretData(secret),
						secret.File,
					))
				}
			}
			for ns, releases := range ctx.Releases {
				for _, release := range releases {
					lines = append(lines, fmt.Sprintf("%s | %s | %s | %s | %s | %s",
						env,
						ns,
						"release",
						fmt.Sprintf("%s:%s", release.Component.Name, release.Component.Release.Version),
						release.Component.Release.Chart,
						release.File,
					))
				}
			}
		}
	}

	fmt.Println(columnize.SimpleFormat(lines))
}

func formatSecretTypes(secret galaxy.SecretManifest) string {
	var types []string
	for _, data := range secret.Manifest.Secrets {
		types = append(types, data.Type)
	}

	return strings.Join(types, ", ")
}

func formatSecretData(secret galaxy.SecretManifest) string {
	var secrets []string

	for group, secretData := range secret.Manifest.Secrets {
		for _, data := range secretData.Data {
			secrets = append(secrets, fmt.Sprintf("%s.%s", group, data.Name))
		}
	}

	return strings.Join(secrets, ", ")
}

func init() {
	flags := compareCmd.PersistentFlags()

	flags.String("releases", "", "show release information")
	flags.String("files", "", "show files")

	rootCmd.AddCommand(compareCmd)
}
