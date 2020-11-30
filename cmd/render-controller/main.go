package main

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/render-controller/provider/linode"
	"github.com/voltaire/render-controller/renderer"
	"github.com/voltaire/render-controller/renderer/events"
)

type config struct {
	BackupTarballURI string `evconfig:"BACKUP_TARBALL_URI" required:"true"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	var rendererConfig renderer.Config
	envconfig.MustProcess("renderer", &rendererConfig)

	var providerCfg linode.Config
	envconfig.MustProcess("LINODE", &providerCfg)

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
		Config: rendererConfig,
	}

	log.Printf("starting render run for tarball '%s'", cfg.BackupTarballURI)
	err = renderer.Render(context.TODO(), instance, cfg.BackupTarballURI)
	if err != nil {
		log.Fatal(err)
	}

	events.HandleContainerEvents(instance, rendererConfig)
}
