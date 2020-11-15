package generic

import (
	"context"

	"github.com/voltaire/render-controller/provider"
)

type Provider struct {
	Docker provider.DockerClient
}

func (svc *Provider) GetRendererInstance(ctx context.Context) (provider.RendererInstance, error) {
	return &RendererInstance{
		DockerClient: svc.Docker,
	}, nil
}
