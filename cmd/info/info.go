package info

import (
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
)

func NewCmdInfo(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "info",
			Short: "info",
			Long:  `info`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewInfoClient(app).RunCmdInfo(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running info command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
