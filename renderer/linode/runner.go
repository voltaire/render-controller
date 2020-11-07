package linode

import "context"

type Runner struct{}

func (runner *Runner) Destroy(ctx context.Context) error {
	return nil
}
