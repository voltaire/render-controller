package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
	"github.com/moby/moby/client"
)

type Config struct {
	Listen string `default:":8080"`

	AWSRegion          string `split_words:"true" default:"us-west-2"`
	AWSAccessKeyId     string `split_words:"true" required:"true"`
	AWSSecretAccessKey string `split_words:"true" required:"true"`

	SourceBucketName       string `default:"mc.sep.gg-backups"`
	SourceBucketAccountId  string `default:"006851364659"`
	SourceBucketPathPrefix string `default:"newworld"`

	DestinationBucketURI       string `split_words:"true" default:"s3://map-tonkat-su/"`
	DestinationBucketEndpoint  string `split_words:"true"`
	DestinationAccessKeyId     string `split_words:"true"`
	DestinationSecretAccessKey string `split_words:"true"`

	OverworldName     string `envconfig:"OVERWORLD_DIR" default:"pumpcraft"`
	NetherName        string `envconfig:"NETHER_DIR" default:"pumpcraft_nether"`
	TheEndName        string `envconfig:"THE_END_DIR" default:"pumpcraft_the_end"`
	RendererImage     string `default:"ghcr.io/voltaire/renderer:latest"`
	DiscordWebhookUrl string `split_words:"true"`

	RunnerName string `split_words:"true" default:"renderer"`

	GithubActionsPublicKey string `split_words:"true" required:"true"`
}

func decodePublicKey(encoded string) (key ed25519.PublicKey, err error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(data), nil
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	githubActionsPublicKey, err := decodePublicKey(cfg.GithubActionsPublicKey)
	if err != nil {
		log.Fatalf("error decoding github actions public key: %s", err.Error())
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("error setting up docker client: %s", err.Error())
	}
	sess := session.Must(session.NewSession())
	server := &server{
		cfg:                    cfg,
		sns:                    sns.New(sess),
		s3:                     s3.New(sess),
		docker:                 docker,
		githubActionsPublicKey: githubActionsPublicKey,
	}

	server.start()
}
