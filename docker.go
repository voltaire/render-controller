package main

import (
	"context"
	"crypto/ed25519"
	"io"
	"io/ioutil"
	"log"
	"time"

	update_docker_image "github.com/bsdlp/update-docker-image"
	"github.com/docker/docker/api/types"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/twitchtv/twirp"
)

func (svc *server) UpdateImage(ctx context.Context, req *update_docker_image.UpdateImageReq) (_ *empty.Empty, err error) {
	if !ed25519.Verify(svc.githubActionsPublicKey, []byte(req.GetImage()), req.GetSignature()) {
		return nil, twirp.NewError(twirp.Unauthenticated, "invalid signature")
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		resp, err := svc.docker.ImagePull(ctx, req.Image, types.ImagePullOptions{})
		if err != nil {
			log.Println("error pulling docker image: " + err.Error())
			return
		}
		defer resp.Close()
		_, err = io.Copy(ioutil.Discard, resp)
		if err != nil {
			log.Printf("error discarding docker image pull response: %s", err.Error())
		}
	}()
	return &empty.Empty{}, nil
}
