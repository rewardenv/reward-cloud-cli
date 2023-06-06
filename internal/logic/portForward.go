package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	"github.com/spf13/cobra"
	corev1 "golang.org/x/build/kubernetes/api"
)

type PortForwardClient struct {
	*Client
}

func NewPortForwardClient(c *config.App) *PortForwardClient {
	return &PortForwardClient{New(c)}
}

//nolint:funlen,cyclop
func (c *PortForwardClient) RunCmdPortForwardDB(cmd *cobra.Command, args []string) error {
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

	_, _, err = c.RewardCloud.ProjectTypeApi.ApiProjectTypesIdGet(
		ctx, GetIDFromPath(projectTypeVersion.GetProjectType())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting project type")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	state, _, err := c.RewardCloud.StateApi.ApiStatesIdGet(ctx, GetIDFromPath(environment.GetState())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting environment state")
	}
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

	out, err := c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"get",
		"pod",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		"--selector", "reward.itg.cloud/component=db",
		"-o", "jsonpath={.items[0].metadata.name}",
	},
		shell.WithSuppressOutput(true), shell.WithCatchOutput(true))
	if err != nil {
		return errors.Wrap(err, "running kubectl get pod")
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

	access, _, err := c.RewardCloud.EnvironmentAccessApi.ApiEnvironmentAccessesIdGet(
		ctx, GetIDFromPath(environment.GetEnvironmentAccess())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting accesses")
	}

	database, _, err := c.RewardCloud.EnvironmentAccessDatabaseApi.ApiEnvironmentAccessDatabasesIdGet(
		ctx, GetIDFromPath(access.GetDatabase())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting database")
	}

	dbPassword, err := base64.StdEncoding.DecodeString(database.GetPassword())
	if err != nil {
		return errors.Wrap(err, "decoding database password")
	}

	t := NewTableWriter(WithTableWidthMax(80))
	t.AppendHeader(table.Row{"Host", "Port", "Schema", "User", "Password"})
	t.AppendRow(table.Row{
		"127.0.0.1",
		c.Get("local_port"),
		database.GetScheme(),
		database.GetUsername(),
		string(dbPassword),
	})
	t.Render()

	podname := obj.Items[0].Name
	_, err = c.Kubectl.RunCommand([]string{
		"--kubeconfig", kubeconfigFile.Name(),
		"--cache-dir", filepath.Join(c.CacheDir(), "kubectl"),
		"port-forward",
		"-n", fmt.Sprintf("%s-%s", project.GetCodeName(), environment.GetCodeName()),
		podname,
		fmt.Sprintf("%v:3306", c.Get("local_port")),
	},
		shell.WithSuppressOutput(false), shell.WithCatchOutput(false))

	if err != nil {
		return errors.Wrap(err, "running port-forward")
	}

	return nil
}
