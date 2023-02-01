package logic

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func (c *Client) RunCmdLogin(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := c.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %w", err)
	}

	return nil
}
