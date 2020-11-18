package renderer

import (
	"context"
	"errors"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/voltaire/render-controller/provider"
)

type Service struct {
	Config           Config
	RendererProvider provider.Renderer
}

var ErrAlreadyRunningRender = errors.New("renderer: a render is already running")

func (svc *Service) Render(ctx context.Context, backupTarballURI string) error {
	instance, err := svc.RendererProvider.GetRendererInstance(ctx)
	if err != nil {
		return err
	}

	log.Println("checking for running container instance")
	alreadyRunning, err := checkForAlreadyRunningContainer(ctx, instance, svc.Config)
	if err != nil {
		return err
	}

	if alreadyRunning {
		return ErrAlreadyRunningRender
	}

	log.Println("starting renderer")
	return svc.startRenderer(ctx, instance, backupTarballURI)
}

func checkForAlreadyRunningContainer(ctx context.Context, instance provider.RendererInstance, cfg Config) (bool, error) {
	args := filters.NewArgs()
	args.Add("label", "service=renderer")
	args.Add("label", "function=renderer")
	args.Add("label", "world="+cfg.OverworldName)
	containers, err := instance.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return false, err
	}
	if len(containers) > 0 {
		return true, nil
	}
	return false, nil
}

func (svc *Service) startRenderer(ctx context.Context, instance provider.RendererInstance, backupTarballURI string) error {
	log.Printf("getting volume for '%s'", svc.Config.OverworldName)
	renderVolume, err := svc.getOrCreateVolume(ctx, instance, svc.Config.OverworldName)
	if err != nil {
		return err
	}

	log.Printf("creating container to render '%s' using image '%s'", backupTarballURI, svc.Config.RendererImage)
	container, err := instance.ContainerCreate(ctx, &container.Config{
		Image: svc.Config.RendererImage,
		Env:   BuildContainerEnv(svc.Config, backupTarballURI),
		Labels: map[string]string{
			"service":  "renderer",
			"world":    svc.Config.OverworldName,
			"function": "renderer",
		},
	}, &container.HostConfig{
		AutoRemove: true,
		LogConfig: container.LogConfig{
			Type: "awslogs",
			Config: map[string]string{
				"awslogs-group":        "renderer",
				"awslogs-create-group": "true",
				"awslogs-region":       svc.Config.AwsRegion,
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: renderVolume.Name,
				Target: "/output",
			},
		},
	}, &network.NetworkingConfig{}, "")
	if err != nil {
		return err
	}

	log.Printf("starting container id '%s'", container.ID)
	return instance.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
