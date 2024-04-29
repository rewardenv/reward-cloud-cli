package env

import (
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdEnv(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "env",
			Short: "environment",
			Long:  `environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help()
			},
		},
		App: app,
	}

	cmd.AddCommands(
		NewCmdEnvBuildAndDeploy(app),
		NewCmdEnvExportDB(app),
		NewCmdEnvExportMedia(app),
		NewCmdEnvLogs(app),
	)

	return cmd
}

func NewCmdEnvBuildAndDeploy(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "build-and-deploy",
			Short: "build and deploy environment",
			Long:  `build and deploy environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewEnvClient(app).RunCmdEnvBuildAndDeploy(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running build and deploy")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}

func NewCmdEnvExportDB(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "export-db",
			Short: "export database",
			Long:  `export database`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewEnvClient(app).RunCmdEnvExportDB(cmd, args)
				if err != nil {
					return errors.Wrap(err, "exporting database")
				}

				return nil
			},
		},
		App: app,
	}

	cmd.Flags().Bool("strip-database", false, "remove sensitve data from database dump")
	cmd.App.BindPFlag("strip_database", cmd.Flags().Lookup("strip-database"))

	return cmd
}

func NewCmdEnvExportMedia(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "export-media",
			Short: "export media",
			Long:  `export media`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewEnvClient(app).RunCmdEnvExportMedia(cmd, args)
				if err != nil {
					return errors.Wrap(err, "exporting media")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}

func NewCmdEnvLogs(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "logs",
			Short: "stream logs for the selected environment",
			Long:  `stream logs for the selected environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewEnvClient(app).RunCmdLogs(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running logs command")
				}

				return nil
			},
		},
		App: app,
	}

	logsCmdAddFlags(cmd)

	err := logsCmdCheckFlags(cmd)
	if err != nil {
		log.Fatalf(`%s`, err)
	}

	return cmd
}

func logsCmdAddFlags(cmd *cmdpkg.Command) {
	cmd.Flags().StringSlice(
		"filter",
		[]string{"application"},
		"comma separated list of components (eg: --filter=application,database)",
	)
	_ = cmd.App.BindPFlag("filter", cmd.Flags().Lookup("filter"))

	cmd.Flags().String(
		"offset", "1m", "read logs from the last offset time (eg: --offset=10s, --offset=5m, --offset=1h)",
	)
	_ = cmd.App.BindPFlag("offset", cmd.Flags().Lookup("offset"))

	cmd.ParseFlags(os.Args)
}

func logsCmdCheckFlags(cmd *cmdpkg.Command) error {
	filter, err := cmd.Flags().GetStringSlice("filter")
	if err != nil {
		return errors.Wrap(err, "getting filter flag")
	}

	if len(filter) == 0 {
		return errors.New("filter flag is required")
	}

	for _, f := range filter {
		switch f {
		case "application", "database", "message-queue", "search-engine", "redis", "http-cache", "build", "export-db", "import-db", "export-media", "import-media":
			continue
		default:
			return errors.Errorf(`invalid filter value: %s
Valid values are:
	application
	database
	message-queue
	search-engine
	redis
	http-cache
	build
	export-db
	import-db
	export-media
	import-media`, f)
		}
	}

	// the offset should not be longer than 1 day
	val, _ := cmd.Flags().GetString("offset")
	offset, err := time.ParseDuration(val)
	if err != nil {
		return errors.Wrap(err, "getting offset flag")
	}

	if offset > time.Duration(24*time.Hour) {
		return errors.New("offset should not be longer than 24h")
	}

	if offset < 0 {
		return errors.New("offset should be a positive value")
	}

	return nil
}
