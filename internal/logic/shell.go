package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	"github.com/spf13/cobra"
	corev1 "golang.org/x/build/kubernetes/api"
)

type ShellClient struct {
	*Client
}

func NewShellClient(c *config.App) *ShellClient {
	return &ShellClient{New(c)}
}

func (c *ShellClient) RunCmdShell(cmd *cobra.Command, args []string) error {
	err := c.CheckKubectl()
	if err != nil {
		return errors.Wrap(err, "checking kubectl")
	}

	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(context.Background())
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	project, err := c.getProject(ctx)
	if err != nil {
		return errors.Wrap(err, "getting project")
	}

	projectTypeVersion, _, err := c.RewardCloud.ProjectTypeVersionApi.ApiProjectTypeVersionsIdGet(
		ctx, GetIDFromPath(project.GetProjectTypeVersion())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type version")
	}

	projectType, _, err := c.RewardCloud.ProjectTypeApi.ApiProjectTypesIdGet(
		ctx, GetIDFromPath(projectTypeVersion.GetProjectType())).Execute()
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

	out, err := c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"get",
		"pod",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		"--selector", "reward.itg.cloud/component=main",
		"-o", "json",
	},
		shell.WithSuppressOutput(true), shell.WithCatchOutput(true))
	if err != nil {
		return errors.Wrapf(err, "running kubectl command, command output: %s", string(out))
	}

	// Remove "Opening in existing browser session.\n" from output
	if strings.Contains(string(out), "Opening in existing browser session.") {
		newout := strings.TrimPrefix(string(out), "Opening in existing browser session.\n")
		out = []byte(newout)
	}

	var obj corev1.PodList
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling pod list, command output: %s", string(out))
	}

	if len(obj.Items) != 1 {
		return errors.Errorf("cannot find shell pod, command output: %s", string(out))
	}

	podname := obj.Items[0].Name

	_, err = c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"exec",
		"-i",
		"-t",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		podname,
		"-c", strings.ToLower(projectType.GetName()),
		"--",
		"sh", "-c", "bash || sh",
	},
		shell.WithSuppressOutput(false), shell.WithCatchOutput(false))

	if err != nil {
		return errors.Wrap(err, "running shell")
	}

	return nil
}
