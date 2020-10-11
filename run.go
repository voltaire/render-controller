package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

func buildContainerEnv(cfg Config, backupTarballURI string) []string {
	return []string{
		fmt.Sprintf("AWS_REGION=%s", cfg.AWSRegion),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", cfg.AWSAccessKeyId),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", cfg.AWSSecretAccessKey),
		fmt.Sprintf("OVERWORLD_DIR=%s", cfg.OverworldName),
		fmt.Sprintf("NETHER_DIR=%s", cfg.NetherName),
		fmt.Sprintf("THE_END_DIR=%s", cfg.TheEndName),
		fmt.Sprintf("BACKUP_TARBALL_URI=%s", backupTarballURI),
		fmt.Sprintf("MAP_OUTPUT_URI=%s", cfg.MapOutputURI),
	}
}

func (svc *server) checkForAlreadyRunningContainer(ctx context.Context) (bool, error) {
	args := filters.NewArgs()
	args.Add("label", "service=renderer")
	args.Add("label", fmt.Sprintf("world=%s", svc.cfg.OverworldName))
	containers, err := svc.docker.ContainerList(ctx, types.ContainerListOptions{
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

func (svc *server) startRenderer(ctx context.Context, backupTarballURI string) error {
	container, err := svc.docker.ContainerCreate(ctx, &container.Config{
		Image: svc.cfg.RendererImage,
		Env:   buildContainerEnv(svc.cfg, backupTarballURI),
		Labels: map[string]string{
			"service": "renderer",
			"world":   svc.cfg.OverworldName,
		},
	}, &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: "render_output",
				Target: "/output",
			},
		},
	}, &network.NetworkingConfig{}, "")
	if err != nil {
		return err
	}

	return svc.docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
