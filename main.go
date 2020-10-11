package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
	"github.com/moby/moby/client"
)

type Config struct {
	Listen string `default:":8080"`

	AWSRegion          string `envconfig:"AWS_REGION" default:"us-west-2"`
	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`

	SourceBucketName       string `default:"mc.sep.gg-backups"`
	SourceBucketAccountId  string `default:"006851364659"`
	SourceBucketPathPrefix string `default:"pumpcraft/"`
	MapOutputURI           string `envconfig:"MAP_OUTPUT_URI" default:"s3://map.tonkat.su/"`
	OverworldName          string `envconfig:"OVERWORLD_DIR" default:"pumpcraft"`
	NetherName             string `envconfig:"NETHER_DIR" default:"pumpcraft_nether"`
	TheEndName             string `envconfig:"THE_END_DIR" default:"pumpcraft_the_end"`
	RendererImage          string `default:"docker.io/voltairemc/renderer"`
	DiscordWebhookUrl      string `envconfig:"DISCORD_WEBHOOK_URL"`
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
		s3:     s3.New(sess),
		docker: docker,
	}

	server.start()
}
