package generic

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
	"github.com/voltaire/render-controller/provider"
)

type Provider struct {
	sync.Once
	docker provider.DockerClient
}

func (svc *Provider) GetRendererInstance(ctx context.Context) (provider.RendererInstance, error) {
	var err error
	svc.Once.Do(func() {
		svc.docker, err = client.NewClientWithOpts(client.FromEnv)
	})
	if err != nil {
		return nil, err
	}

	return &RendererInstance{
		DockerClient: svc.docker,
	}, nil
}
