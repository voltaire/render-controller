package renderer

import "github.com/docker/docker/client"

type Service struct {
	Docker *client.Client
}
