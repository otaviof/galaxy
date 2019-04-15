package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Run:   runApplyCmd,
	Short: ``,
	Long:  ``,
}

func runApplyCmd(cmd *cobra.Command, args []string) {
	g := galaxyPlan()

	if log.GetLevel() < log.InfoLevel {
		galaxy.SetLogLevel("info")
	}

	if err := g.Apply(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	flags := applyCmd.PersistentFlags()

	flags.String("disable", "", "actions to disable, as in \"create\", \"update\" or \"delete\"")
	flags.String("helm-home", "${HOME}/.helm", "helm home folder path")
	flags.Bool("in-cluster", false, "running inside a Kubernetes cluster")
	flags.String("kube-config", "", "alternative kube-config path")
	flags.String("kube-context", "", "alternative Kubernetes context")
	flags.String("override-file", "", "Landscaper configuration override file")
	flags.String("tiller-namespace", "kube-system", "Helm's Tiller namespace")
	flags.Int("tiller-port", 44134, "Helm's Tiller service port")
	flags.Int64("tiller-timeout", 30, "timeout on trying to reach tiller, in seconds")
	flags.Bool("wait", false, "wait for resources to be ready")
	flags.Int64("wait-timeout", 120, "timeout on waiting for resources, in seconds")

	cobra.MarkFlagRequired(flags, "env")
	rootCmd.AddCommand(applyCmd)

	if err := viper.BindPFlags(flags); err != nil {
		log.Fatal(err)
	}
}
