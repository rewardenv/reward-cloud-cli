package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/kubectl"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	"github.com/rewardenv/reward-cloud-sdk-go/rewardcloud"
	"github.com/rewardenv/reward/pkg/util"
	"os"
	"path/filepath"
	"strconv"
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
	cluster, _, err := c.RewardCloud.ClusterApi.ApiClustersIdGet(ctx, clusterID).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster")
	}

	return cluster, nil
}

func (c *Client) getClusterByID(ctx context.Context, id int32) (*rewardcloud.Cluster, error) {
	idstr := strconv.FormatInt(int64(id), 10)
	cluster, _, err := c.RewardCloud.ClusterApi.ApiClustersIdGet(ctx, idstr).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster")
	}

	return cluster, nil
}

func (c *Client) getOrganizationByID(ctx context.Context, id string) (name string, err error) {
	org, _, err := c.RewardCloud.OrganisationApi.ApiOrganisationsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting organization")
	}

	return org.GetName(), nil
}

func (c *Client) getTeamByID(ctx context.Context, id string) (name string, err error) {
	team, _, err := c.RewardCloud.TeamApi.ApiTeamsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting team")
	}

	return team.GetName(), nil
}

func (c *Client) getProject(ctx context.Context) (*rewardcloud.ProjectProjectGet, error) {
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
	project, _, err := c.RewardCloud.ProjectApi.ApiProjectsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting project")
	}

	return project.GetName(), nil
}

func (c *Client) getProjectByID(ctx context.Context, id string) (get *rewardcloud.ProjectProjectGet, err error) {
	project, _, err := c.RewardCloud.ProjectApi.ApiProjectsIdGet(ctx, id).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting project")
	}

	return project, nil
}

func (c *Client) getEnvironment(ctx context.Context) (*rewardcloud.EnvironmentEnvironmentGet, error) {
	ctx, err := c.prepareContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "preparing context")
	}

	environment, err := c.getEnvironmentByID(ctx, c.getRcContext(ctx).Environment)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}

	return environment, nil
}

func (c *Client) getEnvironmentNameByID(ctx context.Context, id string) (name string, err error) {
	environment, _, err := c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdGet(ctx, id).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting environment")
	}

	return environment.GetName(), nil
}

func (c *Client) getEnvironmentByID(ctx context.Context, id string) (*rewardcloud.EnvironmentEnvironmentGet, error) {
	environment, _, err := c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdGet(ctx, id).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}

	return environment, nil
}

func (c *Client) getStateNameByID(ctx context.Context, id string) (string, error) {
	state, _, err := c.RewardCloud.StateApi.ApiStatesIdGet(ctx, id).Execute()
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
