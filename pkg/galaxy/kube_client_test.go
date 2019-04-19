package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var kubeClient *KubeClient

func TestKubeClientNew(t *testing.T) {
	cfg := NewConfig()
	kubeClient = NewKubeClient(cfg.KubernetesConfig)
}

func TestKubeClientLoad(t *testing.T) {
	err := kubeClient.Load()
	assert.Nil(t, err)
	assert.NotNil(t, kubeClient.Client)
}
