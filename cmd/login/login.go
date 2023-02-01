package login

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdLogin(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "login",
			Short: "login to reward cloud",
			Long:  `login to reward cloud`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewLoginClient(app).RunCmdLogin(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running login command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
