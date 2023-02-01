package logic

import (
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"os"

	"github.com/rewardenv/reward/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CacheClient struct {
	*Client
}

func NewCacheClient(c *config.App) *CacheClient {
	return &CacheClient{
		new(c),
	}
}

func (c *CacheClient) RunCmdCacheClean(cmd *cobra.Command, args []string) error {
	log.Println("Cleaning cache...")

	err := os.RemoveAll(c.App.CacheDir())
	if err != nil {
		return errors.Wrap(err, "removing cache dir")
	}

	err = util.CreateDir(c.App.CacheDir(), nil)
	if err != nil {
		return errors.Wrap(err, "creating cache dir")
	}

	log.Println("...cache cleaned successfully")

	return nil
}
