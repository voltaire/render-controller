package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type subscriptionConfirmation struct {
	MessageId *string   `json:"MessageId"`
	TopicArn  *string   `json:"TopicArn"`
	Timestamp time.Time `json:"Timestamp"`

	Token        *string `json:"Token"`
	Message      *string `json:"Message"`
	SubscribeURL *string `json:"SubscribeURL"`
}

func confirmUnsubscribe(msg subscriptionConfirmation) error {
	res, err := http.Get(aws.StringValue(msg.SubscribeURL))
	if err != nil {
		return err
	}
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		log.Printf("error discarding body: %s", err.Error())
	}
	res.Body.Close()
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		return fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}
	return nil
}
