package cache

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdCache(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:     "cache",
			Short:   "cache",
			Long:    `cache`,
			Aliases: []string{"c"},
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return cmd.Help() //nolint:wrapcheck
			},
		},
		App: app,
	}

	cmd.AddCommands(
		NewCmdCacheClean(app),
	)

	return cmd
}

func NewCmdCacheClean(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:     "clean",
			Short:   "clean",
			Long:    `clean`,
			Aliases: []string{"c", "clear", "flush"},
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewCacheClient(app).RunCmdCacheClean(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running clean command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
