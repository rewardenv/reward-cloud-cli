package config

import "strconv"

type RcContext struct {
	Name         string `json:"name" yaml:"name"`
	Organization string `json:"organization" yaml:"organization"`
	Team         string `json:"team" yaml:"team"`
	Project      string `json:"project" yaml:"project"`
	Environment  string `json:"environment" yaml:"environment"`
}

func (c *RcContext) OrganizationID() int32 {
	id, err := strconv.ParseUint(c.Organization, 10, 32)
	if err != nil {
		return -1
	}

	return int32(id)
}

func (c *RcContext) TeamID() int32 {
	id, err := strconv.ParseUint(c.Team, 10, 32)
	if err != nil {
		return -1
	}

	return int32(id)
}

func (c *RcContext) ProjectID() int32 {
	id, err := strconv.ParseUint(c.Project, 10, 32)
	if err != nil {
		return -1
	}

	return int32(id)
}

func (c *RcContext) EnvironmentID() int32 {
	id, err := strconv.ParseUint(c.Environment, 10, 32)
	if err != nil {
		return -1
	}

	return int32(id)
}

type ContextKey struct{}

type Config struct {
	RewardCloudEndpoint string       `json:"endpoint" yaml:"endpoint"`
	RewardCloudID       string       `json:"id" yaml:"id"`
	RewardCloudPassword string       `json:"password" yaml:"password"`
	Contexts            []*RcContext `json:"contexts" yaml:"contexts"`
	CurrentContext      string       `json:"currentContext" yaml:"currentContext"`
}
