package logic

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/ui"
	"github.com/rewardenv/reward-cloud-cli/pkg/api/rewardcloud"
)

type EnvClient struct {
	*Client
}

func NewEnvClient(c *config.App) *EnvClient {
	return &EnvClient{new(c)}
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

	bdp, resp, err := c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdBuildAndDeployPatch(ctx, c.getRcContext(ctx).EnvironmentID()).EnvironmentEnvironmentOutput(patch).Execute()
	if err != nil {
		_ = bdp
		_ = resp
		return errors.Wrap(err, "building environment")
	}

	p := tea.NewProgram(ui.NewModel("Building environment..."))

	go func() error {
		// Initial delay
		initialized := false
		for {
			pause := 3 * time.Second
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

			if strings.ToLower(state) != "running" {
				initialized = true
			}

			if strings.ToLower(state) == "running" && initialized {
				p.Send(ui.ResultMsg{Ready: true})

				return nil
			}

			p.Send(ui.ResultMsg{Msg: fmt.Sprintf("Environment status: %s", state)})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
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
	post := rewardcloud.EnvironmentEnvironmentOutput{
		IsStripDatabase: *stripDatabase,
	}

	_, _, err = c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdExportDatabasePut(ctx, c.getRcContext(ctx).EnvironmentID()).EnvironmentEnvironmentOutput(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting database")
	}

	p := tea.NewProgram(ui.NewModel(fmt.Sprintf("Exporting %s...", datatype)))
	result := ""

	go func() error {
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

			if strings.ToLower(state) != "running" {
				initialized = true
			}

			if strings.ToLower(state) == "running" && initialized {
				datatypeID, err := c.GetDatatransferDataTypeID(ctx, datatype)
				if err != nil {
					return errors.Wrap(err, "getting data type id")
				}

				res, _, err := c.RewardCloud.ExportedDataAPI.ApiExportedDatasGetCollection(ctx).
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
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
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

	post := rewardcloud.EnvironmentEnvironmentOutput{}

	_, _, err = c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdExportMediaPut(ctx, c.getRcContext(ctx).EnvironmentID()).EnvironmentEnvironmentOutput(post).Execute()
	if err != nil {
		return errors.Wrap(err, "exporting media")
	}

	p := tea.NewProgram(ui.NewModel(fmt.Sprintf("Exporting %s...", datatype)))
	result := ""

	go func() error {
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

			if strings.ToLower(state) != "running" {
				initialized = true
			}

			if strings.ToLower(state) == "running" && initialized {
				datatypeID, err := c.GetDatatransferDataTypeID(ctx, datatype)
				if err != nil {
					return errors.Wrap(err, "getting data type id")
				}

				res, _, err := c.RewardCloud.ExportedDataAPI.ApiExportedDatasGetCollection(ctx).
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
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	log.Infof("Export finished: %s", result)

	return nil
}

func (c *EnvClient) GetDatatransferDataTypeID(ctx context.Context, s string) (string, error) {
	dts, _, err := c.RewardCloud.DataTransferDataTypeAPI.ApiDataTransferDataTypesGetCollection(ctx).Execute()
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

func (c *EnvClient) RunCmdLogs(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(ctx)
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	environment, err := c.getEnvironment(ctx)
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}

	f := c.GetStringSlice("filter")
	filter := struct {
		Filters []string `json:"filters"`
	}{}
	for _, f := range f {
		filter.Filters = append(filter.Filters, fmt.Sprintf("component=%s", f))
	}

	jsonFilters, err := json.Marshal(filter)
	if err != nil {
		return errors.Wrap(err, "marshaling filters")
	}

	logstream, resp, err := c.RewardCloud.EnvironmentAPI.ApiEnvironmentsIdLogsGet(ctx, environment.GetId()).
		// XFilters(`{"filters":["component=application"]}`).
		XFilters(string(jsonFilters)).
		XOffset(int32(c.GetDuration("offset").Seconds())).
		Execute()
	if err != nil {
		_ = resp
		return errors.Wrap(err, "getting log stream")
	}

	topic := url.QueryEscape(logstream.GetTopic())

	r, err := http.NewRequestWithContext(ctx, "GET", c.Endpoint()+"/.well-known/mercure?topic="+topic, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	r.Header.Set("Authorization", "Bearer "+logstream.GetToken())

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer func() { _ = res.Body.Close() }()

	log.Info("Receiving logs from stream, press Ctrl+C to exit...")

	scanner := bufio.NewScanner(res.Body)
	for {
		if !scanner.Scan() {
			break
		}

		data := strings.TrimLeft(scanner.Text(), `data: `)
		foo := struct {
			Data string `json:"data"`
		}{}
		_ = json.Unmarshal([]byte(data), &foo)
		if foo.Data != "" {
			decoded, _ := base64.StdEncoding.DecodeString(foo.Data)
			if decoded != nil {
				fmt.Printf("%s\n", decoded)
			}
		}
	}

	return nil
}
