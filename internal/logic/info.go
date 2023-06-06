package logic

import (
	"context"
	"encoding/base64"
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/spf13/cobra"
)

type InfoClient struct {
	*Client
}

func NewInfoClient(c *config.App) *InfoClient {
	return &InfoClient{New(c)}
}

//nolint:funlen,cyclop
func (c *InfoClient) RunCmdInfo(cmd *cobra.Command, args []string) error {
	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(context.Background())
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	project, err := c.getProject(ctx)
	if err != nil {
		return errors.Wrap(err, "getting project")
	}

	projectState, err := c.getStateNameByID(ctx, GetIDFromPath(project.GetState()))
	if err != nil {
		return errors.Wrap(err, "getting project state")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	envState, err := c.getStateNameByID(ctx, GetIDFromPath(environment.GetState()))
	if err != nil {
		return errors.Wrap(err, "getting state")
	}

	t := NewTableWriter(WithTableWidthMax(120))
	t.AppendHeader(table.Row{"INFO", ""})
	t.AppendRow(table.Row{"Project Name", project.GetName()})
	t.AppendRow(table.Row{"Project State", projectState})
	t.AppendRow(table.Row{"Environment Name", environment.GetName()})
	t.AppendRow(table.Row{"Environment State", envState})

	accessID := environment.GetEnvironmentAccess()
	if accessID == "" {
		t.Render()

		return nil
	}

	access, resp, err := c.RewardCloud.EnvironmentAccessApi.ApiEnvironmentAccessesIdGet(
		ctx, GetIDFromPath(accessID)).Execute()
	_ = resp
	if err != nil {
		return errors.Wrap(err, "getting accesses")
	}

	frontend, _, err := c.RewardCloud.EnvironmentAccessFrontendApi.ApiEnvironmentAccessFrontendsIdGet(
		ctx, GetIDFromPath(access.GetFrontend())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting frontend")
	}

	backend, _, err := c.RewardCloud.EnvironmentAccessBackendApi.ApiEnvironmentAccessBackendsIdGet(
		ctx, GetIDFromPath(access.GetBackend())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting backend")
	}

	backendPassword, err := base64.StdEncoding.DecodeString(backend.GetPassword())
	if err != nil {
		return errors.Wrap(err, "decoding backend password")
	}

	devtools, _, err := c.RewardCloud.EnvironmentAccessDevToolsApi.ApiEnvironmentAccessDevToolsIdGet(
		ctx, GetIDFromPath(access.GetDevTools())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting devtools")
	}

	devtoolsPassword, err := base64.StdEncoding.DecodeString(devtools.GetPassword())
	if err != nil {
		return errors.Wrap(err, "decoding devtools password")
	}

	devtoolsURL, err := url.Parse(devtools.GetUrl())
	if err != nil {
		return errors.Wrap(err, "parsing devtools url")
	}

	devtoolsURL.User = url.UserPassword(devtools.GetUsername(), string(devtoolsPassword))

	mailhog, _, err := c.RewardCloud.EnvironmentAccessMailhogApi.ApiEnvironmentAccessMailhogsIdGet(
		ctx, GetIDFromPath(access.GetMailhog())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting mailhog")
	}

	phpmyadmin, _, err := c.RewardCloud.EnvironmentAccessDatabaseApi.ApiEnvironmentAccessDatabasesIdGet(
		ctx, GetIDFromPath(access.GetDatabase())).Execute()
	if err != nil {
		return errors.Wrap(err, "getting database")
	}

	phpmyadminPassword, err := base64.StdEncoding.DecodeString(phpmyadmin.GetPassword())
	if err != nil {
		return errors.Wrap(err, "decoding database password")
	}

	phpmyadminURL, err := url.Parse(phpmyadmin.GetUrl())
	if err != nil {
		return errors.Wrap(err, "parsing phpmyadmin url")
	}

	phpmyadminURL.User = url.UserPassword(devtools.GetUsername(), string(devtoolsPassword))

	t.AppendSeparator()
	t.AppendRow(table.Row{"Frontend URL", frontend.GetUrl()})
	t.AppendRow(table.Row{"Backend URL", backend.GetUrl()})
	t.AppendRow(table.Row{"Username", backend.GetUsername()})
	t.AppendRow(table.Row{"Password", string(backendPassword)})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Devtools"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Devtools URL", devtoolsURL.String()})
	t.AppendRow(table.Row{"Devtools Username", devtools.GetUsername()})
	t.AppendRow(table.Row{"Devtools Password", string(devtoolsPassword)})
	t.AppendRow(table.Row{"PHPMyAdmin URL", phpmyadminURL.String()})
	t.AppendRow(table.Row{"PHPMyAdmin Username", phpmyadmin.GetUsername()})
	t.AppendRow(table.Row{"PHPMyAdmin Password", string(phpmyadminPassword)})
	t.AppendRow(table.Row{"Mailhog URL", mailhog.GetUrl()})

	t.Render()

	return nil
}
