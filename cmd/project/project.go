package project

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdpkg "reward-cloud-cli/cmd"
	"reward-cloud-cli/internal/config"
	"reward-cloud-cli/internal/logic"
)

func NewCmdProject(c *config.Config) *cmdpkg.Command {
	var cmd = &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "project",
			Short: "manipulate projects",
			Long:  `manipulate projects`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				// TODO: print current project

				return nil
			},
		},
		Config: c,
	}

	cmd.AddCommands(
		NewCmdProjectList(c),
	)

	return cmd
}

func NewCmdProjectList(c *config.Config) *cmdpkg.Command {
	var cmd = &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "list",
			Short: "list projects",
			Long:  `list projects`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.New(c).RunCmdProjectList(cmd, args)
				if err != nil {
					return fmt.Errorf("error running project list command: %w", err)
				}

				return nil
			},
		},
		Config: c,
	}

	return cmd
}
