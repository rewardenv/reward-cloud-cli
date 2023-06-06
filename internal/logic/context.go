package logic

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ContextClient struct {
	*Client
}

func NewContextClient(c *config.App) *ContextClient {
	return &ContextClient{
		New(c),
	}
}

func (c *ContextClient) RunCmdContextList(cmd *cobra.Command, args []string) error {
	conf, err := c.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	if len(conf.Contexts) == 0 {
		log.Infof("No contexts configured, please run `%s %s context create` to create one.", c.ParentAppName(), c.AppName())

		return nil
	}

	if c.GetBool("full") {
		return c.listContexts(conf, WithFull())
	}

	log.Info("Listing available contexts...")

	return c.listContexts(conf)
}

func (c *ContextClient) RunCmdContextCreate(cmd *cobra.Command, args []string) error {
	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(context.Background())
	if err != nil {
		return errors.Wrap(err, "checking token")
	}

	conf, err := c.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}
	oldConf := *conf

	log.Info("Creating a New context...")

	ctx, err = c.createCloudContext(ctx)
	if err != nil {
		return errors.Wrap(err, "creating cloud context")
	}

	c.overWriteContext(ctx, conf)

	if reflect.DeepEqual(oldConf, *conf) {
		log.Info("No changes in configuration. Exiting...")

		return nil
	}

	err = c.saveContext(conf)
	if err != nil {
		return errors.Wrap(err, "saving context")
	}

	return nil
}

func (c *ContextClient) RunCmdContextDelete(cmd *cobra.Command, args []string) error {
	log.Info("Select a context to delete...")

	conf, err := c.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	if len(conf.Contexts) == 0 {
		log.Info("No contexts to delete. Exiting...")

		return nil
	}

	err = c.listContexts(conf)
	if err != nil {
		return errors.Wrap(err, "listing contexts")
	}

	val, err := GetValueFromPrompt("Enter the number of the context you want to delete")
	if err != nil {
		return errors.Wrap(err, "getting context number")
	}

	ival, err := strconv.Atoi(val)
	if err != nil {
		return errors.Wrap(err, "converting context number")
	}

	if ival > len(conf.Contexts) || ival < 1 {
		return errors.New("context number out of range")
	}

	conf.Contexts = append(conf.Contexts[:ival-1], conf.Contexts[ival:]...)

	err = c.saveContext(conf)
	if err != nil {
		return errors.Wrap(err, "saving context")
	}

	return nil
}

func (c *ContextClient) RunCmdContextSelect(cmd *cobra.Command, args []string) error {
	conf, err := c.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	log.Info("Select a context to use...")

	err = c.listContexts(conf)
	if err != nil {
		return errors.Wrap(err, "listing contexts")
	}

	val, err := GetValueFromPrompt("Enter the number of the context you want to use")
	if err != nil {
		return errors.Wrap(err, "getting context number")
	}

	ival, err := strconv.Atoi(val)
	if err != nil {
		return errors.Wrap(err, "converting context number")
	}

	if ival > len(conf.Contexts) || ival < 1 {
		return errors.New("context number out of range")
	}

	conf.CurrentContext = conf.Contexts[ival-1].Name

	err = c.saveContext(conf)
	if err != nil {
		return errors.Wrap(err, "saving context")
	}

	log.Infof("RcContext changed to: %s", conf.CurrentContext)

	return nil
}

func (c *ContextClient) RunCmdContextCheck(cmd *cobra.Command, args []string) error {
	log.Info("Checking context...")

	conf, err := c.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	if len(conf.Contexts) == 0 {
		log.Info("No context selected.")

		return nil
	}

	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	valid := true
	_, err = c.getOrganizationByID(ctx, c.getRcContext(ctx).Organization)
	if err != nil {
		valid = false
	}

	_, err = c.getTeamByID(ctx, c.getRcContext(ctx).Team)
	if err != nil {
		valid = false
	}
	_, err = c.getProjectNameByID(ctx, c.getRcContext(ctx).Project)
	if err != nil {
		valid = false
	}
	_, err = c.getEnvironmentNameByID(ctx, c.getRcContext(ctx).Environment)
	if err != nil {
		valid = false
	}

	if !valid {
		log.Warn("Context is invalid.")
		val, err := GetValueFromPrompt("Would you like to delete it? (y/n)")
		if err != nil {
			return errors.Wrap(err, "getting context number")
		}

		var newContexts []*config.RcContext
		if strings.ToLower(val) == "y" || strings.ToLower(val) == "yes" {
			for _, confCtx := range conf.Contexts {
				if confCtx.Name != conf.CurrentContext {
					newContexts = append(newContexts, confCtx)
				}
			}

			conf.Contexts = newContexts
			err = c.saveContext(conf)
			if err != nil {
				return errors.Wrap(err, "saving context")
			}

			log.Info("Context deleted")
		}

		return nil
	}

	log.Info("Context is valid")

	return nil
}

type ListContextOptions struct {
	Full bool
}

type ListContextOption func(*ListContextOptions)

func WithFull() ListContextOption {
	return func(o *ListContextOptions) {
		o.Full = true
	}
}

func (c *ContextClient) listContexts(conf *config.Config, opts ...ListContextOption) error {
	o := &ListContextOptions{}
	for _, opt := range opts {
		opt(o)
	}

	t := NewTableWriter()
	if o.Full {
		t.AppendHeader(table.Row{"#", "Name", "Current", "Organization", "Team", "Project", "Environment"})
	} else {
		t.AppendHeader(table.Row{"#", "Name", "Current"})
	}

	for i, confCtx := range conf.Contexts {
		var current string
		if confCtx.Name == conf.CurrentContext {
			current = "*"
		}

		if o.Full {
			ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(context.Background())
			if err != nil {
				return errors.Wrap(err, "checking token")
			}

			orgname, err := c.getOrganizationByID(ctx, confCtx.Organization)
			if err != nil {
				return errors.Wrap(err, "getting organization by ID")
			}
			teamname, err := c.getTeamByID(ctx, confCtx.Team)
			if err != nil {
				return errors.Wrap(err, "getting team by ID")
			}
			projectname, err := c.getProjectNameByID(ctx, confCtx.Project)
			if err != nil {
				return errors.Wrap(err, "getting project by ID")
			}
			environmentname, err := c.getEnvironmentNameByID(ctx, confCtx.Environment)
			if err != nil {
				return errors.Wrap(err, "getting environment by ID")
			}

			t.AppendRow(table.Row{i + 1, confCtx.Name, current, orgname, teamname, projectname, environmentname})

			continue
		}

		t.AppendRow(table.Row{i + 1, confCtx.Name, current})
	}

	t.Render()

	return nil
}

func (c *ContextClient) createCloudContext(ctx context.Context) (context.Context, error) {
	ctx, orgname, err := c.selectOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting organization")
	}

	ctx, teamname, err := c.selectTeam(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting team")
	}

	ctx, projectname, err := c.selectProject(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting project")
	}

	ctx, envname, err := c.selectEnvironment(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "selecting environment")
	}

	rcContext := c.getRcContext(ctx)
	rcContext.Name = fmt.Sprintf("%s/%s:%s/%s", orgname, teamname, projectname, envname)

	val, err := GetValueFromPrompt(fmt.Sprintf("Enter the name of the context: [%s]", rcContext.Name), WithAllowEmpty())
	if err != nil {
		return nil, errors.Wrap(err, "getting context name")
	}

	if val != "" {
		rcContext.Name = val
	}

	ctx = context.WithValue(ctx, config.ContextKey{}, rcContext)

	return ctx, nil
}

func (c *ContextClient) overWriteContext(ctx context.Context, conf *config.Config) {
	rcContext := c.getRcContext(ctx)

	if len(conf.Contexts) == 0 {
		conf.Contexts = []*config.RcContext{rcContext}

		conf.CurrentContext = rcContext.Name

		return
	}

	for i, confContext := range conf.Contexts {
		if confContext.Name == rcContext.Name {
			if reflect.DeepEqual(confContext, ctx) {
				return
			}

			prompt, _ := GetValueFromPrompt(fmt.Sprintf("RcContext %s already exists. Overwrite? [y/n]", rcContext.Name))
			if prompt == "y" || prompt == "yes" {
				conf.Contexts[i] = rcContext
			}

			return
		}
	}

	conf.Contexts = append(conf.Contexts, rcContext)
}

func (c *ContextClient) saveContext(conf *config.Config) error {
	configBytes, err := yaml.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, "marshalling config")
	}

	configPath := c.App.GetString(fmt.Sprintf("%s_config_file", c.App.ConfigPrefix()))
	if configPath == "" {
		return errors.New("config path not set")
	}

	err = util.CreateDirAndWriteToFile(configBytes, configPath)
	if err != nil {
		return errors.Wrap(err, "writing config file")
	}

	return nil
}

func (c *ContextClient) selectOrganization(ctx context.Context) (_ context.Context, name string, err error) {
	rcContext := c.getRcContext(ctx)

	orgs, _, err := c.RewardCloud.OrganisationApi.ApiOrganisationsGetCollection(ctx).Execute()
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting organizations")
	}

	if len(orgs) < 1 {
		return ctx, "", errors.New("no organizations found")
	}

	log.Info("Select an organization...")

	t := NewTableWriter()
	t.AppendHeader(table.Row{"#", "Name", "Codename"})
	for i, org := range orgs {
		t.AppendRow(table.Row{i + 1, org.GetName(), org.GetCodeName()})
	}
	t.Render()

	val, err := GetValueFromPrompt("Enter the number of the organization",
		WithMinimumValue(1),
		WithMaximumValue(len(orgs)),
	)
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting organization number")
	}

	iVal, err := strconv.Atoi(val)
	if err != nil {
		return ctx, "", errors.Wrap(err, "parsing organization id")
	}

	rcContext.Organization = strconv.FormatInt(int64(orgs[iVal-1].GetId()), 10)

	return context.WithValue(ctx, config.ContextKey{}, rcContext),
		orgs[iVal-1].GetCodeName(),
		nil
}

func (c *ContextClient) selectTeam(ctx context.Context) (_ context.Context, name string, err error) {
	rcContext := c.getRcContext(ctx)

	err = c.checkOrganization(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking organization")
	}

	teams, _, err := c.RewardCloud.TeamApi.ApiTeamsGetCollection(ctx).
		Organisation(rcContext.Organization).
		Execute()
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting teams")
	}

	if len(teams) < 1 {
		return ctx, "", errors.New("no teams found")
	}

	log.Info("Select a team...")

	t := NewTableWriter()
	t.AppendHeader(table.Row{"#", "Name", "Codename"})
	for i, team := range teams {
		t.AppendRow(table.Row{i + 1, team.GetName(), team.GetCodeName()})
	}
	t.Render()

	val, err := GetValueFromPrompt("Enter the number of the team", WithMinimumValue(1), WithMaximumValue(len(teams)))
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting team number")
	}

	iVal, err := strconv.Atoi(val)
	if err != nil {
		return ctx, "", errors.Wrap(err, "parsing team id")
	}

	rcContext.Team = strconv.FormatInt(int64(teams[iVal-1].GetId()), 10)

	return context.WithValue(ctx, config.ContextKey{}, rcContext),
		teams[iVal-1].GetCodeName(),
		nil
}

func (c *ContextClient) selectProject(ctx context.Context) (_ context.Context, name string, err error) {
	rcContext := c.getRcContext(ctx)

	err = c.checkOrganization(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking organization")
	}

	err = c.checkTeam(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking team")
	}

	projects, _, err := c.RewardCloud.ProjectApi.ApiProjectsGetCollection(ctx).
		TeamOrganisationId(rcContext.OrganizationID()).
		Team(rcContext.Team).
		Execute()
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting projects")
	}

	if len(projects) < 1 {
		return ctx, "", errors.New("no projects found")
	}

	log.Info("Select a project...")

	t := NewTableWriter()
	t.AppendHeader(table.Row{"#", "Name", "Codename"})
	for i, project := range projects {
		t.AppendRow(table.Row{i + 1, project.GetName(), project.GetCodeName()})
	}
	t.Render()

	val, err := GetValueFromPrompt("Enter the number of the project",
		WithMinimumValue(1),
		WithMaximumValue(len(projects)),
	)
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting project number")
	}

	iVal, err := strconv.Atoi(val)
	if err != nil {
		return ctx, "", errors.Wrap(err, "parsing project id")
	}

	rcContext.Project = strconv.FormatInt(int64(projects[iVal-1].GetId()), 10)

	return context.WithValue(ctx, config.ContextKey{}, rcContext), projects[iVal-1].GetName(), nil
}

func (c *ContextClient) selectEnvironment(ctx context.Context) (_ context.Context, name string, err error) {
	rcContext := c.getRcContext(ctx)

	err = c.checkOrganization(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking organization")
	}

	err = c.checkTeam(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking team")
	}

	err = c.checkProject(ctx)
	if err != nil {
		return ctx, "", errors.Wrap(err, "checking project")
	}

	environments, _, err := c.RewardCloud.EnvironmentApi.ApiEnvironmentsGetCollection(ctx).
		Project(rcContext.Project).
		ProjectTeamId(rcContext.TeamID()).
		Execute()
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting environments")
	}

	if len(environments) < 1 {
		return ctx, "", errors.New("no environments found")
	}

	log.Info("Select an environment...")

	t := NewTableWriter()
	t.AppendHeader(table.Row{"#", "Name", "Codename"})
	for i, environment := range environments {
		t.AppendRow(table.Row{i + 1, environment.GetName(), environment.GetCodeName()})
	}
	t.Render()

	val, err := GetValueFromPrompt("Enter the number of the environment",
		WithMinimumValue(1),
		WithMaximumValue(len(environments)),
	)
	if err != nil {
		return ctx, "", errors.Wrap(err, "getting environment number")
	}

	iVal, err := strconv.Atoi(val)
	if err != nil {
		return ctx, "", errors.Wrap(err, "parsing environment id")
	}

	rcContext.Environment = strconv.FormatInt(int64(environments[iVal-1].GetId()), 10)

	return context.WithValue(ctx, config.ContextKey{}, rcContext), environments[iVal-1].GetName(), nil
}

func (c *ContextClient) checkOrganization(ctx context.Context) error {
	rcContext := c.getRcContext(ctx)

	if rcContext.Organization == "0" || rcContext.Organization == "" {
		return errors.New("no organization selected")
	}

	return nil
}

func (c *ContextClient) checkTeam(ctx context.Context) error {
	rcContext := c.getRcContext(ctx)

	if rcContext.Team == "0" || rcContext.Team == "" {
		return errors.New("no team selected")
	}

	return nil
}

func (c *ContextClient) checkProject(ctx context.Context) error {
	rcContext := c.getRcContext(ctx)

	if rcContext.Project == "0" || rcContext.Project == "" {
		return errors.New("no project selected")
	}

	return nil
}

func (c *ContextClient) checkEnvironment(ctx context.Context) error {
	rcContext := c.getRcContext(ctx)

	if rcContext.Environment == "0" || rcContext.Environment == "" {
		return errors.New("no environment selected")
	}

	return nil
}

func (c *ContextClient) CurrentContext(conf *config.Config) (*config.RcContext, error) {
	var currentContext *config.RcContext
	for _, clusterContext := range conf.Contexts {
		if clusterContext.Name == conf.CurrentContext {
			currentContext = clusterContext
		}
	}

	if currentContext == nil {
		return nil, errors.New("could not find current context")
	}

	return currentContext, nil
}
