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

	log.Println("rendering using generic provider")
	renderer := &renderer.Service{
		Config:           cfg,
		RendererProvider: &generic.Provider{},
	}

	log.Printf("starting render run for tarball '%s'", cfg.BackupTarballUri)
	err = renderer.Render(context.TODO(), cfg.BackupTarballUri)
	if err != nil {
		log.Fatal(err)
	}
}
