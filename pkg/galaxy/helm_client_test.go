package galaxy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var helmClient *HelmClient

func TestHelmClientNew(t *testing.T) {
	cfg = NewConfig()
	cfg.KubeConfig = os.Getenv("KUBECONFIG")

	k := NewKubeClient(cfg.KubeConfig, cfg.KubeContext, cfg.InCluster)
	_ = k.Load()
	helmClient = NewHelmClient(cfg.HelmHome, cfg.TillerNamespace, cfg.TillerPort, cfg.TillerTimeout, k)
}

func TestHelmClientLoad(t *testing.T) {
	err := helmClient.Load()
	assert.Nil(t, err)
	assert.NotNil(t, helmClient.Client)
}
