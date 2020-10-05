package main

import (
	"encoding/json"
	"log"
	"net/http"

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
		log.Printf("error confirming subscription: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (svc *server) start() {
	http.Handle("/", http.HandlerFunc(svc.handler))
	log.Fatal(http.ListenAndServe(svc.cfg.Listen, nil))
}
