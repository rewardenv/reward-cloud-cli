package logic

import (
	"fmt"
	"strings"

	"github.com/rewardenv/reward-cloud-cli/internal/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	cmdpkg "github.com/rewardenv/reward-cloud-cli/cmd"
)

type RootClient struct {
	*Client
}

func NewRootClient(c *config.App) *RootClient {
	return &RootClient{New(c)}
}

// RunCmdRoot is the default command. If no additional args passed print the help.
func (c *RootClient) RunCmdRoot(cmd *cmdpkg.Command) error {
	if cmd.App.GetBool(fmt.Sprintf("%s_print_environment", cmd.App.ConfigPrefix())) {
		for i, v := range viper.AllSettings() {
			log.Printf("%s=%v", strings.ToUpper(i), v)
		}

		return nil
	}

	_ = cmd.Help()

	return nil
}
