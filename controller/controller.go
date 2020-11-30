package controller

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/voltaire/render-controller/renderer"
)

type Client struct {
	Docker *client.Client
}

func (ctrl *Client) StartForRender(ctx context.Context, cfg renderer.Config, backupTarballUri string) error {
	container, err := ctrl.Docker.ContainerCreate(ctx, &container.Config{
		Image: cfg.RenderControllerImage,
		Env:   cfg.BuildEnvironment("renderer", backupTarballUri),
		Labels: map[string]string{
			"service":  "renderer",
			"world":    cfg.OverworldName,
			"function": "render-controller",
		},
	}, &container.HostConfig{
		AutoRemove: true,
		LogConfig: container.LogConfig{
			Type: "awslogs",
			Config: map[string]string{
				"awslogs-group":        "render-controller/controller",
				"awslogs-create-group": "true",
				"awslogs-region":       cfg.AwsRegion,
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	}, &network.NetworkingConfig{}, "")
	if err != nil {
		return err
	}

	return ctrl.Docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
