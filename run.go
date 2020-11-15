package main

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

func buildContainerEnv(cfg Config, backupTarballURI string) []string {
	return []string{
		"AWS_REGION=" + cfg.AWSRegion,
		"AWS_ACCESS_KEY_ID=" + cfg.AWSAccessKeyId,
		"AWS_SECRET_ACCESS_KEY=" + cfg.AWSSecretAccessKey,
		"OVERWORLD_DIR=" + cfg.OverworldName,
		"NETHER_DIR=" + cfg.NetherName,
		"THE_END_DIR=" + cfg.TheEndName,
		"BACKUP_TARBALL_URI=" + backupTarballURI,
		"DESTINATION_BUCKET_URI=" + cfg.DestinationBucketURI,
		"DESTINATION_BUCKET_ENDPOINT=" + cfg.DestinationBucketEndpoint,
		"DESTINATION_ACCESS_KEY_ID=" + cfg.DestinationAccessKeyId,
		"DESTINATION_SECRET_ACCESS_KEY=" + cfg.DestinationSecretAccessKey,
		"DISCORD_WEBHOOK_URL=" + cfg.DiscordWebhookUrl,
		"RUNNER_NAME=" + cfg.RunnerName,
	}
}

func (svc *server) checkForAlreadyRunningContainer(ctx context.Context) (bool, error) {
	args := filters.NewArgs()
	args.Add("label", "service=renderer")
	args.Add("label", "world="+svc.cfg.OverworldName)
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
	log.Println("creating container to render '" + backupTarballURI + "'")
	renderVolume, err := svc.renderer.GetOrCreateVolume(ctx, svc.cfg.OverworldName)
	if err != nil {
		return err
	}

	container, err := svc.docker.ContainerCreate(ctx, &container.Config{
		Image: svc.cfg.RendererImage,
		Env:   buildContainerEnv(svc.cfg, backupTarballURI),
		Labels: map[string]string{
			"service": "renderer",
			"world":   svc.cfg.OverworldName,
		},
	}, &container.HostConfig{
		AutoRemove: true,
		LogConfig: container.LogConfig{
			Type: "awslogs",
			Config: map[string]string{
				"awslogs-group":        "renderer",
				"awslogs-create-group": "true",
				"awslogs-region":       svc.cfg.AWSRegion,
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

	log.Println("starting container id '" + container.ID + "'")
	return svc.docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
