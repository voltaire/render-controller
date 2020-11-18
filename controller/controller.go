package controller

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/voltaire/render-controller/renderer"
)

type Controller struct {
	Docker *client.Client
}

func (ctrl *Controller) StartForRender(ctx context.Context, cfg renderer.Config, backupTarballUri string) error {
	container, err := ctrl.Docker.ContainerCreate(ctx, &container.Config{
		Image: cfg.RenderControllerImage,
		Env:   renderer.BuildContainerEnv(cfg, backupTarballUri),
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
	}, &network.NetworkingConfig{}, "")
	if err != nil {
		return err
	}

	return ctrl.Docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
