package logic

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/rewardenv/reward-cloud-cli/internal/config"
	"github.com/rewardenv/reward-cloud-cli/internal/globals"
)

type LocalClient struct {
	*Client
}

func NewLocalClient(c *config.App) *LocalClient {
	return &LocalClient{new(c)}
}

func (c *LocalClient) RunCmdLocalBuild(cmd *cobra.Command, args []string) error {
	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(cmd.Context())
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	dockerfile, err := c.generateDockerfile(ctx)
	if err != nil {
		return err
	}

	// Write dockerfile to a temporary file
	tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-", globals.PARENTAPPNAME, globals.APPNAME))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(dockerfile.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	//
	envVars, err := c.getBuildEnvVars(ctx)
	if err != nil {
		return errors.Wrap(err, "getting build env vars")
	}

	buildArgs := []string{"build", "-f", tmpFile.Name(), "."}
	for k, v := range envVars {
		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	out, err := c.Shell.Execute("docker", buildArgs...)
	if err != nil {
		return errors.Wrap(err, "building docker image")
	}

	_ = out

	return nil
}

func (c *LocalClient) RunCmdLocalDumpDockerfile(cmd *cobra.Command, args []string) error {
	ctx, err := NewLoginClient(c.App).CheckTokenAndLogin(cmd.Context())
	if err != nil {
		return errors.Wrap(err, "logging in")
	}

	dockerfile, err := c.generateDockerfile(ctx)
	if err != nil {
		return err
	}

	fmt.Println(dockerfile.String())

	return nil
}

func (c *LocalClient) generateDockerfile(ctx context.Context) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}

	projectType, err := c.getProjectType(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting project type")
	}

	tpl, err := c.Templates.LoadTemplate("templates/build/*")
	if err != nil {
		return nil, errors.Wrap(err, "loading template")
	}

	envVars, err := c.getBuildEnvVars(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting build env vars")
	}

	mainComponentVersion, err := c.getMainComponentVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting main component version")

	}

	data := map[string]interface{}{
		"Values": map[string]interface{}{
			"envVars":     envVars,
			"projectType": projectType,
			"phpVersion":  mainComponentVersion,
		},
	}

	err = c.Templates.ExecuteTemplate(tpl, "Dockerfile.tpl", data, buf)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	return buf, nil
}
