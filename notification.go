package main

import (
	"context"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func (svc *server) handleNotification(ctx context.Context, event events.S3Event) error {
	key, err := svc.parseLatestObjectKey(event)
	if err != nil {
		return err
	}

	objecturi := url.URL{
		Scheme: "s3",
		Host:   svc.cfg.SourceBucketName,
		Path:   key,
	}
	return svc.startRenderer(ctx, objecturi.String())
}

func (svc *server) parseLatestObjectKey(event events.S3Event) (key string, err error) {
	// only render the newest event
	var record events.S3EventRecord
	for _, e := range event.Records {
		if e.S3.Object.Key == "" {
			log.Println("empty event key, skipping")
			continue
		}

		// ignore events not from the bucket that we care about
		if e.S3.Bucket.Name != svc.cfg.SourceBucketName {
			log.Printf("event specifies bucket '%s', expected '%s'", e.S3.Bucket.Name, svc.cfg.SourceBucketName)
			continue
		}

		log.Printf("held event time: %s, new event time: %s", record.EventTime.Format(time.RFC3339), e.EventTime.Format(time.RFC3339))
		if e.EventTime.After(record.EventTime) {
			record = e
		}
	}

	if record.S3.Object.Key == "" {
		return "", errors.New("unable to parse object key from event")
	}

	return record.S3.Object.Key, nil
}
