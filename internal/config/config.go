package config

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/rewardenv/reward-cloud-go/rewardcloud"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"reward-cloud-cli/internal/util"
)

var (
	// FS is the implementation of Afero Filesystem. It's a filesystem wrapper and used for testing.
	FS = &afero.Afero{Fs: afero.NewOsFs()}
)

type Config struct {
	*viper.Viper
	appName string
	token   string
	Client  *rewardcloud.APIClient
}

func New(name, ver string) *Config {
	c := &Config{
		Viper:   viper.GetViper(),
		appName: name,
	}

	c.SetDefault(fmt.Sprintf("%s_version", name), version.Must(version.NewVersion(ver)).String())

	return c
}

func (c *Config) Init() *Config {
	c.AddConfigPath(".")

	cfg := c.GetString(fmt.Sprintf("%s_config_file", c.appName))
	if cfg != "" {
		c.AddConfigPath(filepath.Dir(cfg))
		c.SetConfigName(filepath.Base(cfg))
		c.SetConfigType("yaml")
	}

	c.AutomaticEnv()

	if err := c.ReadInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	c.AddConfigPath(".")
	c.SetConfigName(".env")
	c.SetConfigType("dotenv")

	if err := c.MergeInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	c.SetDefault(fmt.Sprintf("%s_parent_app_name", c.AppName()), "reward")
	c.SetDefault(fmt.Sprintf("%s_dir", c.AppName()), filepath.Join(c.AppHomeDir()))
	c.SetDefault(fmt.Sprintf("%s_endpoint", c.AppName()), "staging.rewardcloud.itg.cloud")

	c.SetLogging()

	endpoint := rewardcloud.ServerConfigurations{
		{
			URL:         "https://" + c.GetString(fmt.Sprintf("%s_endpoint", c.AppName())),
			Description: "",
		},
	}
	conf := &rewardcloud.Configuration{
		UserAgent: "reward-cloud-cli",
		Debug:     false,
		Servers:   endpoint,
		OperationServers: map[string]rewardcloud.ServerConfigurations{
			"default": endpoint,
		},
	}
	c.Client = rewardcloud.NewAPIClient(conf)

	return c
}

// SetLogging sets the logging level based on the command line flags and environment variables.
func (c *Config) SetLogging() {
	switch {
	case c.GetString("log_level") == "trace":
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	case c.IsDebug(), c.GetString("log_level") == "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case c.GetString("log_level") == "info":
		log.SetLevel(log.InfoLevel)
	case c.GetString("log_level") == "warning":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors:          c.GetBool("disable_colors"),
			ForceColors:            true,
			DisableLevelTruncation: true,
			FullTimestamp:          true,
			DisableTimestamp:       !c.GetBool("debug"),
			QuoteEmptyFields:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := strings.ReplaceAll(path.Base(f.File), "reward/", "")

				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
			},
		},
	)
}

func (c *Config) Login(ctx context.Context) error {
	var (
		err          error
		id, password string
	)

	if c.GetString(fmt.Sprintf("%s_%s_token", c.ParentAppName(), c.AppName())) != "" {
		token := rewardcloud.Token{
			Token: rewardcloud.PtrString(
				c.GetString(
					fmt.Sprintf("%s_%s_token",
						c.ParentAppName(),
						c.AppName(),
					),
				),
			),
		}
		c.Set(fmt.Sprintf("%s_%s_token", c.ParentAppName(), c.AppName()), token.Data)

		return nil
	}

	id, password, err = util.GetCredentials()
	if err != nil {
		log.Fatal(err)
	}
	creds := rewardcloud.Credentials{
		Id:       rewardcloud.PtrString(id),
		Password: rewardcloud.PtrString(password),
	}

	log.Infof("Logging in to %s...", c.GetString(fmt.Sprintf("%s_endpoint", c.AppName())))

	token, _, err := c.Client.TokenApi.PostCredentialsItem(ctx).Credentials(creds).Execute()
	if err != nil {
		log.Fatalf("Error while getting token: %v", err)
	}

	if token.Token != nil {
		c.SetToken(*token.Token)

		log.Infof("...successfully logged in.\n\n")

		return nil
	}

	return fmt.Errorf("token is empty")
}

func (c *Config) Token() string {
	if c.GetString(fmt.Sprintf("%s_%s_token", c.ParentAppName(), c.AppName())) != "" {
		c.token = c.GetString(fmt.Sprintf("%s_%s_token", c.ParentAppName(), c.AppName()))
	}

	return c.token
}

func (c *Config) SetToken(token string) {
	c.token = token
	c.Set(fmt.Sprintf("%s_%s_token", c.ParentAppName(), c.AppName()), token)
}

func (c *Config) AppName() string {
	return c.appName
}

func (c *Config) ParentAppName() string {
	return c.GetString(fmt.Sprintf("%s_parent_app_name", c.appName))
}

func (c *Config) AppVersion() string {
	return c.GetString(fmt.Sprintf("%s_version", c.appName))
}

// AppHomeDir returns the application's home directory.
func (c *Config) AppHomeDir() string {
	return c.GetString(fmt.Sprintf("%s_home_dir", c.AppName()))
}

// IsDebug returns true if debug mode is set.
func (c *Config) IsDebug() bool {
	return c.GetBool("debug")
}
