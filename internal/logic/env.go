package logic

import (
	"context"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-sdk-go/rewardcloud"
	"github.com/spf13/cobra"
)

type EnvClient struct {
	*Client
}

func NewEnvClient(c *config.App) *EnvClient {
	return &EnvClient{new(c)}
}

func (c *EnvClient) RunCmdEnvStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	state, err := c.getStateNameByID(ctx, GetIDFromPath(environment.GetState()))
	if err != nil {
		return errors.Wrap(err, "getting state")
	}

	t := NewTableWriter()
	t.AppendHeader(table.Row{"Environment", "State"})
	t.AppendRow(table.Row{environment.GetName(), state})
	t.Render()

	access, _, err := c.RewardCloud.EnvironmentAccessApi.ApiEnvironmentAccessesIdGet(ctx, GetIDFromPath(environment.GetEnvironmentAccess())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting accesses")
	}

	frontend, _, err := c.RewardCloud.EnvironmentAccessFrontendApi.ApiEnvironmentAccessFrontendsIdGet(ctx, GetIDFromPath(access.GetFrontend())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting frontend")
	}

	backend, _, err := c.RewardCloud.EnvironmentAccessBackendApi.ApiEnvironmentAccessBackendsIdGet(ctx, GetIDFromPath(access.GetBackend())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting backend")
	}

	t = NewTableWriter()
	t.AppendHeader(table.Row{"Frontend URL"})
	t.AppendRow(table.Row{frontend.GetUrl()})
	t.Render()

	t = NewTableWriter(WithTableWidthMax(80))
	t.AppendHeader(table.Row{"Backend URL", "Username", "Password"})
	t.AppendRow(table.Row{backend.GetUrl(), backend.GetUsername(), backend.GetPassword()})
	t.Render()

	return nil
}

func (c *EnvClient) RunCmdEnvBuildAndDeploy(cmd *cobra.Command, args []string) error {
	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	env, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	patch := rewardcloud.EnvironmentEnvironmentGet{
		Id: env.Id,
	}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdbuildAndDeployPatch(ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentGet(patch).Execute()
	if err != nil {
		return errors.Wrap(err, "building environment")
	}

	return nil
}

func (c *EnvClient) RunCmdEnvExportDB(cmd *cobra.Command, args []string) error {
	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	stripDatabase := rewardcloud.NewNullableBool(rewardcloud.PtrBool(c.GetBool("strip_database")))
	post := rewardcloud.EnvironmentEnvironmentPost{
		IsStripDatabase: *stripDatabase,
	}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdexportDatabasePut(ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentPost(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting database")
	}

	return nil
}

func (c *EnvClient) RunCmdEnvExportMedia(cmd *cobra.Command, args []string) error {
	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	post := rewardcloud.EnvironmentEnvironmentPost{}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdexportMediaPut(ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentPost(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting media")
	}

	return nil
}
