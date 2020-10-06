package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
	"github.com/moby/moby/client"
)

type Config struct {
	Listen string `default:":8080"`

	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`

	SourceBucketName string `default:"mc.sep.gg-backups"`
	MapOutputURI     string `default:"s3://map.tonkat.su/"`
	OverworldName    string `default:"pumpcraft"`
	NetherName       string `default:"pumpcraft_nether"`
	TheEndName       string `default:"pumpcraft_the_end"`
	RendererImage    string `default:"voltairemc/renderer"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("error setting up docker client: %s", err.Error())
	}
	sess := session.Must(session.NewSession())
	server := &server{
		cfg:    cfg,
		sns:    sns.New(sess),
		docker: docker,
	}

	server.start()
}
