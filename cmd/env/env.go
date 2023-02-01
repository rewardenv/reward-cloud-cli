package env

import (
	"github.com/pkg/errors"
	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
	"github.com/spf13/cobra"
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
		NewCmdEnvStatus(app),
		NewCmdEnvBuildAndDeploy(app),
		NewCmdEnvExportDB(app),
		NewCmdEnvExportMedia(app),
	)

	return cmd
}

func NewCmdEnvStatus(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "status",
			Short: "environment status",
			Long:  `environment status`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewEnvClient(app).RunCmdEnvStatus(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running db port-forward command")
				}

				return nil
			},
		},
		App: app,
	}

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
