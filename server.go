package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/moby/moby/client"
)

type server struct {
	docker client.APIClient
	sns    snsiface.SNSAPI
	cfg    Config
}

func (svc *server) handler(w http.ResponseWriter, r *http.Request) {
	// sns times out after 15 seconds. https://docs.aws.amazon.com/sns/latest/dg/SendMessageToHttp.prepare.html
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	var err error
	switch r.Header.Get("X-Amz-Sns-Message-Type") {
	case "Notification":
		var event events.S3Event
		err = json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			log.Printf("error decoding s3 bucket event: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var alreadyRunning bool
		alreadyRunning, err = svc.checkForAlreadyRunningContainer(ctx)
		if err != nil {
			log.Printf("error checking for running container: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if alreadyRunning {
			log.Printf("previous render container still running, skipping this run")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		err = svc.handleNotification(r.Context(), event)
	case "SubscriptionConfirmation":
		var msg subscriptionConfirmation
		err = json.NewDecoder(r.Body).Decode(&msg)
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
		err = json.NewDecoder(r.Body).Decode(&msg)
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

func (svc *server) start() {
	http.Handle("/callback", http.HandlerFunc(svc.handler))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "ok"); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Printf("error writing healthcheck response: %s", err.Error())
		}
	})
	log.Fatal(http.ListenAndServe(svc.cfg.Listen, nil))
}
