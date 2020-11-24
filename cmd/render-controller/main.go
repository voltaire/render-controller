package main

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/render-controller/provider/linode"
	"github.com/voltaire/render-controller/renderer"
	"github.com/voltaire/render-controller/renderer/events"
)

func main() {
	var cfg renderer.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var providerCfg linode.Config
	err = envconfig.Process("LINODE", &providerCfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("rendering using linode provider")
	provider, err := linode.New(&providerCfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	instance, err := provider.GetRendererInstance(ctx)
	if err != nil {
		log.Fatal(err)
	}

	renderer := &renderer.Service{
		Config: cfg,
	}

	log.Printf("starting render run for tarball '%s'", cfg.BackupTarballUri)
	err = renderer.Render(context.TODO(), instance, cfg.BackupTarballUri)
	if err != nil {
		log.Fatal(err)
	}

	events.HandleContainerEvents(instance, cfg)
}
