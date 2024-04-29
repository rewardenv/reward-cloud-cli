package logic

import (
	"github.com/pkg/errors"

	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
)

type PrerequisitesClient struct {
	*Client
}

func NewPrerequisitesClient(c *config.App) *PrerequisitesClient {
	return &PrerequisitesClient{new(c)}
}

// Check if kubectl oidc-login plugin is installed
func (c *PrerequisitesClient) Check() error {
	_, err := c.Kubectl.RunCommand([]string{"oidc-login", "--help"}, shell.WithSuppressOutput(true))
	if err != nil {
		return errors.Wrap(err, "please install oidc-plugin for kubectl: https://github.com/int128/kubelogin?tab=readme-ov-file#setup")
	}

	return nil
}
