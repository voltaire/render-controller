package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/docker/docker/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/render-controller/controller"
	"github.com/voltaire/render-controller/renderer"
)

func decodePublicKey(encoded string) (key ed25519.PublicKey, err error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(data), nil
}

func main() {
	var cfg renderer.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	githubActionsPublicKey, err := decodePublicKey(cfg.GithubActionsPublicKey)
	if err != nil {
		log.Fatalf("error decoding github actions public key: %s", err.Error())
	}

	docker, err := client.NewClientWithOpts(client.FromEnv)
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
		controller: &controller.Controller{
			Docker: docker,
		},
	}

	server.start()
}
