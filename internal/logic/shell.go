package logic

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

type ShellClient struct {
	*Client
}

func NewShellClient(c *config.App) *ShellClient {
	return &ShellClient{new(c)}
}

func (c *ShellClient) RunCmdShell(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	project, err := c.getProject(ctx)
	if err != nil {
		return errors.Wrap(err, "getting project")
	}

	projectTypeVersion, _, err := c.RewardCloud.ProjectTypeVersionApi.ApiProjectTypeVersionsIdGet(ctx, GetIDFromPath(project.GetProjectTypeVersion())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type version")
	}

	projectType, _, err := c.RewardCloud.ProjectTypeApi.ApiProjectTypesIdGet(ctx, GetIDFromPath(projectTypeVersion.GetProjectType())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	cluster, err := c.getCluster(ctx)
	if err != nil {
		return errors.Wrap(err, "getting cluster")
	}

	kubeconfigFile, err := c.prepareKubeconfig(cluster)
	if err != nil {
		return errors.Wrap(err, "preparing kubeconfig")
	}

	podname, err := c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"get",
		"pod",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		"--selector", "reward.itg.cloud/component=main",
		"-o", "jsonpath={.items[0].metadata.name}",
	},
		shell.WithSuppressOutput(true), shell.WithCatchOutput(true))

	if string(podname) == "" {
		return errors.New("cannot find shell pod")
	}

	_, err = c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"exec",
		"-i",
		"-t",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		string(podname),
		"-c", strings.ToLower(projectType.GetName()),
		"--",
		"sh", "-c", "bash || sh",
	},
		shell.WithSuppressOutput(false), shell.WithCatchOutput(false))

	if err != nil {
		return err
	}

	return nil
}
