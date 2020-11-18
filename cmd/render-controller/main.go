package main

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/render-controller/provider/generic"
	"github.com/voltaire/render-controller/renderer"
)

func main() {
	var cfg renderer.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	renderer := &renderer.Service{
		Config:           cfg,
		RendererProvider: &generic.Provider{},
	}

	err = renderer.Render(context.TODO(), cfg.BackupTarballUri)
	if err != nil {
		log.Fatal(err)
	}
}
