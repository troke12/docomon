package container

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type ContainerService interface {
	ListContainers(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

type DockerClient struct {
	cli *client.Client
}

func NewContainerService(cli *client.Client) ContainerService {
	return &DockerClient{
		cli: cli,
	}
}

func (dc *DockerClient) ListContainers(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return dc.cli.ContainerList(ctx, options)
}

func (dc *DockerClient) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return dc.cli.ContainerInspect(ctx, containerID)
}
