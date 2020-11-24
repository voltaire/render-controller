package linode

import (
	"context"

	"github.com/docker/machine/libmachine/host"
	"github.com/voltaire/render-controller/provider"
)

type RendererInstance struct {
	provider.DockerClient
	host *host.Host
}

func (instance *RendererInstance) Destroy(ctx context.Context) error {
	return instance.host.Stop()
}
