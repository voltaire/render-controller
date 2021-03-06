package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	update_docker_image "github.com/bsdlp/update-docker-image"
	"github.com/docker/docker/client"
	"github.com/voltaire/render-controller/controller"
	"github.com/voltaire/render-controller/renderer"
	"golang.org/x/sync/errgroup"
)

type server struct {
	docker client.APIClient
	sns    snsiface.SNSAPI
	s3     s3iface.S3API
	cfg    renderer.Config

	controller *controller.Controller

	githubActionsPublicKey ed25519.PublicKey
}

func (svc *server) renderLatestMap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.Println("error discarding body")
	}
	r.Body.Close()

	listed, err := svc.s3.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket:              aws.String(svc.cfg.SourceBucketName),
		ExpectedBucketOwner: aws.String(svc.cfg.SourceBucketAccountId),
		Prefix:              aws.String(svc.cfg.SourceBucketPathPrefix),
		RequestPayer:        aws.String("requester"),
	})
	if err != nil {
		log.Printf("error listing s3 objects: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var latestObj *s3.Object
	for _, obj := range listed.Contents {
		if latestObj == nil {
			latestObj = obj
			continue
		}
		timestamp := aws.TimeValue(obj.LastModified)
		if timestamp.After(aws.TimeValue(latestObj.LastModified)) {
			latestObj = obj
		}
	}

	objecturi := url.URL{
		Scheme: "s3",
		Host:   svc.cfg.SourceBucketName,
		Path:   aws.StringValue(latestObj.Key),
	}
	log.Println("starting render run for: " + objecturi.String())
	err = svc.controller.StartForRender(ctx, svc.cfg, objecturi.String())
	if err != nil {
		log.Printf("error starting render-controller: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (svc *server) handleSNSMessage(w http.ResponseWriter, r *http.Request) {
	// sns times out after 15 seconds. https://docs.aws.amazon.com/sns/latest/dg/SendMessageToHttp.prepare.html
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	defer r.Body.Close()
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading sns message: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	switch r.Header.Get("X-Amz-Sns-Message-Type") {
	case "Notification":
		var event events.SNSEntity
		err = json.Unmarshal(bs, &event)
		if err != nil {
			log.Printf("error decoding s3 bucket event: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = svc.handleNotification(r.Context(), event)
	case "SubscriptionConfirmation":
		var msg subscriptionConfirmation
		err = json.Unmarshal(bs, &msg)
		if err != nil {
			log.Printf("error decoding subscription confirmation request: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		_, err = svc.sns.ConfirmSubscriptionWithContext(r.Context(), &sns.ConfirmSubscriptionInput{
			AuthenticateOnUnsubscribe: aws.String("true"),
			Token:                     msg.Token,
			TopicArn:                  msg.TopicArn,
		})
	case "UnsubscribeConfirmation":
		var msg subscriptionConfirmation
		err = json.Unmarshal(bs, &msg)
		if err != nil {
			log.Printf("error decoding unsubscribe confirmation request: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = confirmUnsubscribe(msg)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("error handling message: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (svc *server) ensureDockerImages(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	for _, img := range []string{
		svc.cfg.RendererImage,
		svc.cfg.RenderControllerImage,
	} {
		group.Go(func() error {
			return svc.pullDockerImage(ctx, img)
		})
	}
	return group.Wait()
}

func (svc *server) start() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	err := svc.ensureDockerImages(ctx)
	if err != nil {
		log.Fatal(err)
	}
	cancel()

	mux := http.NewServeMux()
	updateImageHandler := update_docker_image.NewUpdateDockerImageServer(svc)
	mux.Handle(updateImageHandler.PathPrefix(), updateImageHandler)
	mux.Handle("/callback", http.HandlerFunc(svc.handleSNSMessage))
	mux.Handle("/render_latest_map", http.HandlerFunc(svc.renderLatestMap))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "ok\n"); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Printf("error writing healthcheck response: %s", err.Error())
		}
	})
	log.Fatal(http.ListenAndServe(":80", mux))
}
