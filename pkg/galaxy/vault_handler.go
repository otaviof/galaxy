package galaxy

import (
	log "github.com/sirupsen/logrus"

	vh "github.com/otaviof/vault-handler/pkg/vault-handler"
)

// VaultHandler manage copying data from Vault to Kubernetes secrets.
type VaultHandler struct {
	logger     *log.Entry          // logger
	cfg        *VaultHandlerConfig // vault-handler configuration
	kubeCfg    *KubernetesConfig   // kubernetes configuration
	handlerCfg *vh.Config          // handler configuration
	handler    *vh.Handler         // handler instance
	ctxs       []*Context          // slice of context instances
}

// Apply rollout secrets copy from Vault to Kubernetes.
func (v *VaultHandler) Apply() error {
	var err error

	for _, manifest := range v.pickManifests(v.handlerCfg.Namespace) {
		if err = v.handler.Copy(manifest); err != nil {
			return err
		}
	}
	return nil
}

// Bootstrap instantiate handler and execute configuration validation and authentication steps.
func (v *VaultHandler) Bootstrap(ns string, dryRun bool) error {
	var err error

	v.handlerCfg = v.setupVaultHandlerConfig(ns, dryRun)
	if err = v.handlerCfg.Validate(); err != nil {
		return err
	}
	if err = v.handlerCfg.ValidateKubernetes(); err != nil {
		return err
	}

	if v.handler, err = vh.NewHandler(v.handlerCfg); err != nil {
		return err
	}
	return v.handler.Authenticate()
}

// setupVaultHandlerConfig create a vault-handler configuration object based on input config.
func (v *VaultHandler) setupVaultHandlerConfig(ns string, dryRun bool) *vh.Config {
	return &vh.Config{
		Context:       v.kubeCfg.KubeContext,
		DryRun:        dryRun,
		InCluster:     v.kubeCfg.InCluster,
		KubeConfig:    v.kubeCfg.KubeConfig,
		Namespace:     ns,
		VaultAddr:     v.cfg.VaultAddr,
		VaultRoleID:   v.cfg.VaultRoleID,
		VaultSecretID: v.cfg.VaultSecretID,
		VaultToken:    v.cfg.VaultToken,
	}
}

func (v *VaultHandler) pickManifests(ns string) []*vh.Manifest {
	var secretManifests []SecretManifest
	var manifests []*vh.Manifest
	var found bool

	for _, ctx := range v.ctxs {
		if secretManifests, found = ctx.Secrets[ns]; !found {
			continue
		}

		for _, secret := range secretManifests {
			manifests = append(manifests, secret.Manifest)
		}
	}

	return manifests
}

// NewVaultHandler creates a new vault-handler instance.
func NewVaultHandler(cfg *VaultHandlerConfig, kubeCfg *KubernetesConfig, ctxs []*Context) *VaultHandler {
	return &VaultHandler{
		logger:  log.WithField("type", "vaultHandler"),
		cfg:     cfg,
		kubeCfg: kubeCfg,
		ctxs:    ctxs,
	}
}
