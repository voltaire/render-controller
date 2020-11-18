package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"path/filepath"

	"github.com/aws/aws-lambda-go/events"
)

func (svc *server) handleNotification(ctx context.Context, event events.SNSEntity) error {
	s3Event, err := extractS3Event(event)
	if err != nil {
		return err
	}

	key, err := parseLatestObject(s3Event, svc.cfg.SourceBucketName)
	if err != nil {
		return err
	}

	if !filepath.HasPrefix(key, svc.cfg.SourceBucketPathPrefix) {
		return errors.New("received notification for the wrong path prefix")
	}

	objecturi := url.URL{
		Scheme: "s3",
		Host:   svc.cfg.SourceBucketName,
		Path:   key,
	}
	return svc.controller.StartForRender(ctx, svc.cfg, objecturi.String())
}

func extractS3Event(event events.SNSEntity) (events.S3Event, error) {
	var s3Event events.S3Event
	err := json.Unmarshal([]byte(event.Message), &s3Event)
	return s3Event, err
}

func parseLatestObject(event events.S3Event, sourceBucketName string) (string, error) {
	// only render the newest event
	var record events.S3EventRecord
	for _, e := range event.Records {
		if e.S3.Object.Key == "" {
			log.Println("empty event key, skipping")
			continue
		}

		// ignore events not from the bucket that we care about
		if e.S3.Bucket.Name != sourceBucketName {
			log.Printf("event specifies bucket '%s', expected '%s'", e.S3.Bucket.Name, sourceBucketName)
			continue
		}

		if e.EventTime.After(record.EventTime) {
			record = e
		}
	}

	key := record.S3.Object.Key
	if key == "" {
		return "", errors.New("received empty event object key")
	}

	return key, nil
}
