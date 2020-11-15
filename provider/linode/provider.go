package linode

import (
	"context"

	"github.com/docker/machine/libmachine"
	"github.com/voltaire/render-controller/provider"
)

type Provider struct {
	Machine libmachine.API
}

func (provider *Provider) GetRendererInstance(ctx context.Context) (provider.RendererInstance, error) {
	return nil, nil
}
