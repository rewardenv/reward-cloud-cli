package local

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdLocal(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "local",
			Short: "simulate reward cloud commands locally",
			Long:  `simulate reward cloud commands locally`,
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
		NewCmdLocalBuild(app),
		NewCmdLocalDumpDockerfile(app),
	)

	return cmd
}

func NewCmdLocalBuild(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "build",
			Short: "build environment locally",
			Long:  `build environment locally`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewLocalClient(app).RunCmdLocalBuild(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running local build")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}

func NewCmdLocalDumpDockerfile(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "dump-dockerfile",
			Short: "dump the dockerfile which is used for cloud build",
			Long:  `dump the dockerfile which is used for cloud build`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewLocalClient(app).RunCmdLocalDumpDockerfile(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running local build")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
