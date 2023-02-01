package context

import (
	"github.com/pkg/errors"
	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/logic"
	"github.com/spf13/cobra"
)

func NewCmdContext(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "context",
			Short: "context",
			Long:  `configure context`,
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
		NewCmdContextList(app),
		NewCmdContextCreate(app),
		NewCmdContextDelete(app),
		NewCmdContextSelect(app),
	)

	return cmd
}

func NewCmdContextList(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "list",
			Short: "list",
			Long:  `list configured contexts`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewContextClient(app).RunCmdContextList(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running context list command")
				}

				return nil
			},
		},
		App: app,
	}

	cmd.Flags().Bool(
		"full",
		false,
		"print entities in full format",
	)
	_ = cmd.App.BindPFlag("full", cmd.Flags().Lookup("full"))

	return cmd
}

func NewCmdContextCreate(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "create",
			Short: "create",
			Long:  `create context`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewContextClient(app).RunCmdContextCreate(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running context create command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}

func NewCmdContextDelete(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "delete",
			Short: "delete",
			Long:  `delete context(s) by name`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewContextClient(app).RunCmdContextDelete(cmd, args)
				if err != nil {
					return errors.Wrap(err, "running context delete command")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}

func NewCmdContextSelect(app *config.App) *cmdpkg.Command {
	cmd := &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "select",
			Short: "select",
			Long:  `select a specified context`,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (
				[]string, cobra.ShellCompDirective,
			) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				err := logic.NewContextClient(app).RunCmdContextSelect(cmd, args)
				if err != nil {
					return errors.Wrap(err, "select a specified context")
				}

				return nil
			},
		},
		App: app,
	}

	return cmd
}
