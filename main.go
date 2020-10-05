package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Listen string `default:":8080"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sess := session.Must(session.NewSession())
	server := &server{
		cfg: cfg,
		sns: sns.New(sess),
	}

	server.start()
}
