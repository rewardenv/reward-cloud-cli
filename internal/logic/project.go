package logic

import (
	"context"
	"fmt"

	"github.com/rewardenv/reward-cloud-go/rewardcloud"
	"github.com/spf13/cobra"
)

func (c *Client) RunCmdProjectList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := c.Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %w", err)
	}

	ctx = context.WithValue(ctx, rewardcloud.ContextAccessToken, c.Token())

	projects, _, err := c.Client.ProjectApi.ApiProjectsGetCollection(ctx).Execute()
	if err != nil {
		return err
	}

	if len(projects) > 0 {
		fmt.Println("The following projects are available:")
	} else {
		fmt.Println("No projects available.")
	}

	for _, project := range projects {
		fmt.Printf("- %s\n", project.GetName())
	}

	return nil
}
