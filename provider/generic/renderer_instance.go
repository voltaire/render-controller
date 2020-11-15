package generic

import (
	"context"

	"github.com/voltaire/render-controller/provider"
)

type RendererInstance struct {
	provider.DockerClient
}

// this is a noop for now
func (instance *RendererInstance) Destroy(ctx context.Context) error {
	return nil
}
