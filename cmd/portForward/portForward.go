package portForward

import (
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
)

func NewCmdPortForward(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "port-forward",
			Short: "port-forward services",
			Long:  `port-forward services`,
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
		NewCmdPortForwardDB(app),
	)

	return cmd
}

func NewCmdPortForwardDB(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "db",
			Short: "port-forward database port",
			Long:  `port-forward database port`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewPortForwardClient(app).RunCmdPortForwardDB(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running db port-forward command")
				}

				return nil
			},
		},
		App: app,
	}

	cmd.Flags().Int(
		"local-port",
		3306,
		"",
	)
	_ = cmd.App.BindPFlag("local_port", cmd.Flags().Lookup("local-port"))

	return cmd
}
