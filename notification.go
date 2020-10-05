package main

import (
	"context"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
)

func (svc *server) handleNotification(ctx context.Context, event events.S3Event) error {
	key := svc.parseLatestObjectKey(event)
	objecturi := url.URL{
		Scheme: "s3",
		Host:   svc.cfg.SourceBucketName,
		Path:   key,
	}
	return svc.startRenderer(ctx, objecturi.String())
}

func (svc *server) parseLatestObjectKey(event events.S3Event) (key string) {
	// only render the newest event
	var record events.S3EventRecord
	for _, e := range event.Records {
		// ignore events not from the bucket that we care about
		if e.S3.Bucket.Name != svc.cfg.SourceBucketName {
			continue
		}
		if e.EventTime.After(record.EventTime) {
			record = e
		}
	}

	return record.S3.Object.Key
}
