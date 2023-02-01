package root

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	cmdpkg "reward-cloud-cli/cmd"
	"reward-cloud-cli/cmd/login"
	"reward-cloud-cli/cmd/project"
	"reward-cloud-cli/internal/config"
	"reward-cloud-cli/internal/logic"
	"reward-cloud-cli/internal/util"
)

func NewCmdRoot(c *config.Config) *cmdpkg.Command {
	var cmd = &cmdpkg.Command{
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
			Version: c.AppVersion(),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return logic.New(c).RunCmdRoot(cmd)
			},
		},
		Config: c,
	}

	addFlags(cmd)

	cmd.AddCommands(
		login.NewCmdLogin(c),
		project.NewCmdProject(c),
	)

	return cmd
}

func addFlags(cmd *cmdpkg.Command) {
	// --app-dir
	cmd.PersistentFlags().String(
		"app-dir",
		filepath.Join(
			util.HomeDir(),
			fmt.Sprintf(".%s", cmd.Config.ParentAppName()),
			"plugins.d",
			cmd.Config.AppName(),
		),
		"app home directory",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_home_dir", cmd.Config.AppName()),
		cmd.PersistentFlags().Lookup("app-dir"))

	// --log-level
	cmd.PersistentFlags().String(
		"log-level", "info", "logging level (options: trace, debug, info, warning, error)",
	)
	_ = cmd.Config.BindPFlag("log_level", cmd.PersistentFlags().Lookup("log-level"))

	// --debug
	cmd.PersistentFlags().Bool(
		"debug", false, "enable debug mode (same as --log-level=debug)",
	)
	_ = cmd.Config.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	// --disable-colors
	cmd.PersistentFlags().Bool(
		"disable-colors", false, "disable colors in output",
	)
	_ = cmd.Config.BindPFlag("disable_colors", cmd.PersistentFlags().Lookup("disable-colors"))

	// --config
	cmd.PersistentFlags().StringP(
		"config",
		"c",
		filepath.Join(util.HomeDir(), fmt.Sprintf(".%s.yml", cmd.Config.AppName())),
		"config file",
	)
	_ = cmd.Config.BindPFlag(fmt.Sprintf("%s_config_file", cmd.Config.AppName()),
		cmd.PersistentFlags().Lookup("config"))
}
