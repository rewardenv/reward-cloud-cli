package kubectl

import (
	"container/list"
	"fmt"
	"github.com/rewardenv/reward-cloud-cli/internal/shell"
	log "github.com/sirupsen/logrus"
	kube "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
	"strings"
)

type Client struct {
	shell.Shell
	TmpFiles *list.List
}

func NewClient(s shell.Shell, tmpFiles *list.List) *Client {
	return &Client{
		Shell:    s,
		TmpFiles: tmpFiles,
	}
}

// RunCommand runs the passed parameters with docker-compose and returns the output.
func (c *Client) RunCommand(args []string, opts ...shell.Opt) (output []byte, err error) {
	log.Debugf("Running command: kubectl %s", strings.Join(args, " "))

	return c.ExecuteWithOptions("kubectl", args, opts...)
}

type Options struct {
	TokenCacheDir    string
	ClusterServer    string
	ClusterCAData    []byte
	OidcIssuerURL    string
	OidcClientID     string
	OidcClientSecret string
}

func (c *Client) NewKubeConfig(opts *Options) ([]byte, error) {
	conf := kube.Config{
		APIVersion:     "v1",
		CurrentContext: "default",
		Contexts: []kube.NamedContext{
			{
				Name: "default",
				Context: kube.Context{
					Cluster:  "default",
					AuthInfo: "default",
				},
			},
		},
		Clusters: []kube.NamedCluster{
			{
				Name: "default",
				Cluster: kube.Cluster{
					Server:                   opts.ClusterServer,
					CertificateAuthorityData: opts.ClusterCAData,
				},
			},
		},
		AuthInfos: []kube.NamedAuthInfo{
			{
				Name: "default",
				AuthInfo: kube.AuthInfo{
					Exec: &kube.ExecConfig{
						Command: "kubectl",
						Args: []string{
							"oidc-login",
							"get-token",
							"--oidc-extra-scope=openid",
							"--oidc-extra-scope=email",
							"--oidc-extra-scope=profile",
							fmt.Sprintf("--token-cache-dir=%s", opts.TokenCacheDir),
							fmt.Sprintf("--oidc-issuer-url=%s", opts.OidcIssuerURL),
							fmt.Sprintf("--oidc-client-id=%s", opts.OidcClientID),
							fmt.Sprintf("--oidc-client-secret=%s", opts.OidcClientSecret),
						},
						APIVersion:         "client.authentication.k8s.io/v1beta1",
						InteractiveMode:    "IfAvailable",
						ProvideClusterInfo: false,
					},
				},
			},
		},
	}

	kubeconfigdata, err := yaml.Marshal(conf)
	if err != nil {
		return nil, err
	}

	return kubeconfigdata, nil
}
