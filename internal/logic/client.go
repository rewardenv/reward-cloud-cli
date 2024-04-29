package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rewardenv/reward/pkg/util"

	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/kubectl"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	"github.com/rewardenv/reward-cloud-cli/pkg/api/rewardcloud"
)

type Client struct {
	*config.App
	Kubectl *kubectl.Client
}

func new(c *config.App) *Client {
	return &Client{
		c,
		kubectl.NewClient(shell.NewLocalShellWithOpts(), c.TmpFiles),
	}
}

var ErrUnauthorized = GenericOpenAPIError{
	body:  []byte(`{"code":401,"message":"JWT Token not found"}`),
	error: "401 Unauthorized",
	model: nil,
}

// GenericOpenAPIError Provides access to the body, error and model on returned errors.
type GenericOpenAPIError struct {
	body  []byte
	error string
	model interface{}
}

// Error returns non-empty string if there was an error.
func (e GenericOpenAPIError) Error() string {
	return e.error
}

func (c *Client) getCluster(ctx context.Context) (*rewardcloud.Cluster, error) {
	ctx, err := c.prepareContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "preparing context")
	}

	env, err := c.getEnvironment(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}

	clusterID := GetIDFromPath(env.GetCluster())
	cluster, _, err := c.RewardCloud.ClusterAPI.ApiClustersIdGet(ctx, clusterID).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster")
	}

	return cluster, nil
}

func (c *Client) getClusterByID(ctx context.Context, id int32) (*rewardcloud.Cluster, error) {
	idstr := strconv.FormatInt(int64(id), 10)
	cluster, _, err := c.RewardCloud.ClusterAPI.ApiClustersIdGet(ctx, idstr).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster")
	}

	return cluster, nil
}

func (c *Client) getOrganizationByID(ctx context.Context, id string) (name string, err error) {
	org, resp, err := c.RewardCloud.OrganisationAPI.ApiOrganisationsIdGet(ctx, id).Execute()
	_ = resp
	if err != nil {
		return "", errors.Wrap(err, "getting organization")
	}

	return org.GetName(), nil
}

func (c *Client) getTeamByID(ctx context.Context, id string) (name string, err error) {
	team, _, err := c.RewardCloud.TeamAPI.ApiTeamsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting team")
	}

	return team.GetName(), nil
}

func (c *Client) getProject(ctx context.Context) (*rewardcloud.ProjectProjectOutput, error) {
	ctx, err := c.prepareContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "preparing context")
	}

	project, err := c.getProjectByID(ctx, c.getRcContext(ctx).Project)
	if err != nil {
		return nil, errors.Wrap(err, "getting project")
	}

	return project, nil
}

func (c *Client) getProjectNameByID(ctx context.Context, id string) (name string, err error) {
	project, _, err := c.RewardCloud.ProjectAPI.ApiProjectsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting project")
	}

	return project.GetName(), nil
}

func (c *Client) getProjectByID(ctx context.Context, id string) (get *rewardcloud.ProjectProjectOutput, err error) {
	project, _, err := c.RewardCloud.ProjectAPI.ApiProjectsIdGet(ctx, id).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting project")
	}

	return project, nil
}

func (c *Client) getEnvironment(ctx context.Context) (*rewardcloud.EnvironmentEnvironmentOutput, error) {
	ctx, err := c.prepareContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "preparing context")
	}

	environment, err := c.getEnvironmentByID(ctx, c.getRcContext(ctx).Environment)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment by id")
	}

	return environment, nil
}

func (c *Client) getEnvironmentNameByID(ctx context.Context, id string) (name string, err error) {
	environment, _, err := c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting environment from api")
	}

	return environment.GetName(), nil
}

func (c *Client) getEnvironmentByID(ctx context.Context, id string) (*rewardcloud.EnvironmentEnvironmentOutput, error) {
	environment, resp, err := c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdGet(ctx, id).Execute()
	_ = resp
	if err != nil {
		return nil, errors.Wrap(err, "getting environment from api")
	}

	return environment, nil
}

func (c *Client) getStateNameByID(ctx context.Context, id string) (string, error) {
	state, _, err := c.RewardCloud.StateAPI.ApiStatesIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting state")
	}

	return state.GetName(), nil
}

func (c *Client) prepareKubeconfig(cluster *rewardcloud.Cluster) (*os.File, error) {
	cacert, err := base64.StdEncoding.DecodeString(cluster.GetClusterCertificateAuthorityData())
	if err != nil {
		return nil, errors.Wrap(err, "decoding cluster CA data")
	}

	kubeconfig, err := c.Kubectl.NewKubeConfig(&kubectl.Options{
		TokenCacheDir:    filepath.Join(c.CacheDir(), "kubectl", "oidc-login"),
		ClusterServer:    cluster.GetClusterServer(),
		ClusterCAData:    cacert,
		OidcIssuerURL:    cluster.GetOidcIssuerUrl(),
		OidcClientID:     cluster.GetOidcClientID(),
		OidcClientSecret: cluster.GetOidcClientSecret(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating kube config")
	}

	tmpFile, err := os.CreateTemp(c.CacheDir(), fmt.Sprintf("%s-", c.AppName()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	c.TmpFiles.PushBack(tmpFile.Name())

	err = util.CreateDirAndWriteToFile(kubeconfig, tmpFile.Name())
	if err != nil {
		return nil, errors.Wrap(err, "creating kube config file")
	}

	return tmpFile, nil
}

func (c *Client) getRcContext(ctx context.Context) *config.RcContext {
	rcContext, ok := ctx.Value(config.ContextKey{}).(*config.RcContext)
	if !ok {
		return &config.RcContext{}
	}

	return rcContext
}

func (c *Client) prepareContext(ctx context.Context) (context.Context, error) {
	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "logging in")
	}

	if ctx.Value(config.ContextKey{}) != nil {
		return ctx, nil
	}

	conf, err := c.ReadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading config")
	}

	rcContext, err := NewContextClient(c.App).CurrentContext(conf)
	if err != nil {
		return nil, errors.Wrap(err, "getting current context")
	}

	if rcContext == nil {
		return nil, errors.New("no context set")
	}

	return context.WithValue(ctx, config.ContextKey{}, rcContext), nil
}

func (c *Client) getBuildEnvVars(ctx context.Context) (map[string]string, error) {
	buildEnvVars := make(map[string]string)

	// TODO: Fetch organization, team build variables
	project, err := c.getProject(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting project")
	}

	projectEnvVars := project.GetProjectEnvVar()
	projectEnvVarsMap, err := c.envVarsToMap(ctx, projectEnvVars, []string{"build", "all"})
	if err != nil {
		return nil, errors.Wrap(err, "getting project env vars")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}

	envVars := environment.GetEnvVar()
	environmentEnvVarsMap, err := c.envVarsToMap(ctx, envVars, []string{"build", "all"})
	if err != nil {
		return nil, errors.Wrap(err, "getting environment env vars")
	}
	_, _ = projectEnvVarsMap, environmentEnvVarsMap

	for _, v := range envVars {
		t := v.GetEnvVarType()
		envVarType, _, err := c.RewardCloud.EnvVarTypeAPI.ApiEnvVarTypesIdGet(ctx, GetIDFromPath(t)).Execute()
		if err != nil {
			continue
		}
		if envVarType.GetCodeName() == "build" || envVarType.GetCodeName() == "all" {
			buildEnvVars[v.GetKey()] = v.GetValue()
		}
	}

	return buildEnvVars, nil
}

func (c *Client) envVarsToMap(ctx context.Context, envVars interface{}, t []string) (map[string]string, error) {
	ids, err := c.getEnvVarTypeIDs(ctx, t)
	if err != nil {
		return nil, errors.Wrap(err, "getting env var type ids")
	}

	m := make(map[string]string)
	switch t := envVars.(type) {
	case []rewardcloud.EnvironmentEnvVarEnvironmentOutput:
		for _, v := range t {
			if ContainsAny(ids, fmt.Sprintf("%v", GetIDFromPath(v.GetEnvVarType()))) {
				m[v.GetKey()] = v.GetValue()
			}
		}
	case []rewardcloud.ProjectEnvVarProjectOutput:
		for _, v := range t {
			if ContainsAny(ids, fmt.Sprintf("%v", GetIDFromPath(v.GetEnvVarType()))) {
				m[v.GetKey()] = v.GetValue()
			}
		}
	case []rewardcloud.TeamEnvVarTeamOutput:
		for _, v := range t {
			if ContainsAny(ids, fmt.Sprintf("%v", GetIDFromPath(v.GetEnvVarType()))) {
				m[v.GetKey()] = v.GetValue()
			}
		}
	case []rewardcloud.OrganisationEnvVarOrganisationOutput:
		for _, v := range t {
			if ContainsAny(ids, fmt.Sprintf("%v", GetIDFromPath(v.GetEnvVarType()))) {
				m[v.GetKey()] = v.GetValue()
			}
		}
	}
	return m, nil
}

func (c *Client) getProjectType(ctx context.Context) (string, error) {
	project, err := c.getProject(ctx)
	if err != nil {
		return "", errors.Wrap(err, "getting project")
	}

	// Get project type version
	projectTypeVersion, resp, err := c.RewardCloud.ProjectTypeVersionAPI.ApiProjectTypeVersionsIdGet(ctx, GetIDFromPath(project.GetProjectTypeVersion())).Execute()
	if err != nil {
		_ = resp
		return "", errors.Wrap(err, "getting project type version by id")
	}

	// Get project type
	projectType, resp, err := c.RewardCloud.ProjectTypeAPI.ApiProjectTypesIdGet(ctx, GetIDFromPath(projectTypeVersion.GetProjectType())).Execute()
	if err != nil {
		_ = resp
		return "", errors.Wrap(err, "getting project type by id")
	}

	projectTypeName := projectType.GetName()
	if projectTypeName == "" {
		return "", errors.New("project type name is empty")
	}

	return strings.ToLower(projectTypeName), nil
}

func (c *Client) getMainComponentVersion(ctx context.Context) (string, error) {
	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return "", errors.Wrap(err, "getting project")
	}

	components := environment.GetEnvironmentComponent()

	projectTypeName, err := c.getProjectType(ctx)
	if err != nil {
		return "", errors.Wrap(err, "getting project type")
	}

	mainComponentName := ""
	switch projectTypeName {
	case "magento", "wordpress", "shopware":
		mainComponentName = "php"
	case "magento-pwa", "shopware-pwa":
		mainComponentName = "node"
	default:
		return "", errors.New("project type not supported")
	}

	mainComponentVersion, err := c.getComponentVersion(ctx, components, mainComponentName)
	if err != nil {
		return "", errors.Wrap(err, "getting php component version")
	}

	return mainComponentVersion, nil
}

func (c *Client) getComponentVersion(ctx context.Context, components []rewardcloud.EnvironmentComponentEnvironmentOutput, selected string) (string, error) {
	for _, v := range components {
		cv, _, err := c.RewardCloud.ComponentVersionAPI.ApiComponentVersionsIdGet(ctx, GetIDFromPath(v.GetComponentVersion())).Execute()
		if err != nil {
			return "", err
		}
		componentID := cv.GetComponent()
		component, _, err := c.RewardCloud.ComponentAPI.ApiComponentsIdGet(ctx, GetIDFromPath(componentID)).Execute()
		if err != nil {
			return "", err
		}

		if component.GetName() == selected {
			return cv.GetVersion(), nil
		}
	}

	return "", errors.New("component not found")
}

func (c *Client) getEnvVarTypeIDs(ctx context.Context, t []string) ([]string, error) {
	var ids []string
	envVarTypes, _, err := c.RewardCloud.EnvVarTypeAPI.ApiEnvVarTypesGetCollection(ctx).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting env var types")
	}

	for _, v := range envVarTypes {
		for _, tt := range t {
			if v.GetCodeName() == tt {
				ids = append(ids, fmt.Sprintf("%v", v.GetId()))
			}
		}
	}

	return ids, nil
}
