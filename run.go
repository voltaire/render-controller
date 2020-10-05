package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func buildContainerEnv(cfg Config, backupTarballURI string) []string {
	return []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", os.Getenv("AWS_ACCESS_KEY_ID")),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", os.Getenv("AWS_SECRET_ACCESS_KEY")),
		fmt.Sprintf("OVERWORLD_DIR=%s", cfg.OverworldName),
		fmt.Sprintf("NETHER_DIR=%s", cfg.NetherName),
		fmt.Sprintf("THE_END_DIR=%s", cfg.TheEndName),
		fmt.Sprintf("BACKUP_TARBALL_URI=%s", backupTarballURI),
		fmt.Sprintf("MAP_OUTPUT_URI=%s", cfg.MapOutputURI),
	}
}

func (svc *server) startRenderer(ctx context.Context, backupTarballURI string) error {
	r, err := svc.docker.ImagePull(ctx, svc.cfg.RendererImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(ioutil.Discard, r)
	if err != nil {
		return err
	}

	container, err := svc.docker.ContainerCreate(ctx, &container.Config{
		Image: svc.cfg.RendererImage,
		Env:   buildContainerEnv(svc.cfg, backupTarballURI),
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{}, "")
	if err != nil {
		return err
	}

	return svc.docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
