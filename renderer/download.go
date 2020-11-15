package renderer

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

func (svc *Service) GetOrCreateVolume(ctx context.Context, world string) (*types.Volume, error) {
	output, err := svc.Docker.VolumeList(ctx, filters.NewArgs(filters.KeyValuePair{
		Key:   "WorldName",
		Value: world,
	}, filters.KeyValuePair{
		Key:   "Service",
		Value: "Renderer",
	}))
	if err != nil {
		return nil, err
	}
	if len(output.Volumes) != 0 {
		return output.Volumes[0], nil
	}

	vol, err := svc.Docker.VolumeCreate(ctx, volume.VolumeCreateBody{
		Labels: map[string]string{
			"WorldName": world,
			"Service":   "Renderer",
		},
	})
	return &vol, err
}
