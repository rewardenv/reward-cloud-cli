package logic

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/ui"
	"github.com/rewardenv/reward-cloud-sdk-go/rewardcloud"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	EnvStateRunning = "running"
)

type EnvClient struct {
	*Client
}

func NewEnvClient(c *config.App) *EnvClient {
	return &EnvClient{New(c)}
}

func (c *EnvClient) RunCmdEnvBuildAndDeploy(cmd *cobra.Command, args []string) error {
	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	env, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	patch := rewardcloud.EnvironmentEnvironmentOutput{
		Id: env.Id,
	}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdbuildAndDeployPatch(
		ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentOutput(patch).Execute()
	if err != nil {
		return errors.Wrap(err, "building environment")
	}

	p := tea.NewProgram(ui.NewModel("Building environment..."))

	var g errgroup.Group
	g.Go(func() error {
		// Initial delay
		initialized := false
		for {
			pause := time.Duration(3000) * time.Millisecond
			time.Sleep(pause)

			ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
			if err != nil {
				return errors.Wrap(err, "logging in")
			}

			environment, err := c.getEnvironment(ctx)
			if err != nil {
				return errors.Wrap(err, "getting environment")
			}

			state, err := c.getStateNameByID(ctx, GetIDFromPath(environment.GetState()))
			if err != nil {
				return errors.Wrap(err, "getting state")
			}

			if strings.ToLower(state) != EnvStateRunning {
				initialized = true
			}

			if strings.ToLower(state) == EnvStateRunning && initialized {
				p.Send(ui.ResultMsg{Ready: true})

				return nil
			}

			p.Send(ui.ResultMsg{Msg: fmt.Sprintf("Environment status: %s", state)})
		}
	})

	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %s", err)
		os.Exit(1)
	}

	if err := g.Wait(); err != nil {
		log.Errorf("Error waiting for program: %s", err)
		os.Exit(1)
	}

	log.Infof("Build and deploy finished")

	return nil
}

func (c *EnvClient) RunCmdEnvExportDB(cmd *cobra.Command, args []string) error {
	const datatype = "Database"

	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	stripDatabase := rewardcloud.NewNullableBool(rewardcloud.PtrBool(c.GetBool("strip_database")))
	post := rewardcloud.EnvironmentEnvironmentInput{
		IsStripDatabase: *stripDatabase,
	}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdexportDatabasePut(
		ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentInput(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting database")
	}

	p := tea.NewProgram(ui.NewModel(fmt.Sprintf("Exporting %s...", datatype)))
	result := ""

	var g errgroup.Group
	g.Go(func() error {
		// Initial delay
		initialized := false
		for {
			pause := time.Duration(3000) * time.Millisecond
			time.Sleep(pause)

			ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
			if err != nil {
				return errors.Wrap(err, "logging in")
			}

			environment, err := c.getEnvironment(ctx)
			if err != nil {
				return errors.Wrap(err, "getting environment")
			}

			state, err := c.getStateNameByID(ctx, GetIDFromPath(environment.GetState()))
			if err != nil {
				return errors.Wrap(err, "getting state")
			}

			if strings.ToLower(state) != EnvStateRunning {
				initialized = true
			}

			if strings.ToLower(state) == EnvStateRunning && initialized {
				datatypeID, err := c.GetDatatransferDataTypeID(ctx, datatype)
				if err != nil {
					return errors.Wrap(err, "getting data type id")
				}

				res, _, err := c.RewardCloud.ExportedDataApi.ApiExportedDatasGetCollection(ctx).
					Environment(c.getRcContext(ctx).Environment).
					DataTransferDataType(datatypeID).
					OrderCreatedAt("desc").
					Execute()
				if err != nil {
					result = fmt.Sprintf("Error getting exported data: %s", err)
				} else {
					if len(res) < 1 {
						result = "No exported data found"
					}

					result = fmt.Sprintf("Exported data: %s", res[0].GetUrl())
				}

				p.Send(ui.ResultMsg{Ready: true})

				return nil
			}

			p.Send(ui.ResultMsg{Msg: fmt.Sprintf("Environment status: %s", state)})
		}
	})

	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %s", err)
		os.Exit(1)
	}

	if err := g.Wait(); err != nil {
		log.Errorf("Error waiting for program: %s", err)
		os.Exit(1)
	}

	log.Infof("Export finished: %s", result)

	return nil
}

func (c *EnvClient) RunCmdEnvExportMedia(cmd *cobra.Command, args []string) error {
	const datatype = "Media"

	ctx, err := c.prepareContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "preparing context")
	}

	post := rewardcloud.EnvironmentEnvironmentInput{}

	_, _, err = c.RewardCloud.EnvironmentApi.ApiEnvironmentsIdexportMediaPut(
		ctx, c.getRcContext(ctx).Environment).EnvironmentEnvironmentInput(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting media")
	}

	p := tea.NewProgram(ui.NewModel(fmt.Sprintf("Exporting %s...", datatype)))
	result := ""

	var g errgroup.Group
	g.Go(func() error {
		// Initial delay
		initialized := false
		for {
			pause := time.Duration(3000) * time.Millisecond
			time.Sleep(pause)

			ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
			if err != nil {
				return errors.Wrap(err, "logging in")
			}

			environment, err := c.getEnvironment(ctx)
			if err != nil {
				return errors.Wrap(err, "getting environment")
			}

			state, err := c.getStateNameByID(ctx, GetIDFromPath(environment.GetState()))
			if err != nil {
				return errors.Wrap(err, "getting state")
			}

			if strings.ToLower(state) != EnvStateRunning {
				initialized = true
			}

			if strings.ToLower(state) == EnvStateRunning && initialized {
				datatypeID, err := c.GetDatatransferDataTypeID(ctx, datatype)
				if err != nil {
					return errors.Wrap(err, "getting data type id")
				}

				res, _, err := c.RewardCloud.ExportedDataApi.ApiExportedDatasGetCollection(ctx).
					Environment(c.getRcContext(ctx).Environment).
					DataTransferDataType(datatypeID).
					OrderCreatedAt("desc").
					Execute()
				if err != nil {
					result = fmt.Sprintf("Error getting exported data: %s", err)
				} else {
					if len(res) < 1 {
						result = "No exported data found"
					}

					result = fmt.Sprintf("Exported data: %s", res[0].GetUrl())
				}

				p.Send(ui.ResultMsg{Ready: true})

				return nil
			}

			p.Send(ui.ResultMsg{Msg: fmt.Sprintf("Environment status: %s", state)})
		}
	})

	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %s", err)
		os.Exit(1)
	}

	if err := g.Wait(); err != nil {
		log.Errorf("Error waiting for program: %s", err)
		os.Exit(1)
	}

	log.Infof("Export finished: %s", result)

	return nil
}

func (c *EnvClient) GetDatatransferDataTypeID(ctx context.Context, s string) (string, error) {
	dts, _, err := c.RewardCloud.DataTransferDataTypeApi.ApiDataTransferDataTypesGetCollection(ctx).Execute()
	if err != nil {
		return "", errors.Wrap(err, "getting data transfer data types")
	}

	id := ""
	for _, dt := range dts {
		if dt.GetName() == s {
			id = strconv.FormatInt(int64(dt.GetId()), 10)
		}
	}

	return id, nil
}
