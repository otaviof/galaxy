package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/otaviof/galaxy/pkg/galaxy"
)

var rootCmd = &cobra.Command{
	Use:   "galaxy",
	Short: ``,
	Long:  ``,
}

// configFromEnv load runtime configuration from environment, which also includes command-line
// parameters by using Viper.
func configFromEnv() *galaxy.Config {
	return &galaxy.Config{
		DotGalaxyPath: viper.GetString("config"),
		DryRun:        viper.GetBool("dry-run"),
		Environments:  viper.GetString("env"),
		Namespaces:    viper.GetString("namespace"),
		LogLevel:      viper.GetString("log-level"),
		SkipSecrets:   viper.GetBool("skip-secrets"),
		KubernetesConfig: &galaxy.KubernetesConfig{
			InCluster:   viper.GetBool("in-cluster"),
			KubeConfig:  viper.GetString("kube-config"),
			KubeContext: viper.GetString("kube-context"),
		},
		LandscaperConfig: &galaxy.LandscaperConfig{
			DisabledStages:   viper.GetString("disable"),
			OverrideFile:     viper.GetString("override-file"),
			HelmHome:         os.ExpandEnv(viper.GetString("helm-home")),
			TillerNamespace:  viper.GetString("tiller-namespace"),
			TillerPort:       viper.GetInt("tiller-port"),
			TillerTimeout:    viper.GetInt64("tiller-timeout"),
			WaitForResources: viper.GetBool("wait"),
			WaitTimeout:      viper.GetInt64("wait-timeout"),
		},
		VaultHandlerConfig: &galaxy.VaultHandlerConfig{
			VaultAddr:     viper.GetString("vault-addr"),
			VaultToken:    viper.GetString("vault-token"),
			VaultRoleID:   viper.GetString("vault-role-id"),
			VaultSecretID: viper.GetString("vault-secret-id"),
		},
	}
}

// bootstrap reads the configuration from command-line informed place, and set log-level
func bootstrap(cfg *galaxy.Config) *galaxy.DotGalaxy {
	var dotGalaxy *galaxy.DotGalaxy
	var err error

	if dotGalaxy, err = galaxy.NewDotGalaxy(cfg.DotGalaxyPath); err != nil {
		log.Fatalf("[ERROR] Parsing dot-galaxy file ('%s'): %s", cfg.DotGalaxyPath, err)
	}
	return dotGalaxy
}

// galaxyPlan return a planned galaxy object.
func galaxyPlan() *galaxy.Galaxy {
	cfg := configFromEnv()
	galaxy.SetLogLevel(cfg.LogLevel)
	log.Debugf("cfg: %#v", cfg)

	dotGalaxy := bootstrap(cfg)
	g := galaxy.NewGalaxy(dotGalaxy, cfg)

	if err := g.Plan(); err != nil {
		log.Fatal(err)
	}

	return g
}

// init command-line arguments
func init() {
	flags := rootCmd.PersistentFlags()

	viper.SetEnvPrefix("galaxy")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	flags.String("config", ".galaxy.yaml", "alternative Galaxy manifest file")
	flags.Bool("dry-run", false, "dry-run mode")
	flags.String("log-level", "error", "logging level")

	flags.String("env", "", "filter by environments, comma separated list")
	flags.String("namespace", "", "filter by namespaces, comma separated list")

	if err := viper.BindPFlags(flags); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("[ERROR] %s", err)
	}
}
