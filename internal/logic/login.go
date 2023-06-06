package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-sdk-go/rewardcloud"
	"github.com/rewardenv/reward/pkg/util"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

type LoginClient struct {
	*Client
}

func NewLoginClient(c *config.App) *LoginClient {
	return &LoginClient{New(c)}
}

func (c *LoginClient) RunCmdLogin(cmd *cobra.Command, args []string) error {
	_, err := c.CheckTokenAndLogin(context.Background())
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	return nil
}

func (c *LoginClient) CheckTokenAndLogin(ctx context.Context) (context.Context, error) {
	bs := c.readToken()
	if len(bs) > 0 {
		ctx = c.SetToken(ctx, string(bs))
	}

	valid, err := c.ValidateToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "validating token")
	}

	if !valid {
		log.Info("Token is invalid. Please log in...")

		return c.Login(ctx)
	}

	return ctx, nil
}

func (c *LoginClient) Login(ctx context.Context) (context.Context, error) {
	log.Printf("Logging in to %s...", c.Endpoint())

	token, err := c.loginWithUsernameAndPassword(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "logging in")
	}
	ctx = c.SetToken(ctx, token)

	valid, err := c.ValidateToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "validating token")
	}
	if !valid {
		return nil, errors.New("token is not valid")
	}

	err = c.writeToken()
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (c *LoginClient) loginWithUsernameAndPassword(ctx context.Context) (string, error) {
	var (
		err          error
		id, password string
	)

	if c.GetString(fmt.Sprintf("%s_id", c.ConfigPrefix())) != "" {
		id = c.GetString(fmt.Sprintf("%s_id", c.ConfigPrefix()))
	}

	if c.GetString(fmt.Sprintf("%s_password", c.ConfigPrefix())) != "" {
		password = c.GetString(fmt.Sprintf("%s_password", c.ConfigPrefix()))
	}

	if id == "" || password == "" {
		id, password, err = c.getCredentials()
		if err != nil {
			return "", err
		}
	}

	creds := rewardcloud.Credentials{
		Id:       rewardcloud.PtrString(id),
		Password: rewardcloud.PtrString(password),
	}

	token, _, err := c.RewardCloud.TokenApi.PostCredentialsItem(ctx).Credentials(creds).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting token")
	}

	log.Printf("...successfully logged in.\n\n")

	return token.GetToken(), nil
}

func (c *LoginClient) getCredentials() (string, string, error) {
	var (
		err      error
		username = c.ID()
		password = c.Password()
	)

	if username == "" {
		username, err = GetValueFromPrompt("Username or email")
		if err != nil {
			return "", "", errors.Wrap(err, "getting username")
		}
	}

	if password == "" {
		password, err = GetPasswordFromPrompt("Password")
		if err != nil {
			return "", "", errors.Wrap(err, "getting password")
		}
	}

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func (c *LoginClient) writeToken() error {
	token := c.Token()
	str := base64.StdEncoding.EncodeToString([]byte(token))

	err := util.CreateDirAndWriteToFile([]byte(str), c.TokenFile(), 0o600)
	if err != nil {
		return errors.Wrap(err, "writing token to file")
	}

	return nil
}

func (c *LoginClient) readToken() []byte {
	var bs []byte
	tokenfile := c.TokenFile()
	content, _ := os.ReadFile(tokenfile)

	if content != nil {
		bs, _ = base64.StdEncoding.DecodeString(string(content))
	}

	return bs
}

func (c *LoginClient) ValidateToken(ctx context.Context) (bool, error) {
	projects, _, err := c.RewardCloud.ProjectApi.ApiProjectsGetCollection(ctx).Execute()
	if err != nil {
		if err.Error() == ErrUnauthorized.Error() {
			return false, nil
		}

		return false, errors.Wrap(err, "checking token")
	}

	_ = projects

	return true, nil
}
