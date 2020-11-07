package renderer

import "context"

type RunnerProvider interface {
	GetRunner(ctx context.Context) (Runner, error)
}

type Runner interface {
	Destroy(ctx context.Context) error
}
