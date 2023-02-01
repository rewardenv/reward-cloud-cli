package shell

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
)

func NewCmdShell(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "shell",
			Short: "open a shell in a reward cloud environment",
			Long:  `open a shell in a reward cloud environment`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewShellClient(app).RunCmdShell(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running shell command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
