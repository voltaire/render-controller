package renderer

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/voltaire/render-controller/provider"
)

func (svc *Service) getOrCreateVolume(ctx context.Context, instance provider.RendererInstance, world string) (*types.Volume, error) {
	output, err := instance.VolumeList(ctx, filters.NewArgs(filters.KeyValuePair{
		Key:   "label",
		Value: "WorldName=" + world,
	}, filters.KeyValuePair{
		Key:   "label",
		Value: "Service=Renderer",
	}))
	if err != nil {
		return nil, err
	}
	if len(output.Volumes) != 0 {
		return output.Volumes[0], nil
	}

	vol, err := instance.VolumeCreate(ctx, volume.VolumeCreateBody{
		Labels: map[string]string{
			"WorldName": world,
			"Service":   "Renderer",
		},
	})
	return &vol, err
}
