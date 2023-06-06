package root

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/cmd/cache"
	"github.com/rewardenv/reward-cloud-cli/cmd/context"
	"github.com/rewardenv/reward-cloud-cli/cmd/env"
	"github.com/rewardenv/reward-cloud-cli/cmd/info"
	"github.com/rewardenv/reward-cloud-cli/cmd/portforward"

	"github.com/rewardenv/reward/pkg/util"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/cmd/login"
	"github.com/rewardenv/reward-cloud-cli/cmd/shell"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdRoot(conf *config.App) *cmdpkg.Command {
	conf.Init()

	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "cloud [command]",
			Short: "cloud is a cli tool that helps you interact with reward cloud",
			Long: ` ██▀███  ▓█████  █     █░ ▄▄▄       ██▀███  ▓█████▄     ▄████▄   ██▓     ▒█████   █    ██ ▓█████▄ 
▓██ ▒ ██▒▓█   ▀ ▓█░ █ ░█░▒████▄    ▓██ ▒ ██▒▒██▀ ██▌   ▒██▀ ▀█  ▓██▒    ▒██▒  ██▒ ██  ▓██▒▒██▀ ██▌
▓██ ░▄█ ▒▒███   ▒█░ █ ░█ ▒██  ▀█▄  ▓██ ░▄█ ▒░██   █▌   ▒▓█    ▄ ▒██░    ▒██░  ██▒▓██  ▒██░░██   █▌
▒██▀▀█▄  ▒▓█  ▄ ░█░ █ ░█ ░██▄▄▄▄██ ▒██▀▀█▄  ░▓█▄   ▌   ▒▓▓▄ ▄██▒▒██░    ▒██   ██░▓▓█  ░██░░▓█▄   ▌
░██▓ ▒██▒░▒████▒░░██▒██▓  ▓█   ▓██▒░██▓ ▒██▒░▒████▓    ▒ ▓███▀ ░░██████▒░ ████▓▒░▒▒█████▓ ░▒████▓ 
░ ▒▓ ░▒▓░░░ ▒░ ░░ ▓░▒ ▒   ▒▒   ▓▒█░░ ▒▓ ░▒▓░ ▒▒▓  ▒    ░ ░▒ ▒  ░░ ▒░▓  ░░ ▒░▒░▒░ ░▒▓▒ ▒ ▒  ▒▒▓  ▒ 
  ░▒ ░ ▒░ ░ ░  ░  ▒ ░ ░    ▒   ▒▒ ░  ░▒ ░ ▒░ ░ ▒  ▒      ░  ▒   ░ ░ ▒  ░  ░ ▒ ▒░ ░░▒░ ░ ░  ░ ▒  ▒ 
  ░░   ░    ░     ░   ░    ░   ▒     ░░   ░  ░ ░  ░    ░          ░ ░   ░ ░ ░ ▒   ░░░ ░ ░  ░ ░  ░ 
   ░        ░  ░    ░          ░  ░   ░        ░       ░ ░          ░  ░    ░ ░     ░        ░    
                                             ░         ░                                   ░`,
			Version:       conf.AppVersion(),
			SilenceErrors: conf.SilenceErrors(),
			SilenceUsage:  true,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewRootClient(conf).RunCmdRoot(&cmdpkg.Command{Command: cmd, App: conf})
				if err != nil {
					return errors.Wrap(err, "running root command")
				}

				return nil
			},
		},
		App: conf,
	}

	addFlags(cmd)
	// Reinitialize config after command line flags are added
	conf.Init()

	cmd.AddCommands(
		cache.NewCmdCache(conf),
		context.NewCmdContext(conf),
		login.NewCmdLogin(conf),
		shell.NewCmdShell(conf),
		portforward.NewCmdPortForward(conf),
		env.NewCmdEnv(conf),
		info.NewCmdInfo(conf),
	)

	return cmd
}

func addFlags(cmd *cmdpkg.Command) {
	// --app-dir
	cmd.PersistentFlags().String(
		"app-dir",
		filepath.Join(
			util.HomeDir(),
			fmt.Sprintf(".%s", cmd.App.ParentAppName()),
			"plugins.conf.d",
			cmd.App.AppName(),
		),
		"app home directory",
	)
	_ = cmd.App.BindPFlag(fmt.Sprintf("%s_home_dir", cmd.App.ConfigPrefix()),
		cmd.PersistentFlags().Lookup("app-dir"))

	// --log-level
	cmd.PersistentFlags().String(
		"log-level", "info", "logging level (options: trace, debug, info, warning, error)",
	)
	_ = cmd.App.BindPFlag("log_level", cmd.PersistentFlags().Lookup("log-level"))

	// --debug
	cmd.PersistentFlags().Bool(
		"debug", false, "enable debug mode (same as --log-level=debug)",
	)
	_ = cmd.App.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	// --disable-colors
	cmd.PersistentFlags().Bool(
		"disable-colors", false, "disable colors in output",
	)
	_ = cmd.App.BindPFlag("disable_colors", cmd.PersistentFlags().Lookup("disable-colors"))

	// --config
	cmd.PersistentFlags().StringP(
		"config",
		"c",
		filepath.Join(
			util.HomeDir(),
			fmt.Sprintf(".%s", cmd.App.ParentAppName()),
			"plugins.conf.d",
			cmd.App.AppName(),
			"config.yml",
		),
		"config file",
	)
	_ = cmd.App.BindPFlag(fmt.Sprintf("%s_config_file", cmd.App.ConfigPrefix()),
		cmd.PersistentFlags().Lookup("config"))

	// --print-environment
	cmd.Flags().Bool(
		"print-environment", false, "environment vars",
	)
	_ = cmd.App.BindPFlag(fmt.Sprintf("%s_print_environment", cmd.App.ConfigPrefix()),
		cmd.Flags().Lookup("print-environment"))
}
