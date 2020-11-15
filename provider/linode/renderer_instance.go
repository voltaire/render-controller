package linode

import (
	"context"

	"github.com/voltaire/render-controller/provider"
)

type RendererInstance struct {
	provider.DockerClient
}

func (instance *RendererInstance) Destroy(ctx context.Context) error {
	return nil
}
