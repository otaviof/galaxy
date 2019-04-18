package galaxy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var kubeClient *KubeClient

func TestKubeClientNew(t *testing.T) {
	cfg = NewConfig()
	cfg.KubeConfig = os.Getenv("KUBECONFIG")

	kubeClient = NewKubeClient(cfg.KubeConfig, cfg.KubeContext, cfg.InCluster)
}

func TestKubeClientLoad(t *testing.T) {
	err := kubeClient.Load()
	assert.Nil(t, err)
	assert.NotNil(t, kubeClient.Client)
}
