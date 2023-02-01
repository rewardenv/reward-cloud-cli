package config

import (
	"container/list"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/rewardenv/reward-cloud-sdk-go/rewardcloud"
	"github.com/rewardenv/reward/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// FS is the implementation of Afero Filesystem. It's a filesystem wrapper and used for testing.
var FS = &afero.Afero{Fs: afero.NewOsFs()}

type App struct {
	*viper.Viper
	appName       string
	parentAppName string
	token         string
	TmpFiles      *list.List
	RewardCloud   *rewardcloud.APIClient
}

func New(name, parentAppName, ver string) *App {
	a := &App{
		Viper:         viper.GetViper(),
		appName:       name,
		parentAppName: parentAppName,
		TmpFiles:      list.New(),
	}

	a.SetDefault(fmt.Sprintf("%s_%s_version", parentAppName, name), version.Must(version.NewVersion(ver)).String())

	return a
}

func (a *App) Cleanup() error {
	var err error
	for e := a.TmpFiles.Front(); e != nil; e = e.Next() {
		err2 := os.Remove(e.Value.(string))
		if err2 != nil {
			err = err2
		}
	}

	return err
}

func (a *App) Init() *App {
	a.AddConfigPath(".")

	cfg := a.GetString(fmt.Sprintf("%s_%s_config_file", a.parentAppName, a.appName))
	if cfg != "" {
		a.AddConfigPath(filepath.Dir(cfg))
		a.SetConfigName(filepath.Base(cfg))
		a.SetConfigType("yaml")
	}

	a.AutomaticEnv()

	if err := a.ReadInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	a.AddConfigPath(".")
	a.SetConfigName(".env")
	a.SetConfigType("dotenv")

	if err := a.MergeInConfig(); err != nil {
		log.Debugf("%v", err)
	}

	// Configure defaults.
	a.SetDefault("silence_errors", true)
	a.SetDefault(fmt.Sprintf("%s_%s_parent_app_name", a.parentAppName, a.appName), a.parentAppName)
	a.SetDefault(fmt.Sprintf("%s_parent_app_home_dir", a.ConfigPrefix()),
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s", a.ParentAppName())))
	a.SetDefault(fmt.Sprintf("%s_home_dir", a.ParentAppName()),
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s", a.ParentAppName())))
	a.SetDefault(fmt.Sprintf("%s_plugins_config_dir", a.ParentAppName()),
		filepath.Join(a.ParentAppHomeDir(), "plugins.conf.d"))
	a.SetDefault(fmt.Sprintf("%s_home_dir", a.ConfigPrefix()),
		filepath.Join(a.PluginsConfigDir(), a.AppName()))
	a.SetDefault(fmt.Sprintf("%s_cache_dir", a.ConfigPrefix()),
		filepath.Join(a.PluginsConfigDir(), a.AppName(), ".cache"))
	a.SetDefault(fmt.Sprintf("%s_token_file", a.ConfigPrefix()),
		filepath.Join(a.CacheDir(), "token"))

	// Cloud API App
	a.SetDefault(fmt.Sprintf("%s_endpoint", a.ConfigPrefix()), "dev.rewardcloud.itg.cloud")

	a.SetLogging()

	endpoint := rewardcloud.ServerConfigurations{
		{
			URL:         "https://" + a.GetString(fmt.Sprintf("%s_endpoint", a.ConfigPrefix())),
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
	a.RewardCloud = rewardcloud.NewAPIClient(conf)

	return a
}

// SetLogging sets the logging level based on the command line flags and environment variables.
func (a *App) SetLogging() {
	switch {
	case a.GetString("log_level") == "trace":
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	case a.IsDebug(), a.GetString("log_level") == "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case a.GetString("log_level") == "info":
		log.SetLevel(log.InfoLevel)
	case a.GetString("log_level") == "warning":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}

	log.SetFormatter(
		&log.TextFormatter{
			DisableColors:          a.GetBool("disable_colors"),
			ForceColors:            true,
			DisableLevelTruncation: true,
			FullTimestamp:          true,
			DisableTimestamp:       !a.GetBool("debug"),
			QuoteEmptyFields:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := strings.ReplaceAll(path.Base(f.File), "reward/", "")

				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
			},
		},
	)
}

func (a *App) Token() string {
	if a.GetString(fmt.Sprintf("%s_token", a.ConfigPrefix())) != "" {
		a.token = a.GetString(fmt.Sprintf("%s_token", a.ConfigPrefix()))
	}

	return a.token
}

func (a *App) SetToken(ctx context.Context, token string) context.Context {
	a.token = token
	a.Set(fmt.Sprintf("%s_token", a.ConfigPrefix()), token)

	return context.WithValue(ctx, rewardcloud.ContextAccessToken, token)
}

func (a *App) AppName() string {
	return a.appName
}

func (a *App) ParentAppName() string {
	return a.GetString(fmt.Sprintf("%s_%s_parent_app_name", a.parentAppName, a.appName))
}

func (a *App) AppVersion() string {
	return a.GetString(fmt.Sprintf("%s_version", a.ConfigPrefix()))
}

// AppHomeDir returns the application's home directory.
func (a *App) AppHomeDir() string {
	return a.GetString(fmt.Sprintf("%s_home_dir", a.ConfigPrefix()))
}

func (a *App) ParentAppHomeDir() string {
	return a.GetString(fmt.Sprintf("%s_home_dir", a.ParentAppName()))
}

func (a *App) PluginsConfigDir() string {
	return a.GetString(fmt.Sprintf("%s_plugins_config_dir", a.ParentAppName()))
}

// SilenceErrors returns true if errors should be silenced.
func (a *App) SilenceErrors() bool {
	return a.GetBool("silence_errors")
}

// IsDebug returns true if debug mode is set.
func (a *App) IsDebug() bool {
	return a.GetBool("debug")
}

func (a *App) TokenFile() string {
	return a.GetString(fmt.Sprintf("%s_token_file", a.ConfigPrefix()))
}

func (a *App) CacheDir() string {
	return a.GetString(fmt.Sprintf("%s_cache_dir", a.ConfigPrefix()))
}

func (a *App) ConfigPrefix() string {
	return fmt.Sprintf("%s_%s", a.ParentAppName(), a.AppName())
}

func (a *App) ConfigFilePath() string {
	return a.GetString(fmt.Sprintf("%s_config_file", a.ConfigPrefix()))
}

func (a *App) ReadConfig() (*Config, error) {
	configPath := a.ConfigFilePath()
	if configPath == "" {
		return nil, errors.New("config file path is empty")
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	conf := &Config{}

	err = yaml.Unmarshal(configBytes, conf)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling config file")
	}

	return conf, nil
}
