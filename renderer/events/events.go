package events

import (
	"context"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/voltaire/render-controller/provider"
	"github.com/voltaire/render-controller/renderer"
)

func HandleContainerEvents(instance provider.RendererInstance, cfg renderer.Config) {
	for {
		ctx := context.Background()
		messages, errChan := instance.Events(ctx, types.EventsOptions{
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "type",
					Value: "container",
				},
				filters.KeyValuePair{
					Key:   "label",
					Value: "service=renderer",
				},
				filters.KeyValuePair{
					Key:   "label",
					Value: "function=renderer",
				},
				filters.KeyValuePair{
					Key:   "label",
					Value: "world=" + cfg.OverworldName,
				},
			),
		})

		select {
		case msg := <-messages:
			if msg.Type == "stop" {
				err := instance.Destroy(ctx)
				if err != nil {
					log.Printf("error destroying instance: %s", err)
				}
				return
			}
		case err := <-errChan:
			if err != io.EOF {
				log.Println("received EOF from event stream")
				return
			}
			log.Printf("event stream error: %s", err)
		}
	}
}
