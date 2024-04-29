package logic

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
)

type PortForwardClient struct {
	*Client
}

func NewPortForwardClient(c *config.App) *PortForwardClient {
	return &PortForwardClient{new(c)}
}

func (c *PortForwardClient) RunCmdPortForwardDB(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	project, err := c.getProject(ctx)
	if err != nil {
		return errors.Wrap(err, "getting project")
	}

	projectTypeVersion, _, err := c.RewardCloud.ProjectTypeVersionAPI.ApiProjectTypeVersionsIdGet(ctx, GetIDFromPath(project.GetProjectTypeVersion())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type version")
	}

	_, _, err = c.RewardCloud.ProjectTypeAPI.ApiProjectTypesIdGet(ctx, GetIDFromPath(projectTypeVersion.GetProjectType())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	state, _, err := c.RewardCloud.StateAPI.ApiStatesIdGet(ctx, GetIDFromPath(environment.GetState())).Execute()
	if state.GetName() != "Running" {
		return errors.Errorf("environment is not running: state = %s", state.GetName())
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
		"--selector", "reward.itg.cloud/component=db",
		"-o", "jsonpath={.items[0].metadata.name}",
	},
		shell.WithSuppressOutput(true), shell.WithCatchOutput(true))

	if string(podname) == "" {
		return errors.New("cannot find db pod")
	}

	access, _, err := c.RewardCloud.EnvironmentAccessAPI.ApiEnvironmentAccessesIdGet(ctx, GetIDFromPath(environment.GetEnvironmentAccess())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting accesses")
	}

	database, _, err := c.RewardCloud.EnvironmentAccessDatabaseAPI.ApiEnvironmentAccessDatabasesIdGet(ctx, GetIDFromPath(access.GetDatabase())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting database")
	}

	dbPassword, err := base64.StdEncoding.DecodeString(database.GetPassword())
	if err != nil {
		return errors.Wrap(err, "decoding database password")
	}

	t := NewTableWriter(WithTableWidthMax(80))
	t.AppendHeader(table.Row{"Host", "Port", "Schema", "User", "Password"})
	t.AppendRow(table.Row{"127.0.0.1", c.Get("local_port"), database.GetScheme(), database.GetUsername(), string(dbPassword)})
	t.Render()

	_, err = c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"port-forward",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		string(podname),
		fmt.Sprintf("%v:3306", c.Get("local_port")),
	},
		shell.WithSuppressOutput(false), shell.WithCatchOutput(false))

	if err != nil {
		return err
	}

	return nil
}
