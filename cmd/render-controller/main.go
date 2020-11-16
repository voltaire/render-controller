package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/render-controller/renderer"
)

func main() {
	var cfg renderer.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}
}
