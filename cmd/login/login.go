package login

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward-cloud-cli/cmd"
	"reward-cloud-cli/internal/config"
	"reward-cloud-cli/internal/logic"
)

func NewCmdLogin(c *config.Config) *cmdpkg.Command {
	var cmd = &cmdpkg.Command{
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
				err := logic.New(c).RunCmdLogin(cmd, args)
				if err != nil {
					return fmt.Errorf("error running login command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	return cmd
}
