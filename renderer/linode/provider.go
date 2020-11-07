package linode

import (
	"context"

	"github.com/docker/machine/libmachine"
)

type Provider struct {
	machine libmachine.API
}

func (provider *Provider) GetRunner(ctx context.Context) (*Runner, error) {
	return nil, nil
}
