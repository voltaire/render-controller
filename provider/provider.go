package provider

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
)

type Renderer interface {
	GetRendererInstance(ctx context.Context) (RendererInstance, error)
}

type DockerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error
	VolumeCreate(ctx context.Context, options volume.VolumeCreateBody) (types.Volume, error)
	VolumeList(ctx context.Context, filter filters.Args) (volume.VolumeListOKBody, error)
}

type RendererInstance interface {
	DockerClient
	Destroy(ctx context.Context) error
}
