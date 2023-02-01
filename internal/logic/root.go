package logic

import "github.com/spf13/cobra"

func (c *Client) RunCmdRoot(cmd *cobra.Command) error {
	return cmd.Help()
}
